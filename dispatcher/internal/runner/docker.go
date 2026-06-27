package runner

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/auto-patcher/dispatcher/internal/config"
)

// DockerRunner runs each patch cycle in an isolated Docker container.
// The container is pulled, run to completion, and removed automatically.
type DockerRunner struct {
	cfg *config.Config
}

func (r *DockerRunner) Run(ctx context.Context, job Job) error {
	args := []string{
		"run", "--rm",
		"-e", "GITHUB_TOKEN=" + r.cfg.GitHubToken(),
		"-e", "ANTHROPIC_API_KEY=" + r.cfg.AnthropicKey(),
		"-e", "TARGET_REPO=" + job.Repo,
		r.cfg.Runner.Docker.Image,
		job.Repo,
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed for %s: %w\noutput:\n%s", job.Repo, err, out)
	}
	return nil
}
