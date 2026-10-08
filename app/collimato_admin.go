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

func (a *App) canManageCollimato(user model.User) bool {
	return a.SessionHasPermission(user, model.CollimatoSectionPermissions.PermissionManageCollimato)
}

func (a *App) managedWorkspace(user model.User, workspaceID string) (*model.CollimatoWorkspace, *model.AppError) {
	if !a.canManageCollimato(user) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	workspace, err := a.Store.Collimato.GetWorkspaceByID(workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.NewAppError("collimato.not_found", http.StatusNotFound)
	}

	if err != nil {
		tlog.Errorw("Failed to retrieve workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	return workspace, nil
}

func (a *App) ListAllWorkspaces(ctx context.Context, user model.User) ([]model.CollimatoWorkspace, *model.AppError) {
	if !a.canManageCollimato(user) {
		return nil, model.NewAppError("collimato.forbidden", http.StatusForbidden)
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	workspaces, err := a.Store.Collimato.GetWorkspaces()
	if err != nil {
		tlog.Errorw("Failed to retrieve workspaces",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	ids := make([]string, len(workspaces))
	for i, workspace := range workspaces {
		ids[i] = workspace.ID
	}

	members, err := a.Store.Collimato.GetWorkspaceMemberIDs(ctx, ids)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace members",
			"user_id", user.ID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.retrieval_failed", http.StatusInternalServerError)
	}

	for i := range workspaces {
		memberIDs := members[workspaces[i].ID]

		workspaces[i].MemberCount = len(memberIDs)
		workspaces[i].MemberIDs = memberIDs[:min(len(memberIDs), workspaceMemberPreviewSize)]
	}

	return workspaces, nil
}

func (a *App) GetWorkspaceDetails(ctx context.Context, user model.User, workspaceID string) (*model.CollimatoWorkspaceDetails, *model.AppError) {
	workspace, appErr := a.managedWorkspace(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()

	users, err := a.Store.Collimato.GetWorkspaceUsers(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace users",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	viaGroup, err := a.Store.Collimato.GetWorkspaceGroupMembers(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace group members",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	groups, err := a.Store.Collimato.GetWorkspaceGroups(ctx, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to list workspace groups",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.groups_retrieval_failed", http.StatusInternalServerError)
	}

	roles, err := a.Store.Collimato.GetWorkspaceRoles(workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace roles",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.roles_failed", http.StatusInternalServerError)
	}

	return &model.CollimatoWorkspaceDetails{
		Workspace: *workspace,
		Users:     append(users, viaGroup...),
		Groups:    groups,
		Roles:     roles,
	}, nil
}

func (a *App) UpdateManagedWorkspace(user model.User, workspaceID string, patch model.CollimatoWorkspacePatch) (*model.CollimatoWorkspace, *model.AppError) {
	workspace, appErr := a.managedWorkspace(user, workspaceID)
	if appErr != nil {
		return nil, appErr
	}

	name := strings.TrimSpace(patch.Name)
	if name == "" {
		return nil, model.NewAppError("collimato.name_required", http.StatusBadRequest)
	}

	workspace.Name = name
	workspace.Description = strings.TrimSpace(patch.Description)

	updated, err := a.Store.Collimato.UpdateWorkspace(*workspace)
	if err != nil {
		tlog.Errorw("Failed to update workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.update_failed", http.StatusInternalServerError)
	}

	return updated, nil
}

func (a *App) DeleteManagedWorkspace(user model.User, workspaceID string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if err := a.Store.Collimato.DeleteWorkspace(workspaceID); err != nil {
		tlog.Errorw("Failed to delete workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.delete_failed", http.StatusInternalServerError)
	}

	tlog.Infow("Workspace deleted from settings",
		"workspace_id", workspaceID,
		"user_id", user.ID,
	)

	return nil
}

func (a *App) AddManagedWorkspaceUsers(user model.User, workspaceID string, userIDs []string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	err := a.Store.Collimato.AddWorkspaceUsers(workspaceID, DedupeIDs(userIDs), []string{
		model.CollimatoWorkspaceUserRoleId,
	})
	if err != nil {
		tlog.Errorw("Failed to add users to workspace",
			"workspace_id", workspaceID,
			"error", err,
		)
		return model.NewAppError("collimato.member_add_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) RemoveManagedWorkspaceUser(user model.User, workspaceID, userID string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if appErr := a.keepCollimatoWorkspaceAdmin(workspaceID, userID, nil); appErr != nil {
		return appErr
	}

	if err := a.Store.Collimato.RemoveWorkspaceUser(workspaceID, userID); err != nil {
		tlog.Errorw("Failed to remove user from workspace",
			"workspace_id", workspaceID,
			"member_id", userID,
			"error", err,
		)
		return model.NewAppError("collimato.member_remove_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateManagedWorkspaceUserRoles(user model.User, workspaceID, userID string, roles []string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if a.CollimatoRoles == nil {
		return rolesUnavailable()
	}

	member, appErr := a.workspaceMember(workspaceID, userID)
	if appErr != nil {
		return appErr
	}

	if appErr := a.keepCollimatoWorkspaceAdmin(workspaceID, userID, roles); appErr != nil {
		return appErr
	}

	return a.CollimatoRoles.AssignUserRoles(workspaceID, member.ID, roles)
}

func (a *App) AddManagedWorkspaceGroups(user model.User, workspaceID string, groupIDs []string, roles []string) ([]model.CollimatoWorkspaceGroup, *model.AppError) {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return nil, appErr
	}

	if !a.Server.License.HasGroups() {
		return nil, model.NewAppError("groups.license_required", http.StatusPaymentRequired)
	}

	return a.attachWorkspaceGroups(user, workspaceID, groupIDs, roles)
}

func (a *App) RemoveManagedWorkspaceGroup(user model.User, workspaceID, groupID string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if err := a.Store.Collimato.RemoveWorkspaceGroup(context.Background(), workspaceID, groupID); err != nil {
		tlog.Errorw("Failed to remove group from workspace",
			"workspace_id", workspaceID,
			"group_id", groupID,
			"error", err,
		)
		return model.NewAppError("collimato.group_remove_failed", http.StatusInternalServerError)
	}

	return nil
}

func (a *App) UpdateManagedWorkspaceGroupRoles(user model.User, workspaceID, groupID string, roles []string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if a.CollimatoRoles == nil {
		return rolesUnavailable()
	}

	return a.CollimatoRoles.AssignGroupRoles(workspaceID, groupID, roles)
}

func (a *App) JoinManagedWorkspace(user model.User, workspaceID string) *model.AppError {
	if _, appErr := a.managedWorkspace(user, workspaceID); appErr != nil {
		return appErr
	}

	if appErr := a.ensureWorkspaceAdmin(workspaceID, user.ID); appErr != nil {
		return appErr
	}

	tlog.Infow("Joined workspace as admin from settings",
		"workspace_id", workspaceID,
		"user_id", user.ID,
	)

	return nil
}

func (a *App) workspaceMember(workspaceID, userID string) (*model.CollimatoWorkspaceUser, *model.AppError) {
	member, err := a.Store.Collimato.GetWorkspaceUserByUserID(userID, workspaceID)
	if err != nil {
		tlog.Errorw("Failed to retrieve workspace member",
			"workspace_id", workspaceID,
			"member_id", userID,
			"error", err,
		)
		return nil, model.NewAppError("collimato.users_retrieval_failed", http.StatusInternalServerError)
	}

	if member == nil {
		return nil, model.NewAppError("collimato.member_not_found", http.StatusNotFound)
	}

	return member, nil
}

// keepCollimatoWorkspaceAdmin refuses a change that would leave a workspace with
// no direct workspace admin, since only an admin can delete it from inside.
func (a *App) keepCollimatoWorkspaceAdmin(workspaceID, userID string, newRoles []string) *model.AppError {
	if slices.Contains(newRoles, model.CollimatoWorkspaceAdminRoleId) {
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

	target := slices.IndexFunc(members, func(m model.CollimatoWorkspaceUser) bool { return m.UserID == userID })
	if target < 0 {
		return model.NewAppError("collimato.member_not_found", http.StatusNotFound)
	}

	if !isCollimatoWorkspaceAdmin(members[target].Role) {
		return nil
	}

	admins := 0
	for _, m := range members {
		if isCollimatoWorkspaceAdmin(m.Role) {
			admins++
		}
	}

	if admins <= 1 {
		return model.NewAppError("collimato.last_admin", http.StatusBadRequest)
	}

	return nil
}

func isCollimatoWorkspaceAdmin(role string) bool {
	return slices.Contains(strings.Split(role, ","), model.CollimatoWorkspaceAdminRoleId)
}

func (a *App) ensureWorkspaceAdmin(workspaceID, userID string) *model.AppError {
	err := a.Store.Collimato.AddWorkspaceUsers(workspaceID, []string{userID}, []string{
		model.CollimatoWorkspaceAdminRoleId,
	})
	if err != nil {
		tlog.Errorw("Failed to add user to workspace",
			"workspace_id", workspaceID,
			"member_id", userID,
			"error", err,
		)
		return model.NewAppError("collimato.member_add_failed", http.StatusInternalServerError)
	}

	member, appErr := a.workspaceMember(workspaceID, userID)
	if appErr != nil {
		return appErr
	}

	roles := []string{}
	if member.Role != "" {
		roles = strings.Split(member.Role, ",")
	}

	if slices.Contains(roles, model.CollimatoWorkspaceAdminRoleId) {
		return nil
	}

	roles = append(roles, model.CollimatoWorkspaceAdminRoleId)

	if err := a.Store.Collimato.UpdateUserRoles(workspaceID, member.ID, roles); err != nil {
		tlog.Errorw("Failed to update user roles",
			"workspace_id", workspaceID,
			"member_id", userID,
			"error", err,
		)
		return model.NewAppError("collimato.roles_update_failed", http.StatusInternalServerError)
	}

	return nil
}
