package main

import (
	"log"
	"net/http"
	"os"

	"arabica/internal/database/sqlite"
	"arabica/internal/handlers"
)

func main() {
	// Get database path from env or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./arabica.db"
	}

	// Initialize database
	store, err := sqlite.NewSQLiteStore(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	// Initialize handlers
	h := handlers.NewHandler(store)

	// Create router
	mux := http.NewServeMux()

	// Page routes (must come before static files)
	mux.HandleFunc("GET /{$}", h.HandleHome) // {$} means exact match
	mux.HandleFunc("GET /brews", h.HandleBrewList)
	mux.HandleFunc("GET /brews/new", h.HandleBrewNew)
	mux.HandleFunc("POST /brews", h.HandleBrewCreate)
	mux.HandleFunc("DELETE /brews/{id}", h.HandleBrewDelete)
	mux.HandleFunc("GET /brews/export", h.HandleBrewExport)

	// API routes for adding beans/roasters via AJAX
	mux.HandleFunc("POST /api/beans", h.HandleBeanCreate)
	mux.HandleFunc("POST /api/roasters", h.HandleRoasterCreate)

	// Static files (must come after specific routes)
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Get port from env or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Arabica server on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
