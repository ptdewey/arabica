package atproto

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

var scopes = []string{
	"atproto",
	"repo:com.arabica.bean",
	"repo:com.arabica.brew",
	"repo:com.arabica.brewer",
	"repo:com.arabica.grinder",
	"repo:com.arabica.roaster",
}

// OAuthManager wraps indigo's OAuth client for managing user authentication
type OAuthManager struct {
	app *oauth.ClientApp
}

// NewOAuthManager creates a new OAuth manager with the given configuration
func NewOAuthManager(clientID, redirectURI string) (*OAuthManager, error) {
	var config oauth.ClientConfig

	// Check if we should use localhost config
	if clientID == "" || strings.HasPrefix(clientID, "http://localhost") {
		// Use special localhost config for development
		config = oauth.NewLocalhostConfig(redirectURI, scopes)
	} else {
		// Use public config for production (with real domain)
		config = oauth.NewPublicConfig(clientID, redirectURI, scopes)
	}

	// Use in-memory store for development
	// TODO: Replace with persistent store (Redis/SQLite) for production
	store := oauth.NewMemStore()

	// Create the OAuth client app
	app := oauth.NewClientApp(&config, store)

	return &OAuthManager{
		app: app,
	}, nil
}

// InitiateLogin starts the OAuth flow for a user
// Returns the authorization URL to redirect the user to
func (m *OAuthManager) InitiateLogin(ctx context.Context, handle string) (authURL string, err error) {
	// Start the OAuth flow using the handle/username
	redirectURL, err := m.app.StartAuthFlow(ctx, handle)
	if err != nil {
		return "", fmt.Errorf("failed to start OAuth flow: %w", err)
	}

	return redirectURL, nil
}

// SessionData holds session information after OAuth callback
type SessionData struct {
	AccountDID syntax.DID
	SessionID  string
	Scopes     []string
}

// HandleCallback processes the OAuth callback after user authorization
// Returns the session information including DID and session ID
func (m *OAuthManager) HandleCallback(ctx context.Context, params url.Values) (*SessionData, error) {
	// Process the callback parameters (includes code, state, etc.)
	sessData, err := m.app.ProcessCallback(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to process OAuth callback: %w", err)
	}

	return &SessionData{
		AccountDID: sessData.AccountDID,
		SessionID:  sessData.SessionID,
		Scopes:     sessData.Scopes,
	}, nil
}

// GetSession retrieves session information for a given session ID
func (m *OAuthManager) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*oauth.ClientSessionData, error) {
	sessData, err := m.app.Store.GetSession(ctx, did, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return sessData, nil
}

// DeleteSession removes a session (for logout)
func (m *OAuthManager) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	err := m.app.Store.DeleteSession(ctx, did, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ClientMetadata returns the OAuth client metadata document
// This should be served at the client_id URL
func (m *OAuthManager) ClientMetadata() oauth.ClientMetadata {
	return m.app.Config.ClientMetadata()
}

// AuthMiddleware adds authentication context to HTTP requests
func (m *OAuthManager) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get session cookies
		didCookie, err1 := r.Cookie("account_did")
		sessionCookie, err2 := r.Cookie("session_id")

		if err1 != nil || err2 != nil {
			// No session cookies, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Parse DID
		did, err := syntax.ParseDID(didCookie.Value)
		if err != nil {
			// Invalid DID, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Get session from store
		sessData, err := m.GetSession(r.Context(), did, sessionCookie.Value)
		if err != nil {
			// Invalid session, continue without auth
			next.ServeHTTP(w, r)
			return
		}

		// Note: Token refresh is handled automatically by the SDK when making authenticated requests

		// Add authenticated DID and session to context
		ctx := context.WithValue(r.Context(), contextKeyUserDID, did.String())
		ctx = context.WithValue(ctx, contextKeySessionID, sessionCookie.Value)
		ctx = context.WithValue(ctx, contextKeySessionData, sessData)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Context keys for storing auth info
type contextKey string

const (
	contextKeyUserDID     contextKey = "userDID"
	contextKeySessionID   contextKey = "sessionID"
	contextKeySessionData contextKey = "sessionData"
)

// GetAuthenticatedDID retrieves the authenticated user's DID from the request context
func GetAuthenticatedDID(ctx context.Context) (string, error) {
	did, ok := ctx.Value(contextKeyUserDID).(string)
	if !ok || did == "" {
		return "", fmt.Errorf("no authenticated user")
	}
	return did, nil
}

// GetSessionIDFromContext retrieves the session ID from the request context
func GetSessionIDFromContext(ctx context.Context) (string, error) {
	sessionID, ok := ctx.Value(contextKeySessionID).(string)
	if !ok || sessionID == "" {
		return "", fmt.Errorf("no session ID in context")
	}
	return sessionID, nil
}

// ParseDID is a helper to parse a DID string to syntax.DID
func ParseDID(didStr string) (syntax.DID, error) {
	return syntax.ParseDID(didStr)
}

// GetSessionDataFromContext retrieves the full session data from the request context
func GetSessionDataFromContext(ctx context.Context) (*oauth.ClientSessionData, error) {
	sessData, ok := ctx.Value(contextKeySessionData).(*oauth.ClientSessionData)
	if !ok || sessData == nil {
		return nil, fmt.Errorf("no session data in context")
	}
	return sessData, nil
}

// RequireAuth is middleware that requires authentication
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := GetAuthenticatedDID(r.Context())
		if err != nil {
			// Not authenticated, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
