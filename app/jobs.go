// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) GetUserRunningJobs(user model.User) ([]model.Job, *model.AppError) {
	jobs, err := a.Store.Jobs.GetRunningForUser(user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve running jobs",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("job.retrieval_failed", http.StatusInternalServerError)
	}

	return jobs, nil
}

func (a *App) CancelUserJob(user model.User, jobID string) *model.AppError {
	job, err := a.Store.Jobs.GetForUserByID(user.ID, jobID)
	if err != nil {
		tlog.Errorw("Failed to retrieve job for cancellation",
			"user_id", user.ID,
			"job_id", jobID,
			"error", err,
		)
		return model.NewAppError("job.not_found", http.StatusInternalServerError)
	}

	if job.Status != model.JobStatusPending && job.Status != model.JobStatusRunning {
		return model.NewAppError("job.cancel_invalid_status", http.StatusBadRequest)
	}

	if err = a.Store.Jobs.UpdateStatus(job.ID, model.JobStatusCancelled, "Job cancelled by user"); err != nil {
		tlog.Errorw("Failed to cancel job",
			"user_id", user.ID,
			"job_id", jobID,
			"error", err,
		)
		return model.NewAppError("job.cancel_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) AcknowledgeUserJob(user model.User, jobID string) *model.AppError {
	if err := a.Store.Jobs.AcknowledgeForUser(user.ID, jobID); err != nil {
		tlog.Errorw("Failed to acknowledge job",
			"user_id", user.ID,
			"job_id", jobID,
			"error", err,
		)
		return model.NewAppError("job.acknowledge_failed", http.StatusInternalServerError)
	}

	return nil
}
