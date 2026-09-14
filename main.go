package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/config"
	"github.com/Traderong/omnivibe-api/database"
	"github.com/Traderong/omnivibe-api/handlers"
	"github.com/Traderong/omnivibe-api/middleware"
	"github.com/Traderong/omnivibe-api/migrations"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env when available.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	// Validate application configuration before starting the server.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	// The individual packages currently read environment variables directly.
	// Keep the validated configuration referenced for now.
	_ = cfg

	log.Println("Configuration validated successfully")

	// Connect to PostgreSQL.
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Run database migrations before accepting requests.
	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer migrationCancel()

	if err := migrations.Run(migrationCtx, db); err != nil {
		log.Fatal("Database migrations failed:", err)
	}

	log.Println("Database migrations completed")

	// Remove expired and long-revoked refresh tokens at startup.
	cleanupRefreshTokens := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cleanupCancel()

		if err := auth.CleanupRefreshTokens(cleanupCtx); err != nil {
			log.Printf("refresh token cleanup failed: %v", err)
			return
		}

		log.Println("refresh token cleanup completed")
	}

	cleanupRefreshTokens()

	// Run refresh-token cleanup every 24 hours.
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			cleanupRefreshTokens()
		}
	}()

	// Authentication rate limiter:
	// Maximum 10 requests per IP address per minute.
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	// Clean inactive rate-limit entries periodically.
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			authRateLimiter.Cleanup(30 * time.Minute)
		}
	}()

	// Authentication endpoints.

	http.Handle(
		"/api/auth/register",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Register),
		),
	)

	http.Handle(
		"/api/auth/login",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Login),
		),
	)

	http.Handle(
		"/api/auth/verify-email",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.VerifyEmail),
		),
	)

	http.Handle(
		"/api/auth/resend-verification",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ResendVerification),
		),
	)

	http.Handle(
		"/api/auth/forgot-password",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ForgotPassword),
		),
	)

	http.Handle(
		"/api/auth/reset-password",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ResetPassword),
		),
	)

	http.Handle(
		"/api/auth/refresh",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Refresh),
		),
	)

	http.Handle(
		"/api/auth/logout",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Logout),
		),
	)

	// Authenticated endpoints.

	http.Handle(
		"/api/me",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.Me),
		),
	)

	http.Handle(
		"/api/me/profile",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.UpdateProfile),
		),
	)

	http.Handle(
		"/api/me/password",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.ChangePassword),
		),
	)

	// Apply a 1 MB request-body limit to every registered endpoint.
	rootHandler := middleware.BodyLimit(http.DefaultServeMux)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 * 1024,
	}

	log.Println("OmniVibe backend started on http://localhost:8080")
	log.Println("PostgreSQL connected successfully")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
