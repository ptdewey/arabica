package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/rs/zerolog/log"
)

// defaultHTTPClient is a shared HTTP client with connection pooling.
// Reusing http.Client is recommended by the Go documentation as it
// manages connection pooling and is safe for concurrent use.
var defaultHTTPClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	},
}

// HandleLogin redirects to the home page
// The login form is now integrated into the home page
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleLoginSubmit initiates the OAuth flow
func (h *Handler) HandleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	handle := r.FormValue("handle")
	if handle == "" {
		http.Error(w, "Handle is required", http.StatusBadRequest)
		return
	}

	// Initiate OAuth flow
	authURL, err := h.oauth.InitiateLogin(r.Context(), handle)
	if err != nil {
		log.Error().Err(err).Str("handle", handle).Msg("Failed to initiate login")
		http.Error(w, "Failed to initiate login", http.StatusInternalServerError)
		return
	}

	// Redirect to PDS authorization endpoint
	// State and PKCE are handled automatically by the OAuth client
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleOAuthCallback handles the OAuth callback from the PDS
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	// Process the callback with all query parameters
	sessData, err := h.oauth.HandleCallback(r.Context(), r.URL.Query())
	if err != nil {
		log.Error().Err(err).Msg("Failed to complete OAuth flow")
		http.Error(w, "Failed to complete login", http.StatusInternalServerError)
		return
	}

	// Register user in the feed registry for the community feed
	if h.feedRegistry != nil {
		h.feedRegistry.Register(sessData.AccountDID.String())
	}

	// Set session cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "account_did",
		Value:    sessData.AccountDID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   h.config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30, // 30 days
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessData.SessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.config.SecureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30, // 30 days
	})

	log.Info().
		Str("user_did", sessData.AccountDID.String()).
		Str("session_id", sessData.SessionID).
		Msg("User logged in successfully")

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleLogout logs out the user
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	// Get session cookies
	didCookie, err1 := r.Cookie("account_did")
	sessionCookie, err2 := r.Cookie("session_id")

	if err1 == nil && err2 == nil {
		// Parse DID
		did, err := syntax.ParseDID(didCookie.Value)
		if err == nil {
			// Delete session from store
			err = h.oauth.DeleteSession(r.Context(), did, sessionCookie.Value)
			if err != nil {
				log.Warn().Err(err).Str("user_did", did.String()).Msg("Failed to delete session during logout")
			}
		}
	}

	// Clear session cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "account_did",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleClientMetadata serves the OAuth client metadata
