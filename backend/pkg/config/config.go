package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTIssuer     string
	JWTTTLMinutes int
}

// Load reads environment variables, optionally from a .env file if present.
func Load() Config {
	// Try to load .env if it exists; ignore error if file not found
	_ = godotenv.Load()

	cfg := Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-secret-change"),
		JWTIssuer:     getEnv("JWT_ISSUER", "hr-service"),
		JWTTTLMinutes: getEnvInt("JWT_TTL_MINUTES", 60),
	}
	return cfg
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
