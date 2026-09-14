package main

import (
	"log"
	"net/http"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/database"
	"github.com/Traderong/omnivibe-api/handlers"
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

	http.HandleFunc("/api/auth/register", handlers.Register)
	http.HandleFunc("/api/auth/login", handlers.Login)
	http.HandleFunc("/api/auth/verify-email", handlers.VerifyEmail)
	http.HandleFunc("/api/auth/resend-verification", handlers.ResendVerification)
	http.HandleFunc("/api/auth/forgot-password", handlers.ForgotPassword)
	http.HandleFunc("/api/auth/reset-password", handlers.ResetPassword)
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
