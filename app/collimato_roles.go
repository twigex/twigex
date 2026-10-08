// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

// Collimato role management is an enterprise implementation. Reading roles and
// enforcing the built-in workspace_admin / workspace_user stays here, because an
// unlicensed instance still needs both.

import (
	"net/http"

	"github.com/twigex/twigex/model"
)

func rolesUnavailable() *model.AppError {
	return model.NewAppError("collimato.custom_roles_unavailable", http.StatusPaymentRequired)
}

func (a *App) CreateWorkspaceRole(user model.User, workspaceID string, role model.CollimatoRole) (*model.CollimatoRole, *model.AppError) {
	if a.CollimatoRoles == nil {
		return nil, rolesUnavailable()
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionCreateRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	if appErr := a.checkCollimatoRoleNameFree(workspaceID, role.Name); appErr != nil {
		return nil, appErr
	}

	return a.CollimatoRoles.CreateRole(user, workspaceID, role)
}

func (a *App) UpdateWorkspaceRole(user model.User, workspaceID, roleID string, patch model.CollimatoRolePatch) (*model.CollimatoRole, *model.AppError) {
	if a.CollimatoRoles == nil {
		return nil, rolesUnavailable()
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionEditRoles) {
		return nil, model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardCollimatoRoleEdit(user, workspaceID, roleID); appErr != nil {
		return nil, appErr
	}

	if patch.Name != nil {
		if appErr := a.checkCollimatoRoleRename(workspaceID, roleID, *patch.Name); appErr != nil {
			return nil, appErr
		}
	}

	return a.CollimatoRoles.UpdateRole(user, workspaceID, roleID, patch)
}

func (a *App) DeleteWorkspaceRole(user model.User, workspaceID, roleID string) *model.AppError {
	if a.CollimatoRoles == nil {
		return rolesUnavailable()
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionDeleteRoles) {
		return model.NewAppError("role.forbidden", http.StatusForbidden)
	}

	if appErr := a.checkCollimatoRoleDeletable(workspaceID, roleID); appErr != nil {
		return appErr
	}

	return a.CollimatoRoles.DeleteRole(user, workspaceID, roleID)
}

func (a *App) UpdateUserRoles(user model.User, workspaceID, userID string, roles []string) *model.AppError {
	if a.CollimatoRoles == nil {
		return rolesUnavailable()
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAssignRoles) {
		return model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardCollimatoMemberRoles(user, workspaceID, userID, roles); appErr != nil {
		return appErr
	}

	return a.CollimatoRoles.AssignUserRoles(workspaceID, userID, roles)
}

func (a *App) UpdateWorkspaceGroupRoles(user model.User, workspaceID, groupID string, roles []string) *model.AppError {
	if a.CollimatoRoles == nil {
		return rolesUnavailable()
	}

	if !a.CollimatoWorkspaceHasPermission(user, workspaceID, model.CollimatoPermissions.PermissionAssignRoles) {
		return model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardCollimatoGroupRoles(user, workspaceID, []string{groupID}, roles); appErr != nil {
		return appErr
	}

	return a.CollimatoRoles.AssignGroupRoles(workspaceID, groupID, roles)
}
