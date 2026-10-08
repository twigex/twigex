// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeManagedProjectStore struct {
	store.WorkspaceStore
	workspaces []model.Workspace
	listed     map[string][]model.WorkspaceMember
	members    map[string]*model.WorkspaceMember
	tables     []model.WorkspaceTable
	groups     map[string]int
	removed    []string
}

func (f *fakeManagedProjectStore) GetRow(workspaceID string) (*model.Workspace, error) {
	for _, ws := range f.workspaces {
		if ws.ID == workspaceID {
			return &ws, nil
		}
	}

	return nil, sql.ErrNoRows
}

func (f *fakeManagedProjectStore) GetAll(context.Context) ([]model.Workspace, error) {
	return f.workspaces, nil
}

func (f *fakeManagedProjectStore) GetMembersForWorkspaces(context.Context, []string) (map[string][]model.WorkspaceMember, error) {
	return f.listed, nil
}

func (f *fakeManagedProjectStore) GetGroupCounts(context.Context, []string) (map[string]int, error) {
	return f.groups, nil
}

func (f *fakeManagedProjectStore) GetTableMetas([]string) ([]model.WorkspaceTable, error) {
	return f.tables, nil
}

func (f *fakeManagedProjectStore) GetMemberByUserID(_, userID string) (*model.WorkspaceMember, error) {
	member, ok := f.members[userID]
	if !ok {
		return nil, sql.ErrNoRows
	}

	return member, nil
}

func (f *fakeManagedProjectStore) GetMembers(string) ([]model.WorkspaceMember, error) {
	out := make([]model.WorkspaceMember, 0, len(f.members))
	for _, member := range f.members {
		out = append(out, *member)
	}

	return out, nil
}

func (f *fakeManagedProjectStore) GetRolesByName(names []string, _ string) ([]model.ProjectWorkspaceRole, error) {
	defined := []string{model.ProjectWorkspaceAdminRoleID, model.ProjectWorkspaceUserRoleID, "editor"}

	found := []model.ProjectWorkspaceRole{}
	for _, name := range names {
		if slices.Contains(defined, name) {
			found = append(found, model.ProjectWorkspaceRole{Name: name})
		}
	}

	return found, nil
}

func (f *fakeManagedProjectStore) AddGroups(context.Context, string, []string, []string, string) error {
	return nil
}

func (f *fakeManagedProjectStore) GetGroups(context.Context, string) ([]model.WorkspaceGroup, error) {
	return []model.WorkspaceGroup{}, nil
}

func (f *fakeManagedProjectStore) UpdateGroupRoles(context.Context, string, string, []string) error {
	return nil
}

func (f *fakeManagedProjectStore) UserHasAnyGroupAccess(string, string) (bool, error) {
	return true, nil
}

func (f *fakeManagedProjectStore) DeleteMember(_, memberID string) (bool, error) {
	f.removed = append(f.removed, memberID)
	delete(f.members, memberID)

	return true, nil
}

func (f *fakeManagedProjectStore) AddMember(members []model.WorkspaceMember) ([]model.WorkspaceMember, error) {
	for _, m := range members {
		if _, ok := f.members[m.UserID]; !ok {
			member := m
			f.members[m.UserID] = &member
		}
	}

	return members, nil
}

func (f *fakeManagedProjectStore) UpdateMemberRole(_, memberID, role string, _ model.User) (bool, error) {
	member, ok := f.members[memberID]
	if !ok {
		return false, nil
	}

	member.Role = role

	return true, nil
}

func managedProjectApp(ws *fakeManagedProjectStore) *App {
	return &App{
		Store: store.Store{
			Workspace: ws,
			Roles:     &fakeRoleStore{known: map[string]bool{}},
		},
		ConfigStore: photoTestConfig(""),
		Server: Server{
			License: &model.License{
				StartsAt:  time.Now().Add(-time.Hour).Unix(),
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Features:  &model.Features{WorkspaceRoles: boolPtr(true), Groups: boolPtr(true)},
			},
		},
	}
}

