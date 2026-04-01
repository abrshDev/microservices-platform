package config

import (
	"log"
	"os"
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
