package config

import (
	"log"
	"os"
	"time"

	"github.com/MetaDandy/go-fiber-skeleton/config/seed"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	Port string
)

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	Port = os.Getenv("PORT")
	if Port == "" {
		Port = "8001"
	}

	maxRetries := 10
	for i := range maxRetries {
		dns := os.Getenv("DATABASE_URL")
		if dns == "" {
			log.Fatal("DATABASE_URL not set in .env file")
		}

		Migrate(dns)

		DB, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
		if err == nil {
			log.Printf("Database connected successfully after %d attempt(s)", i+1)
			seed.Seeder(DB)
			return
		}

		log.Printf("Failed to connect to database, retrying (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Error connecting to database after %d retries", maxRetries)

	log.Printf("Database connected")
}
