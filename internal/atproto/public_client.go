package atproto

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// PublicAPIBaseURL is the public Bluesky API endpoint
	PublicAPIBaseURL = "https://public.api.bsky.app"
	// PLCDirectoryURL is the PLC directory for resolving DIDs
	PLCDirectoryURL = "https://plc.directory"
)

// ErrSSRFBlocked is returned when a potential SSRF attack is blocked
var ErrSSRFBlocked = errors.New("request blocked: potential SSRF detected")

// isPrivateIP checks if an IP address is in a private/internal range
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Check for loopback
	if ip.IsLoopback() {
		return true
	}

	// Check for link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check for private ranges
	if ip.IsPrivate() {
		return true
	}

	// Check for unspecified (0.0.0.0)
	if ip.IsUnspecified() {
		return true
	}

	// Check for cloud metadata endpoint (169.254.169.254)
	if ip.Equal(net.ParseIP("169.254.169.254")) {
		return true
	}

	return false
}

// validateDomain checks if a domain is safe to connect to (not internal/private)
func validateDomain(domain string) error {
	// Block obviously dangerous patterns
	if domain == "localhost" || strings.HasSuffix(domain, ".local") {
		return ErrSSRFBlocked
	}

	// Check for IP addresses embedded in the domain
	if ip := net.ParseIP(domain); ip != nil {
		if isPrivateIP(ip) {
			return ErrSSRFBlocked
		}
	}

	// Resolve the domain and check all IPs
	ips, err := net.LookupIP(domain)
	if err != nil {
		// If we can't resolve it, let the HTTP request fail later
		return nil
	}

	for _, ip := range ips {
		if isPrivateIP(ip) {
			return ErrSSRFBlocked
		}
	}

	return nil
}

// PublicClient provides unauthenticated access to public ATProto APIs
type PublicClient struct {
	baseURL    string
	httpClient *http.Client
	// Cache PDS endpoints to avoid repeated lookups
	pdsCache   map[string]string
	pdsCacheMu sync.RWMutex
}

// NewPublicClient creates a new public API client
func NewPublicClient() *PublicClient {
	return &PublicClient{
		baseURL: PublicAPIBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		pdsCache: make(map[string]string),
	}
}

