// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

const moveFilesLockTTL = 5 * time.Minute

type FileLocker interface {
	Lock(lockOwner string, fileIDs []string, ttl time.Duration) error
	Unlock(lockOwner string) error
	RenewLock(lockOwner string, ttl time.Duration) error
}

type FileMover interface {
	MoveFilesAcrossDisks(ctx context.Context, jobID string) error
}

type MoveFilesJob struct {
	files FileLocker
	mover FileMover
}

func NewMoveFilesJob(files FileLocker, mover FileMover) MoveFilesJob {
	return MoveFilesJob{files: files, mover: mover}
}

func (j MoveFilesJob) Type() string {
	return model.JobTypeMoveFiles
}

func (j MoveFilesJob) Run(ctx context.Context, job *model.Job) error {
	allFiles, err := job.GetAllFileIDs()
	if err != nil {
		return err
	}

	if err := j.files.Lock(job.ID, allFiles, moveFilesLockTTL); err != nil {
		return err
	}

	defer func() {
		if err := j.files.Unlock(job.ID); err != nil {
			tlog.Warnw("failed to unlock files", "job_id", job.ID, "error", err)
		}
	}()

	// A lock that cannot be renewed may already have expired, so the move is
	// abandoned rather than left running over files another process could take.
	moveCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go j.renewLock(moveCtx, cancel, job.ID)

	return j.mover.MoveFilesAcrossDisks(moveCtx, job.ID)
}

func (j MoveFilesJob) renewLock(ctx context.Context, cancel context.CancelFunc, jobID string) {
	ticker := time.NewTicker(moveFilesLockTTL / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := j.files.RenewLock(jobID, moveFilesLockTTL); err != nil {
				tlog.Errorw("failed to renew file lock", "job_id", jobID, "error", err)
				cancel()
				return
			}
		}
	}
}
