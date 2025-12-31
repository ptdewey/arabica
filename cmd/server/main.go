package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"arabica/internal/database/sqlite"
	"arabica/internal/handlers"
)

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

	// Initialize handlers
	h := handlers.NewHandler(store)

	// Create router
	mux := http.NewServeMux()

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

	// Get port from env or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "18910"
	}

	// TODO: configure port and address via env vars
	log.Printf("Starting Arabica server on http://0.0.0.0:%s", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
