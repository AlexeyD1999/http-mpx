package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server *Server `yaml:"server"`
}

type Server struct {
	Network string `yaml:"network"`
	Address string `yaml:"address"`
}

func LoadConfig() (*Config, error) {
	configFilePath := "config/config.yml"

	cfg := Config{}

	raw, err := os.ReadFile(filepath.Clean(configFilePath))
	if err != nil {
		return nil, fmt.Errorf("unable to read config file %w", err)
	}

	if err = yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("unable to read config as YML: %w", err)
	}

	return &cfg, nil
}
