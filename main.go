package main

import (
	"context"
	"log"
	"net/http"
	"strings"
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

	// Validate application configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	_ = cfg

	log.Println("Configuration validated successfully")

	// Connect to PostgreSQL.
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Run database migrations.
	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer migrationCancel()

	if err := migrations.Run(migrationCtx, db); err != nil {
		log.Fatal("Database migrations failed:", err)
	}

	log.Println("Database migrations completed")

	// Refresh-token cleanup.
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

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			cleanupRefreshTokens()
		}
	}()

	// Authentication rate limiter.
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			authRateLimiter.Cleanup(30 * time.Minute)
		}
	}()

	// ------------------------------------------------------------
	// Authentication endpoints
	// ------------------------------------------------------------

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

	// ------------------------------------------------------------
	// Authenticated account endpoints
	// ------------------------------------------------------------

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

	// ------------------------------------------------------------
	// User/profile/follow routes
	// ------------------------------------------------------------

	http.HandleFunc(
		"/api/users/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimSuffix(r.URL.Path, "/")

			// Follow status:
			// GET /api/users/{username}/follow-status
			if strings.HasSuffix(path, "/follow-status") {
				auth.AuthMiddleware(
					http.HandlerFunc(handlers.FollowStatus),
				).ServeHTTP(w, r)
				return
			}

			// Follow/unfollow:
			// POST   /api/users/{username}/follow
			// DELETE /api/users/{username}/follow
			if strings.HasSuffix(path, "/follow") {
				auth.AuthMiddleware(
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						switch r.Method {
						case http.MethodPost:
							handlers.FollowUser(w, r)

						case http.MethodDelete:
							handlers.UnfollowUser(w, r)

						default:
							http.Error(
								w,
								"method not allowed",
								http.StatusMethodNotAllowed,
							)
						}
					}),
				).ServeHTTP(w, r)
				return
			}

			// Followers:
			// GET /api/users/{username}/followers
			if strings.HasSuffix(path, "/followers") {
				handlers.Followers(w, r)
				return
			}

			// Following:
			// GET /api/users/{username}/following
			if strings.HasSuffix(path, "/following") {
				handlers.Following(w, r)
				return
			}

			// Public profile:
			// GET /api/users/{username}
			handlers.PublicProfile(w, r)
		},
	)

	// ------------------------------------------------------------
	// Global request-body limit
	// ------------------------------------------------------------

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
