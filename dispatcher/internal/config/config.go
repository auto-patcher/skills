package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Org          string        `yaml:"org"`
	Workers      int           `yaml:"workers"`
	ScanInterval time.Duration `yaml:"scan_interval"`
	WorkerDelay  time.Duration `yaml:"worker_delay"`
	Exclude      []string      `yaml:"exclude"`

	Runner RunnerConfig `yaml:"runner"`

	GitHub struct {
		TokenEnv string `yaml:"token_env"`
	} `yaml:"github"`

	Anthropic struct {
		APIKeyEnv string `yaml:"api_key_env"`
	} `yaml:"anthropic"`
}

type RunnerConfig struct {
	Type   string       `yaml:"type"` // "docker" or "nomad"
	Docker DockerConfig `yaml:"docker"`
	Nomad  NomadConfig  `yaml:"nomad"`
}

type DockerConfig struct {
	Image string `yaml:"image"`
}

type NomadConfig struct {
	Address string `yaml:"address"`
	JobName string `yaml:"job_name"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	cfg := &Config{
		Workers:      3,
		ScanInterval: 15 * time.Minute,
		WorkerDelay:  10 * time.Minute,
	}
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func (c *Config) GitHubToken() string { return os.Getenv(c.GitHub.TokenEnv) }
func (c *Config) AnthropicKey() string { return os.Getenv(c.Anthropic.APIKeyEnv) }
