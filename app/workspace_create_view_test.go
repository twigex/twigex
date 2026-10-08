// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeCreateViewStore struct {
	fakeReadAccessStore
	names   map[string]bool
	created []string
}

func (f *fakeCreateViewStore) ViewExistsByName(_, name string) (bool, error) {
	return f.names[name], nil
}

func (f *fakeCreateViewStore) GetCustomFields(_, _ string) ([]model.TaskOrderField, error) {
	return []model.TaskOrderField{{Name: "budget", DisplayName: "Budget"}}, nil
}

func (f *fakeCreateViewStore) CreateView(workspaceID, tableID, order, name, viewType, _, _, viewID string) (*model.WorkspaceView, error) {
	f.created = append(f.created, order)
	return &model.WorkspaceView{ID: viewID, WorkspaceID: workspaceID, TableID: tableID, Name: name, ViewType: viewType, TaskOrder: order}, nil
}

func createViewApp(memberRole string) (*App, *fakeCreateViewStore) {
	ws := &fakeCreateViewStore{
		fakeReadAccessStore: fakeReadAccessStore{
			fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{memberRoleName: memberRole},
			tableOwner:               map[string]string{"t-a": "ws-a"},
		},
		names: map[string]bool{"Taken": true},
	}
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	return a, ws
}

func TestCreateGridViewStartsWithTheTableColumns(t *testing.T) {
	a, ws := createViewApp("member")
	member := model.User{ID: "u1", Role: model.SystemUserRoleId}

	view, appErr := a.CreateView(context.Background(), "ws-a", "t-a", "", "Mine", "grid", true, member, "", "")
	if appErr != nil {
		t.Fatalf("CreateView: %v", appErr)
	}

	var columns []model.TaskOrderField
	if err := json.Unmarshal([]byte(view.TaskOrder), &columns); err != nil {
		t.Fatalf("grid order %q is not a column list: %v", view.TaskOrder, err)
	}
	has := map[string]bool{}
	for _, c := range columns {
		has[c.Name] = true
	}
	for _, name := range []string{"name", "status", "budget"} {
		if !has[name] {
			t.Errorf("grid columns %v lack %q", columns, name)
		}
	}
	if len(ws.created) != 1 {
		t.Errorf("created %d views, want 1", len(ws.created))
	}
}

func TestCreateGridViewRefusesATakenName(t *testing.T) {
	a, ws := createViewApp("member")
	member := model.User{ID: "u1", Role: model.SystemUserRoleId}

	_, appErr := a.CreateView(context.Background(), "ws-a", "t-a", "", "Taken", "grid", true, member, "", "")
	if appErr == nil || appErr.Status != http.StatusConflict {
		t.Errorf("a grid view named like another: got %v, want 409", appErr)
	}
	if len(ws.created) != 0 {
		t.Errorf("created %d views, want none", len(ws.created))
	}
}
