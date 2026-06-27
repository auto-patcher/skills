// Package runner clones a fork repository and invokes claude non-interactively
// with a pre-rendered prompt. Isolation (Docker, Nomad) is a deployment concern
// handled at the infrastructure level, not here.
package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/auto-patcher/skills/dispatcher/internal/config"
)

// Job describes a single patch cycle run.
type Job struct {
	Repo   string // "owner/repo" — cloned into a temp directory
	Prompt string // fully rendered prompt, passed to claude via stdin
}

// Runner clones a repo and runs a Claude cycle against it.
type Runner struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Runner {
	return &Runner{cfg: cfg}
}

// Run clones the fork into a temp directory, passes the rendered prompt to
// claude via stdin, and removes the temp directory when done.
func (r *Runner) Run(ctx context.Context, job Job) error {
	workDir, err := os.MkdirTemp("", "autopatcher-*")
	if err != nil {
		return fmt.Errorf("mkdirtemp: %w", err)
	}
	defer os.RemoveAll(workDir)

	if err := r.clone(ctx, job.Repo, workDir); err != nil {
		return err
	}
	return r.runClaude(ctx, job.Prompt, workDir)
}

func (r *Runner) clone(ctx context.Context, repo, dir string) error {
	cmd := exec.CommandContext(ctx, "gh", "repo", "clone", repo, dir, "--", "--depth=50")
	cmd.Env = append(os.Environ(), "GITHUB_TOKEN="+r.cfg.GitHubToken())
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clone %s: %w\n%s", repo, err, out)
	}
	return nil
}

func (r *Runner) runClaude(ctx context.Context, prompt, dir string) error {
	// TODO: update flag/invocation once the auto-patcher CC fork's
	// non-interactive API is finalised.
	cmd := exec.CommandContext(ctx, "claude", "--print")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_API_KEY="+r.cfg.AnthropicKey(),
		"GITHUB_TOKEN="+r.cfg.GitHubToken(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("claude: %w\noutput:\n%s", err, out)
	}
	return nil
}
