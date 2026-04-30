package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
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

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config %s: %w", filename, err)
	}

	return &cfg, nil
}

// validate rejects configs that would otherwise crash later in less obvious
// places — empty addresses, empty DB host. Caught at boot with a clear error
// instead of a confusing dial / panic at first request.
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
	}
	return nil
}
