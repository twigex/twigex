// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) AddGroupsToProjectWorkspace(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.WorkspaceGroup, *model.AppError) {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionAddMemberToWorkspace) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardProjectGroupRoles(user, workspaceID, groupIDs, roles); appErr != nil {
		return nil, appErr
	}

	return a.attachProjectWorkspaceGroups(user, workspaceID, groupIDs, roles)
}

func (a *App) attachProjectWorkspaceGroups(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.WorkspaceGroup, *model.AppError) {
	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	roles, appErr := a.validProjectWorkspaceRoles(workspaceID, roles)
	if appErr != nil {
		return nil, appErr
	}

	ctx := context.Background()
	if err := a.Store.Workspace.AddGroups(ctx, workspaceID, groupIDs, roles, user.ID); err != nil {
		tlog.Errorw("Failed to add groups to workspace", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("collimato.group_add_failed", http.StatusInternalServerError)
	}

	all, err := a.Store.Workspace.GetGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups after add", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	added := make([]model.WorkspaceGroup, 0, len(groupIDs))
	idSet := make(map[string]struct{}, len(groupIDs))
	for _, gid := range groupIDs {
		idSet[gid] = struct{}{}
	}

	for _, g := range all {
		if _, ok := idSet[g.GroupID]; ok {
			added = append(added, g)
		}
	}

	return added, nil
}

func (a *App) ListProjectWorkspaceGroups(user model.User, workspaceID string) ([]model.WorkspaceGroup, *model.AppError) {
	if !a.IsWorkspaceMemberOrGroupMember(workspaceID, user.ID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	groups, err := a.Store.Workspace.GetGroups(context.Background(), workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	return groups, nil
}

func (a *App) RemoveProjectWorkspaceGroup(user model.User, workspaceID, groupID string) *model.AppError {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionDeleteWorkspaceMember) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return a.detachProjectWorkspaceGroup(workspaceID, groupID)
}

func (a *App) detachProjectWorkspaceGroup(workspaceID, groupID string) *model.AppError {
	if err := a.Store.Workspace.RemoveGroup(context.Background(), workspaceID, groupID); err != nil {
		tlog.Errorw("Failed to remove workspace group", "workspace_id", workspaceID, "group_id", groupID, "error", err)
		return model.NewAppError("collimato.group_remove_failed", http.StatusInternalServerError)
	}

	// Best-effort: clear assignee fields for group members in this workspace
	a.clearGroupAssigneesFromWorkspace(workspaceID, groupID)
	return nil
}

// The built-in "assignee" column on main task tables isn't stored in
// workspace_fields, so it's cleared separately from the custom person fields.
func (a *App) clearAllPersonFields(rt model.WorkspaceTable, userID string) {
	fieldNames := []string{}
	// Built-in assignee column only exists on main task tables
	if !rt.SingleSelect && rt.ParentTableID == "" {
		fieldNames = append(fieldNames, "assignee")
	}

	customFields, _ := a.Store.Workspace.GetPersonFieldNamesForTable(rt.ID)
	fieldNames = append(fieldNames, customFields...)
	for _, fieldName := range fieldNames {
		_ = a.Store.Workspace.ClearAssigneeByUserID(rt.Name, fieldName, userID)
	}
}

// Best-effort: failures are logged as warnings, not returned.
func (a *App) clearGroupAssigneesFromWorkspace(workspaceID, groupID string) {
	ctx := context.Background()

	members, err := a.Store.Groups.GetMembers(ctx, groupID)
	if err != nil {
		tlog.Warnw("Failed to list group members for assignee cleanup",
			"group_id", groupID,
			"workspace_id", workspaceID,
			"error", err,
		)
		return
	}

	tables, err := a.Store.Workspace.GetAllTablesBasic(workspaceID)
	if err != nil {
		tlog.Warnw("Failed to get workspace tables for assignee cleanup",
			"workspace_id", workspaceID,
			"group_id", groupID,
			"error", err,
		)
		return
	}

	for _, m := range members {
		// Skip if the user still has direct workspace membership.
		if direct, _ := a.Store.Workspace.GetMemberByUserID(workspaceID, m.UserID); direct != nil {
			continue
		}
		// Skip if user is still in another group attached to this workspace
		if otherGroup, _ := a.Store.Workspace.UserHasOtherGroupAccess(workspaceID, groupID, m.UserID); otherGroup {
			continue
		}

		for _, rt := range tables {
			a.clearAllPersonFields(rt, m.UserID)
		}
	}
}

// Takes a pre-fetched member list because DeleteGroup loads the members before
// Groups.SoftDelete removes the group_members rows.
func (a *App) clearGroupMembersFromWorkspaces(groupID string, members []model.GroupMember) {
	wsIDs, err := a.Store.Workspace.GetIDsForGroup(groupID)
	if err != nil || len(wsIDs) == 0 {
		return
	}

	for _, wsID := range wsIDs {
		tables, err := a.Store.Workspace.GetAllTablesBasic(wsID)
		if err != nil {
			continue
		}

		for _, m := range members {
			if direct, _ := a.Store.Workspace.GetMemberByUserID(wsID, m.UserID); direct != nil {
				continue
			}

			if other, _ := a.Store.Workspace.UserHasOtherGroupAccess(wsID, groupID, m.UserID); other {
				continue
			}

			for _, rt := range tables {
				a.clearAllPersonFields(rt, m.UserID)
			}
		}
	}
}

func (a *App) SearchProjectWorkspaceMembers(user model.User, workspaceID, query string, limit, offset int) ([]model.User, *model.AppError) {
	if !a.IsWorkspaceMemberOrGroupMember(workspaceID, user.ID) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	users, err := a.Store.Workspace.SearchMembers(context.Background(), []string{workspaceID}, query, limit, offset)
	if err != nil {
		tlog.Errorw("Failed to search workspace members", "workspace_id", workspaceID, "error", err)
		return nil, model.NewAppError("workspace.workspaces_retrieval_failed", http.StatusInternalServerError)
	}

	return users, nil
}

// SearchTaskReportMembers finds the people the task report can show tasks
// for: members of the chosen workspaces, or of any when none are chosen.
func (a *App) SearchTaskReportMembers(ctx context.Context, user model.User, workspaceIDs []string, query string, limit int) ([]model.User, *model.AppError) {
	if user.Role != model.SystemAdminRoleId {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if limit <= 0 || limit > 50 {
		limit = 50
	}

	if len(workspaceIDs) > 500 {
		return nil, model.NewAppError("request.invalid", http.StatusBadRequest)
	}

	dbCtx, cancel := a.dbCtx(ctx)
	defer cancel()
	users, err := a.Store.Workspace.SearchMembers(dbCtx, workspaceIDs, query, limit, 0)
	if err != nil {
		tlog.Errorw("Failed to search task report members", "error", err)
		return nil, model.NewAppError("workspace.workspaces_retrieval_failed", http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) IsWorkspaceMemberOrGroupMember(workspaceID, userID string) bool {
	if ok, _ := a.Store.Workspace.IsMember(workspaceID, userID); ok {
		return true
	}

	hasGroup, _ := a.Store.Workspace.UserHasAnyGroupAccess(workspaceID, userID)
	return hasGroup
}

func (a *App) UpdateProjectWorkspaceGroupRoles(user model.User, workspaceID, groupID string, roles []string) *model.AppError {
	if !a.ProjectWorkspaceHasPermission(user, workspaceID, model.PermissionUpdateRoles) {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	if appErr := a.guardProjectGroupRoles(user, workspaceID, []string{groupID}, roles); appErr != nil {
		return appErr
	}

	return a.setProjectWorkspaceGroupRoles(workspaceID, groupID, roles)
}

func (a *App) setProjectWorkspaceGroupRoles(workspaceID, groupID string, roles []string) *model.AppError {
	// Allow clearing all roles (len==0) so admins can remove restrictions after expiry.
	if len(roles) > 0 && !a.Server.License.HasWorkspaceRoles() {
		return model.NewAppError("workspace.roles_license_required", http.StatusPaymentRequired)
	}

	roles, appErr := a.validProjectWorkspaceRoles(workspaceID, roles)
	if appErr != nil {
		return appErr
	}

	if err := a.Store.Workspace.UpdateGroupRoles(context.Background(), workspaceID, groupID, roles); err != nil {
		tlog.Errorw("Failed to update workspace group roles", "workspace_id", workspaceID, "group_id", groupID, "error", err)
		return model.NewAppError("collimato.group_roles_update_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) validProjectWorkspaceRoles(workspaceID string, roles []string) ([]string, *model.AppError) {
	roles = DedupeIDs(roles)

	if len(roles) == 0 {
		return roles, nil
	}

	found, err := a.Store.Workspace.GetRolesByName(roles, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	if len(found) != len(roles) {
		return nil, model.NewAppError("workspace.role_not_found", http.StatusBadRequest)
	}

	return roles, nil
}
