package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTIssuer          string
	JWTTTLMinutes      int
	OpenRouterAPIKey   string
	OpenRouterBase     string
	OpenRouterModel    string
	OpenRouterReferer  string
	OpenRouterAppTitle string
}

// Load reads environment variables, optionally from a .env file if present.
func Load() Config {
	// Try to load .env if it exists; ignore error if file not found
	_ = godotenv.Load()

	cfg := Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change"),
		JWTIssuer:          getEnv("JWT_ISSUER", "hr-service"),
		JWTTTLMinutes:      getEnvInt("JWT_TTL_MINUTES", 60),
		OpenRouterAPIKey:   os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterBase:     getEnv("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
		OpenRouterModel:    getEnv("OPENROUTER_MODEL", "qwen/qwen2.5-32b-instruct"),
		OpenRouterReferer:  getEnv("OPENROUTER_HTTP_REFERER", ""),
		OpenRouterAppTitle: getEnv("OPENROUTER_APP_TITLE", "hr-service"),
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
