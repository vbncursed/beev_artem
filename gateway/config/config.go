package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"`
	Vacancy  VacancyConfig  `yaml:"vacancy"`
	Resume   ResumeConfig   `yaml:"resume"`
	Analysis AnalysisConfig `yaml:"analysis"`
}

type ServerConfig struct {
	HTTPAddr string `yaml:"http_addr"`
}

type AuthConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type VacancyConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type ResumeConfig struct {
	GRPCAddr string `yaml:"grpc_addr"`
}

type AnalysisConfig struct {
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
// places — missing addresses surface as connection-refused at first
// request instead of an actionable boot error.
func (c *Config) validate() error {
	switch {
	case c.Server.HTTPAddr == "":
		return fmt.Errorf("server.http_addr is required")
	case c.Auth.GRPCAddr == "":
		return fmt.Errorf("auth.grpc_addr is required")
	case c.Vacancy.GRPCAddr == "":
		return fmt.Errorf("vacancy.grpc_addr is required")
	case c.Resume.GRPCAddr == "":
		return fmt.Errorf("resume.grpc_addr is required")
	case c.Analysis.GRPCAddr == "":
		return fmt.Errorf("analysis.grpc_addr is required")
	}
	return nil
}
