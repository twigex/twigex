// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

// AddWorkspaceGroups attaches a batch of groups to a workspace, all with the
// same role set, in one statement. Idempotent: re-attaching a group updates its
// roles and reactivates a soft-deleted attachment (UNIQUE(workspace_id, group_id)
// drives ON DUPLICATE KEY UPDATE).
func (c *collimatoRepository) AddWorkspaceGroups(ctx context.Context, workspaceID string, groupIDs []string, roles []string, addedBy string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	rolesStr := strings.Join(roles, ",")
	now := time.Now().Unix()
	placeholders := make([]string, len(groupIDs))
	args := make([]any, 0, len(groupIDs)*7)
	for i, gid := range groupIDs {
		placeholders[i] = "(?,?,?,?,?,?,?,0)"
		args = append(args, model.NewID(), workspaceID, gid, rolesStr, addedBy, now, now)
	}

	_, err := c.Db.ExecContext(ctx,
		`INSERT INTO collimato_workspace_groups
		 (id, workspace_id, group_id, roles, added_by, added_at, updated_at, deleted_at)
		 VALUES `+strings.Join(placeholders, ",")+`
		 ON DUPLICATE KEY UPDATE roles = VALUES(roles), updated_at = VALUES(updated_at), deleted_at = 0`,
		args...,
	)
	return err
}

// GetWorkspaceGroups returns the active group attachments for a workspace,
// enriched with the linked group's name, description, and current member count.
func (c *collimatoRepository) GetWorkspaceGroups(ctx context.Context, workspaceID string) ([]model.CollimatoWorkspaceGroup, error) {
	rows, err := c.Db.QueryContext(ctx,
		`SELECT cwg.id, cwg.workspace_id, cwg.group_id, cwg.roles, cwg.added_by,
		        cwg.added_at, cwg.updated_at,
		        ug.name, ug.description, `+memberCount("cwg.group_id")+`
		 FROM collimato_workspace_groups cwg
		 JOIN user_groups ug ON ug.id = cwg.group_id AND ug.deleted_at = 0
		 WHERE cwg.workspace_id = ? AND cwg.deleted_at = 0
		 ORDER BY ug.name ASC`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]model.CollimatoWorkspaceGroup, 0)
	for rows.Next() {
		var cwg model.CollimatoWorkspaceGroup
		var roles, description sql.NullString
		var updatedAt, memberCount sql.NullInt64
		if err := rows.Scan(&cwg.ID, &cwg.WorkspaceID, &cwg.GroupID, &roles, &cwg.AddedBy,
			&cwg.AddedAt, &updatedAt, &cwg.Name, &description, &memberCount); err != nil {
			return nil, err
		}

		cwg.Roles = strings.Split(roles.String, ",")
		cwg.UpdatedAt = updatedAt.Int64
		cwg.Description = description.String
		cwg.MemberCount = int(memberCount.Int64)
		out = append(out, cwg)
	}

	return out, rows.Err()
}

// RemoveWorkspaceGroup soft-deletes a group attachment (consistent with
// RemoveWorkspaceUser). Effective roles drop immediately because resolution is
// live; there is nothing materialized to clean up.
func (c *collimatoRepository) RemoveWorkspaceGroup(ctx context.Context, workspaceID, groupID string) error {
	_, err := c.Db.ExecContext(ctx,
		`UPDATE collimato_workspace_groups SET deleted_at = ?
		 WHERE workspace_id = ? AND group_id = ?`,
		time.Now().Unix(), workspaceID, groupID,
	)
	return err
}

// UpdateWorkspaceGroupRoles replaces the role set for an active attachment.
func (c *collimatoRepository) UpdateWorkspaceGroupRoles(ctx context.Context, workspaceID, groupID string, roles []string) error {
	_, err := c.Db.ExecContext(ctx,
		`UPDATE collimato_workspace_groups SET roles = ?, updated_at = ?
		 WHERE workspace_id = ? AND group_id = ? AND deleted_at = 0`,
		strings.Join(roles, ","), time.Now().Unix(), workspaceID, groupID,
	)
	return err
}

