package runner

import (
	"context"
	"fmt"

	"github.com/auto-patcher/dispatcher/internal/config"
)

// NomadRunner dispatches a parameterized Nomad job for each patch cycle.
//
// TODO: implement via the Nomad HTTP API.
// The job should be a parameterized batch job with TARGET_REPO as a meta field.
// See: https://developer.hashicorp.com/nomad/api-docs/jobs#dispatch-job
type NomadRunner struct {
	cfg *config.Config
}

func (r *NomadRunner) Run(ctx context.Context, job Job) error {
	// POST /v1/job/<job_name>/dispatch
	// Body: {"Meta": {"TARGET_REPO": job.Repo}}
	// Then poll /v1/job/<dispatch_id> until terminal.
	return fmt.Errorf("nomad runner: not yet implemented")
}
