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
