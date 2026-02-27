package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	APIPort   int    `yaml:"api_port"`
	ProxyPort int    `yaml:"proxy_port"`
	TargetPort int   `yaml:"target_port"`
	Token     string `yaml:"token"`
}

type ClientConfig struct {
	ServerURL string `yaml:"server_url"`
	Token     string `yaml:"token"`
	Interval  int    `yaml:"interval"`
}

type Config struct {
	Server ServerConfig `yaml:"server"`
	Client ClientConfig `yaml:"client"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			APIPort:    8080,
			ProxyPort:  9090,
			TargetPort: 22,
		},
		Client: ClientConfig{
			Interval: 60,
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return cfg, nil
}
