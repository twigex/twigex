// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"errors"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (r *Registry) run(ctx context.Context, job *model.Job, impl Job) {
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go r.watchCancellation(jobCtx, cancel, job.ID)

	status, detail := statusFromError(impl.Run(jobCtx, job))

	_ = r.store.UpdateStatus(job.ID, status, detail)

	tlog.Infow("job finished",
		"job_id", job.ID,
		"job_type", job.Type,
		"status", status,
	)
}

type RefusedError struct {
	Reason string
}

func (e *RefusedError) Error() string {
	return e.Reason
}

func statusFromError(err error) (string, string) {
	var refused *RefusedError

	switch {
	case err == nil:
		return model.JobStatusCompleted, ""
	case errors.Is(err, context.Canceled):
		return model.JobStatusCancelled, "Job was cancelled"
	case errors.As(err, &refused):
		return model.JobStatusRefused, refused.Reason
	default:
		return model.JobStatusFailed, err.Error()
	}
}

func (r *Registry) watchCancellation(
	ctx context.Context,
	cancel context.CancelFunc,
	jobID string,
) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cancelled, err := r.store.IsCancelled(jobID)
			if err != nil {
				tlog.Errorw("cancel check failed", "job_id", jobID, "error", err.Error())
				continue
			}

			if cancelled {
				tlog.Infow("job cancelled", "job_id", jobID)
				cancel()
				return
			}
		}
	}
}
