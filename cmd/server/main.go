package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"arabica/internal/atproto"
	"arabica/internal/database/sqlite"
	"arabica/internal/handlers"
)

// loggingMiddleware logs HTTP request details to stdout
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log request details
		log.Printf(
			"%s %s %d %s %s",
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func main() {
	// Get database path from env or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Try XDG_DATA_HOME first, then fallback to HOME, then current dir
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			dbPath = filepath.Join(xdgData, "arabica", "arabica.db")
			os.MkdirAll(filepath.Dir(dbPath), 0755)
		} else if home := os.Getenv("HOME"); home != "" {
			dbPath = filepath.Join(home, ".local", "share", "arabica", "arabica.db")
			os.MkdirAll(filepath.Dir(dbPath), 0755)
		} else {
			dbPath = "./arabica.db"
		}
	}

	// Initialize database
	store, err := sqlite.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	log.Printf("Using database: %s", dbPath)

	// Get port from env or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "18910"
	}

	// TODO: address environment variable

	// Initialize OAuth manager
	// For local development, localhost URLs trigger special localhost mode in indigo
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")

	if clientID == "" && redirectURI == "" {
		// Use localhost defaults for development
		redirectURI = fmt.Sprintf("http://127.0.0.1:%s/oauth/callback", port)
		clientID = "" // Empty triggers localhost mode
		log.Printf("Using localhost OAuth mode (for development)")
	}

	// FIX: trim down scopes to not need full bluesky perms?
	scopes := []string{"atproto", "repo:social.arabica.brew", "repo:social.arabica.brewer"}

	oauthManager, err := atproto.NewOAuthManager(clientID, redirectURI, scopes)
	if err != nil {
		log.Fatalf("Failed to initialize OAuth: %v", err)
	}

	log.Printf("OAuth configured:")
	if clientID == "" {
		log.Printf("  Mode: Localhost development")
	} else {
		log.Printf("  Client ID: %s", clientID)
	}
	log.Printf("  Redirect URI: %s", redirectURI)
	log.Printf("  Scopes: %v", scopes)

	// Initialize atproto client
	atprotoClient := atproto.NewClient(oauthManager)
	log.Printf("Atproto client initialized")

	// Initialize handlers
	h := handlers.NewHandler(store)
	h.SetOAuthManager(oauthManager)
	h.SetAtprotoClient(atprotoClient)

	// Create router
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

	// Page routes (must come before static files)
	mux.HandleFunc("GET /{$}", h.HandleHome) // {$} means exact match
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

	// Static files (must come after specific routes)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Apply OAuth middleware to add auth context to all requests
	handler := oauthManager.AuthMiddleware(mux)

	// TODO: configure port and address via env vars
	log.Printf("Starting Arabica server on http://localhost:%s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, loggingMiddleware(handler)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
