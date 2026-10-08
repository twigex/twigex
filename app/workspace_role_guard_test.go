// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/twigex/twigex/model"
)

type fakeProjectRoleGuardStore struct {
	fakeAccessWorkspaceStore
	members map[string]string
	groups  []model.WorkspaceGroup
	changed []string
}

func (f *fakeProjectRoleGuardStore) GetUserByUserID(userID, _ string) (*model.WorkspaceMember, error) {
	role, ok := f.members[userID]
	if !ok {
		return nil, nil
	}

	return &model.WorkspaceMember{UserID: userID, Role: role}, nil
}

func (f *fakeProjectRoleGuardStore) GetGroups(context.Context, string) ([]model.WorkspaceGroup, error) {
	return f.groups, nil
}

func (f *fakeProjectRoleGuardStore) GetRoleByID(_, roleID string) (*model.ProjectWorkspaceRole, error) {
	return &model.ProjectWorkspaceRole{ID: roleID, Name: roleID}, nil
}

func (f *fakeProjectRoleGuardStore) UpdateMemberRole(_, memberID, _ string, _ model.User) (bool, error) {
	f.changed = append(f.changed, memberID)
	return true, nil
}

func (f *fakeProjectRoleGuardStore) UpdateGroupRoles(_ context.Context, _, groupID string, _ []string) error {
	f.changed = append(f.changed, groupID)
	return nil
}

func projectRoleGuardApp() (*App, *fakeProjectRoleGuardStore) {
	admin := *model.MakeDefaultProjectWorkspaceRoles()[model.ProjectWorkspaceAdminRoleID]

	ws := &fakeProjectRoleGuardStore{
		fakeAccessWorkspaceStore: fakeAccessWorkspaceStore{
			roles: []model.ProjectWorkspaceRole{
				admin,
				{Name: "manager", Permissions: []string{model.PermissionUpdateRoles, model.PermissionAddMemberToWorkspace}},
				{Name: "editor", Permissions: []string{model.PermissionUpdateTask}},
				{Name: model.ProjectWorkspaceUserRoleID},
			},
		},
		members: map[string]string{
			"boss":    model.ProjectWorkspaceAdminRoleID,
			"manager": "manager",
			"member":  model.ProjectWorkspaceUserRoleID,
		},
		groups: []model.WorkspaceGroup{
			{GroupID: "mine", Roles: []string{"editor"}},
			{GroupID: "admins", Roles: []string{model.ProjectWorkspaceAdminRoleID}},
			{GroupID: "team", Roles: []string{"editor"}},
		},
	}

	a := accessApp(&ws.fakeAccessWorkspaceStore)
	a.Store.Workspace = ws
	a.Store.Groups = &fakeUserGroups{of: map[string][]string{"manager": {"mine"}}}
	a.Server.License.Features.Groups = boolPtr(true)

	return a, ws
}

var (
	projectManager = model.User{ID: "manager"}
	projectBoss    = model.User{ID: "boss"}
)

func TestProjectRoleAssignersCannotPromoteThemselvesOrHandOutAdmin(t *testing.T) {
	a, ws := projectRoleGuardApp()

	_, appErr := a.UpdateWorkspaceMemberRole("w1", "manager", "manager editor", projectManager)
	wantForbidden(t, "own roles", "workspace.own_roles", appErr)

	_, appErr = a.UpdateWorkspaceMemberRole("w1", "member", model.ProjectWorkspaceAdminRoleID, projectManager)
	wantForbidden(t, "giving admin", "workspace.admin_role_required", appErr)

	_, appErr = a.UpdateWorkspaceMemberRole("w1", "boss", "editor", projectManager)
	wantForbidden(t, "taking admin away", "workspace.admin_role_required", appErr)

	if len(ws.changed) != 0 {
		t.Fatalf("expected nothing changed, got %v", ws.changed)
	}

	if _, appErr := a.UpdateWorkspaceMemberRole("w1", "member", "editor", projectManager); appErr != nil {
		t.Errorf("ordinary role: %v", appErr)
	}

	if _, appErr := a.UpdateWorkspaceMemberRole("w1", "member", model.ProjectWorkspaceAdminRoleID, projectBoss); appErr != nil {
		t.Errorf("admin giving admin: %v", appErr)
	}
}

func TestProjectRoleAssignersCannotUseGroupsToPromoteThemselves(t *testing.T) {
	a, _ := projectRoleGuardApp()

	wantForbidden(t, "own group", "workspace.own_roles",
		a.UpdateProjectWorkspaceGroupRoles(projectManager, "w1", "mine", []string{"manager"}))

	wantForbidden(t, "admin group", "workspace.admin_role_required",
		a.UpdateProjectWorkspaceGroupRoles(projectManager, "w1", "admins", []string{"editor"}))

	_, appErr := a.AddGroupsToProjectWorkspace(projectManager, "w1", []string{"new"}, []string{model.ProjectWorkspaceAdminRoleID})
	wantForbidden(t, "attaching with admin", "workspace.admin_role_required", appErr)

	if appErr := a.UpdateProjectWorkspaceGroupRoles(projectManager, "w1", "team", []string{"manager"}); appErr != nil {
		t.Errorf("another group: %v", appErr)
	}
}

func TestProjectRoleEditorsCannotEditRolesTheyHoldOrTheAdminRole(t *testing.T) {
	a, _ := projectRoleGuardApp()

	wantForbidden(t, "own role", "workspace.own_role_edit", a.guardProjectRoleEdit(projectManager, "w1", "manager"))

	wantForbidden(t, "admin role", "workspace.admin_role_required",
		a.guardProjectRoleEdit(projectManager, "w1", model.ProjectWorkspaceAdminRoleID))

	if appErr := a.guardProjectRoleEdit(projectManager, "w1", "editor"); appErr != nil {
		t.Errorf("another role: %v", appErr)
	}

	if appErr := a.guardProjectRoleEdit(projectBoss, "w1", model.ProjectWorkspaceAdminRoleID); appErr != nil {
		t.Errorf("admin editing the admin role: %v", appErr)
	}
}

func TestProjectRoleNamesStayUnique(t *testing.T) {
	a, _ := projectRoleGuardApp()

	appErr := a.checkProjectRoleNameFree("w1", "manager")
	if appErr == nil || appErr.Status != http.StatusConflict {
		t.Errorf("duplicate name: want 409, got %v", appErr)
	}

	if appErr := a.checkProjectRoleNameFree("w1", "fresh"); appErr != nil {
		t.Errorf("new name: %v", appErr)
	}
}