func (h *Handler) HandleClientMetadata(w http.ResponseWriter, r *http.Request) {
	if h.oauth == nil {
		http.Error(w, "OAuth not configured", http.StatusInternalServerError)
		return
	}

	metadata := h.oauth.ClientMetadata()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		log.Error().Err(err).Msg("Failed to encode client metadata")
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleWellKnownOAuth serves the OAuth client metadata at /.well-known/oauth-client-metadata
func (h *Handler) HandleWellKnownOAuth(w http.ResponseWriter, r *http.Request) {
	h.HandleClientMetadata(w, r)
}

// HandleResolveHandle resolves an AT Protocol handle and returns basic profile info
// This is used for the autocomplete login feature
func (h *Handler) HandleResolveHandle(w http.ResponseWriter, r *http.Request) {
	handle := r.URL.Query().Get("handle")
	if handle == "" {
		http.Error(w, "Handle parameter is required", http.StatusBadRequest)
		return
	}

	// Use a public API client to resolve the handle
	// We don't need authentication for this
	apiClient := defaultHTTPClient

	// First resolve the handle to a DID using the public API
	// Note: public.api.bsky.app is the public Bluesky API endpoint that works for any handle
	resolveURL := fmt.Sprintf("https://public.api.bsky.app/xrpc/com.atproto.identity.resolveHandle?handle=%s", handle)
	resp, err := apiClient.Get(resolveURL)
	if err != nil {
		log.Error().Err(err).Str("handle", handle).Msg("Failed to resolve handle")
		http.Error(w, "Failed to resolve handle", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]string{"error": "Handle not found"}); err != nil {
			log.Error().Err(err).Msg("Failed to encode error response")
		}
		return
	}

	if resp.StatusCode != 200 {
		// Read the error body for better debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn().
			Str("handle", handle).
			Int("status", resp.StatusCode).
			Str("body", string(bodyBytes)).
			Msg("Unexpected status resolving handle")

		// Return a more informative error for 400s
		if resp.StatusCode == 400 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "Invalid handle format"}); err != nil {
				log.Error().Err(err).Msg("Failed to encode error response")
			}
			return
		}

		http.Error(w, "Failed to resolve handle", http.StatusInternalServerError)
		return
	}

	var resolveResult struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&resolveResult); err != nil {
		log.Error().Err(err).Str("handle", handle).Msg("Failed to decode resolve response")
		http.Error(w, "Failed to parse resolve response", http.StatusInternalServerError)
		return
	}

	// Now fetch the profile for this DID using the public API
	profileURL := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.getProfile?actor=%s", resolveResult.DID)
	profileResp, err := apiClient.Get(profileURL)
	if err != nil {
		log.Warn().Err(err).Str("did", resolveResult.DID).Msg("Failed to fetch profile")
		// Return just the DID if we can't get the profile
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"did":    resolveResult.DID,
			"handle": handle,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to encode response")
		}
		return
	}
	defer profileResp.Body.Close()

	if profileResp.StatusCode != 200 {
		// Return just the DID if we can't get the profile
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"did":    resolveResult.DID,
			"handle": handle,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to encode response")
		}
		return
	}

	var profile struct {
		DID         string  `json:"did"`
		Handle      string  `json:"handle"`
		DisplayName *string `json:"displayName,omitempty"`
		Avatar      *string `json:"avatar,omitempty"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		log.Warn().Err(err).Str("did", resolveResult.DID).Msg("Failed to decode profile")
		// Return just the DID if we can't parse the profile
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"did":    resolveResult.DID,
			"handle": handle,
		}); err != nil {
			log.Error().Err(err).Msg("Failed to encode response")
		}
		return
	}

	// Return the profile info
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		log.Error().Err(err).Msg("Failed to encode profile response")
	}
}

// HandleSearchActors searches for actors by handle or display name
// This is used for the autocomplete login feature
func (h *Handler) HandleSearchActors(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	// Use a public API client with timeout
	apiClient := defaultHTTPClient

	// Try using the public API endpoint with typeahead parameter
	// Some PDS instances support public search
	searchURL := fmt.Sprintf("https://public.api.bsky.app/xrpc/app.bsky.actor.searchActorsTypeahead?q=%s&limit=5", query)
	resp, err := apiClient.Get(searchURL)
	if err != nil {
		log.Warn().Err(err).Str("query", query).Msg("Failed to search actors")
		// Return empty results instead of error
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"actors": []interface{}{}}); err != nil {
			log.Error().Err(err).Msg("Failed to encode empty actors response")
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn().
			Str("query", query).
			Int("status", resp.StatusCode).
			Str("body", string(bodyBytes)).
			Msg("Unexpected status searching actors")
		// Return empty results instead of error
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"actors": []interface{}{}}); err != nil {
			log.Error().Err(err).Msg("Failed to encode empty actors response")
		}
		return
	}

	var searchResult struct {
		Actors []struct {
			DID         string  `json:"did"`
			Handle      string  `json:"handle"`
			DisplayName *string `json:"displayName,omitempty"`
			Avatar      *string `json:"avatar,omitempty"`
		} `json:"actors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		log.Warn().Err(err).Str("query", query).Msg("Failed to decode search response")
		// Return empty results instead of error
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"actors": []interface{}{}}); err != nil {
			log.Error().Err(err).Msg("Failed to encode empty actors response")
		}
		return
	}

	// Return the actors
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(searchResult); err != nil {
		log.Error().Err(err).Msg("Failed to encode search result response")
	}
}
