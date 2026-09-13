package main

import (
	"log"

	"github.com/Traderong/omnivibe-api/database"
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

	log.Println("OmniVibe backend started successfully")
	log.Println("PostgreSQL connected successfully")
}