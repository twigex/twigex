// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

// Tests for workspace-group app functions: permission gates, delegation to
// store. Uses fake stores, so no database and no requireDB.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// fakeWSGroupWorkspaceStore fakes the workspace store methods used by
// workspace_groups.go. Embed the interface so unoverridden methods panic if
// accidentally reached.
type fakeWSGroupWorkspaceStore struct {
	store.WorkspaceStore
	isMember       bool
	hasGroupAccess bool
	memberRoleName string // role name the user holds in the workspace; "" = not a member
	roles          []model.ProjectWorkspaceRole
	groups         []model.WorkspaceGroup
	addErr         error
	removeErr      error
	updateErr      error
}

func (f *fakeWSGroupWorkspaceStore) IsMember(_, _ string) (bool, error) {
	return f.isMember, nil
}
func (f *fakeWSGroupWorkspaceStore) UserHasAnyGroupAccess(_, _ string) (bool, error) {
	return f.hasGroupAccess, nil
}

// effectiveWorkspaceRoleNames calls these two.
// memberRoleName controls what role name the user has in the workspace ("" = not a member).
func (f *fakeWSGroupWorkspaceStore) GetUserByUserID(_, _ string) (*model.WorkspaceMember, error) {
	if f.memberRoleName == "" {
		return nil, nil
	}
	return &model.WorkspaceMember{Role: f.memberRoleName}, nil
}
func (f *fakeWSGroupWorkspaceStore) GetGroupRolesForUser(_, _ string) ([]string, error) {
	return nil, nil
}

// ProjectWorkspaceGrantPermission calls this to resolve role → permissions.
func (f *fakeWSGroupWorkspaceStore) GetRolesByName(names []string, _ string) ([]model.ProjectWorkspaceRole, error) {
	var out []model.ProjectWorkspaceRole
	for _, r := range f.roles {
		for _, n := range names {
			if r.Name == n {
				out = append(out, r)
			}
		}
	}
	return out, nil
}

// clearGroupAssigneesFromWorkspace calls this; return empty so cleanup is a no-op.
func (f *fakeWSGroupWorkspaceStore) GetAllTablesBasic(_ string) ([]model.WorkspaceTable, error) {
	return nil, nil
}
func (f *fakeWSGroupWorkspaceStore) AddGroups(_ context.Context, _ string, _ []string, _ []string, _ string) error {
	return f.addErr
}
func (f *fakeWSGroupWorkspaceStore) GetGroups(_ context.Context, _ string) ([]model.WorkspaceGroup, error) {
	return f.groups, nil
}
func (f *fakeWSGroupWorkspaceStore) RemoveGroup(_ context.Context, _, _ string) error {
	return f.removeErr
}
func (f *fakeWSGroupWorkspaceStore) UpdateGroupRoles(_ context.Context, _, _ string, _ []string) error {
	return f.updateErr
}

func wsGroupApp(ws *fakeWSGroupWorkspaceStore) *App {
	return &App{
		Store: store.Store{Workspace: ws, Groups: &fakeGroupStoreWS{}},
		Server: Server{
			License: &model.License{
				StartsAt:  time.Now().Add(-time.Hour).Unix(),
				ExpiresAt: time.Now().Add(time.Hour).Unix(),
				Features: &model.Features{
					WorkspaceRoles: boolPtr(true),
					Groups:         boolPtr(true),
				},
			},
		},
	}
}

// wsUserWithPerm returns a fake that grants permissionId to a regular user
// via a role named "workspace_member".
func wsUserWithPerm(permissionID string) *fakeWSGroupWorkspaceStore {
	return &fakeWSGroupWorkspaceStore{
		isMember:       true,
		memberRoleName: "workspace_member",
		roles: []model.ProjectWorkspaceRole{
			{Name: "workspace_member", Permissions: []string{permissionID}},
		},
	}
}

// wsUserWithoutPerm returns a fake member whose role has no permissions.
func wsUserWithoutPerm() *fakeWSGroupWorkspaceStore {
	return &fakeWSGroupWorkspaceStore{
		memberRoleName: "workspace_member",
		roles: []model.ProjectWorkspaceRole{
			{Name: "workspace_member", Permissions: []string{}},
		},
	}
}

func sysAdmin() model.User {
	return model.User{ID: "admin1", Role: model.SystemAdminRoleId}
}

func wsRegularUser() model.User {
	return model.User{ID: "user1", Role: model.SystemUserRoleId}
}

