package config

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
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
	data, err := readConfig(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	cfg := &Config{
		Workers:      3,
		ScanInterval: 15 * time.Minute,
		WorkerDelay:  10 * time.Minute,
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

// readConfig attempts to decrypt path with sops. If sops is unavailable or
// the file is not a sops-encrypted file, it falls back to reading plaintext.
func readConfig(path string) ([]byte, error) {
	out, err := exec.Command("sops", "--decrypt", path).Output()
	if err == nil {
		slog.Debug("loaded config via sops", "path", path)
		return out, nil
	}
	slog.Debug("sops unavailable or file not encrypted, reading plaintext", "path", path)
	return os.ReadFile(path)
}

func (c *Config) GitHubToken() string  { return os.Getenv(c.GitHub.TokenEnv) }
func (c *Config) AnthropicKey() string { return os.Getenv(c.Anthropic.APIKeyEnv) }
