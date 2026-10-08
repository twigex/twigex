// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeAttachmentStore struct {
	store.WorkspaceStore
	reads int
	fail  bool
}

func (f *fakeAttachmentStore) GetAttachmentsForTasks(_ context.Context, _, fieldID string, taskIDs []string) (map[string][]map[string]interface{}, error) {
	f.reads++
	if f.fail {
		return nil, errors.New("down")
	}
	return map[string][]map[string]interface{}{"t1": {{"name": fieldID + ".pdf"}}}, nil
}

func TestHydrateFileAttachmentsReadsEachFieldOnceForThePage(t *testing.T) {
	ws := &fakeAttachmentStore{}
	a := &App{Store: store.Store{Workspace: ws}}
	headers := []model.WorkspaceHeaders{
		{ID: "docs", Name: "docs", HeaderUsage: "file"},
		{ID: "name", Name: "name"},
		{ID: "scans", Name: "scans", HeaderUsage: "file"},
	}
	items := []map[string]interface{}{{"id": "t1"}, {"id": "t2"}, {"id": "t3"}}

	a.hydrateFileAttachments(context.Background(), items, headers, "tbl")
	if ws.reads != 2 {
		t.Errorf("read attachments %d times, want once per file field", ws.reads)
	}
	if !reflect.DeepEqual(items[0]["docs"], []map[string]interface{}{{"name": "docs.pdf"}}) {
		t.Errorf("t1 docs = %v, want its file", items[0]["docs"])
	}
	if files, ok := items[1]["scans"].([]map[string]interface{}); !ok || len(files) != 0 {
		t.Errorf("t2 scans = %v, want an empty list", items[1]["scans"])
	}

	ws.fail = true
	a.hydrateFileAttachments(context.Background(), items, headers, "tbl")
	if files, ok := items[0]["docs"].([]map[string]interface{}); !ok || len(files) != 0 {
		t.Errorf("after a failed read t1 docs = %v, want an empty list", items[0]["docs"])
	}
}