func TestAddGroupsToProjectWorkspace_Forbidden(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{isMember: false})
	_, appErr := a.AddGroupsToProjectWorkspace(regularUser(), "ws1", []string{"g1"}, []string{"member"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestAddGroupsToProjectWorkspace_SystemAdminOutsideForbidden(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{})
	_, appErr := a.AddGroupsToProjectWorkspace(sysAdmin(), "ws1", []string{"g1"}, []string{"member"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestAddGroupsToProjectWorkspace_MemberWithPermissionSucceeds(t *testing.T) {
	ws := wsUserWithPerm(model.PermissionAddMemberToWorkspace)
	ws.roles = append(ws.roles, model.ProjectWorkspaceRole{Name: "member"})
	ws.groups = []model.WorkspaceGroup{{ID: "r1", GroupID: "g1", WorkspaceID: "ws1"}}
	_, appErr := wsGroupApp(ws).AddGroupsToProjectWorkspace(wsRegularUser(), "ws1", []string{"g1"}, []string{"member"})
	if appErr != nil {
		t.Fatalf("expected success, got %v", appErr)
	}
}

func TestAddGroupsToProjectWorkspace_MemberWithoutPermissionForbidden(t *testing.T) {
	_, appErr := wsGroupApp(wsUserWithoutPerm()).AddGroupsToProjectWorkspace(wsRegularUser(), "ws1", []string{"g1"}, []string{"member"})
	if appErr == nil || appErr.Status != 403 {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestListProjectWorkspaceGroups_Forbidden_NonMember(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{isMember: false, hasGroupAccess: false})
	_, appErr := a.ListProjectWorkspaceGroups(regularUser(), "ws1")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestListProjectWorkspaceGroups_AllowedForGroupMember(t *testing.T) {
	ws := &fakeWSGroupWorkspaceStore{
		isMember:       false,
		hasGroupAccess: true,
		groups:         []model.WorkspaceGroup{{ID: "r1", GroupID: "g1"}},
	}
	a := wsGroupApp(ws)
	got, appErr := a.ListProjectWorkspaceGroups(regularUser(), "ws1")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if len(got) != 1 {
		t.Errorf("want 1 group, got %d", len(got))
	}
}

func TestListProjectWorkspaceGroups_Admin(t *testing.T) {
	ws := &fakeWSGroupWorkspaceStore{
		isMember: true, // ListProjectWorkspaceGroups checks membership, no admin shortcut
		groups: []model.WorkspaceGroup{
			{ID: "r1", GroupID: "g1"},
			{ID: "r2", GroupID: "g2"},
		},
	}
	a := wsGroupApp(ws)
	got, appErr := a.ListProjectWorkspaceGroups(sysAdmin(), "ws1")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if len(got) != 2 {
		t.Errorf("want 2 groups, got %d", len(got))
	}
}

func TestRemoveProjectWorkspaceGroup_MemberWithPermissionSucceeds(t *testing.T) {
	ws := wsUserWithPerm(model.PermissionDeleteWorkspaceMember)
	a := &App{Store: store.Store{Workspace: ws, Groups: &fakeGroupStoreWS{}}}
	if appErr := a.RemoveProjectWorkspaceGroup(wsRegularUser(), "ws1", "g1"); appErr != nil {
		t.Fatalf("expected success, got %v", appErr)
	}
}

func TestRemoveProjectWorkspaceGroup_MemberWithoutPermissionForbidden(t *testing.T) {
	a := &App{Store: store.Store{Workspace: wsUserWithoutPerm(), Groups: &fakeGroupStoreWS{}}}
	appErr := a.RemoveProjectWorkspaceGroup(wsRegularUser(), "ws1", "g1")
	if appErr == nil || appErr.Status != 403 {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestRemoveProjectWorkspaceGroup_Forbidden(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{isMember: false})
	appErr := a.RemoveProjectWorkspaceGroup(regularUser(), "ws1", "g1")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestRemoveProjectWorkspaceGroup_SystemAdminOutsideForbidden(t *testing.T) {
	a := &App{Store: store.Store{Workspace: &fakeWSGroupWorkspaceStore{}, Groups: &fakeGroupStoreWS{}}}
	appErr := a.RemoveProjectWorkspaceGroup(sysAdmin(), "ws1", "g1")
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUpdateProjectWorkspaceGroupRoles_MemberWithPermissionSucceeds(t *testing.T) {
	ws := wsUserWithPerm(model.PermissionUpdateRoles)
	ws.roles = append(ws.roles, model.ProjectWorkspaceRole{Name: "editor"})
	if appErr := wsGroupApp(ws).UpdateProjectWorkspaceGroupRoles(wsRegularUser(), "ws1", "g1", []string{"editor"}); appErr != nil {
		t.Fatalf("expected success, got %v", appErr)
	}
}

func TestUpdateProjectWorkspaceGroupRoles_MemberWithoutPermissionForbidden(t *testing.T) {
	appErr := wsGroupApp(wsUserWithoutPerm()).UpdateProjectWorkspaceGroupRoles(wsRegularUser(), "ws1", "g1", []string{"editor"})
	if appErr == nil || appErr.Status != 403 {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUpdateProjectWorkspaceGroupRoles_Forbidden(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{isMember: false})
	appErr := a.UpdateProjectWorkspaceGroupRoles(regularUser(), "ws1", "g1", []string{"admin"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

func TestUpdateProjectWorkspaceGroupRoles_SystemAdminOutsideForbidden(t *testing.T) {
	a := wsGroupApp(&fakeWSGroupWorkspaceStore{})
	appErr := a.UpdateProjectWorkspaceGroupRoles(sysAdmin(), "ws1", "g1", []string{"editor"})
	if appErr == nil || appErr.Status != http.StatusForbidden {
		t.Fatalf("want 403, got %v", appErr)
	}
}

// fakeGroupStoreWS provides a minimal GroupStore for the remove test's
// best-effort cleanup path (clearGroupAssigneesFromWorkspace calls GetMembers).
type fakeGroupStoreWS struct {
	store.GroupStore
}

func (f *fakeGroupStoreWS) GetMembers(_ context.Context, _ string) ([]model.GroupMember, error) {
	return nil, nil // no members → cleanup is a no-op
}

func (f *fakeGroupStoreWS) GetIDsForUser(context.Context, string) ([]string, error) {
	return nil, nil
}
