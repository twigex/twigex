// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"encoding/json"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type GraphRebuilder interface {
	RebuildGraph(ctx context.Context) (model.GraphRebuildResult, error)
}

type GraphRebuildJob struct {
	rebuilder GraphRebuilder
	store     store.JobsStore
}

func NewGraphRebuildJob(rebuilder GraphRebuilder, store store.JobsStore) GraphRebuildJob {
	return GraphRebuildJob{rebuilder: rebuilder, store: store}
}

func (j GraphRebuildJob) Type() string {
	return model.JobTypeGraphRebuild
}

func (j GraphRebuildJob) Run(ctx context.Context, job *model.Job) error {
	result, rebuildErr := j.rebuilder.RebuildGraph(ctx)

	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}

	err = j.store.UpdatePayload(job.ID, payload)
	if err != nil {
		return err
	}

	return rebuildErr
}
