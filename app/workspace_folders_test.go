// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeFolderMoveStore struct {
	fakeReadAccessStore
	moved []string
	err   error
}

func (f *fakeFolderMoveStore) MoveTable(_ context.Context, _, tableID, folderID string) error {
	f.moved = append(f.moved, "table "+tableID+" to "+folderID)
	return f.err
}

func (f *fakeFolderMoveStore) MoveFolder(_ context.Context, _, folderID, parentID string) error {
	f.moved = append(f.moved, "folder "+folderID+" to "+parentID)
	return f.err
}

func (f *fakeFolderMoveStore) DeleteFolder(_ context.Context, _, folderID string, withTables bool) ([]string, []map[string]interface{}, error) {
	f.moved = append(f.moved, fmt.Sprintf("delete %s with tables %t", folderID, withTables))
	return nil, nil, f.err
}

func folderMoveApp(err error) (*App, *fakeFolderMoveStore) {
	ws := &fakeFolderMoveStore{err: err}
	ws.fakeAccessWorkspaceStore = workspaceAdminAccess()
	a := readAccessApp(&ws.fakeReadAccessStore)
	a.Store.Workspace = ws
	return a, ws
}

func TestMoveWorkspaceItemSendsEachKindToItsStoreMethod(t *testing.T) {
	admin := workspaceAdmin
	a, ws := folderMoveApp(nil)

	if appErr := a.MoveWorkspaceItem(context.Background(), "ws", "table", "t1", "f1", admin); appErr != nil {
		t.Fatal(appErr)
	}
	if appErr := a.MoveWorkspaceItem(context.Background(), "ws", "folder", "f2", "", admin); appErr != nil {
		t.Fatal(appErr)
	}
	if appErr := a.MoveWorkspaceItem(context.Background(), "ws", "view", "v1", "", admin); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("moving a view gave %v, want a bad request", appErr)
	}

	want := []string{"table t1 to f1", "folder f2 to "}
	if len(ws.moved) != len(want) || ws.moved[0] != want[0] || ws.moved[1] != want[1] {
		t.Errorf("store calls = %q, want %q", ws.moved, want)
	}
}

func TestDeleteFolderTakesItsTablesOnlyForWhoMayDeleteTables(t *testing.T) {
	folderOnly := model.ProjectWorkspaceRole{
		Name:        "folders",
		Permissions: []string{model.PermissionDeleteWorkspaceFolder},
	}
	both := model.ProjectWorkspaceRole{
		Name:        "both",
		Permissions: []string{model.PermissionDeleteWorkspaceFolder, model.PermissionDeleteWorkspaceTable},
	}
	member := model.User{ID: "u1", Role: model.SystemUserRoleId}

	for _, tc := range []struct {
		role string
		want string
	}{
		{"folders", "delete f1 with tables false"},
		{"both", "delete f1 with tables true"},
	} {
		a, ws := folderMoveApp(nil)
		ws.memberRoleName = tc.role
		ws.onlyMember = member.ID
		ws.roles = []model.ProjectWorkspaceRole{folderOnly, both}

		if _, appErr := a.DeleteWorkspaceFolder(context.Background(), "ws", "f1", member); appErr != nil {
			t.Fatal(appErr)
		}
		if len(ws.moved) != 1 || ws.moved[0] != tc.want {
			t.Errorf("role %s: store calls = %q, want %q", tc.role, ws.moved, tc.want)
		}
	}
}

func TestMoveAndDeleteAnswerForWhatTheStoreRefused(t *testing.T) {
	admin := workspaceAdmin
	for _, tc := range []struct {
		err        error
		move, drop int
	}{
		{sql.ErrNoRows, http.StatusNotFound, http.StatusNotFound},
		{model.ErrFolderCycle, http.StatusBadRequest, http.StatusInternalServerError},
		{model.ErrFolderHasTables, http.StatusInternalServerError, http.StatusBadRequest},
		{errors.New("connection lost"), http.StatusInternalServerError, http.StatusInternalServerError},
	} {
		a, _ := folderMoveApp(tc.err)

		if appErr := a.MoveWorkspaceItem(context.Background(), "ws", "folder", "f1", "f2", admin); appErr == nil || appErr.Status != tc.move {
			t.Errorf("move with store error %v gave %v, want status %d", tc.err, appErr, tc.move)
		}
		if _, appErr := a.DeleteWorkspaceFolder(context.Background(), "ws", "f1", admin); appErr == nil || appErr.Status != tc.drop {
			t.Errorf("delete with store error %v gave %v, want status %d", tc.err, appErr, tc.drop)
		}
	}
}

func TestBuildFolderStructureKeepsEveryLevelAndEveryOrphan(t *testing.T) {
	folders := []model.WorkspaceFolder{
		{ID: "c", ParentFolderID: "b"},
		{ID: "b", ParentFolderID: "a"},
		{ID: "a"},
		{ID: "lost", ParentFolderID: "deleted"},
	}
	tables := []model.WorkspaceTable{
		{ID: "in-c", FolderID: "c"},
		{ID: "in-deleted", FolderID: "deleted"},
	}

	for range 20 {
		top := buildFolderStructure(folders, tables)

		ids := map[string]model.WorkspaceFolder{}
		for _, f := range top {
			ids[f.ID] = f
		}
		if len(top) != 2 || ids["a"].ID == "" || ids["lost"].ID == "" {
			t.Fatalf("top level = %v, want a and the folder whose parent is gone", top)
		}

		a := ids["a"]
		if len(a.Children) != 1 || len(a.Children[0].Children) != 1 {
			t.Fatalf("a folder two levels down went missing: %+v", a)
		}
		c := a.Children[0].Children[0]
		if c.ID != "c" || len(c.Tables) != 1 || c.Tables[0].ID != "in-c" {
			t.Fatalf("c = %+v, want it holding in-c", c)
		}
	}
}