// GetWorkspaceGroupMembers returns the users who have access to a workspace
// only through an attached group (i.e. they have no direct
// collimato_workspace_users row). Each returned entry is synthesized with
// ViaGroup=true and Role set to the union of role names granted by the attached
// group(s) the user belongs to. Direct members are excluded so the caller can
// concatenate this with the direct list without duplicates.
func (c *collimatoRepository) GetWorkspaceGroupMembers(ctx context.Context, workspaceID string) ([]model.CollimatoWorkspaceUser, error) {
	rows, err := c.Db.QueryContext(ctx,
		`SELECT `+userColumns("u")+`, cwg.roles
		 FROM collimato_workspace_groups cwg
		 JOIN group_members gm ON gm.group_id = cwg.group_id
		 JOIN user_groups ug ON ug.id = cwg.group_id AND ug.deleted_at = 0
		 JOIN users u ON u.id = gm.user_id
		 WHERE cwg.workspace_id = ? AND cwg.deleted_at = 0
		   AND NOT EXISTS (
		     SELECT 1 FROM collimato_workspace_users cwu
		     WHERE cwu.workspace_id = cwg.workspace_id AND cwu.user_id = gm.user_id
		       AND cwu.deleted_at = 0
		   )`,
		workspaceID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// A user can be in several attached groups; collect their role sets and emit
	// one synthesized member per user.
	byUser := make(map[string]*model.CollimatoWorkspaceUser)
	roleSets := make(map[string]map[string]struct{})
	order := make([]string, 0)
	for rows.Next() {
		u := &model.User{}
		var timezone []byte
		var roles sql.NullString
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.Password,
			&u.AuthService, &u.Username, &u.Name, &u.LastName,
			&u.StorageLimit, &u.Photo, &timezone, &u.MfaActive, &u.MfaSecret,
			&u.CreatedAt, &u.UpdatedAt, &u.DeactivatedAt, &u.AuthData,
			&roles); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(timezone, &u.Timezone); err != nil {
			return nil, err
		}

		if _, ok := byUser[u.ID]; !ok {
			byUser[u.ID] = &model.CollimatoWorkspaceUser{
				WorkspaceID: workspaceID,
				UserID:      u.ID,
				UserInfo:    *u,
				ViaGroup:    true,
			}
			roleSets[u.ID] = make(map[string]struct{})
			order = append(order, u.ID)
		}

		for _, name := range strings.Split(roles.String, ",") {
			if name != "" {
				roleSets[u.ID][name] = struct{}{}
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.CollimatoWorkspaceUser, 0, len(order))
	for _, uid := range order {
		names := make([]string, 0, len(roleSets[uid]))
		for n := range roleSets[uid] {
			names = append(names, n)
		}

		sort.Strings(names)
		byUser[uid].Role = strings.Join(names, ",")
		out = append(out, *byUser[uid])
	}

	return out, nil
}

// GetWorkspaceMemberIDs returns, per workspace, the IDs of everyone who can
// reach it directly or through an attached group, the same set
// GetWorkspaceUsers and GetWorkspaceGroupMembers return together. UNION
// de-duplicates a user who has both paths.
func (c *collimatoRepository) GetWorkspaceMemberIDs(ctx context.Context, workspaceIDs []string) (map[string][]string, error) {
	members := make(map[string][]string)

	if len(workspaceIDs) == 0 {
		return members, nil
	}

	in := sqlPlaceholders(len(workspaceIDs))
	args := make([]any, 0, 2*len(workspaceIDs))
	for range 2 {
		for _, id := range workspaceIDs {
			args = append(args, id)
		}
	}

	rows, err := c.Db.QueryContext(ctx,
		`SELECT cwu.workspace_id, cwu.user_id
		 FROM collimato_workspace_users cwu
		 JOIN users u ON u.id = cwu.user_id
		 WHERE cwu.workspace_id IN (`+in+`) AND cwu.deleted_at = 0
		 UNION
		 SELECT cwg.workspace_id, gm.user_id
		 FROM collimato_workspace_groups cwg
		 JOIN group_members gm ON gm.group_id = cwg.group_id
		 JOIN user_groups ug ON ug.id = cwg.group_id AND ug.deleted_at = 0
		 JOIN users u ON u.id = gm.user_id
		 WHERE cwg.workspace_id IN (`+in+`) AND cwg.deleted_at = 0
		 ORDER BY 1, 2`,
		args...,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var workspaceID, userID string
		if err := rows.Scan(&workspaceID, &userID); err != nil {
			return nil, err
		}

		members[workspaceID] = append(members[workspaceID], userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// GetEffectiveRolesForUser returns the de-duplicated set of workspace role
// names that apply to a user: their direct collimato_workspace_users.role plus
// the roles from every non-deleted group attached to the workspace that the
// user belongs to. This is the single source of truth for "which roles does
// this user have here". Every permission check resolves through it.
func (c *collimatoRepository) GetEffectiveRolesForUser(ctx context.Context, workspaceID, userID string) ([]string, error) {
	roleStrings := make([]string, 0)

	// Direct membership role(s).
	var directRoles sql.NullString
	err := c.Db.QueryRowContext(ctx,
		`SELECT role FROM collimato_workspace_users
		 WHERE workspace_id = ? AND user_id = ? AND deleted_at = 0`,
		workspaceID, userID,
	).Scan(&directRoles)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if directRoles.Valid {
		roleStrings = append(roleStrings, directRoles.String)
	}

	// Roles from attached (non-deleted) groups the user belongs to.
	rows, err := c.Db.QueryContext(ctx,
		`SELECT cwg.roles FROM collimato_workspace_groups cwg
		 JOIN group_members gm ON gm.group_id = cwg.group_id
		 JOIN user_groups ug ON ug.id = cwg.group_id AND ug.deleted_at = 0
		 WHERE cwg.workspace_id = ? AND cwg.deleted_at = 0 AND gm.user_id = ?`,
		workspaceID, userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var r sql.NullString
		if err := rows.Scan(&r); err != nil {
			return nil, err
		}

		if r.Valid {
			roleStrings = append(roleStrings, r.String)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Split each comma-separated role string and de-dupe, preserving order.
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, rs := range roleStrings {
		for _, name := range strings.Split(rs, ",") {
			if name == "" {
				continue
			}

			if _, ok := seen[name]; ok {
				continue
			}

			seen[name] = struct{}{}
			out = append(out, name)
		}
	}

	return out, nil
}