func projectWithMembers(members ...model.WorkspaceMember) *fakeManagedProjectStore {
	ws := &fakeManagedProjectStore{
		workspaces: []model.Workspace{{ID: "w1", Title: "Roadmap"}},
		members:    map[string]*model.WorkspaceMember{},
		groups:     map[string]int{},
	}

	for _, m := range members {
		member := m
		ws.members[m.UserID] = &member
	}

	return ws
}

func TestProjectManagementRefusesUsersWithoutPermission(t *testing.T) {
	ws := projectWithMembers(model.WorkspaceMember{UserID: "anna", Role: model.ProjectWorkspaceAdminRoleID})
	a := managedProjectApp(ws)
	user := regularUser()

	if _, appErr := a.ListAllProjectWorkspaces(context.Background(), user); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("ListAllProjectWorkspaces: want 403, got %v", appErr)
	}

	if _, appErr := a.GetManagedProjectWorkspace(context.Background(), user, "w1"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("GetManagedProjectWorkspace: want 403, got %v", appErr)
	}

	if appErr := a.JoinManagedProjectWorkspace(user, "w1"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("JoinManagedProjectWorkspace: want 403, got %v", appErr)
	}

	if appErr := a.RemoveManagedProjectWorkspaceMember(user, "w1", "anna"); appErr == nil || appErr.Status != http.StatusForbidden {
		t.Errorf("RemoveManagedProjectWorkspaceMember: want 403, got %v", appErr)
	}

	if _, ok := ws.members["u1"]; ok || len(ws.removed) > 0 {
		t.Error("expected nothing to change for a user without the permission")
	}
}

func TestProjectManagementGrantedThroughRole(t *testing.T) {
	ws := projectWithMembers()
	ws.tables = []model.WorkspaceTable{{WorkspaceID: "w1"}, {WorkspaceID: "w1"}}
	ws.groups = map[string]int{"w1": 3}
	ws.listed = map[string][]model.WorkspaceMember{"w1": make([]model.WorkspaceMember, 6)}

	a := managedProjectApp(ws)
	a.Store.Roles = &permFakeRoleStore{perms: map[string][]string{
		"project_managers": {model.ProjectSectionPermissions.PermissionManageProjects.Id},
	}}

	got, appErr := a.ListAllProjectWorkspaces(context.Background(), model.User{ID: "u1", Role: "project_managers"})
	if appErr != nil {
		t.Fatalf("ListAllProjectWorkspaces: %v", appErr)
	}

	if len(got) != 1 || got[0].TableCount != 2 || got[0].GroupCount != 3 {
		t.Fatalf("expected one workspace with 2 tables and 3 groups, got %+v", got)
	}

	if got[0].MemberCount != 6 || len(got[0].Members) != workspaceMemberPreviewSize {
		t.Errorf("expected 6 members counted and %d previewed, got %d and %d",
			workspaceMemberPreviewSize, got[0].MemberCount, len(got[0].Members))
	}
}

func TestManagedProjectWorkspaceNotFound(t *testing.T) {
	a := managedProjectApp(projectWithMembers())

	if _, appErr := a.GetManagedProjectWorkspace(context.Background(), admin(), "missing"); appErr == nil || appErr.Status != http.StatusNotFound {
		t.Errorf("want 404, got %v", appErr)
	}
}

