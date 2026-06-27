package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/auto-patcher/skills/internal/config"
	"github.com/auto-patcher/skills/internal/github"
	"github.com/auto-patcher/skills/internal/runner"
	"github.com/auto-patcher/skills/internal/scheduler"
)

func main() {
	configPath := flag.String("config", envOr("AUTOPATCHER_CONFIG", "config.yaml"), "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	if cfg.GitHubToken() == "" {
		slog.Error("missing GITHUB_TOKEN environment variable")
		os.Exit(1)
	}
	if cfg.AnthropicKey() == "" {
		slog.Error("missing ANTHROPIC_API_KEY environment variable")
		os.Exit(1)
	}

	client := github.NewClient(cfg)
	r := runner.New(cfg)
	sched := scheduler.New(cfg, client, r)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	slog.Info("autopatcher run starting",
		"org", cfg.Org,
		"workers", cfg.Workers,
		"worker_delay", cfg.WorkerDelay,
	)

	if err := sched.RunOnce(ctx); err != nil {
		slog.Error("autopatcher run failed", "err", err)
		os.Exit(1)
	}

	slog.Info("autopatcher run complete")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
