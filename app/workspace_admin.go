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
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

func (a *App) canManageProjects(user model.User) bool {
	return a.SessionHasPermission(user, model.ProjectSectionPermissions.PermissionManageProjects)
}

func (a *App) managedProjectWorkspace(user model.User, workspaceID string) (*model.Workspace, *model.AppError) {
	if !a.canManageProjects(user) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Workspace.GetRow(workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.NewAppError("workspace.workspace_not_found", http.StatusNotFound)
	}

	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.workspaces_retrieval_failed", http.StatusInternalServerError)
	}

	return workspace, nil
}

func (a *App) ListAllProjectWorkspaces(ctx context.Context, user model.User) ([]model.Workspace, *model.AppError) {
	if !a.canManageProjects(user) {
		return nil, model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	failed := model.NewAppError("workspace.workspaces_retrieval_failed", http.StatusInternalServerError)

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaces, err := a.Store.Workspace.GetAll(ctx)
	if err != nil {
		tlog.Errorw("Failed to list project workspaces", "user_id", user.ID, "error", err)
		return nil, failed
	}

	if len(workspaces) == 0 {
		return workspaces, nil
	}

	ids := make([]string, len(workspaces))
	for i, ws := range workspaces {
		ids[i] = ws.ID
	}

	members, err := a.Store.Workspace.GetMembersForWorkspaces(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace members", "user_id", user.ID, "error", err)
		return nil, failed
	}

	groupCounts, err := a.Store.Workspace.GetGroupCounts(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace group counts", "user_id", user.ID, "error", err)
		return nil, failed
	}

	tables, err := a.Store.Workspace.GetTableMetas(ids)
	if err != nil {
		tlog.Errorw("Failed to get workspace tables", "user_id", user.ID, "error", err)
		return nil, failed
	}

	tableCounts := make(map[string]int, len(ids))
	for _, t := range tables {
		tableCounts[t.WorkspaceID]++
	}

	for i := range workspaces {
		ws := &workspaces[i]

		all := members[ws.ID]

		ws.MemberCount = len(all)
		ws.Members = all[:min(len(all), workspaceMemberPreviewSize)]
		if ws.Members == nil {
			ws.Members = []model.WorkspaceMember{}
		}

		ws.GroupCount = groupCounts[ws.ID]
		ws.TableCount = tableCounts[ws.ID]
	}

	return workspaces, nil
}

func (a *App) GetManagedProjectWorkspace(ctx context.Context, user model.User, workspaceID string) (*model.ProjectWorkspaceDetails, *model.AppError) {
	workspace, appErr := a.managedProjectWorkspace(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	members, err := a.Store.Workspace.GetMembers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace members",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	groups, err := a.Store.Workspace.GetGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	roles, err := a.Store.Workspace.GetRoles(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("workspace.role_retrieval_failed", http.StatusInternalServerError)
	}

	return &model.ProjectWorkspaceDetails{
		Workspace: *workspace,
		Members:   members,
		Groups:    groups,
		Roles:     roles,
	}, nil
}

func (a *App) UpdateManagedProjectWorkspace(user model.User, workspaceID string, patch model.ProjectWorkspacePatch) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	name := strings.TrimSpace(patch.Name)
	if name == "" {
		return model.NewAppError("workspace.workspace_name_empty", http.StatusBadRequest)
	}

	description := strings.TrimSpace(patch.Description)

	if _, err := a.Store.Workspace.Update(workspaceID, name, &description); err != nil {
		tlog.Errorw("Failed to update workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("workspace.workspace_update_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) DeleteManagedProjectWorkspace(ctx context.Context, user model.User, workspaceID string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if _, appErr := a.deleteProjectWorkspace(ctx, workspaceID, user); appErr != nil {
		return appErr
	}

	tlog.Infow("Project workspace deleted from settings",
		"workspace_id", workspaceID,
		"user_id", user.ID,
	)

	return nil
}

func (a *App) AddManagedProjectWorkspaceMembers(user model.User, workspaceID string, userIDs []string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	_, appErr := a.addProjectWorkspaceMembers(workspaceID, DedupeIDs(userIDs), user)

	return appErr
}

func (a *App) RemoveManagedProjectWorkspaceMember(user model.User, workspaceID, memberID string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if appErr := a.keepProjectWorkspaceAdmin(workspaceID, memberID, ""); appErr != nil {
		return appErr
	}

	_, appErr := a.removeProjectWorkspaceMember(workspaceID, memberID)

	return appErr
}

func (a *App) UpdateManagedProjectWorkspaceMemberRole(user model.User, workspaceID, memberID, role string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if appErr := a.keepProjectWorkspaceAdmin(workspaceID, memberID, role); appErr != nil {
		return appErr
	}

	updated, appErr := a.setProjectWorkspaceMemberRole(workspaceID, memberID, role, user)
	if appErr != nil {
		return appErr
	}

	if !updated {
		return model.NewAppError("workspace.member_not_found", http.StatusNotFound)
	}

	return nil
}

func (a *App) AddManagedProjectWorkspaceGroups(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.WorkspaceGroup, *model.AppError) {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return nil, appErr
	}

	return a.attachProjectWorkspaceGroups(user, workspaceID, DedupeIDs(groupIDs), roles)
}

func (a *App) RemoveManagedProjectWorkspaceGroup(user model.User, workspaceID, groupID string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	return a.detachProjectWorkspaceGroup(workspaceID, groupID)
}

func (a *App) UpdateManagedProjectWorkspaceGroupRoles(user model.User, workspaceID, groupID string, roles []string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	return a.setProjectWorkspaceGroupRoles(workspaceID, groupID, roles)
}

func (a *App) JoinManagedProjectWorkspace(user model.User, workspaceID string) *model.AppError {
	if _, appErr := a.managedProjectWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	member := model.WorkspaceMember{
		ID:          model.NewID(),
		UserID:      user.ID,
		WorkspaceID: workspaceID,
		Role:        model.ProjectWorkspaceAdminRoleID,
		DateJoined:  time.Now().Unix(),
	}

	if _, err := a.Store.Workspace.AddMember([]model.WorkspaceMember{member}); err != nil {
		tlog.Errorw("Failed to add user to workspace",
			"workspace_id", workspaceID,
			"member_id", user.ID,
			"error", err,
		)
		return model.NewAppError("workspace.member_add_failed", http.StatusInternalServerError)
	}

	stored, err := a.Store.Workspace.GetMemberByUserID(workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"member_id", user.ID,
			"error", err,
		)
		return model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	roles := strings.Fields(stored.Role)

	if !slices.Contains(roles, model.ProjectWorkspaceAdminRoleID) {
		roles = append(roles, model.ProjectWorkspaceAdminRoleID)

		if _, err := a.Store.Workspace.UpdateMemberRole(workspaceID, user.ID, strings.Join(roles, " "), user); err != nil {
			tlog.Errorw("Failed to update workspace member role",
				"workspace_id", workspaceID,
				"member_id", user.ID,
				"error", err,
			)
			return model.NewAppError("workspace.member_role_update_failed", http.StatusInternalServerError)
		}
	}

	tlog.Infow("Joined project workspace as admin from settings",
		"workspace_id", workspaceID,
		"user_id", user.ID,
	)

	return nil
}

// keepProjectWorkspaceAdmin refuses a change that would leave a workspace with
// no direct admin, since only an admin can manage the workspace from inside it.
func (a *App) keepProjectWorkspaceAdmin(workspaceID, memberID, newRole string) *model.AppError {
	if isProjectWorkspaceAdmin(newRole) {
		return nil
	}

	members, err := a.Store.Workspace.GetMembers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace members",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("workspace.member_retrieval_failed", http.StatusInternalServerError)
	}

	target := slices.IndexFunc(members, func(m model.WorkspaceMember) bool { return m.UserID == memberID })
	if target < 0 {
		return model.NewAppError("workspace.member_not_found", http.StatusNotFound)
	}

	if !isProjectWorkspaceAdmin(members[target].Role) {
		return nil
	}

	admins := 0
	for _, m := range members {
		if isProjectWorkspaceAdmin(m.Role) {
			admins++
		}
	}

	if admins <= 1 {
		return model.NewAppError("workspace.last_admin", http.StatusBadRequest)
	}

	return nil
}

func isProjectWorkspaceAdmin(role string) bool {
	return slices.Contains(strings.Fields(role), model.ProjectWorkspaceAdminRoleID)
}
