// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeRoleGuardStore struct {
	fakeWorkspaceManagementStore
	groups []model.CollimatoWorkspaceGroup
}

func (f *fakeRoleGuardStore) GetRolesByName(names []string, _ string) ([]model.CollimatoRole, error) {
	definitions := map[string][]string{
		"manager": {
			model.CollimatoPermissions.PermissionAssignRoles.Id,
			model.CollimatoPermissions.PermissionEditRoles.Id,
			model.CollimatoPermissions.PermissionAddUsers.Id,
		},
		"analyst":                           {model.CollimatoPermissions.PermissionViewCharts.Id},
		model.CollimatoWorkspaceAdminRoleId: model.MakeDefaultCollimatoWorkspaceRoles()[model.CollimatoWorkspaceAdminRoleId].Permissions,
		model.CollimatoWorkspaceUserRoleId:  {model.CollimatoPermissions.PermissionViewCharts.Id},
	}

	roles := []model.CollimatoRole{}
	for _, name := range names {
		if permissions, ok := definitions[name]; ok {
			roles = append(roles, model.CollimatoRole{Name: name, Permissions: permissions})
		}
	}

	return roles, nil
}

func (f *fakeRoleGuardStore) GetWorkspaceGroups(context.Context, string) ([]model.CollimatoWorkspaceGroup, error) {
	return f.groups, nil
}

func (f *fakeRoleGuardStore) GetWorkspaceRoleByID(_, roleID string) (*model.CollimatoRole, error) {
	return &model.CollimatoRole{ID: roleID, Name: roleID}, nil
}

type fakeUserGroups struct {
	store.GroupStore
	of map[string][]string
}

func (f *fakeUserGroups) GetIDsForUser(_ context.Context, userID string) ([]string, error) {
	return f.of[userID], nil
}

func roleGuardApp() (*App, *fakeRoleGuardStore) {
	collimato := &fakeRoleGuardStore{
		fakeWorkspaceManagementStore: fakeWorkspaceManagementStore{
			workspace: &model.CollimatoWorkspace{ID: "ws"},
			members: map[string]*model.CollimatoWorkspaceUser{
				"boss":    {ID: "row-boss", UserID: "boss", Role: model.CollimatoWorkspaceAdminRoleId},
				"manager": {ID: "row-manager", UserID: "manager", Role: "manager"},
				"member":  {ID: "row-member", UserID: "member", Role: model.CollimatoWorkspaceUserRoleId},
			},
		},
		groups: []model.CollimatoWorkspaceGroup{
			{GroupID: "mine", Roles: []string{"analyst"}},
			{GroupID: "admins", Roles: []string{model.CollimatoWorkspaceAdminRoleId}},
			{GroupID: "team", Roles: []string{"analyst"}},
		},
	}

	a := workspaceManagementApp(&collimato.fakeWorkspaceManagementStore)
	a.Store.Collimato = collimato
	a.Store.Groups = &fakeUserGroups{of: map[string][]string{"manager": {"mine"}}}
	a.Server.License = &model.License{
		StartsAt:  time.Now().Add(-time.Hour).Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Features:  &model.Features{CollimatoRoles: boolPtr(true), Groups: boolPtr(true)},
	}

	return a, collimato
}

var (
	roleManager = model.User{ID: "manager"}
	roleBoss    = model.User{ID: "boss"}
)

func wantForbidden(t *testing.T, what, id string, appErr *model.AppError) {
	t.Helper()

	if appErr == nil || appErr.Status != http.StatusForbidden || appErr.ID != id {
		t.Errorf("%s: want 403 %s, got %v", what, id, appErr)
	}
}

func TestRoleAssignersCannotPromoteThemselvesOrHandOutAdmin(t *testing.T) {
	a, _ := roleGuardApp()
	assigner := a.CollimatoRoles.(*fakeRoleAssigner)

	wantForbidden(t, "own roles", "collimato.own_roles",
		a.UpdateUserRoles(roleManager, "ws", "row-manager", []string{"manager", "analyst"}))

	wantForbidden(t, "giving admin", "collimato.admin_role_required",
		a.UpdateUserRoles(roleManager, "ws", "row-member", []string{model.CollimatoWorkspaceAdminRoleId}))

	wantForbidden(t, "taking admin away", "collimato.admin_role_required",
		a.UpdateUserRoles(roleManager, "ws", "row-boss", []string{"analyst"}))

	if len(assigner.assigned) != 0 {
		t.Fatalf("expected nothing assigned, got %v", assigner.assigned)
	}

	if appErr := a.UpdateUserRoles(roleManager, "ws", "row-member", []string{"analyst"}); appErr != nil {
		t.Errorf("ordinary role: %v", appErr)
	}

	if appErr := a.UpdateUserRoles(roleBoss, "ws", "row-member", []string{model.CollimatoWorkspaceAdminRoleId}); appErr != nil {
		t.Errorf("admin giving admin: %v", appErr)
	}
}

