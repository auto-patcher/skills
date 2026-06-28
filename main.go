// Command auto-patcher is a multipurpose CLI for keeping a GitHub org of forks
// synchronized with their upstreams. It exposes independent subcommands that
// compose into a GitHub Actions pipeline:
//
//	auto-patcher scan   — list the org's actionable repos, one "owner/repo" per
//	                      line on stdout (the "plan" job).
//	auto-patcher patch  — run a full patch cycle against a single repo, with the
//	                      upstream coordinates supplied via flags or computed
//	                      from the repo's PATCHER.md (one invocation per repo).
//
// Concurrency is owned by the caller: the workflow pipes `scan` into GNU
// `parallel`, which fans out `patch` invocations with a bounded job count.
//
// Secrets are never read from disk — they come from the environment, which the
// workflow populates from GitHub secrets:
//
//	GITHUB_TOKEN      — org-wide repo + issues access (both subcommands)
//	ANTHROPIC_API_KEY — key for the claude subprocess (patch only)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd, args := os.Args[1], os.Args[2:]
	var err error
	switch cmd {
	case "scan":
		err = runScan(args)
	case "patch":
		err = runPatch(args)
	case "help", "-h", "--help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usage()
		os.Exit(2)
	}

	if err != nil {
		slog.Error(cmd+" failed", "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `auto-patcher — keep a GitHub org of forks synced with upstream

usage:
  auto-patcher scan  --org <org> [--exclude a,b]
  auto-patcher patch --repo <owner/repo> [--upstream <owner/repo>] [--last-patched <tag>]

secrets are read from the environment:
  GITHUB_TOKEN       (scan, patch)
  ANTHROPIC_API_KEY  (patch)
`)
}

// signalContext returns a context cancelled on SIGINT/SIGTERM so an in-flight
// cycle unwinds cleanly when a workflow job is cancelled or times out.
func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

// splitCSV parses a comma-separated flag value into a trimmed, non-empty slice.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// requireEnv returns the value of key, or an error naming it if unset.
func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("missing %s environment variable", key)
	}
	return v, nil
}
