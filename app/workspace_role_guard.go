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
// roles, hand out or take away admin, or edit a role they hold.

func (a *App) holdsProjectWorkspaceAdmin(user model.User, workspaceID string) bool {
	roles, _ := a.effectiveWorkspaceRoleNames(user.ID, workspaceID)

	return slices.Contains(roles, model.ProjectWorkspaceAdminRoleID)
}

func (a *App) guardProjectMemberRole(user model.User, workspaceID, memberID, role string) *model.AppError {
	if a.holdsProjectWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	if memberID == user.ID {
		return model.NewAppError("workspace.own_roles", http.StatusForbidden)
	}

	member, err := a.Store.Workspace.GetUserByUserID(memberID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"member_id", memberID,
			"error", err,
		)
		return model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	if member == nil {
		return model.NewAppError("workspace.member_not_found", http.StatusNotFound)
	}

	if changesProjectAdminRole(strings.Fields(member.Role), strings.Fields(role)) {
		return model.NewAppError("workspace.admin_role_required", http.StatusForbidden)
	}

	return nil
}

func (a *App) guardProjectGroupRoles(user model.User, workspaceID string, groupIDs, roles []string) *model.AppError {
	if a.holdsProjectWorkspaceAdmin(user, workspaceID) {
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

	attached, err := a.Store.Workspace.GetGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	for _, groupID := range groupIDs {
		if slices.Contains(mine, groupID) {
			return model.NewAppError("workspace.own_roles", http.StatusForbidden)
		}

		var current []string
		if i := slices.IndexFunc(attached, func(g model.WorkspaceGroup) bool { return g.GroupID == groupID }); i >= 0 {
			current = attached[i].Roles
		}

		if changesProjectAdminRole(current, roles) {
			return model.NewAppError("workspace.admin_role_required", http.StatusForbidden)
		}
	}

	return nil
}

func (a *App) guardProjectRoleEdit(user model.User, workspaceID, roleID string) *model.AppError {
	if a.holdsProjectWorkspaceAdmin(user, workspaceID) {
		return nil
	}

	role, err := a.Store.Workspace.GetRoleByID(workspaceID, roleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}

	if err != nil {
		tlog.Errorw("Failed to retrieve workspace role",
			"workspace_id", workspaceID,
			"role_id", roleID,
			"error", err,
		)
		return model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	if role.Name == model.ProjectWorkspaceAdminRoleID {
		return model.NewAppError("workspace.admin_role_required", http.StatusForbidden)
	}

	held, _ := a.effectiveWorkspaceRoleNames(user.ID, workspaceID)
	if slices.Contains(held, role.Name) {
		return model.NewAppError("workspace.own_role_edit", http.StatusForbidden)
	}

	return nil
}

// Roles are resolved by name, so a second role with a held name would add its
// permissions to everyone holding the first.
func (a *App) checkProjectRoleNameFree(workspaceID, name string) *model.AppError {
	existing, err := a.Store.Workspace.GetRolesByName([]string{name}, workspaceID)
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

func changesProjectAdminRole(before, after []string) bool {
	return slices.Contains(before, model.ProjectWorkspaceAdminRoleID) != slices.Contains(after, model.ProjectWorkspaceAdminRoleID)
}
