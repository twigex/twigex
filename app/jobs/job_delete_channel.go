// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"

	"github.com/twigex/twigex/model"
)

type ChannelEraser interface {
	EraseChannel(ctx context.Context, channelID string) error
}

type DeleteChannelJob struct {
	eraser ChannelEraser
}

func NewDeleteChannelJob(eraser ChannelEraser) DeleteChannelJob {
	return DeleteChannelJob{eraser: eraser}
}

func (j DeleteChannelJob) Type() string {
	return model.JobTypeDeleteChannel
}

func (j DeleteChannelJob) Run(ctx context.Context, job *model.Job) error {
	payload, err := job.GetDeleteChannelPayload()
	if err != nil {
		return err
	}

	return j.eraser.EraseChannel(ctx, payload.ChannelID)
}
