package config

import (
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database  DatabaseConfig  `yaml:"database"`
	Server    ServerConfig    `yaml:"server"`
	Yandex    YandexConfig    `yaml:"yandex"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
}

type ServerConfig struct {
	GRPCAddr string    `yaml:"grpc_addr"`
	TLS      TLSConfig `yaml:"tls"`
}

// TLSConfig is opt-in: leave both fields empty to keep plaintext gRPC (the
// default for docker-compose, where the network boundary is the trust
// boundary). Multiagent is internal-only so plaintext is the norm; this
// toggle exists only for environments that mandate encryption everywhere.
type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// Enabled reports whether both halves of a usable TLS keypair are configured.
func (t TLSConfig) Enabled() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

// YandexConfig points at Yandex Cloud Foundation Models. APIKey is loaded
// from the YANDEX_API_KEY env var by LoadConfig — never put it in YAML, the
// file is committed and routinely shared. FolderID and Model are safe to
// keep in YAML (FolderID is not a secret on its own; Model is a public
// identifier).
type YandexConfig struct {
	FolderID        string        `yaml:"folder_id"`
	Model           string        `yaml:"model"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
	MaxOutputTokens int           `yaml:"max_output_tokens"`
	// APIKey is populated from env, never YAML — see LoadConfig.
	APIKey string `yaml:"-"`
}

// RateLimitConfig caps how many LLM calls per second multiagent will make.
// Defends Yandex billing from a runaway loop in analysis. Burst absorbs
// short bursts without dropping calls.
type RateLimitConfig struct {
	RPS   float64 `yaml:"rps"`
	Burst int     `yaml:"burst"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Secrets must come from env — keeps the API key out of git.
	cfg.Yandex.APIKey = os.Getenv("YANDEX_API_KEY")

	// Defaults so YAML stays minimal.
	if cfg.Yandex.RequestTimeout == 0 {
		cfg.Yandex.RequestTimeout = 60 * time.Second
	}
	if cfg.Yandex.MaxOutputTokens == 0 {
		cfg.Yandex.MaxOutputTokens = 1500
	}
	if cfg.RateLimit.RPS == 0 {
		cfg.RateLimit.RPS = 10
	}
	if cfg.RateLimit.Burst == 0 {
		cfg.RateLimit.Burst = 5
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", filename, err)
	}

	return &cfg, nil
}

// validate rejects configs that would otherwise crash later in less obvious
// places. Caught at boot with a clear error instead of a confusing dial /
// panic at first request.
func (c *Config) validate() error {
	switch {
	case c.Server.GRPCAddr == "":
		return fmt.Errorf("server.grpc_addr is required")
	case c.Database.Host == "":
		return fmt.Errorf("database.host is required")
	case c.Database.Port == 0:
		return fmt.Errorf("database.port is required")
	case c.Database.DBName == "":
		return fmt.Errorf("database.name is required")
	case c.Yandex.FolderID == "":
		return fmt.Errorf("yandex.folder_id is required")
	case c.Yandex.Model == "":
		return fmt.Errorf("yandex.model is required")
	case c.Yandex.APIKey == "":
		return fmt.Errorf("YANDEX_API_KEY env var is required")
	}
	return nil
}
