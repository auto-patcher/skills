package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the non-secret operational settings for a single patch run.
//
// Secrets (the GitHub and Anthropic tokens) are intentionally NOT part of this
// struct: in the GitHub Actions deployment they are injected as environment
// variables from repository/organization secrets and read via GitHubToken /
// AnthropicKey below. The YAML file is committed in the clear and carries only
// settings that are safe to read in a public repository.
type Config struct {
	Org         string        `yaml:"org"`
	Workers     int           `yaml:"workers"`
	WorkerDelay time.Duration `yaml:"worker_delay"`
	Exclude     []string      `yaml:"exclude"`
}

// Load reads the plaintext config file. The org may also be supplied via the
// AUTOPATCHER_ORG environment variable, which takes precedence over the file —
// handy for overriding a run from the workflow without editing the repo.
func Load(path string) (*Config, error) {
	cfg := &Config{
		Workers: 3,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.NewDecoder(bytes.NewReader(data)).Decode(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if org := os.Getenv("AUTOPATCHER_ORG"); org != "" {
		cfg.Org = org
	}
	if cfg.Org == "" {
		return nil, fmt.Errorf("config: org is required (set it in %s or via AUTOPATCHER_ORG)", path)
	}
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}

	return cfg, nil
}

// GitHubToken returns the token used for all GitHub API calls. In the workflow
// this is the AUTOPATCHER_GITHUB_TOKEN secret, exported as GITHUB_TOKEN.
func (c *Config) GitHubToken() string { return os.Getenv("GITHUB_TOKEN") }

// AnthropicKey returns the key passed to the claude subprocess. In the workflow
// this is the ANTHROPIC_API_KEY secret.
func (c *Config) AnthropicKey() string { return os.Getenv("ANTHROPIC_API_KEY") }
