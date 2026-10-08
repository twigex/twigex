// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/twigex/twigex/model"
)

// AddGroups attaches a batch of groups to a channel in one statement.
// Idempotent: re-attaching a group that's already present is a no-op. The
// UNIQUE KEY unique_channel_group (channel_id, group_id) from migration 76
// drives ON DUPLICATE KEY UPDATE id = id. Returns only an error; callers
// re-read the enriched join rows via GetGroups.
func (c *channelsRepository) AddGroups(ctx context.Context, channelID string, groupIDs []string, addedBy string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	now := time.Now().Unix()
	placeholders := make([]string, len(groupIDs))
	args := make([]any, 0, len(groupIDs)*5)
	for i, gid := range groupIDs {
		placeholders[i] = "(?,?,?,?,?)"
		args = append(args, model.NewID(), channelID, gid, addedBy, now)
	}

	_, err := c.Db.ExecContext(ctx,
		`INSERT INTO channel_groups (id, channel_id, group_id, added_by, added_at)
		 VALUES `+strings.Join(placeholders, ",")+`
		 ON DUPLICATE KEY UPDATE id = id`,
		args...,
	)
	return err
}

func (c *channelsRepository) RemoveGroup(ctx context.Context, channelID, groupID string) error {
	_, err := c.Db.ExecContext(ctx,
		`DELETE FROM channel_groups WHERE channel_id = ? AND group_id = ?`,
		channelID, groupID,
	)
	return err
}

// GetGroups returns the channel_groups join rows for a channel,
// enriched with the linked group's name, description, and current member count
// so callers don't need a second round-trip per group.
func (c *channelsRepository) GetGroups(ctx context.Context, channelID string) ([]model.ChannelGroup, error) {
	rows, err := c.Db.QueryContext(ctx,
		`SELECT cg.id, cg.channel_id, cg.group_id, cg.added_by, cg.added_at,
		        ug.name, ug.description, `+memberCount("cg.group_id")+`
		 FROM channel_groups cg
		 JOIN user_groups ug ON ug.id = cg.group_id AND ug.deleted_at = 0
		 WHERE cg.channel_id = ?
		 ORDER BY ug.name ASC`,
		channelID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]model.ChannelGroup, 0)
	for rows.Next() {
		var cg model.ChannelGroup
		var description sql.NullString
		var memberCount sql.NullInt64
		if err := rows.Scan(&cg.ID, &cg.ChannelID, &cg.GroupID, &cg.AddedBy, &cg.AddedAt,
			&cg.Name, &description, &memberCount); err != nil {
			return nil, err
		}

		cg.Description = description.String
		cg.MemberCount = int(memberCount.Int64)
		out = append(out, cg)
	}

	return out, rows.Err()
}

