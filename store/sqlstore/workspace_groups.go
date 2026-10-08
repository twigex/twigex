// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

// AddGroups attaches a batch of groups to a workspace, all with the
// same role set. Idempotent: re-attaching an existing group updates its roles
// (ON DUPLICATE KEY UPDATE).
func (w *workspaceRepository) AddGroups(ctx context.Context, workspaceID string, groupIDs []string, roles []string, addedBy string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	rolesStr := strings.Join(roles, " ")
	now := time.Now().Unix()
	placeholders := make([]string, len(groupIDs))
	args := make([]any, 0, len(groupIDs)*6)
	for i, gid := range groupIDs {
		placeholders[i] = "(?,?,?,?,?,?)"
		args = append(args, model.NewID(), workspaceID, gid, rolesStr, addedBy, now)
	}

	_, err := w.Db.ExecContext(ctx,
		`INSERT INTO workspace_groups (id, workspace_id, group_id, roles, added_by, added_at)
		 VALUES `+strings.Join(placeholders, ",")+`
		 ON DUPLICATE KEY UPDATE roles = VALUES(roles)`,
		args...,
	)
	return err
}

// GetGroups returns the group attachments for a workspace, enriched
// with group name, description, and current member count.
func (w *workspaceRepository) GetGroups(ctx context.Context, workspaceID string) ([]model.WorkspaceGroup, error) {
	rows, err := w.Db.QueryContext(ctx,
		`SELECT wg.id, wg.workspace_id, wg.group_id, wg.roles, wg.added_by, wg.added_at,
		        ug.name, ug.description, `+memberCount("wg.group_id")+`
		 FROM workspace_groups wg
		 JOIN user_groups ug ON ug.id = wg.group_id AND ug.deleted_at = 0
		 WHERE wg.workspace_id = ?
		 ORDER BY ug.name ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]model.WorkspaceGroup, 0)
	for rows.Next() {
		var wg model.WorkspaceGroup
		var rolesStr, description sql.NullString
		var memberCount sql.NullInt64
		if err := rows.Scan(&wg.ID, &wg.WorkspaceID, &wg.GroupID, &rolesStr, &wg.AddedBy, &wg.AddedAt,
			&wg.Name, &description, &memberCount); err != nil {
			return nil, err
		}

		wg.Roles = strings.Split(rolesStr.String, " ")
		wg.Description = description.String
		wg.MemberCount = int(memberCount.Int64)
		out = append(out, wg)
	}

	return out, rows.Err()
}

// RemoveGroup hard-deletes the group attachment (matching the hard-delete
// pattern used for workspace_members).
func (w *workspaceRepository) RemoveGroup(ctx context.Context, workspaceID, groupID string) error {
	_, err := w.Db.ExecContext(ctx,
		`DELETE FROM workspace_groups WHERE workspace_id = ? AND group_id = ?`,
		workspaceID, groupID,
	)
	return err
}

// SearchMembers returns paginated users who are members of any of
// workspaceIDs, or of any live workspace when there are none, either directly
// (workspace_members) or via an attached group (workspace_groups →
// group_members). LIKE search on name/lastname/email/username.
func (w *workspaceRepository) SearchMembers(ctx context.Context, workspaceIDs []string, query string, limit, offset int) ([]model.User, error) {
	like := "%" + escapeLike(query) + "%"

	in := "(SELECT id FROM workspaces WHERE deleted_at IS NULL OR deleted_at = 0)"
	var ids []any
	if len(workspaceIDs) > 0 {
		in = "(" + sqlPlaceholders(len(workspaceIDs)) + ")"
		ids = toInterfaceSlice(workspaceIDs)
	}

	args := append([]any{like, like, like, like}, ids...)
	args = append(args, ids...)
	args = append(args, limit, offset)
	rows, err := w.Db.QueryContext(ctx,
		`SELECT DISTINCT u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
		 FROM users u
		 WHERE u.deactivated_at = 0
		   AND (u.name LIKE ? OR u.lastname LIKE ? OR u.email LIKE ? OR u.username LIKE ?)
		   AND (
		     EXISTS (
		       SELECT 1 FROM workspace_members wm
		       WHERE wm.workspace_id IN `+in+` AND wm.user_id = u.id
		     ) OR EXISTS (
		       SELECT 1 FROM workspace_groups wg
		       JOIN group_members gm ON gm.group_id = wg.group_id
		       JOIN user_groups ug  ON ug.id = wg.group_id AND ug.deleted_at = 0
		       WHERE wg.workspace_id IN `+in+` AND gm.user_id = u.id
		     )
		   )
		 ORDER BY u.name ASC, u.lastname ASC
		 LIMIT ? OFFSET ?`,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]model.User, 0, limit)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		out = append(out, u)
	}

	return out, rows.Err()
}

// UpdateGroupRoles replaces the role set for an existing attachment.
func (w *workspaceRepository) UpdateGroupRoles(ctx context.Context, workspaceID, groupID string, roles []string) error {
	_, err := w.Db.ExecContext(ctx,
		`UPDATE workspace_groups SET roles = ? WHERE workspace_id = ? AND group_id = ?`,
		strings.Join(roles, " "), workspaceID, groupID,
	)
	return err
}
