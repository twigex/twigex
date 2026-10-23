// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeGraphRebuilder struct {
	result model.GraphRebuildResult
	err    error
}

func (f fakeGraphRebuilder) RebuildGraph(context.Context) (model.GraphRebuildResult, error) {
	return f.result, f.err
}

type fakePayloadStore struct {
	store.JobsStore
	payload json.RawMessage
}

func (f *fakePayloadStore) UpdatePayload(_ string, payload json.RawMessage) error {
	f.payload = payload

	return nil
}

func TestGraphRebuildJobKeepsTheResultWhenItFails(t *testing.T) {
	jobsStore := &fakePayloadStore{}
	rebuilder := fakeGraphRebuilder{
		result: model.GraphRebuildResult{Workspaces: 3, RelationsAdded: 12},
		err:    errors.New("database went away"),
	}

	job := NewGraphRebuildJob(rebuilder, jobsStore)

	err := job.Run(context.Background(), &model.Job{ID: "j1"})
	if err == nil {
		t.Fatalf("want the rebuild error returned")
	}

	saved := model.Job{Payload: jobsStore.payload}

	result, err := saved.GraphRebuildPayload()
	if err != nil {
		t.Fatalf("GraphRebuildPayload: %v", err)
	}

	if result.Workspaces != 3 || result.RelationsAdded != 12 {
		t.Errorf("want the partial result saved, got %+v", result)
	}
}
