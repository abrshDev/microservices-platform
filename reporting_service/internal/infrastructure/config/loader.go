package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv attempts to load the .env file from the root
func LoadEnv() {
	paths := []string{
		".env",
		"../.env",
	}

	for _, path := range paths {
		if err := godotenv.Load(path); err == nil {
			log.Printf("Loaded environment from %s", path)
			return
		}
	}

	log.Println("Using existing environment variables")
}
