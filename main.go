package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/database"
	"github.com/Traderong/omnivibe-api/handlers"
	"github.com/Traderong/omnivibe-api/middleware"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	cleanupRefreshTokens := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := auth.CleanupRefreshTokens(ctx); err != nil {
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

	// Authentication rate limiter:
	// Maximum 10 requests per IP address per minute.
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	http.Handle(
		"/api/auth/register",
		authRateLimiter.Middleware(http.HandlerFunc(handlers.Register)),
	)

	http.Handle(
		"/api/auth/login",
		authRateLimiter.Middleware(http.HandlerFunc(handlers.Login)),
	)

	http.HandleFunc("/api/auth/verify-email", handlers.VerifyEmail)

	http.Handle(
		"/api/auth/resend-verification",
		authRateLimiter.Middleware(http.HandlerFunc(handlers.ResendVerification)),
	)

	http.Handle(
		"/api/auth/forgot-password",
		authRateLimiter.Middleware(http.HandlerFunc(handlers.ForgotPassword)),
	)

	http.Handle(
		"/api/auth/reset-password",
		authRateLimiter.Middleware(http.HandlerFunc(handlers.ResetPassword)),
	)

	http.HandleFunc("/api/auth/refresh", handlers.Refresh)
	http.HandleFunc("/api/auth/logout", handlers.Logout)

	meHandler := http.HandlerFunc(handlers.Me)
	http.Handle("/api/me", auth.AuthMiddleware(meHandler))

	updateProfileHandler := http.HandlerFunc(handlers.UpdateProfile)
	http.Handle("/api/me/profile", auth.AuthMiddleware(updateProfileHandler))

	changePasswordHandler := http.HandlerFunc(handlers.ChangePassword)
	http.Handle("/api/me/password", auth.AuthMiddleware(changePasswordHandler))

	log.Println("OmniVibe backend started on http://localhost:8080")
	log.Println("PostgreSQL connected successfully")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
