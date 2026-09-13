package main

import (
	"log"
	"net/http"

	"github.com/Traderong/omnivibe-api/database"
	"github.com/Traderong/omnivibe-api/handlers"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	http.HandleFunc("/api/auth/register", handlers.Register)

	log.Println("OmniVibe backend started on http://localhost:8080")
	log.Println("PostgreSQL connected successfully")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}