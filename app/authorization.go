// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

// user.Role is space-separated (direct roles plus any folded in from group
// membership by GetCurrentUser), so it is tokenized before the role lookup.
func (a *App) rolesForUser(user model.User) ([]model.Role, error) {
	roleNames := strings.Split(user.Role, " ")

	roles, err := a.Store.Roles.GetByNames(roleNames)
	if err != nil {
		return nil, err
	}

	if roles == nil {
		return nil, nil
	}

	return roles, nil
}

func (a *App) RolesGrantPermission(user model.User, permissionId string) bool {
	roles, err := a.rolesForUser(user)
	if err != nil {
		tlog.Errorw("Failed to get roles from database", "user_id", user.ID, "error", err)
		return false
	}

	for _, role := range roles {
		for _, permission := range role.Permissions {
			if permission == permissionId {
				return true
			}
		}
	}

	return false
}

// Roles come from the space-separated user.Role (direct plus group-granted), so
// it must be tokenized via rolesForUser before lookup: passing the whole string
// as one role name matches nothing once a user has more than one role.
func (a *App) GetMyPermissions(user model.User) ([]string, *model.AppError) {
	if user.Role == model.SystemAdminRoleId {
		sections := model.GetSystemPermissionSections()
		all := make([]string, 0)
		for _, section := range sections {
			for _, p := range section.Permissions {
				all = append(all, p.Id)
			}
		}

		return all, nil
	}

	roles, err := a.rolesForUser(user)
	if err != nil {
		tlog.Errorw("Failed to get roles from database", "user_id", user.ID, "error", err)
		return nil, model.NewAppError("role.retrieval_failed", http.StatusInternalServerError)
	}

	seen := make(map[string]struct{})
	permissions := make([]string, 0)
	for _, role := range roles {
		for _, p := range role.Permissions {
			if _, ok := seen[p]; ok {
				continue
			}

			seen[p] = struct{}{}
			permissions = append(permissions, p)
		}
	}

	return permissions, nil
}

func (a *App) SessionHasPermission(user model.User, permission *model.SystemPermission) bool {
	if user.Role == model.SystemAdminRoleId {
		return true
	}

	return a.RolesGrantPermission(user, permission.Id)
}

func (a *App) CollimatoWorkspaceGrantPermission(user model.User, workspaceId, permissionId string) bool {
	// No effective roles (direct or via an attached group) means the user isn't a
	// member of the workspace, so they get no permissions.
	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceId, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles", "user_id", user.ID, "workspace_id", workspaceId, "error", err)
		return false
	}

	if len(roleNames) == 0 {
		return false
	}

	roles, err := a.Store.Collimato.GetRolesByName(roleNames, workspaceId)
	if err != nil {
		tlog.Errorw("Failed to get roles from database", "user_id", user.ID, "workspace_id", workspaceId, "error", err)
		return false
	}

	for _, role := range roles {
		permissions := role.Permissions
		for _, permission := range permissions {
			if permission == permissionId {
				return true
			}
		}
	}

	return false
}

// System admins are deliberately not exempt here. Administering the instance
// does not grant access to a workspace's contents; an admin who needs in joins
// the workspace, which is visible to its members.
func (a *App) CollimatoWorkspaceHasPermission(user model.User, workspaceId string, permission *model.CollimatoPermission) bool {
	return a.CollimatoWorkspaceGrantPermission(user, workspaceId, permission.Id)
}

func (a *App) holdsCollimatoWorkspaceAdmin(user model.User, workspaceID string) bool {
	roleNames, err := a.Store.Collimato.GetEffectiveRolesForUser(context.Background(), workspaceID, user.ID)
	if err != nil {
		tlog.Errorw("Failed to resolve effective workspace roles", "user_id", user.ID, "workspace_id", workspaceID, "error", err)
		return false
	}

	return slices.Contains(roleNames, model.CollimatoWorkspaceAdminRoleId)
}

// Combines a user's direct workspace roles with the roles from any groups they
// belong to, deduplicated.
func (a *App) effectiveWorkspaceRoleNames(userID, workspaceID string) ([]string, bool) {
	seen := map[string]bool{}
	var roles []string

	workspaceUser, _ := a.Store.Workspace.GetUserByUserID(userID, workspaceID)
	if workspaceUser != nil {
		for _, r := range strings.Split(workspaceUser.Role, " ") {
			if !seen[r] {
				seen[r] = true
				roles = append(roles, r)
			}
		}
	}

	groupRoles, _ := a.Store.Workspace.GetGroupRolesForUser(userID, workspaceID)
	for _, r := range groupRoles {
		if !seen[r] {
			seen[r] = true
			roles = append(roles, r)
		}
	}

	return roles, len(roles) > 0
}

func (a *App) ProjectWorkspaceGrantPermission(user model.User, workspaceId, permissionId string) bool {
	roleNames, ok := a.effectiveWorkspaceRoleNames(user.ID, workspaceId)
	if !ok {
		return false
	}

	roles, err := a.Store.Workspace.GetRolesByName(roleNames, workspaceId)
	if err != nil {
		tlog.Errorw("Failed to get roles from database", "user_id", user.ID, "workspace_id", workspaceId, "error", err)
		return false
	}

	for _, role := range roles {
		permissions := role.Permissions
		for _, permission := range permissions {
			if permission == permissionId {
				return true
			}
		}
	}

	return false
}

func (a *App) ProjectWorkspaceHasPermission(user model.User, workspaceId string, permissionId string) bool {
	return a.ProjectWorkspaceGrantPermission(user, workspaceId, permissionId)
}