// MaterializeGroups inserts a materialized channel_members row
// (is_direct=false) for every distinct member of any group in groupIDs who
// does not already have a row in channelID. Soft-deleted groups contribute no
// members. Existing rows (direct or materialized) are left untouched,
// preserving their per-user state.
//
// One SELECT for the absent members across all groups + one multi-row INSERT.
// ON DUPLICATE KEY UPDATE id = id is a deliberate no-op guarding the race where
// a concurrent materialize path inserts the same (channel_id, user_id) between
// this SELECT and INSERT. The UNIQUE constraint from migration 78 is the
// conflict target.
//
// Atomic: the INSERT commits as a single statement. Idempotent: re-running with
// the same input is a no-op.
func (c *channelsRepository) MaterializeGroups(ctx context.Context, channelID string, groupIDs []string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	// 1. Distinct members across all (non-deleted) groups who don't already
	//    have a row in the channel.
	placeholders := make([]string, len(groupIDs))
	selArgs := make([]any, 0, len(groupIDs)+1)
	for i, gid := range groupIDs {
		placeholders[i] = "?"
		selArgs = append(selArgs, gid)
	}

	selArgs = append(selArgs, channelID)

	rows, err := c.Db.QueryContext(ctx,
		`SELECT DISTINCT gm.user_id FROM group_members gm
		 JOIN user_groups ug ON ug.id = gm.group_id AND ug.deleted_at = 0
		 WHERE gm.group_id IN (`+strings.Join(placeholders, ",")+`)
		   AND NOT EXISTS (
		     SELECT 1 FROM channel_members cm
		     WHERE cm.channel_id = ? AND cm.user_id = gm.user_id
		   )`, selArgs...)
	if err != nil {
		return err
	}

	userIDs := make([]string, 0)
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			rows.Close()
			return err
		}

		userIDs = append(userIDs, uid)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(userIDs) == 0 {
		return nil
	}

	// 2. Multi-row INSERT of materialized rows. Each row gets its own
	//    app-generated UUID so id stays unique across the batch. Chunked so a
	//    large group cannot exceed MySQL's 65,535 placeholder cap per prepared
	//    statement (10 placeholders per row).
	emptyProps, _ := json.Marshal(model.ChannelNotifyProps{})
	now := time.Now().UnixMilli()

	const chunkSize = 1000
	for lo := 0; lo < len(userIDs); lo += chunkSize {
		hi := lo + chunkSize
		if hi > len(userIDs) {
			hi = len(userIDs)
		}

		valuePlaceholders := make([]string, 0, hi-lo)
		args := make([]any, 0, (hi-lo)*10)
		for _, uid := range userIDs[lo:hi] {
			valuePlaceholders = append(valuePlaceholders, "(?,?,?,?,?,?,?,?,?,?, FALSE)")
			args = append(args,
				model.NewID(), channelID, uid, model.ChannelRoleMember, emptyProps,
				0, 0, now, now, now,
			)
		}

		query := `INSERT INTO channel_members
		          (id, channel_id, user_id, role, notify_props, msg_count, mention_count,
		           last_viewed_at, updated_at, date_joined, is_direct)
		          VALUES ` + strings.Join(valuePlaceholders, ",") + `
		          ON DUPLICATE KEY UPDATE id = id`

		if _, err := c.Db.ExecContext(ctx, query, args...); err != nil {
			return err
		}
	}

	return nil
}

// DematerializeGroup deletes the materialized rows (is_direct=false)
// for users whose only access to channelID was via groupID, i.e. users that
// have no remaining anchored path. Direct rows (is_direct=true) and users
// reachable via another attached group are left intact.
//
// Must run AFTER the channel_groups row is deleted, since the "any other path"
// check looks at channel_groups (excluding the now-removed row).
func (c *channelsRepository) DematerializeGroup(ctx context.Context, channelID, groupID string) error {
	_, err := c.Db.ExecContext(ctx,
		`DELETE cm FROM channel_members cm
		 WHERE cm.channel_id = ?
		   AND cm.is_direct = FALSE
		   AND cm.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		   AND NOT EXISTS (
		     SELECT 1 FROM channel_groups cg
		     JOIN group_members gm ON gm.group_id = cg.group_id
		     WHERE cg.channel_id = cm.channel_id AND gm.user_id = cm.user_id
		   )`,
		channelID, groupID)
	return err
}

