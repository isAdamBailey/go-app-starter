// Command migrate applies database migrations.
package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/isAdamBailey/go-app-starter/backend/internal/db"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if err := db.Migrate(dsn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("migrations applied")
}
