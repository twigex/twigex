// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/twigex/twigex/model"
)

// GetMembersForWorkspaces returns the direct members of each workspace.
func (w *workspaceRepository) GetMembersForWorkspaces(ctx context.Context, workspaceIDs []string) (map[string][]model.WorkspaceMember, error) {
	members := make(map[string][]model.WorkspaceMember, len(workspaceIDs))
	if len(workspaceIDs) == 0 {
		return members, nil
	}

	rows, err := w.Db.QueryContext(ctx, `
		SELECT wm.id, wm.user_id, wm.workspace_id, wm.role, wm.date_joined,
		       u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
		FROM workspace_members wm
		JOIN users u ON u.id = wm.user_id
		WHERE wm.workspace_id IN (`+sqlPlaceholders(len(workspaceIDs))+`)`,
		toInterfaceSlice(workspaceIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var member model.WorkspaceMember
		if err := rows.Scan(
			&member.ID,
			&member.UserID,
			&member.WorkspaceID,
			&member.Role,
			&member.DateJoined,
			&member.UserInfo.Email,
			&member.UserInfo.Username,
			&member.UserInfo.Name,
			&member.UserInfo.LastName,
			&member.UserInfo.Photo,
		); err != nil {
			return nil, err
		}

		member.UserInfo.ID = member.UserID
		members[member.WorkspaceID] = append(members[member.WorkspaceID], member)
	}

	return members, rows.Err()
}

// GetGroupCounts returns how many live groups each workspace is shared with.
func (w *workspaceRepository) GetGroupCounts(ctx context.Context, workspaceIDs []string) (map[string]int, error) {
	counts := make(map[string]int, len(workspaceIDs))
	if len(workspaceIDs) == 0 {
		return counts, nil
	}

	rows, err := w.Db.QueryContext(ctx, `
		SELECT wg.workspace_id, COUNT(*)
		FROM workspace_groups wg
		JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		WHERE wg.workspace_id IN (`+sqlPlaceholders(len(workspaceIDs))+`)
		GROUP BY wg.workspace_id`,
		toInterfaceSlice(workspaceIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var workspaceID string
		var count int
		if err := rows.Scan(&workspaceID, &count); err != nil {
			return nil, err
		}

		counts[workspaceID] = count
	}

	return counts, rows.Err()
}

// GetUserRoleNamesByWorkspace returns the names of the roles a user holds in
// each workspace, directly or through a group.
func (w *workspaceRepository) GetUserRoleNamesByWorkspace(ctx context.Context, userID string, workspaceIDs []string) (map[string][]string, error) {
	names := make(map[string][]string, len(workspaceIDs))
	if len(workspaceIDs) == 0 {
		return names, nil
	}

	seen := map[string]map[string]bool{}
	add := func(workspaceID, roles string) {
		if seen[workspaceID] == nil {
			seen[workspaceID] = map[string]bool{}
		}

		for _, r := range strings.Fields(roles) {
			if !seen[workspaceID][r] {
				seen[workspaceID][r] = true
				names[workspaceID] = append(names[workspaceID], r)
			}
		}
	}

	in := sqlPlaceholders(len(workspaceIDs))
	args := append([]any{userID}, toInterfaceSlice(workspaceIDs)...)
	rows, err := w.Db.QueryContext(ctx, `
		SELECT workspace_id, role FROM workspace_members
		WHERE user_id = ? AND workspace_id IN (`+in+`)
		UNION ALL
		SELECT wg.workspace_id, wg.roles
		FROM workspace_groups wg
		JOIN group_members gm ON gm.group_id = wg.group_id
		JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		WHERE gm.user_id = ? AND wg.workspace_id IN (`+in+`) AND wg.roles IS NOT NULL AND wg.roles != ''`,
		append(args, args...)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var workspaceID, roles string
		if err := rows.Scan(&workspaceID, &roles); err != nil {
			return nil, err
		}

		add(workspaceID, roles)
	}

	return names, rows.Err()
}

// GetRolesForWorkspaces returns the roles defined in each workspace.
func (w *workspaceRepository) GetRolesForWorkspaces(ctx context.Context, workspaceIDs []string) (map[string][]model.ProjectWorkspaceRole, error) {
	roles := make(map[string][]model.ProjectWorkspaceRole, len(workspaceIDs))
	if len(workspaceIDs) == 0 {
		return roles, nil
	}

	rows, err := w.Db.QueryContext(ctx, `
		SELECT workspace_id, id, name, displayname, description, permissions, per_table_mode, created_at, updated_at
		FROM workspace_roles
		WHERE workspace_id IN (`+sqlPlaceholders(len(workspaceIDs))+`)`,
		toInterfaceSlice(workspaceIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var workspaceID, permissions string
		var role model.ProjectWorkspaceRole
		if err := rows.Scan(
			&workspaceID,
			&role.ID,
			&role.Name,
			&role.DisplayName,
			&role.Description,
			&permissions,
			&role.PerTableMode,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(permissions), &role.Permissions); err != nil {
			role.Permissions = strings.Split(permissions, " ")
		}

		roles[workspaceID] = append(roles[workspaceID], role)
	}

	return roles, rows.Err()
}

// GetTablePermissionsForRoleIDs returns the table permissions of the roles,
// whichever workspaces they belong to.
func (w *workspaceRepository) GetTablePermissionsForRoleIDs(ctx context.Context, roleIDs []string) ([]model.TablePermission, error) {
	perms := make([]model.TablePermission, 0)
	if len(roleIDs) == 0 {
		return perms, nil
	}

	rows, err := w.Db.QueryContext(ctx, `
		SELECT id, role_id, workspace_id, table_id, action, created_at, updated_at
		FROM workspace_table_permissions
		WHERE role_id IN (`+sqlPlaceholders(len(roleIDs))+`)`,
		toInterfaceSlice(roleIDs)...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var p model.TablePermission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.WorkspaceID, &p.TableID, &p.Action, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		perms = append(perms, p)
	}

	return perms, rows.Err()
}
