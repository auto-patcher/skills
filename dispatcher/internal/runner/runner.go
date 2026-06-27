package runner

import (
	"context"
	"fmt"

	"github.com/auto-patcher/dispatcher/internal/config"
)

// Job describes a single patch cycle run against a fork repository.
type Job struct {
	Repo string // "owner/repo"
}

// Runner executes a patch cycle job in an isolated environment.
type Runner interface {
	Run(ctx context.Context, job Job) error
}

// New returns the Runner indicated by cfg.Runner.Type.
func New(cfg *config.Config) (Runner, error) {
	switch cfg.Runner.Type {
	case "docker", "":
		return &DockerRunner{cfg: cfg}, nil
	case "nomad":
		return &NomadRunner{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown runner type: %q", cfg.Runner.Type)
	}
}
