package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/auto-patcher/skills/internal/github"
	"github.com/auto-patcher/skills/internal/prompts"
	"github.com/auto-patcher/skills/internal/runner"
	"github.com/auto-patcher/skills/internal/state"
)

// runPatch runs a full patch cycle against a single fork. --repo is required;
// --upstream and --last-patched may be passed by the caller (fully computed by
// the scan job) or left empty, in which case they are computed here from the
// repo's PATCHER.md. State is re-checked at run time, since it may have changed
// since the scan that scheduled this invocation.
func runPatch(args []string) error {
	fs := flag.NewFlagSet("patch", flag.ExitOnError)
	repo := fs.String("repo", "", "fork to patch, owner/repo (required)")
	upstream := fs.String("upstream", "", "upstream owner/repo (computed from PATCHER.md if empty)")
	lastPatched := fs.String("last-patched", "", "last patched upstream tag (computed from PATCHER.md if empty)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *repo == "" {
		return fmt.Errorf("patch: --repo is required")
	}
	githubToken, err := requireEnv("GITHUB_TOKEN")
	if err != nil {
		return err
	}
	if !hasLLMKey() {
		return fmt.Errorf("set one of ANTHROPIC_API_KEY, OPENAI_API_KEY, or OPENROUTER_API_KEY")
	}

	ctx, cancel := signalContext()
	defer cancel()

	client := github.NewClient(githubToken)
	r := runner.New(githubToken)

	// Re-read repo info at run time — state may have changed since the scan.
	info, err := client.RepoInfo(ctx, *repo)
	if err != nil {
		return fmt.Errorf("read repo info: %w", err)
	}
	if !state.Determine(info).Actionable() {
		slog.Info("repo no longer actionable, skipping", "repo", *repo)
		return nil
	}

	// Fill any flag the caller left empty from the repo's own metadata.
	up, lp := *upstream, *lastPatched
	if up == "" {
		up = info.Upstream
	}
	if lp == "" {
		lp = info.LastPatched
	}

	prompt, err := prompts.RenderCycle(prompts.Context{
		Repo:        *repo,
		Upstream:    up,
		LastPatched: lp,
	})
	if err != nil {
		return fmt.Errorf("render prompt: %w", err)
	}

	if _, err := client.AcquireLock(ctx, *repo); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer client.ReleaseLock(ctx, *repo)

	slog.Info("starting patch cycle", "repo", *repo, "upstream", up, "last_patched", lp)
	if err := r.Run(ctx, runner.Job{Repo: *repo, Prompt: prompt}); err != nil {
		client.PostFailure(ctx, *repo, err)
		return fmt.Errorf("cycle failed: %w", err)
	}
	slog.Info("patch cycle complete", "repo", *repo)
	return nil
}
