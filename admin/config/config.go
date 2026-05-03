// Package config loads admin's YAML config + env overrides. Same shape
// as other services: APP_ENV / configPath selector, validate() on boot.
package config

import (
	"errors"
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthClientCfg  `yaml:"auth"`
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

type AuthClientCfg struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type ServerConfig struct {
	GRPCAddr string    `yaml:"grpc_addr"`
	TLS      TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", filename, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Database.Host == "" {
		return errors.New("database.host is required")
	}
	if c.Database.Port == 0 {
		return errors.New("database.port is required")
	}
	if c.Database.Username == "" {
		return errors.New("database.username is required")
	}
	if c.Database.DBName == "" {
		return errors.New("database.name is required")
	}
	if c.Auth.GRPCAddr == "" {
		return errors.New("auth.grpc_addr is required")
	}
	if c.Server.GRPCAddr == "" {
		return errors.New("server.grpc_addr is required")
	}
	return nil
}