func TestRoleAssignersCannotUseGroupsToPromoteThemselves(t *testing.T) {
	a, _ := roleGuardApp()

	wantForbidden(t, "own group", "collimato.own_roles",
		a.UpdateWorkspaceGroupRoles(roleManager, "ws", "mine", []string{"manager"}))

	wantForbidden(t, "admin group", "collimato.admin_role_required",
		a.UpdateWorkspaceGroupRoles(roleManager, "ws", "admins", []string{"analyst"}))

	wantForbidden(t, "attaching with admin", "collimato.admin_role_required",
		a.guardCollimatoGroupRoles(roleManager, "ws", []string{"new"}, []string{model.CollimatoWorkspaceAdminRoleId}))

	if appErr := a.UpdateWorkspaceGroupRoles(roleManager, "ws", "team", []string{"manager"}); appErr != nil {
		t.Errorf("another group: %v", appErr)
	}
}

func TestRoleEditorsCannotEditRolesTheyHoldOrTheAdminRole(t *testing.T) {
	a, _ := roleGuardApp()

	wantForbidden(t, "own role", "collimato.own_role_edit",
		a.guardCollimatoRoleEdit(roleManager, "ws", "manager"))

	wantForbidden(t, "admin role", "collimato.admin_role_required",
		a.guardCollimatoRoleEdit(roleManager, "ws", model.CollimatoWorkspaceAdminRoleId))

	if appErr := a.guardCollimatoRoleEdit(roleManager, "ws", "analyst"); appErr != nil {
		t.Errorf("another role: %v", appErr)
	}

	if appErr := a.guardCollimatoRoleEdit(roleBoss, "ws", model.CollimatoWorkspaceAdminRoleId); appErr != nil {
		t.Errorf("admin editing the admin role: %v", appErr)
	}
}

func TestAddingMembersCannotBeUsedToPromote(t *testing.T) {
	a, collimato := roleGuardApp()

	wantForbidden(t, "adding yourself", "collimato.own_roles",
		a.AddWorkspaceUsers(roleManager, "ws", []string{"manager"}, []string{"analyst"}))

	wantForbidden(t, "adding as admin", "collimato.admin_role_required",
		a.AddWorkspaceUsers(roleManager, "ws", []string{"newcomer"}, []string{model.CollimatoWorkspaceAdminRoleId}))

	if collimato.members["newcomer"] != nil {
		t.Fatal("expected nobody added")
	}

	if appErr := a.AddWorkspaceUsers(roleManager, "ws", []string{"newcomer"}, []string{"analyst"}); appErr != nil {
		t.Errorf("ordinary member: %v", appErr)
	}
}

func TestRoleNamesStayUniqueAndBuiltInRolesStayPut(t *testing.T) {
	a, _ := roleGuardApp()

	if appErr := a.checkCollimatoRoleNameFree("ws", "manager"); appErr == nil || appErr.ID != "role.name_taken" {
		t.Errorf("duplicate name: want role.name_taken, got %v", appErr)
	}

	if appErr := a.checkCollimatoRoleNameFree("ws", "fresh"); appErr != nil {
		t.Errorf("new name: %v", appErr)
	}

	if appErr := a.checkCollimatoRoleRename("ws", "analyst", "manager"); appErr == nil || appErr.ID != "role.name_taken" {
		t.Errorf("renaming onto a held name: want role.name_taken, got %v", appErr)
	}

	if appErr := a.checkCollimatoRoleRename("ws", "analyst", "analyst"); appErr != nil {
		t.Errorf("keeping the name: %v", appErr)
	}

	if appErr := a.checkCollimatoRoleRename("ws", model.CollimatoWorkspaceAdminRoleId, "boss"); appErr == nil || appErr.ID != "role.cannot_rename_builtin" {
		t.Errorf("renaming a built-in role: want role.cannot_rename_builtin, got %v", appErr)
	}

	if appErr := a.checkCollimatoRoleDeletable("ws", model.CollimatoWorkspaceUserRoleId); appErr == nil || appErr.ID != "role.cannot_delete_builtin" {
		t.Errorf("deleting a built-in role: want role.cannot_delete_builtin, got %v", appErr)
	}

	if appErr := a.checkCollimatoRoleDeletable("ws", "analyst"); appErr != nil {
		t.Errorf("deleting a custom role: %v", appErr)
	}
}