func TestManagedProjectWorkspaceKeepsAnAdmin(t *testing.T) {
	ws := projectWithMembers(
		model.WorkspaceMember{UserID: "anna", Role: model.ProjectWorkspaceAdminRoleID},
		model.WorkspaceMember{UserID: "bob", Role: model.ProjectWorkspaceUserRoleID},
	)
	a := managedProjectApp(ws)

	if appErr := a.RemoveManagedProjectWorkspaceMember(admin(), "w1", "anna"); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("removing the last admin: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "anna", model.ProjectWorkspaceUserRoleID); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("demoting the last admin: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "bob", model.ProjectWorkspaceAdminRoleID); appErr != nil {
		t.Fatalf("promoting a second admin: %v", appErr)
	}

	if appErr := a.RemoveManagedProjectWorkspaceMember(admin(), "w1", "anna"); appErr != nil {
		t.Fatalf("removing an admin once another exists: %v", appErr)
	}

	if !slices.Equal(ws.removed, []string{"anna"}) {
		t.Errorf("expected only anna removed, got %v", ws.removed)
	}
}

func TestRemoveManagedProjectWorkspaceMemberNotFound(t *testing.T) {
	a := managedProjectApp(projectWithMembers())

	if appErr := a.RemoveManagedProjectWorkspaceMember(admin(), "w1", "ghost"); appErr == nil || appErr.Status != http.StatusNotFound {
		t.Errorf("want 404, got %v", appErr)
	}
}

func TestJoinManagedProjectWorkspaceMakesTheCallerAdmin(t *testing.T) {
	ws := projectWithMembers(model.WorkspaceMember{UserID: "admin", Role: model.ProjectWorkspaceUserRoleID})
	a := managedProjectApp(ws)

	if appErr := a.JoinManagedProjectWorkspace(admin(), "w1"); appErr != nil {
		t.Fatalf("JoinManagedProjectWorkspace: %v", appErr)
	}

	if got := ws.members["admin"].Role; got != "user admin" {
		t.Errorf("expected admin added to the existing role, got %q", got)
	}
}

func TestManagedProjectWorkspaceRolesMustExist(t *testing.T) {
	ws := projectWithMembers(
		model.WorkspaceMember{UserID: "anna", Role: model.ProjectWorkspaceAdminRoleID},
		model.WorkspaceMember{UserID: "bob", Role: model.ProjectWorkspaceUserRoleID},
	)
	a := managedProjectApp(ws)

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "bob", "made_up"); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("unknown member role: want 400, got %v", appErr)
	}

	if ws.members["bob"].Role != model.ProjectWorkspaceUserRoleID {
		t.Errorf("expected bob to keep his role, got %q", ws.members["bob"].Role)
	}

	if _, appErr := a.AddManagedProjectWorkspaceGroups(admin(), "w1", []string{"g1"}, []string{"made_up"}); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("unknown role on a new group: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceGroupRoles(admin(), "w1", "g1", []string{model.ProjectWorkspaceUserRoleID, "made_up"}); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("unknown group role: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceGroupRoles(admin(), "w1", "g1", []string{model.ProjectWorkspaceUserRoleID, model.ProjectWorkspaceUserRoleID}); appErr != nil {
		t.Errorf("known group roles with a duplicate: %v", appErr)
	}
}

func TestManagedProjectWorkspaceCountsAdminsInRoleLists(t *testing.T) {
	ws := projectWithMembers(
		model.WorkspaceMember{UserID: "anna", Role: "admin editor"},
		model.WorkspaceMember{UserID: "bob", Role: "editor"},
	)
	a := managedProjectApp(ws)

	if appErr := a.RemoveManagedProjectWorkspaceMember(admin(), "w1", "anna"); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("removing the only admin with several roles: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "anna", "editor"); appErr == nil || appErr.Status != http.StatusBadRequest {
		t.Errorf("dropping admin from the only admin: want 400, got %v", appErr)
	}

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "bob", "editor  admin editor"); appErr != nil {
		t.Fatalf("giving bob several roles: %v", appErr)
	}

	if got := ws.members["bob"].Role; got != "editor admin" {
		t.Errorf("expected the roles cleaned up to %q, got %q", "editor admin", got)
	}

	if appErr := a.UpdateManagedProjectWorkspaceMemberRole(admin(), "w1", "anna", "editor"); appErr != nil {
		t.Errorf("dropping admin once bob is also an admin: %v", appErr)
	}
}
