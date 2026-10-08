// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net/http"
	"slices"
	"strings"

	"github.com/twigex/twigex/model"
)

// Holding a user or role permission must not let anyone promote themselves or
// take over an administrator. System admins are not restricted; everyone else
// may not change their own role, give or take away system_admin, manage a
// system admin's account, or edit a role they hold.

func holdsSystemAdmin(role string) bool {
	return slices.Contains(strings.Fields(role), model.SystemAdminRoleId)
}

func guardNewUserRole(actor model.User, role string) *model.AppError {
	if actor.Role == model.SystemAdminRoleId || !holdsSystemAdmin(role) {
		return nil
	}

	return model.NewAppError("user.admin_role_required", http.StatusForbidden)
}

func guardUserRoleChange(actor model.User, target *model.User, role string) *model.AppError {
	if actor.Role == model.SystemAdminRoleId {
		return nil
	}

	if appErr := guardSystemAdminAccount(actor, target); appErr != nil {
		return appErr
	}

	if target.ID == actor.ID && strings.Join(strings.Fields(role), " ") != strings.Join(strings.Fields(target.Role), " ") {
		return model.NewAppError("user.own_role", http.StatusForbidden)
	}

	return guardNewUserRole(actor, role)
}

func guardSystemAdminAccount(actor model.User, target *model.User) *model.AppError {
	if actor.Role == model.SystemAdminRoleId || !holdsSystemAdmin(target.Role) {
		return nil
	}

	return model.NewAppError("user.admin_role_required", http.StatusForbidden)
}

func guardSystemRoleEdit(actor model.User, role *model.Role) *model.AppError {
	if actor.Role == model.SystemAdminRoleId {
		return nil
	}

	if role.Name == model.SystemAdminRoleId {
		return model.NewAppError("user.admin_role_required", http.StatusForbidden)
	}

	if slices.Contains(strings.Fields(actor.Role), role.Name) {
		return model.NewAppError("role.own_role_edit", http.StatusForbidden)
	}

	return nil
}