// CanPerformRowAction asks for membership before resolving roles because,
// without enterprise roles, resolving grants everything to anyone.
func (a *App) CanPerformRowAction(user model.User, workspaceID, tableID, workspacePerm, tableAction string) bool {
	if !a.IsWorkspaceMemberOrGroupMember(workspaceID, user.ID) {
		return false
	}

	in := a.buildAccessInput(user, workspaceID)
	return a.resolveAccess(in).CanPerformRowAction(tableID, workspacePerm, tableAction)
}

// requireTaskAccess refuses a task the user may not see. On a table where
// they see only the tasks assigned to them, that is any task assigned to
// someone else, or to no one. Lists leave such tasks out; this covers a task
// named by its id. A task that does not exist is refused the same way, so
// the answer does not tell which ids are tasks.
func (a *App) requireTaskAccess(user model.User, workspaceID, tableID, taskID string) *model.AppError {
	if a.assignedOnlyFilter(user, workspaceID, tableID).SQL == "" {
		return nil
	}

	tableName, err := a.Store.Workspace.GetTableName(tableID)
	if err != nil {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	assignee, err := a.Store.Workspace.GetTaskAssigneeID(tableName, taskID)
	if err != nil || assignee != user.ID {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return nil
}

// restrictLinkedItems hides, in the link cells of rows, the linked tasks the
// user may not see. Where they see only their own tasks in the linked table,
// a task assigned to someone else keeps its id, so the link stays, but is
// sent without its name. A cell whose owners cannot be read is hidden whole.
func (a *App) restrictLinkedItems(ctx context.Context, rows []map[string]interface{}, headers []model.WorkspaceHeaders, user model.User, workspaceID string) {
	access := a.resolveAccess(a.buildAccessInput(user, workspaceID))
	a.hideUnseenLinks(ctx, rows, headers, user.ID, access.AssignedOnly)
}

// hideUnseenLinks is restrictLinkedItems for a caller that already knows
// which tables show the user only their own tasks.
func (a *App) hideUnseenLinks(ctx context.Context, rows []map[string]interface{}, headers []model.WorkspaceHeaders, userID string, ownTasksOnly func(tableID string) bool) {
	for _, header := range headers {
		if header.LinkedID == "" || header.SingleSelect || header.ParentTableID == "" {
			continue
		}

		if !ownTasksOnly(header.ParentTableID) {
			continue
		}

		var ids []string
		for _, row := range rows {
			links, _ := row[header.Name].([]model.LinkedItem)
			for _, link := range links {
				ids = append(ids, link.ID)
			}
		}

		if len(ids) == 0 {
			continue
		}

		owners := map[string]string{}
		tableName, err := a.Store.Workspace.GetTableName(header.ParentTableID)
		if err == nil {
			owners, err = a.Store.Workspace.GetTaskAssignees(ctx, tableName, ids)
		}

		if err != nil {
			tlog.Warnw("Failed to read who linked tasks are assigned to", "table_id", header.ParentTableID, "error", err)
		}

		for _, row := range rows {
			links, ok := row[header.Name].([]model.LinkedItem)
			if !ok {
				continue
			}

			shown := make([]model.LinkedItem, len(links))
			for i, link := range links {
				if err == nil && owners[link.ID] == userID {
					shown[i] = link
				} else {
					shown[i] = model.LinkedItem{ID: link.ID, Restricted: true}
				}
			}

			row[header.Name] = shown
		}
	}
}

// The task report's saved filters are stored under these ids, which are not a
// real workspace and table.
const (
	taskReportWorkspaceID = "all"
	taskReportTableID     = "detailed-task-report"
)

// CheckTableInWorkspace refuses a table that is not in the workspace a request
// names. Permission checks look only at the workspace, so without this a
// member of one workspace could name another workspace's table.
func (a *App) CheckTableInWorkspace(ctx context.Context, workspaceID, tableID string) *model.AppError {
	if workspaceID == taskReportWorkspaceID && tableID == taskReportTableID {
		return nil
	}

	ctx, cancel := a.dbCtx(ctx)
	defer cancel()
	owner, err := a.Store.Workspace.GetTableWorkspaceID(ctx, tableID)
	if err != nil {
		tlog.Errorw("Failed to resolve table workspace", "workspace_id", workspaceID, "table_id", tableID, "error", err)
		return model.NewAppError("workspace.tables_retrieval_failed", http.StatusInternalServerError)
	}

	if owner == "" || owner != workspaceID {
		return model.NewAppError("workspace.forbidden", http.StatusForbidden)
	}

	return nil
}

// canReadTable reports whether user may read tableID's rows through
// workspaceID: the table belongs to that workspace, the user is a member of it
// directly or through a group, and their roles do not hide the table.
func (a *App) canReadTable(ctx context.Context, user model.User, workspaceID, tableID string) bool {
	if workspaceID == taskReportWorkspaceID || a.CheckTableInWorkspace(ctx, workspaceID, tableID) != nil {
		return false
	}

	if !a.IsWorkspaceMemberOrGroupMember(workspaceID, user.ID) {
		return false
	}

	access := a.resolveAccess(a.buildAccessInput(user, workspaceID))
	return len(access.VisibleTables([]model.WorkspaceTable{{ID: tableID}})) == 1
}

func (a *App) ProjectWorkspaceTableHasPermission(user model.User, workspaceId, tableId, action string) bool {
	if !a.IsWorkspaceMemberOrGroupMember(workspaceId, user.ID) {
		return false
	}

	in := a.buildAccessInput(user, workspaceId)
	return a.resolveAccess(in).TableHasPermission(tableId, action)
}
