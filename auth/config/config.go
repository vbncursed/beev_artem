package config

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Auth     AuthConfig     `yaml:"auth"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type AuthConfig struct {
	JWTSecret                  string `yaml:"jwt_secret"`
	AccessTTLSeconds           int64  `yaml:"access_ttl_seconds"`
	RefreshTTLSeconds          int64  `yaml:"refresh_ttl_seconds"`
	RateLimitLoginPerMinute    int    `yaml:"rate_limit_login_per_minute"`
	RateLimitRegisterPerMinute int    `yaml:"rate_limit_register_per_minute"`
	RateLimitRefreshPerMinute  int    `yaml:"rate_limit_refresh_per_minute"`
}

type ServerConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

const (
	envJWTSecret     = "AUTH_JWT_SECRET"
	envDBPassword    = "AUTH_DB_PASSWORD"
	envRedisPassword = "AUTH_REDIS_PASSWORD"

	jwtSecretMinLen     = 32
	jwtSecretPlaceholder = "CHANGE_ME_IN_PRODUCTION"
)

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal YAML: %w", err)
	}

	overlaySecretsFromEnv(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func overlaySecretsFromEnv(cfg *Config) {
	if v := os.Getenv(envJWTSecret); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv(envDBPassword); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv(envRedisPassword); v != "" {
		cfg.Redis.Password = v
	}
}

func validate(cfg *Config) error {
	switch {
	case cfg.Auth.JWTSecret == "":
		return fmt.Errorf("auth.jwt_secret is empty: set %s env var", envJWTSecret)
	case cfg.Auth.JWTSecret == jwtSecretPlaceholder:
		return fmt.Errorf("auth.jwt_secret is the literal placeholder %q: set %s to a real secret", jwtSecretPlaceholder, envJWTSecret)
	case len(cfg.Auth.JWTSecret) < jwtSecretMinLen:
		return fmt.Errorf("auth.jwt_secret is too short: %d bytes, need >= %d", len(cfg.Auth.JWTSecret), jwtSecretMinLen)
	}

	if cfg.Auth.AccessTTLSeconds <= 0 {
		return errors.New("auth.access_ttl_seconds must be > 0")
	}
	if cfg.Auth.RefreshTTLSeconds <= 0 {
		return errors.New("auth.refresh_ttl_seconds must be > 0")
	}

	return nil
}