// GetPDSEndpoint resolves a DID to find the user's PDS endpoint
func (c *PublicClient) GetPDSEndpoint(ctx context.Context, did string) (string, error) {
	// Check cache first
	c.pdsCacheMu.RLock()
	if pds, ok := c.pdsCache[did]; ok {
		c.pdsCacheMu.RUnlock()
		return pds, nil
	}
	c.pdsCacheMu.RUnlock()

	// Resolve DID document from PLC directory
	var pdsEndpoint string

	if strings.HasPrefix(did, "did:plc:") {
		// PLC DID - resolve from plc.directory
		reqURL := fmt.Sprintf("%s/%s", PLCDirectoryURL, did)
		req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
		if err != nil {
			return "", fmt.Errorf("creating request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("fetching DID document: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("DID resolution failed with status %d", resp.StatusCode)
		}

		var didDoc struct {
			Service []struct {
				ID              string `json:"id"`
				Type            string `json:"type"`
				ServiceEndpoint string `json:"serviceEndpoint"`
			} `json:"service"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&didDoc); err != nil {
			return "", fmt.Errorf("decoding DID document: %w", err)
		}

		// Find the atproto_pds service
		for _, svc := range didDoc.Service {
			if svc.ID == "#atproto_pds" || svc.Type == "AtprotoPersonalDataServer" {
				pdsEndpoint = svc.ServiceEndpoint
				break
			}
		}
	} else if strings.HasPrefix(did, "did:web:") {
		// Web DID - the domain is the PDS
		// Validate domain to prevent SSRF attacks
		domain := strings.TrimPrefix(did, "did:web:")
		// Handle percent-encoded colons for ports (e.g., did:web:example.com%3A8080)
		domain = strings.ReplaceAll(domain, "%3A", ":")

		// Extract just the host part (without path)
		if idx := strings.Index(domain, "/"); idx != -1 {
			domain = domain[:idx]
		}

		// Validate the domain is safe
		host := domain
		if hostPart, _, err := net.SplitHostPort(domain); err == nil {
			host = hostPart
		}
		if err := validateDomain(host); err != nil {
			return "", err
		}

		pdsEndpoint = "https://" + domain
	}

	if pdsEndpoint == "" {
		return "", fmt.Errorf("could not resolve PDS endpoint for %s", did)
	}

	// Cache the result
	c.pdsCacheMu.Lock()
	c.pdsCache[did] = pdsEndpoint
	c.pdsCacheMu.Unlock()

	return pdsEndpoint, nil
}

// Profile represents a user's public profile
type Profile struct {
	DID         string  `json:"did"`
	Handle      string  `json:"handle"`
	DisplayName *string `json:"displayName,omitempty"`
	Avatar      *string `json:"avatar,omitempty"`
}

// PublicListRecordsOutput represents the response from public listRecords API
type PublicListRecordsOutput struct {
	Records []PublicRecordEntry `json:"records"`
	Cursor  *string             `json:"cursor,omitempty"`
}

// PublicRecordEntry represents a single record in the public listRecords response
type PublicRecordEntry struct {
	URI   string                 `json:"uri"`
	CID   string                 `json:"cid"`
	Value map[string]interface{} `json:"value"`
}

// GetProfile fetches a user's public profile by DID or handle
func (c *PublicClient) GetProfile(ctx context.Context, actor string) (*Profile, error) {
	reqURL := fmt.Sprintf("%s/xrpc/app.bsky.actor.getProfile?actor=%s",
		c.baseURL, url.QueryEscape(actor))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile request failed with status %d", resp.StatusCode)
	}

	var profile Profile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("decoding profile: %w", err)
	}

	return &profile, nil
}

// ListRecords fetches public records from a user's repository
// Records are returned in reverse chronological order (newest first)
// This queries the user's PDS directly to support custom collections
func (c *PublicClient) ListRecords(ctx context.Context, did, collection string, limit int) (*PublicListRecordsOutput, error) {
	// Resolve the user's PDS endpoint
	pdsEndpoint, err := c.GetPDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("resolving PDS: %w", err)
	}

	reqURL := fmt.Sprintf("%s/xrpc/com.atproto.repo.listRecords?repo=%s&collection=%s&limit=%d&reverse=true",
		pdsEndpoint, url.QueryEscape(did), url.QueryEscape(collection), limit)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("listing records: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list records request failed with status %d", resp.StatusCode)
	}

	var output PublicListRecordsOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decoding records: %w", err)
	}

	return &output, nil
}

// ResolveHandle resolves an AT Protocol handle to a DID
func (c *PublicClient) ResolveHandle(ctx context.Context, handle string) (string, error) {
	reqURL := fmt.Sprintf("%s/xrpc/com.atproto.identity.resolveHandle?handle=%s",
		c.baseURL, url.QueryEscape(handle))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("resolving handle: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("handle not found: %s", handle)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolve handle failed with status %d", resp.StatusCode)
	}

	var result struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return result.DID, nil
}

// GetRecord fetches a single public record from the user's PDS
func (c *PublicClient) GetRecord(ctx context.Context, did, collection, rkey string) (*PublicRecordEntry, error) {
	// Resolve the user's PDS endpoint
	pdsEndpoint, err := c.GetPDSEndpoint(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("resolving PDS: %w", err)
	}

	reqURL := fmt.Sprintf("%s/xrpc/com.atproto.repo.getRecord?repo=%s&collection=%s&rkey=%s",
		pdsEndpoint, url.QueryEscape(did), url.QueryEscape(collection), url.QueryEscape(rkey))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get record request failed with status %d", resp.StatusCode)
	}

	var entry PublicRecordEntry
	if err := json.NewDecoder(resp.Body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("decoding record: %w", err)
	}

	return &entry, nil
}
