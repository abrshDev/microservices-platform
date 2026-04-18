package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AuthConfig struct {
	JWTSecret string
}

func LoadAuthConfig() *AuthConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {

		log.Fatal("FATAL: JWT_SECRET environment variable is missing")
	}

	return &AuthConfig{
		JWTSecret: secret,
	}
}
func LoadEnv() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Note: .env not found, using system environment variables")
	}
}
