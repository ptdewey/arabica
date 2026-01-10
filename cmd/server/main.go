package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"arabica/internal/atproto"
	"arabica/internal/handlers"
	"arabica/internal/routing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Configure zerolog
	// Set log level from environment (default: info)
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info", "":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Use pretty console logging in development, JSON in production
	if os.Getenv("LOG_FORMAT") == "json" {
		// Production: JSON logs
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Development: pretty console logs
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}

	log.Info().Msg("Starting Arabica Coffee Tracker")

	// Get port from env or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "18910"
	}

	// Initialize OAuth manager
	// For local development, localhost URLs trigger special localhost mode in indigo
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_REDIRECT_URI")

	if clientID == "" && redirectURI == "" {
		// Use localhost defaults for development
		redirectURI = fmt.Sprintf("http://127.0.0.1:%s/oauth/callback", port)
		clientID = "" // Empty triggers localhost mode
		log.Info().Msg("Using localhost OAuth mode (for development)")
	}

	oauthManager, err := atproto.NewOAuthManager(clientID, redirectURI)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize OAuth")
	}

	if clientID == "" {
		log.Info().
			Str("mode", "localhost development").
			Str("redirect_uri", redirectURI).
			Msg("OAuth configured")
	} else {
		log.Info().
			Str("client_id", clientID).
			Str("redirect_uri", redirectURI).
			Msg("OAuth configured")
	}

	// Initialize atproto client
	atprotoClient := atproto.NewClient(oauthManager)
	log.Info().Msg("ATProto client initialized")

	// Determine if we should use secure cookies (default: false for development)
	// Set SECURE_COOKIES=true in production with HTTPS
	secureCookies := os.Getenv("SECURE_COOKIES") == "true"

	// Initialize handlers
	h := handlers.NewHandler()
	h.SetConfig(handlers.Config{
		SecureCookies: secureCookies,
	})
	h.SetOAuthManager(oauthManager)
	h.SetAtprotoClient(atprotoClient)

	// Setup router with middleware
	handler := routing.SetupRouter(routing.Config{
		Handlers:     h,
		OAuthManager: oauthManager,
		Logger:       log.Logger,
	})

	// Start HTTP server
	log.Info().
		Str("address", "0.0.0.0:"+port).
		Str("url", "http://localhost:"+port).
		Bool("secure_cookies", secureCookies).
		Msg("Starting HTTP server")

	if err := http.ListenAndServe("0.0.0.0:"+port, handler); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
