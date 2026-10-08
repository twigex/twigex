// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

const (
	jobRetention    = 24 * time.Hour
	cleanupInterval = time.Hour
)

type CleanupJob struct {
	store store.JobsStore
}

func NewCleanupJob(store store.JobsStore) CleanupJob {
	return CleanupJob{store: store}
}

func (j CleanupJob) Type() string {
	return model.JobTypeCleanupJobs
}

func (j CleanupJob) Schedule() (time.Duration, bool) {
	return cleanupInterval, true
}

func (j CleanupJob) Run(ctx context.Context, job *model.Job) error {
	return j.store.DeleteOld(time.Now().Add(-jobRetention).Unix())
}
