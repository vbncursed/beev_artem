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

	return &cfg, nil
}
