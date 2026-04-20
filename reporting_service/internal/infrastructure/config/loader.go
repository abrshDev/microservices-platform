package config

import (
	"log"

	"github.com/joho/godotenv"
)

// LoadEnv attempts to load the .env file from the root
func LoadEnv() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}
}