// MaterializeUserInGroupChannels handles "users joined group X": for every
// channel that has X attached, insert a materialized channel_members row for
// each user who doesn't already have one. Existing rows (direct or
// materialized) are left untouched.
//
// One SELECT for the attached channels + one multi-row INSERT for the full
// (channel × user) cross product. The INSERT uses ON DUPLICATE KEY UPDATE as
// a no-op (id = id) so concurrent inserts via another materialize path don't
// surface the UNIQUE constraint as a 1062 error.
//
// Atomic: the INSERT commits as a single statement. Idempotent: re-running
// with the same input is a no-op.
func (c *channelsRepository) MaterializeUserInGroupChannels(ctx context.Context, userIDs []string, groupID string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// 1. Channels the (non-deleted) group is currently attached to.
	rows, err := c.Db.QueryContext(ctx,
		`SELECT cg.channel_id FROM channel_groups cg
		 JOIN user_groups ug ON ug.id = cg.group_id AND ug.deleted_at = 0
		 WHERE cg.group_id = ?`, groupID)
	if err != nil {
		return err
	}

	channelIDs := make([]string, 0)
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			rows.Close()
			return err
		}

		channelIDs = append(channelIDs, cid)
	}

	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(channelIDs) == 0 {
		return nil
	}

	// 2. Build the (channel × user) cross product as multi-row VALUES. Each
	//    row gets its own app-generated UUID so id stays unique across the
	//    batch.
	emptyProps, _ := json.Marshal(model.ChannelNotifyProps{})
	now := time.Now().UnixMilli()

	placeholders := make([]string, 0, len(channelIDs)*len(userIDs))
	args := make([]any, 0, len(channelIDs)*len(userIDs)*10)
	for _, cid := range channelIDs {
		for _, uid := range userIDs {
			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?, FALSE)")
			args = append(args,
				model.NewID(), cid, uid, model.ChannelRoleMember, emptyProps,
				0, 0, now, now, now,
			)
		}
	}

	// 3. ON DUPLICATE KEY UPDATE id = id is a deliberate no-op: if a row
	//    already exists for (channel_id, user_id), don't touch it. The
	//    UNIQUE constraint from migration 78 is the conflict target.
	query := `INSERT INTO channel_members
	          (id, channel_id, user_id, role, notify_props, msg_count, mention_count,
	           last_viewed_at, updated_at, date_joined, is_direct)
	          VALUES ` + strings.Join(placeholders, ",") + `
	          ON DUPLICATE KEY UPDATE id = id`

	_, err = c.Db.ExecContext(ctx, query, args...)
	return err
}

// RemoveGroupFromAllChannels cascades a group deletion across every channel
// the group was attached to. Direct rows survive (the user may have been
// promoted to direct). Materialized rows whose only path was this group get
// deleted; rows still reachable via another attached group stay.
func (c *channelsRepository) RemoveGroupFromAllChannels(ctx context.Context, groupID string) error {
	tx, err := c.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	// Delete materialized rows that lose their last anchor when this group goes
	// away. The NOT EXISTS clause excludes the group being removed from the
	// "other path" search.
	if _, err := tx.ExecContext(ctx,
		`DELETE cm FROM channel_members cm
		 JOIN channel_groups cg_self ON cg_self.channel_id = cm.channel_id AND cg_self.group_id = ?
		 WHERE cm.is_direct = FALSE
		   AND cm.user_id IN (SELECT user_id FROM group_members WHERE group_id = ?)
		   AND NOT EXISTS (
		     SELECT 1 FROM channel_groups cg
		     JOIN group_members gm ON gm.group_id = cg.group_id
		     WHERE cg.channel_id = cm.channel_id
		       AND cg.group_id != ?
		       AND gm.user_id = cm.user_id
		   )`, groupID, groupID, groupID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM channel_groups WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	return tx.Commit()
}

// DematerializeUserFromGroupChannels handles "user left group X": for every
// channel that has X attached, delete the user's materialized row IFF they
// have no remaining anchored path. Direct rows are untouched.
//
// Run BEFORE the group_members row is removed so the "any other path" check
// excludes the leaving group naturally, or run after, and let the existing
// channel_groups + group_members JOIN do the work. Implementation here checks
// the current state of channel_groups/group_members, so caller must remove the
// group_members row first.
func (c *channelsRepository) DematerializeUserFromGroupChannels(ctx context.Context, userID, groupID string) error {
	_, err := c.Db.ExecContext(ctx,
		`DELETE cm FROM channel_members cm
		 WHERE cm.user_id = ?
		   AND cm.is_direct = FALSE
		   AND cm.channel_id IN (SELECT channel_id FROM channel_groups WHERE group_id = ?)
		   AND NOT EXISTS (
		     SELECT 1 FROM channel_groups cg
		     JOIN group_members gm ON gm.group_id = cg.group_id
		     WHERE cg.channel_id = cm.channel_id AND gm.user_id = ?
		   )`,
		userID, groupID, userID)
	return err
}
