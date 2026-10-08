// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/twigex/twigex/interfaces"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeBatchAccessStore struct {
	store.WorkspaceStore
	memberRoles map[string]string
	groupRoles  map[string][]string
	roles       []model.ProjectWorkspaceRole
	tablePerms  []model.TablePermission
	shared      map[string][]string
}

func (f fakeBatchAccessStore) GetUserByUserID(userID, _ string) (*model.WorkspaceMember, error) {
	role, ok := f.memberRoles[userID]
	if !ok {
		return nil, nil
	}
	return &model.WorkspaceMember{Role: role}, nil
}

func (f fakeBatchAccessStore) GetGroupRolesForUser(userID, _ string) ([]string, error) {
	return f.groupRoles[userID], nil
}

func (f fakeBatchAccessStore) GetRoleNamesForUsers(_ context.Context, _ string, userIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	for _, id := range userIDs {
		names, _ := batchAccessApp(f).effectiveWorkspaceRoleNames(id, "")
		if len(names) > 0 {
			out[id] = names
		}
	}
	return out, nil
}

func (f fakeBatchAccessStore) GetRolesByName(names []string, _ string) ([]model.ProjectWorkspaceRole, error) {
	var out []model.ProjectWorkspaceRole
	for _, r := range f.roles {
		if slices.Contains(names, r.Name) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f fakeBatchAccessStore) GetTablePermissionsForRoles(roleIDs []string, _ string) ([]model.TablePermission, error) {
	var out []model.TablePermission
	for _, p := range f.tablePerms {
		if slices.Contains(roleIDs, p.RoleID) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f fakeBatchAccessStore) GetSharedViewIDsForUser(userID, _ string) ([]string, error) {
	return f.shared[userID], nil
}

func batchAccessApp(ws fakeBatchAccessStore) *App {
	return &App{Store: store.Store{Workspace: ws}}
}

func TestBuildAccessInputsMatchesBuildAccessInput(t *testing.T) {
	ws := fakeBatchAccessStore{
		memberRoles: map[string]string{
			"viewer":   "viewer",
			"per":      "per viewer",
			"mixed":    "viewer",
			"unknown":  "gone",
			"empty":    "",
			"withview": "viewer",
		},
		groupRoles: map[string][]string{
			"mixed":      {"per"},
			"group-only": {"per"},
		},
		roles: []model.ProjectWorkspaceRole{
			{ID: "r-viewer", Name: "viewer"},
			{ID: "r-per", Name: "per", PerTableMode: true},
		},
		tablePerms: []model.TablePermission{
			{RoleID: "r-per", TableID: "t1", Action: "read"},
			{RoleID: "r-viewer", TableID: "t2", Action: "read"},
			{RoleID: "r-other", TableID: "t3", Action: "read"},
		},
		shared: map[string][]string{
			"withview": {"v1", "v2"},
			"unknown":  {"v3"},
			"stranger": {"v4"},
		},
	}
	app := batchAccessApp(ws)

	var users []model.User
	for _, id := range []string{"viewer", "per", "mixed", "unknown", "empty", "withview", "group-only", "stranger"} {
		users = append(users, model.User{ID: id, Role: model.SystemUserRoleId})
	}
	users = append(users, model.User{ID: "admin", Role: model.SystemAdminRoleId})

	got := app.buildAccessInputs(context.Background(), users, "ws")
	if len(got) != len(users) {
		t.Fatalf("inputs for %d users, want %d", len(got), len(users))
	}
	for _, u := range users {
		want := app.buildAccessInput(u, "ws")
		if !reflect.DeepEqual(normalized(got[u.ID]), normalized(want)) {
			t.Errorf("%s: batched %+v, want %+v", u.ID, got[u.ID], want)
		}
	}

	if perms := got["mixed"].TablePerms; len(perms) != 2 {
		t.Errorf("mixed has table permissions of every role it holds, got %+v", perms)
	}
	if perms := got["viewer"].TablePerms; perms != nil {
		t.Errorf("viewer has no per-table role, so no table permissions, got %+v", perms)
	}
}

// normalized treats an empty list and a missing one alike, as the access
// rules do.
func normalized(in interfaces.AccessInput) interfaces.AccessInput {
	if len(in.Roles) == 0 {
		in.Roles = nil
	}
	if len(in.TablePerms) == 0 {
		in.TablePerms = nil
	}
	return in
}
