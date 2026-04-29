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
	BcryptCost                 int    `yaml:"bcrypt_cost"`
}

type ServerConfig struct {
	GRPCAddr string    `yaml:"grpc_addr"`
	TLS      TLSConfig `yaml:"tls"`
}

// TLSConfig is opt-in: leave both fields empty to keep plaintext gRPC (the
// default for docker-compose and any deployment where mTLS is provided by a
// service mesh / sidecar). When both cert_file and key_file are set, the
// gRPC server upgrades to TLS — all clients (the gateway in particular) must
// be reconfigured to dial with matching credentials.
//
// For real production, prefer service-mesh-managed mTLS (Istio, Linkerd,
// k8s-cilium) over hand-rolled cert files in YAML. This toggle exists as an
// escape hatch when no mesh is available.
type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// Enabled reports whether both halves of a usable TLS keypair are configured.
func (t TLSConfig) Enabled() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

const (
	envJWTSecret     = "AUTH_JWT_SECRET"
	envDBPassword    = "AUTH_DB_PASSWORD"
	envRedisPassword = "AUTH_REDIS_PASSWORD"

	jwtSecretMinLen      = 32
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
	// bcrypt cost: 0 means "fall back to bcrypt.DefaultCost"; anything outside
	// bcrypt's accepted range (4..31) would crash at hashing time, so reject
	// it loudly at boot.
	if cfg.Auth.BcryptCost != 0 && (cfg.Auth.BcryptCost < 4 || cfg.Auth.BcryptCost > 31) {
		return fmt.Errorf("auth.bcrypt_cost must be 0 (default) or in [4..31], got %d", cfg.Auth.BcryptCost)
	}

	return nil
}
