package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/auto-patcher/skills/internal/github"
)

// runScan lists every actionable repo in the org, most-stale-first, and writes
// one "owner/repo" per line to stdout. Logs go to stderr, so the stdout stream
// can be piped straight into GNU parallel as the job list.
func runScan(args []string) error {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	org := fs.String("org", os.Getenv("AUTOPATCHER_ORG"), "GitHub organization to scan (or set AUTOPATCHER_ORG)")
	exclude := fs.String("exclude", "", "comma-separated repo names to skip")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *org == "" {
		return fmt.Errorf("scan: --org is required (or set AUTOPATCHER_ORG)")
	}
	token, err := requireEnv("GITHUB_TOKEN")
	if err != nil {
		return err
	}

	ctx, cancel := signalContext()
	defer cancel()

	client := github.NewClient(token)
	ranked, err := client.RankedRepos(ctx, *org, splitCSV(*exclude))
	if err != nil {
		return err
	}
	slog.Info("scan complete", "org", *org, "actionable", len(ranked))

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for _, r := range ranked {
		fmt.Fprintln(w, r.Repo)
	}
	return nil
}
