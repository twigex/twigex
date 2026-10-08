// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// Holding a role-management permission must not let anyone promote themselves.
// Workspace admins are not restricted; everyone else may not change their own
// roles, hand out or take away workspace_admin, or edit a role they hold.

func (a *App) guardCollimatoNewMembers(user model.User, workspaceID string, userIDs, roles []string) *model.AppError {
	if a.holdsCollimatoWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	if slices.Contains(userIDs, user.ID) {
		return model.NewAppError("collimato.own_roles", http.StatusForbidden)
	}

	if slices.Contains(roles, model.CollimatoWorkspaceAdminRoleId) {
		return model.NewAppError("collimato.admin_role_required", http.StatusForbidden)
	}

	return nil
}

func (a *App) guardCollimatoMemberRoles(user model.User, workspaceID, memberID string, roles []string) *model.AppError {
	if a.holdsCollimatoWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	members, err := a.Store.Collimato.GetWorkspaceUsers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace users",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	target := slices.IndexFunc(members, func(m model.CollimatoWorkspaceUser) bool { return m.ID == memberID })
	if target < 0 {
		return model.NewAppError("collimato.member_not_found", http.StatusNotFound)
	}

	if members[target].UserID == user.ID {
		return model.NewAppError("collimato.own_roles", http.StatusForbidden)
	}

	if changesAdminRole(strings.Split(members[target].Role, ","), roles) {
		return model.NewAppError("collimato.admin_role_required", http.StatusForbidden)
	}

	return nil
}

func (a *App) guardCollimatoGroupRoles(user model.User, workspaceID string, groupIDs, roles []string) *model.AppError {
	if a.holdsCollimatoWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	ctx := context.Background()

	mine, err := a.Store.Groups.GetIDsForUser(ctx, user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve the user's groups",
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("group.retrieval_failed", http.StatusInternalServerError)
	}

	attached, err := a.Store.Collimato.GetWorkspaceGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	for _, groupID := range groupIDs {
		if slices.Contains(mine, groupID) {
			return model.NewAppError("collimato.own_roles", http.StatusForbidden)
		}

		var current []string
		if i := slices.IndexFunc(attached, func(g model.CollimatoWorkspaceGroup) bool { return g.GroupID == groupID }); i >= 0 {
			current = attached[i].Roles
		}

		if changesAdminRole(current, roles) {
			return model.NewAppError("collimato.admin_role_required", http.StatusForbidden)
		}
	}

	return nil
}

func (a *App) guardCollimatoRoleEdit(user model.User, workspaceID, roleID string) *model.AppError {
	if a.holdsCollimatoWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	role, appErr := a.collimatoRole(workspaceID, roleID)
	if appErr != nil || role == nil {
		return appErr
	}

	if role.Name == model.CollimatoWorkspaceAdminRoleId {
		return model.NewAppError("collimato.admin_role_required", http.StatusForbidden)
	}

	held, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles",
			"workspace_id", workspaceID,
			"user_id", user.ID,
			"error", err,
		)
		return model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	if slices.Contains(held, role.Name) {
		return model.NewAppError("collimato.own_role_edit", http.StatusForbidden)
	}

	return nil
}

// Roles are resolved by name, so a second role with a held name would add its
// permissions to everyone holding the first.
func (a *App) checkCollimatoRoleNameFree(workspaceID, name string) *model.AppError {
	existing, err := a.Store.Collimato.GetRolesByName([]string{name}, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to check role name",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("role.check_failed", http.StatusInternalServerError)
	}

	if len(existing) > 0 {
		return model.NewAppError("role.name_taken", http.StatusConflict)
	}

	return nil
}

func (a *App) checkCollimatoRoleRename(workspaceID, roleID, name string) *model.AppError {
	role, appErr := a.collimatoRole(workspaceID, roleID)
	if appErr != nil || role == nil || role.Name == name {
		return appErr
	}

	if isBuiltInCollimatoRole(role.Name) {
		return model.NewAppError("role.cannot_rename_builtin", http.StatusConflict)
	}

	return a.checkCollimatoRoleNameFree(workspaceID, name)
}

func (a *App) checkCollimatoRoleDeletable(workspaceID, roleID string) *model.AppError {
	role, appErr := a.collimatoRole(workspaceID, roleID)
	if appErr != nil || role == nil {
		return appErr
	}

	if isBuiltInCollimatoRole(role.Name) {
		return model.NewAppError("role.cannot_delete_builtin", http.StatusConflict)
	}

	return nil
}

func (a *App) collimatoRole(workspaceID, roleID string) (*model.CollimatoRole, *model.AppError) {
	role, err := a.Store.Collimato.GetWorkspaceRoleByID(workspaceID, roleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		tlog.Errorw("Failed to retrieve workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	return role, nil
}

func isBuiltInCollimatoRole(name string) bool {
	return name == model.CollimatoWorkspaceAdminRoleId || name == model.CollimatoWorkspaceUserRoleId
}

func changesAdminRole(before, after []string) bool {
	return slices.Contains(before, model.CollimatoWorkspaceAdminRoleId) != slices.Contains(after, model.CollimatoWorkspaceAdminRoleId)
}
