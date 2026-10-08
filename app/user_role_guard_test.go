// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/model"
)

var (
	userManager = model.User{ID: "manager", Role: "user_managers"}
	systemAdmin = model.User{ID: "root", Role: model.SystemAdminRoleId}
)

func wantAppErr(t *testing.T, what, id string, appErr *model.AppError) {
	t.Helper()

	if appErr == nil || appErr.ID != id {
		t.Errorf("%s: want %s, got %v", what, id, appErr)
	}
}

func TestUserManagersCannotCreateSystemAdmins(t *testing.T) {
	wantAppErr(t, "creating an admin", "user.admin_role_required", guardNewUserRole(userManager, model.SystemAdminRoleId))

	if appErr := guardNewUserRole(userManager, model.SystemUserRoleId); appErr != nil {
		t.Errorf("creating a user: %v", appErr)
	}

	if appErr := guardNewUserRole(systemAdmin, model.SystemAdminRoleId); appErr != nil {
		t.Errorf("admin creating an admin: %v", appErr)
	}
}

func TestUserManagersCannotPromoteThemselvesOrTouchAdmins(t *testing.T) {
	self := &model.User{ID: "manager", Role: "user_managers"}
	admin := &model.User{ID: "root", Role: model.SystemAdminRoleId}
	member := &model.User{ID: "member", Role: model.SystemUserRoleId}

	wantAppErr(t, "own role", "user.own_role", guardUserRoleChange(userManager, self, "user_managers role_managers"))
	wantAppErr(t, "promoting", "user.admin_role_required", guardUserRoleChange(userManager, member, model.SystemAdminRoleId))
	wantAppErr(t, "editing an admin", "user.admin_role_required", guardUserRoleChange(userManager, admin, model.SystemAdminRoleId))
	wantAppErr(t, "resetting an admin", "user.admin_role_required", guardSystemAdminAccount(userManager, admin))

	if appErr := guardUserRoleChange(userManager, self, "user_managers"); appErr != nil {
		t.Errorf("editing own profile without a role change: %v", appErr)
	}

	if appErr := guardUserRoleChange(userManager, member, "editors"); appErr != nil {
		t.Errorf("changing a member's role: %v", appErr)
	}

	if appErr := guardUserRoleChange(systemAdmin, admin, model.SystemUserRoleId); appErr != nil {
		t.Errorf("admin demoting an admin: %v", appErr)
	}
}

func TestRoleManagersCannotEditRolesTheyHoldOrTheAdminRole(t *testing.T) {
	wantAppErr(t, "own role", "role.own_role_edit", guardSystemRoleEdit(userManager, &model.Role{Name: "user_managers"}))
	wantAppErr(t, "admin role", "user.admin_role_required", guardSystemRoleEdit(userManager, &model.Role{Name: model.SystemAdminRoleId}))

	if appErr := guardSystemRoleEdit(userManager, &model.Role{Name: "editors"}); appErr != nil {
		t.Errorf("another role: %v", appErr)
	}

	if appErr := guardSystemRoleEdit(systemAdmin, &model.Role{Name: model.SystemAdminRoleId}); appErr != nil {
		t.Errorf("admin editing the admin role: %v", appErr)
	}
}
