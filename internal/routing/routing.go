package routing

import (
	"net/http"

	"arabica/internal/atproto"
	"arabica/internal/handlers"
	"arabica/internal/middleware"

	"github.com/rs/zerolog"
)

// Config holds the configuration needed for setting up routes
type Config struct {
	Handlers     *handlers.Handler
	OAuthManager *atproto.OAuthManager
	Logger       zerolog.Logger
}

// SetupRouter creates and configures the HTTP router with all routes and middleware
func SetupRouter(cfg Config) http.Handler {
	h := cfg.Handlers
	mux := http.NewServeMux()

	// OAuth routes
	mux.HandleFunc("GET /login", h.HandleLogin)
	mux.HandleFunc("POST /auth/login", h.HandleLoginSubmit)
	mux.HandleFunc("GET /oauth/callback", h.HandleOAuthCallback)
	mux.HandleFunc("POST /logout", h.HandleLogout)
	mux.HandleFunc("GET /client-metadata.json", h.HandleClientMetadata)
	mux.HandleFunc("GET /.well-known/oauth-client-metadata", h.HandleWellKnownOAuth)

	// API routes for handle resolution (used by login autocomplete)
	mux.HandleFunc("GET /api/resolve-handle", h.HandleResolveHandle)
	mux.HandleFunc("GET /api/search-actors", h.HandleSearchActors)

	// API route for fetching all user data (used by client-side cache)
	mux.HandleFunc("GET /api/data", h.HandleAPIListAll)

	// Community feed partial (loaded async via HTMX)
	mux.HandleFunc("GET /api/feed", h.HandleFeedPartial)

	// Brew list partial (loaded async via HTMX)
	mux.HandleFunc("GET /api/brews", h.HandleBrewListPartial)

	// Manage page partial (loaded async via HTMX)
	mux.HandleFunc("GET /api/manage", h.HandleManagePartial)

	// Page routes (must come before static files)
	mux.HandleFunc("GET /{$}", h.HandleHome) // {$} means exact match
	mux.HandleFunc("GET /about", h.HandleAbout)
	mux.HandleFunc("GET /terms", h.HandleTerms)
	mux.HandleFunc("GET /manage", h.HandleManage)
	mux.HandleFunc("GET /brews", h.HandleBrewList)
	mux.HandleFunc("GET /brews/new", h.HandleBrewNew)
	mux.HandleFunc("GET /brews/{id}", h.HandleBrewEdit)
	mux.HandleFunc("POST /brews", h.HandleBrewCreate)
	mux.HandleFunc("PUT /brews/{id}", h.HandleBrewUpdate)
	mux.HandleFunc("DELETE /brews/{id}", h.HandleBrewDelete)
	mux.HandleFunc("GET /brews/export", h.HandleBrewExport)

	// API routes for CRUD operations
	mux.HandleFunc("POST /api/beans", h.HandleBeanCreate)
	mux.HandleFunc("PUT /api/beans/{id}", h.HandleBeanUpdate)
	mux.HandleFunc("DELETE /api/beans/{id}", h.HandleBeanDelete)

	mux.HandleFunc("POST /api/roasters", h.HandleRoasterCreate)
	mux.HandleFunc("PUT /api/roasters/{id}", h.HandleRoasterUpdate)
	mux.HandleFunc("DELETE /api/roasters/{id}", h.HandleRoasterDelete)

	mux.HandleFunc("POST /api/grinders", h.HandleGrinderCreate)
	mux.HandleFunc("PUT /api/grinders/{id}", h.HandleGrinderUpdate)
	mux.HandleFunc("DELETE /api/grinders/{id}", h.HandleGrinderDelete)

	mux.HandleFunc("POST /api/brewers", h.HandleBrewerCreate)
	mux.HandleFunc("PUT /api/brewers/{id}", h.HandleBrewerUpdate)
	mux.HandleFunc("DELETE /api/brewers/{id}", h.HandleBrewerDelete)

	// Profile routes (public user profiles)
	mux.HandleFunc("GET /profile/{actor}", h.HandleProfile)

	// Static files (must come after specific routes)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Catch-all 404 handler - must be last, catches any unmatched routes
	mux.HandleFunc("/", h.HandleNotFound)

	// Apply middleware in order (outermost first, innermost last)
	var handler http.Handler = mux

	// 1. Limit request body size (innermost - runs first on request)
	handler = middleware.LimitBodyMiddleware(handler)

	// 2. Apply OAuth middleware to add auth context
	handler = cfg.OAuthManager.AuthMiddleware(handler)

	// 3. Apply rate limiting
	rateLimitConfig := middleware.NewDefaultRateLimitConfig()
	handler = middleware.RateLimitMiddleware(rateLimitConfig)(handler)

	// 4. Apply security headers
	handler = middleware.SecurityHeadersMiddleware(handler)

	// 5. Apply logging middleware (outermost - wraps everything)
	handler = middleware.LoggingMiddleware(cfg.Logger)(handler)

	return handler
}
