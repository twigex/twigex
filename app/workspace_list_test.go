// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"slices"
	"testing"

	"github.com/twigex/twigex/config"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeWorkspaceListStore struct {
	store.WorkspaceStore
	metaCalls int
}

func (fakeWorkspaceListStore) GetAllForUser(string) ([]model.Workspace, error) {
	return []model.Workspace{{ID: "ws-a"}, {ID: "ws-b"}}, nil
}

func (fakeWorkspaceListStore) GetMembersForWorkspaces(context.Context, []string) (map[string][]model.WorkspaceMember, error) {
	return map[string][]model.WorkspaceMember{"ws-b": {{UserID: "u1"}, {UserID: "u2"}}}, nil
}

func (fakeWorkspaceListStore) GetGroupCounts(context.Context, []string) (map[string]int, error) {
	return map[string]int{"ws-a": 3}, nil
}

func (f *fakeWorkspaceListStore) GetTableMetas(ids []string) ([]model.WorkspaceTable, error) {
	f.metaCalls++
	return []model.WorkspaceTable{
		{ID: "t1", WorkspaceID: "ws-a"},
		{ID: "t2", WorkspaceID: "ws-a"},
		{ID: "t3", WorkspaceID: "ws-b"},
	}, nil
}

func (fakeWorkspaceListStore) GetUserRoleNamesByWorkspace(context.Context, string, []string) (map[string][]string, error) {
	return map[string][]string{"ws-a": {"viewer"}, "ws-b": {"editor"}}, nil
}

func (fakeWorkspaceListStore) GetRolesForWorkspaces(context.Context, []string) (map[string][]model.ProjectWorkspaceRole, error) {
	return map[string][]model.ProjectWorkspaceRole{
		"ws-a": {
			{ID: "r1", Name: "viewer", DisplayName: "Viewer", Permissions: []string{"view_workspace"}},
			{ID: "r2", Name: "owner", DisplayName: "Owner", Permissions: []string{"delete_workspace"}},
		},
		"ws-b": {{ID: "r3", Name: "editor", DisplayName: "Editor", Permissions: []string{"update_workspace"}}},
	}, nil
}

func (fakeWorkspaceListStore) GetTablePermissionsForRoleIDs(context.Context, []string) ([]model.TablePermission, error) {
	return nil, nil
}

func TestWorkspaceListCountsOnlyTheTablesTheUserSees(t *testing.T) {
	ws := &fakeWorkspaceListStore{}
	a := &App{
		Store: store.Store{Workspace: ws},
		ConfigStore: config.ConfigStore{Config: &model.ServerConfig{
			SqlSettings: model.SqlSettings{QueryTimeout: model.NewInt(25)},
		}},
		WorkspaceRoles: changeRoles{hidden: map[string]bool{"Viewer": true}},
	}

	list, appErr := a.GetProjectWorkspaces(context.Background(), model.User{ID: "u1", Role: model.SystemUserRoleId})
	if appErr != nil {
		t.Fatal(appErr)
	}
	byID := map[string]model.Workspace{}
	for _, w := range list {
		byID[w.ID] = w
	}

	a1, b := byID["ws-a"], byID["ws-b"]
	if a1.TableCount != 0 || b.TableCount != 1 {
		t.Errorf("table counts = %d and %d, want 0 where the role hides tables and 1", a1.TableCount, b.TableCount)
	}
	if !slices.Equal(a1.UserRoles, []string{"Viewer"}) || !slices.Equal(a1.UserPermissions, []string{"view_workspace"}) {
		t.Errorf("ws-a roles %v permissions %v, want only the role the user holds", a1.UserRoles, a1.UserPermissions)
	}
	if a1.GroupCount != 3 || len(b.Members) != 2 || a1.Members == nil {
		t.Errorf("group count %d, members %d and %v", a1.GroupCount, len(b.Members), a1.Members)
	}
	if ws.metaCalls != 1 {
		t.Errorf("tables were read %d times, want once for every workspace", ws.metaCalls)
	}
}
