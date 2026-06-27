package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/auto-patcher/dispatcher/internal/config"
	"github.com/auto-patcher/dispatcher/internal/github"
	"github.com/auto-patcher/dispatcher/internal/runner"
)

// Scheduler drives the worker pool. A background goroutine scans the org on
// a fixed interval and enqueues actionable repos in staleness order (most
// overdue first). Workers pace themselves with a configurable inter-job delay.
type Scheduler struct {
	cfg    *config.Config
	client *github.Client
	runner runner.Runner
	queue  chan string   // buffered; repos ready to process
	active sync.Map     // string → struct{}: repos currently in a container
}

func New(cfg *config.Config, client *github.Client, r runner.Runner) *Scheduler {
	return &Scheduler{
		cfg:    cfg,
		client: client,
		runner: r,
		queue:  make(chan string, 256),
	}
}

// Run starts workers and the scan loop. Blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	for i := 0; i < s.cfg.Workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s.worker(ctx, id)
		}(i)
	}

	// Scan immediately on start, then on each tick.
	s.scan(ctx)
	ticker := time.NewTicker(s.cfg.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(s.queue)
			wg.Wait()
			return nil
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan queries GitHub for all actionable repos, sorted by staleness, and
// enqueues any not already active or buffered.
func (s *Scheduler) scan(ctx context.Context) {
	slog.Info("scanning org", "org", s.cfg.Org)
	ranked, err := s.client.RankedRepos(ctx)
	if err != nil {
		slog.Error("scan failed", "err", err)
		return
	}

	enqueued := 0
	for _, entry := range ranked {
		if _, active := s.active.Load(entry.Repo); active {
			continue // already being processed
		}
		select {
		case s.queue <- entry.Repo:
			enqueued++
		default:
			slog.Warn("queue full; repo will be picked up on next scan", "repo", entry.Repo)
		}
	}
	slog.Info("scan complete", "actionable", len(ranked), "enqueued", enqueued)
}

func (s *Scheduler) worker(ctx context.Context, id int) {
	log := slog.With("worker", id)
	for {
		select {
		case <-ctx.Done():
			return
		case repo, ok := <-s.queue:
			if !ok {
				return
			}
			s.active.Store(repo, struct{}{})
			log.Info("starting cycle", "repo", repo)
			s.process(ctx, repo, log)
			s.active.Delete(repo)

			// Pace: sleep before taking the next job.
			select {
			case <-ctx.Done():
				return
			case <-time.After(s.cfg.WorkerDelay):
			}
		}
	}
}

func (s *Scheduler) process(ctx context.Context, repo string, log *slog.Logger) {
	if _, err := s.client.AcquireLock(ctx, repo); err != nil {
		log.Error("failed to acquire lock", "repo", repo, "err", err)
		return
	}
	defer s.client.ReleaseLock(ctx, repo)

	if err := s.runner.Run(ctx, runner.Job{Repo: repo}); err != nil {
		log.Error("cycle failed", "repo", repo, "err", err)
		s.client.PostFailure(ctx, repo, err)
	} else {
		log.Info("cycle complete", "repo", repo)
	}
}
