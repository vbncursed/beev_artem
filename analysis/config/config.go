package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database   DatabaseConfig   `yaml:"database"`
	Server     ServerConfig     `yaml:"server"`
	Auth       AuthConfig       `yaml:"auth"`
	MultiAgent MultiAgentConfig `yaml:"multiagent"`
}

// AuthConfig points at the auth gRPC service. Analysis calls
// auth.ValidateAccessToken on every protected RPC instead of trusting
// caller-supplied identity headers — defense in depth, since anyone with
// network access to analysis:50054 could otherwise impersonate any user
// just by setting x-user-id in metadata.
type AuthConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
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
// boundary). When both cert_file and key_file are set, the gRPC server
// upgrades to TLS — every client (the gateway, in particular) must be
// reconfigured to dial with matching credentials.
//
// For real production, prefer service-mesh-managed mTLS (Istio, Linkerd, …)
// over hand-rolled cert files in YAML. This toggle exists as an escape hatch
// when no mesh is available.
type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// Enabled reports whether both halves of a usable TLS keypair are configured.
func (t TLSConfig) Enabled() bool {
	return t.CertFile != "" && t.KeyFile != ""
}

// MultiAgentConfig points at the multiagent gRPC service. Analysis calls it
// for LLM-backed HR decisions during StartAnalysis.
type MultiAgentConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
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
	case c.Auth.GRPCAddr == "":
		return fmt.Errorf("auth.grpc_addr is required")
	case c.MultiAgent.GRPCAddr == "":
		return fmt.Errorf("multiagent.grpc_addr is required")
	case c.Database.Host == "":
		return fmt.Errorf("database.host is required")
	case c.Database.Port == 0:
		return fmt.Errorf("database.port is required")
	case c.Database.DBName == "":
		return fmt.Errorf("database.name is required")
	}
	return nil
}
