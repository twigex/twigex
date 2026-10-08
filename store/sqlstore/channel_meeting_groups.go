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

func (c *channelsRepository) AddMeetingGroups(ctx context.Context, meetingID string, groupIDs []string, addedBy string) error {
	if len(groupIDs) == 0 {
		return nil
	}

	now := time.Now().Unix()
	placeholders := make([]string, len(groupIDs))
	args := make([]any, 0, len(groupIDs)*5)
	for i, gid := range groupIDs {
		placeholders[i] = "(?,?,?,?,?)"
		args = append(args, model.NewID(), meetingID, gid, addedBy, now)
	}

	_, err := c.Db.ExecContext(ctx,
		`INSERT INTO channel_meeting_groups (id, meeting_id, group_id, added_by, added_at)
		 VALUES `+strings.Join(placeholders, ",")+`
		 ON DUPLICATE KEY UPDATE id = id`,
		args...,
	)
	return err
}

func (c *channelsRepository) DeleteMeetingGroups(ctx context.Context, meetingID string) error {
	_, err := c.Db.ExecContext(ctx,
		`DELETE FROM channel_meeting_groups WHERE meeting_id = ?`,
		meetingID,
	)
	return err
}

func (c *channelsRepository) GetMeetingGroups(ctx context.Context, meetingID string) ([]model.ChannelMeetingGroup, error) {
	rows, err := c.Db.QueryContext(ctx,
		`SELECT cmg.id, cmg.meeting_id, cmg.group_id, cmg.added_by, cmg.added_at,
		        ug.name, ug.description, `+memberCount("cmg.group_id")+`
		 FROM channel_meeting_groups cmg
		 JOIN user_groups ug ON ug.id = cmg.group_id AND ug.deleted_at = 0
		 WHERE cmg.meeting_id = ?
		 ORDER BY ug.name ASC`,
		meetingID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	out := make([]model.ChannelMeetingGroup, 0)
	for rows.Next() {
		var cmg model.ChannelMeetingGroup
		var description sql.NullString
		var memberCount sql.NullInt64
		if err := rows.Scan(&cmg.ID, &cmg.MeetingID, &cmg.GroupID, &cmg.AddedBy, &cmg.AddedAt,
			&cmg.Name, &description, &memberCount); err != nil {
			return nil, err
		}

		cmg.Description = description.String
		cmg.MemberCount = int(memberCount.Int64)
		out = append(out, cmg)
	}

	return out, rows.Err()
}

func (c *channelsRepository) GetMeetingGroupMemberIDs(ctx context.Context, meetingID string) ([]string, error) {
	rows, err := c.Db.QueryContext(ctx,
		`SELECT DISTINCT gm.user_id
		 FROM channel_meeting_groups cmg
		 JOIN group_members gm ON gm.group_id = cmg.group_id
		 JOIN user_groups ug ON ug.id = cmg.group_id AND ug.deleted_at = 0
		 WHERE cmg.meeting_id = ?`,
		meetingID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (c *channelsRepository) UserInMeetingGroup(ctx context.Context, meetingID, userID string) (bool, error) {
	var count int
	err := c.Db.QueryRowContext(ctx,
		`SELECT COUNT(*)
		 FROM channel_meeting_groups cmg
		 JOIN group_members gm ON gm.group_id = cmg.group_id
		 JOIN user_groups ug ON ug.id = cmg.group_id AND ug.deleted_at = 0
		 WHERE cmg.meeting_id = ? AND gm.user_id = ?`,
		meetingID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
