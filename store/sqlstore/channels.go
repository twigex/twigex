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

type channelsRepository struct {
	Db *sql.DB
}

const (
	channelColumns = `id, type, display_name, name, header, description, last_post, msg_count, created_by, created_at, updated_at, deleted_at`

	channelMeetingColumns = `id, channel_id, host_id, created_by, title, guest_password, duration_minutes, sequence, timezone, invite_only, scheduled_at, started_at, ended_at, created_at`
)

func NewChannelsRepository(Db *sql.DB) (*channelsRepository, error) {
	repo := &channelsRepository{}

	repo.Db = Db
	return repo, nil
}

func (c *channelsRepository) Create(ch *model.Channel) (*model.Channel, error) {
	_, err := c.Db.Exec(
		`INSERT INTO channels (`+channelColumns+`) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		ch.ID, ch.Type, ch.DisplayName, ch.Name, ch.Header, ch.Description,
		ch.LastPost, ch.MessageCount, ch.CreatedBy, ch.CreatedAt, ch.UpdatedAt, ch.DeletedAt)
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (c *channelsRepository) Update(ch *model.Channel) error {
	t := time.Now().Unix()

	_, err := c.Db.Exec(`UPDATE channels SET display_name=?, name=?, header=?, description=?, updated_at=? WHERE id=?`,
		ch.DisplayName, ch.Name, ch.Header, ch.Description, t, ch.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *channelsRepository) GetAll() ([]model.Channel, error) {
	results, err := c.Db.Query("SELECT "+channelColumns+" FROM channels WHERE type!=?", model.ChannelTypeDirect)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	channels := make([]model.Channel, 0)

	for results.Next() {
		channel := model.Channel{}

		err := results.Scan(&channel.ID, &channel.Type, &channel.DisplayName, &channel.Name,
			&channel.Header, &channel.Description, &channel.LastPost, &channel.MessageCount,
			&channel.CreatedBy, &channel.CreatedAt, &channel.UpdatedAt, &channel.DeletedAt)
		if err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	if err := c.attachMembers(channels, "cm.is_direct = TRUE"); err != nil {
		return nil, err
	}

	return channels, nil
}

func (c *channelsRepository) GetAllByType(channelType string) ([]model.Channel, error) {
	results, err := c.Db.Query("SELECT "+channelColumns+" FROM channels WHERE type=? AND deleted_at=0", channelType)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	channels := make([]model.Channel, 0)

	for results.Next() {
		channel := model.Channel{}

		err := results.Scan(&channel.ID, &channel.Type, &channel.DisplayName, &channel.Name,
			&channel.Header, &channel.Description, &channel.LastPost, &channel.MessageCount,
			&channel.CreatedBy, &channel.CreatedAt, &channel.UpdatedAt, &channel.DeletedAt)
		if err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	if err := c.attachMembers(channels, "cm.is_direct = TRUE"); err != nil {
		return nil, err
	}

	return channels, nil
}

func (c *channelsRepository) GetAllForUser(userID string) ([]model.Channel, error) {
	results, err := c.Db.Query(`SELECT channels.*
								FROM channels, channel_members
								WHERE channel_members.user_id=?
								AND channels.id=channel_members.channel_id
								AND channels.deleted_at=0`, userID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	channels := make([]model.Channel, 0)

	for results.Next() {
		channel := model.Channel{}

		err := results.Scan(&channel.ID, &channel.Type, &channel.DisplayName, &channel.Name,
			&channel.Header, &channel.Description, &channel.LastPost, &channel.MessageCount,
			&channel.CreatedBy, &channel.CreatedAt, &channel.UpdatedAt, &channel.DeletedAt)
		if err != nil {
			return nil, err
		}

		channels = append(channels, channel)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	if err := c.attachMembers(channels, "(cm.is_direct = TRUE OR cm.user_id = ?)", userID); err != nil {
		return nil, err
	}

	return channels, nil
}

// attachMembers loads the members of every channel in one query, keeping the
// rows that match filter, instead of one query per channel.
func (c *channelsRepository) attachMembers(channels []model.Channel, filter string, filterArgs ...any) error {
	if len(channels) == 0 {
		return nil
	}

	args := make([]any, 0, len(channels)+len(filterArgs))
	for _, ch := range channels {
		args = append(args, ch.ID)
	}

	args = append(args, filterArgs...)

	results, err := c.Db.Query(`SELECT cm.id, cm.channel_id, cm.user_id, cm.role, cm.notify_props,
		                               cm.msg_count, cm.mention_count, cm.last_viewed_at,
		                               cm.updated_at, cm.date_joined, cm.is_direct,
		                               u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
		                        FROM channel_members cm
		                        JOIN users u ON u.id = cm.user_id
		                        WHERE cm.channel_id IN (`+sqlPlaceholders(len(channels))+`) AND `+filter, args...)
	if err != nil {
		return err
	}

	defer results.Close()

	byChannel := make(map[string][]model.ChannelMember, len(channels))
	for results.Next() {
		member := model.ChannelMember{}
		user := model.User{}
		var props []byte
		err := results.Scan(&member.ID, &member.ChannelID, &member.UserID, &member.Role, &props,
			&member.MsgCount, &member.MentionCount, &member.LastViewedAt, &member.UpdatedAt,
			&member.DateJoined, &member.IsDirect,
			&user.ID, &user.Email, &user.Username, &user.Name, &user.LastName, &user.Photo)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(props, &member.NotifyProps); err != nil {
			return err
		}

		member.UserInfo = &user
		byChannel[member.ChannelID] = append(byChannel[member.ChannelID], member)
	}

	if err := results.Err(); err != nil {
		return err
	}

	for i := range channels {
		members := byChannel[channels[i].ID]
		if members == nil {
			members = make([]model.ChannelMember, 0)
		}

		channels[i].ChannelMembers = members
	}

	return nil
}

func (c *channelsRepository) GetDirectMessage(userID, userID2 string) (*model.Channel, error) {
	name1 := userID + "__" + userID2
	name2 := userID2 + "__" + userID

	row := c.Db.QueryRow("SELECT "+channelColumns+" FROM channels WHERE name = ? OR name = ?", name1, name2)

	channel := model.Channel{}
	err := row.Scan(&channel.ID, &channel.Type, &channel.DisplayName, &channel.Name,
		&channel.Header, &channel.Description, &channel.LastPost, &channel.MessageCount,
		&channel.CreatedBy, &channel.CreatedAt, &channel.UpdatedAt, &channel.DeletedAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &channel, nil
}

// archivedAt is seconds, matching created_at and updated_at on this table.
func (c *channelsRepository) UpdateArchived(ctx context.Context, channelID string, archivedAt int64) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channels SET deleted_at=?, updated_at=? WHERE id=?",
		archivedAt, time.Now().Unix(), channelID)

	return err
}

func (c *channelsRepository) Get(channelID string) (*model.Channel, error) {
	row := c.Db.QueryRow("SELECT "+channelColumns+" FROM channels WHERE id=?", channelID)

	channel := model.Channel{}
	err := row.Scan(&channel.ID, &channel.Type, &channel.DisplayName, &channel.Name,
		&channel.Header, &channel.Description, &channel.LastPost, &channel.MessageCount,
		&channel.CreatedBy, &channel.CreatedAt, &channel.UpdatedAt, &channel.DeletedAt)
	if err != nil {
		return nil, err
	}

	members, err := c.getDirectMembers(channel.ID)
	if err != nil {
		return nil, err
	}

	channel.ChannelMembers = members

	return &channel, nil
}

// CreateOrPromoteMembers inserts a batch of direct channel_members rows
// in a single statement. For each row, if a row already exists for the same
// (channel_id, user_id), whether it was direct or materialized via a group,
// the existing row is flipped to is_direct=TRUE with updated_at bumped, while
// preserving all other per-user state (notify_props, msg_count, mention_count,
// last_viewed_at, date_joined). The state preservation is encoded in the
// statement itself: the UPDATE branch only touches is_direct and updated_at.
//
// Atomic: either every row is inserted/promoted or none are. Idempotent:
// rerunning with the same input is a no-op aside from updated_at being bumped.
//
// Requires UNIQUE KEY unique_channel_member (channel_id, user_id) on
// channel_members (migration 78). The ON DUPLICATE KEY UPDATE conflict
// target is this constraint.
func (c *channelsRepository) CreateOrPromoteMembers(ctx context.Context, members []model.ChannelMember) error {
	if len(members) == 0 {
		return nil
	}

	placeholders := make([]string, len(members))
	args := make([]any, 0, len(members)*10)
	for i, m := range members {
		props, err := json.Marshal(m.NotifyProps)
		if err != nil {
			return err
		}

		placeholders[i] = "(?,?,?,?,?,?,?,?,?,?, TRUE)"
		args = append(args,
			m.ID, m.ChannelID, m.UserID, m.Role, props,
			m.MsgCount, m.MentionCount, m.LastViewedAt, m.UpdatedAt, m.DateJoined,
		)
	}

	query := `INSERT INTO channel_members
	          (id, channel_id, user_id, role, notify_props, msg_count, mention_count,
	           last_viewed_at, updated_at, date_joined, is_direct)
	          VALUES ` + strings.Join(placeholders, ",") + `
	          ON DUPLICATE KEY UPDATE is_direct = TRUE, updated_at = VALUES(updated_at)`

	_, err := c.Db.ExecContext(ctx, query, args...)
	return err
}

// RemoveUser removes a user's direct membership. If the user still
// has access to the channel via an attached group, their row is demoted to a
// materialized (is_direct=false) row so their per-user state survives. If
// they have no remaining group path, the row is deleted.
func (c *channelsRepository) RemoveUser(channelID, userID string) error {
	tx, err := c.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var hasGroupPath bool
	err = tx.QueryRow(
		`SELECT EXISTS(
		   SELECT 1 FROM channel_groups cg
		   JOIN group_members gm ON gm.group_id = cg.group_id
		   WHERE cg.channel_id = ? AND gm.user_id = ?
		 )`, channelID, userID).Scan(&hasGroupPath)
	if err != nil {
		return err
	}

	if hasGroupPath {
		// The role was granted to the direct membership being removed, so it
		// goes with it. Left behind it is invisible: the member list shows
		// direct rows only, but IsChannelAdmin reads every row.
		_, err = tx.Exec(
			`UPDATE channel_members SET is_direct = FALSE, role = ?, updated_at = ?
			 WHERE channel_id = ? AND user_id = ?`,
			model.ChannelRoleMember, time.Now().UnixMilli(), channelID, userID)
	} else {
		_, err = tx.Exec(
			`DELETE FROM channel_members WHERE channel_id = ? AND user_id = ?`,
			channelID, userID)
	}

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *channelsRepository) UpdateUserRole(channelID, userID, role string) error {
	t := time.Now().UnixMilli()

	tx, err := c.Db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec("UPDATE channel_members SET role=?, updated_at=? WHERE channel_id=? AND user_id=?", role, t, channelID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (c *channelsRepository) IsMember(channelID, userID string) (bool, error) {
	var id string
	err := c.Db.QueryRow("SELECT user_id FROM channel_members WHERE channel_id=? AND user_id=?", channelID, userID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// GetMembers returns every channel_members row for channelID. Rows
// materialized via a group (is_direct = false) are kept in sync by cascade
// handlers in the app layer, so the table is the canonical truth.
func (c *channelsRepository) GetMembers(channelID string) ([]model.ChannelMember, error) {
	results, err := c.Db.Query(`SELECT id, channel_id, user_id, role, notify_props,
		                                 msg_count, mention_count, last_viewed_at,
		                                 updated_at, date_joined, is_direct
		                          FROM channel_members WHERE channel_id = ?`, channelID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	members := make([]model.ChannelMember, 0)
	for results.Next() {
		member := model.ChannelMember{}
		var props []byte
		err := results.Scan(&member.ID, &member.ChannelID, &member.UserID, &member.Role, &props, &member.MsgCount,
			&member.MentionCount, &member.LastViewedAt, &member.UpdatedAt, &member.DateJoined, &member.IsDirect)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(props, &member.NotifyProps)
		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// SearchMembers returns members of a channel whose name, lastname,
// username or email matches query, capped at limit. It reads channel_members
// (which materialization keeps complete for both directly-added and group
// members), so group members are included while the result stays bounded
// regardless of channel size.
func (c *channelsRepository) SearchMembers(ctx context.Context, channelID, query string, limit int) ([]model.User, error) {
	q := `SELECT u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
	      FROM channel_members cm
	      JOIN users u ON u.id = cm.user_id
	      WHERE cm.channel_id = ?
	        AND u.deactivated_at = 0`
	args := []any{channelID}

	if query != "" {
		like := "%" + escapeLike(query) + "%"
		q += ` AND (u.name LIKE ? OR u.lastname LIKE ? OR u.username LIKE ? OR u.email LIKE ?)`
		args = append(args, like, like, like, like)
	}

	q += ` ORDER BY u.name ASC, u.lastname ASC LIMIT ?`
	args = append(args, limit)

	rows, err := c.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, limit)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetMembersByUsernames resolves legacy @name mention handles to
// channel members, matching usernames exactly (unlike SearchMembers).
// Reads channel_members so members reachable via a group are included.
func (c *channelsRepository) GetMembersByUsernames(ctx context.Context, channelID string, usernames []string) ([]model.User, error) {
	return c.findChannelMembersBy(ctx, channelID, "u.username", usernames)
}

// GetMembersByIDs applies the same filters as the username form, so
// both mention forms notify identically.
func (c *channelsRepository) GetMembersByIDs(ctx context.Context, channelID string, ids []string) ([]model.User, error) {
	return c.findChannelMembersBy(ctx, channelID, "u.id", ids)
}

// column is a caller-supplied literal, never user input.
func (c *channelsRepository) findChannelMembersBy(ctx context.Context, channelID, column string, values []string) ([]model.User, error) {
	if len(values) == 0 {
		users := make([]model.User, 0)
		return users, nil
	}

	args := make([]any, 0, len(values)+1)
	args = append(args, channelID)
	for _, v := range values {
		args = append(args, v)
	}

	q := `SELECT u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
	      FROM channel_members cm
	      JOIN users u ON u.id = cm.user_id
	      WHERE cm.channel_id = ?
	        AND u.deactivated_at = 0
	        AND ` + column + ` IN (?` + strings.Repeat(",?", len(values)-1) + `)`

	rows, err := c.Db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := make([]model.User, 0, len(values))
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.LastName, &u.Photo); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// IncrementMentionCount raises the unread-mention badge for specific members,
// used for targeted @user mentions.
func (c *channelsRepository) IncrementMentionCount(ctx context.Context, channelID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	args := make([]any, 0, len(userIDs)+1)
	args = append(args, channelID)
	for _, id := range userIDs {
		args = append(args, id)
	}

	q := `UPDATE channel_members SET mention_count = mention_count + 1
	      WHERE channel_id = ?
	        AND user_id IN (?` + strings.Repeat(",?", len(userIDs)-1) + `)`

	_, err := c.Db.ExecContext(ctx, q, args...)
	return err
}

// IncrementAllMentionCounts raises the badge for every member except the
// author in one statement, so @all/@here stays cheap on large channels.
func (c *channelsRepository) IncrementAllMentionCounts(ctx context.Context, channelID, excludeUserID string) error {
	_, err := c.Db.ExecContext(ctx,
		`UPDATE channel_members SET mention_count = mention_count + 1
		 WHERE channel_id = ? AND user_id != ?`,
		channelID, excludeUserID)
	return err
}

// getDirectMembers returns only the directly-added members of a channel
// (is_direct = TRUE), each hydrated with the member's user details. Members
// materialized via an attached group are intentionally excluded so the payload
// stays bounded even when a large group is attached; those members are
// represented by the channel's groups instead.
func (c *channelsRepository) getDirectMembers(channelID string) ([]model.ChannelMember, error) {
	results, err := c.Db.Query(`SELECT cm.id, cm.channel_id, cm.user_id, cm.role, cm.notify_props,
		                               cm.msg_count, cm.mention_count, cm.last_viewed_at,
		                               cm.updated_at, cm.date_joined, cm.is_direct,
		                               u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
		                        FROM channel_members cm
		                        JOIN users u ON u.id = cm.user_id
		                        WHERE cm.channel_id = ? AND cm.is_direct = TRUE`, channelID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	members := make([]model.ChannelMember, 0)
	for results.Next() {
		member := model.ChannelMember{}
		user := model.User{}
		var props []byte
		err := results.Scan(&member.ID, &member.ChannelID, &member.UserID, &member.Role, &props,
			&member.MsgCount, &member.MentionCount, &member.LastViewedAt, &member.UpdatedAt,
			&member.DateJoined, &member.IsDirect,
			&user.ID, &user.Email, &user.Username, &user.Name, &user.LastName, &user.Photo)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(props, &member.NotifyProps); err != nil {
			return nil, err
		}

		member.UserInfo = &user
		members = append(members, member)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

// GetDirectMembersForUser is like getDirectMembers but always
// includes the given user's own row even when they are in the channel only via
// a group (is_direct = FALSE). The frontend relies on the current user's
// membership row being present for per-user state (unread counts, last viewed,
// mentions); the user is still filtered out of the displayed members list by
// is_direct on the client.
func (c *channelsRepository) GetDirectMembersForUser(channelID, userID string) ([]model.ChannelMember, error) {
	results, err := c.Db.Query(`SELECT cm.id, cm.channel_id, cm.user_id, cm.role, cm.notify_props,
		                               cm.msg_count, cm.mention_count, cm.last_viewed_at,
		                               cm.updated_at, cm.date_joined, cm.is_direct,
		                               u.id, u.email, u.username, u.name, u.lastname, COALESCE((SELECT photo_id FROM user_photos WHERE user_id = u.id), '')
		                        FROM channel_members cm
		                        JOIN users u ON u.id = cm.user_id
		                        WHERE cm.channel_id = ? AND (cm.is_direct = TRUE OR cm.user_id = ?)`, channelID, userID)
	if err != nil {
		return nil, err
	}

	defer results.Close()

	members := make([]model.ChannelMember, 0)
	for results.Next() {
		member := model.ChannelMember{}
		user := model.User{}
		var props []byte
		err := results.Scan(&member.ID, &member.ChannelID, &member.UserID, &member.Role, &props,
			&member.MsgCount, &member.MentionCount, &member.LastViewedAt, &member.UpdatedAt,
			&member.DateJoined, &member.IsDirect,
			&user.ID, &user.Email, &user.Username, &user.Name, &user.LastName, &user.Photo)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(props, &member.NotifyProps); err != nil {
			return nil, err
		}

		member.UserInfo = &user
		members = append(members, member)
	}

	if err := results.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func scanChannelMeeting(s interface{ Scan(...any) error }) (*model.ChannelMeeting, error) {
	var m model.ChannelMeeting
	var guestPassword sql.NullString
	if err := s.Scan(&m.ID, &m.ChannelID, &m.HostID, &m.CreatedBy, &m.Title,
		&guestPassword, &m.DurationMinutes, &m.Sequence, &m.Timezone, &m.InviteOnly, &m.ScheduledAt, &m.StartedAt, &m.EndedAt, &m.CreatedAt); err != nil {
		return nil, err
	}

	m.GuestPassword = guestPassword.String
	m.SetStatus()
	return &m, nil
}

func (c *channelsRepository) CreateMeeting(ctx context.Context, meeting model.ChannelMeeting) error {
	tx, err := c.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	var guestPassword any
	if meeting.GuestPassword != "" {
		guestPassword = meeting.GuestPassword
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO channel_meetings
		(id, channel_id, host_id, created_by, title, guest_password, duration_minutes, sequence, timezone, invite_only, scheduled_at, started_at, ended_at, created_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		meeting.ID, meeting.ChannelID, meeting.HostID, meeting.CreatedBy, meeting.Title,
		guestPassword, meeting.DurationMinutes, meeting.Sequence, meeting.Timezone, meeting.InviteOnly, meeting.ScheduledAt, meeting.StartedAt, meeting.EndedAt, meeting.CreatedAt); err != nil {
		return err
	}

	for _, userID := range meeting.Invitees {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO channel_meeting_invitees (meeting_id, user_id) VALUES (?, ?)",
			meeting.ID, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Idempotent on the (meeting_id, user_id) PK.
func (c *channelsRepository) AddMeetingInvitees(ctx context.Context, meetingID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	args := make([]any, 0, len(userIDs)*2)
	for _, userID := range userIDs {
		args = append(args, meetingID, userID)
	}

	_, err := c.Db.ExecContext(ctx,
		`INSERT IGNORE INTO channel_meeting_invitees (meeting_id, user_id)
		 VALUES `+rowPlaceholders(len(userIDs), 2), args...)

	return err
}

func (c *channelsRepository) getMeetingInvitees(ctx context.Context, meetingIDs []string) (map[string][]string, error) {
	result := map[string][]string{}
	if len(meetingIDs) == 0 {
		return result, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(meetingIDs)), ",")
	args := make([]any, 0, len(meetingIDs))
	for _, id := range meetingIDs {
		args = append(args, id)
	}

	rows, err := c.Db.QueryContext(ctx,
		"SELECT meeting_id, user_id FROM channel_meeting_invitees WHERE meeting_id IN ("+placeholders+")",
		args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var meetingID, userID string
		if err := rows.Scan(&meetingID, &userID); err != nil {
			return nil, err
		}

		result[meetingID] = append(result[meetingID], userID)
	}

	return result, rows.Err()
}

func (c *channelsRepository) GetMeeting(ctx context.Context, id string) (*model.ChannelMeeting, error) {
	row := c.Db.QueryRowContext(ctx, `SELECT `+channelMeetingColumns+` FROM channel_meetings WHERE id = ?`, id)
	m, err := scanChannelMeeting(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	invitees, err := c.getMeetingInvitees(ctx, []string{id})
	if err != nil {
		return nil, err
	}

	m.Invitees = invitees[id]
	m.SetStatus()

	return m, nil
}

func (c *channelsRepository) getChannelMeetingsWhere(ctx context.Context, where string, args ...any) ([]model.ChannelMeeting, error) {
	return c.getChannelMeetingsWhereOrdered(ctx, where, "created_at DESC", args...)
}

func (c *channelsRepository) getChannelMeetingsWhereOrdered(ctx context.Context, where, orderBy string, args ...any) ([]model.ChannelMeeting, error) {
	rows, err := c.Db.QueryContext(ctx, `SELECT `+channelMeetingColumns+` FROM channel_meetings WHERE `+where+` ORDER BY `+orderBy, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	meetings := []model.ChannelMeeting{}
	ids := []string{}
	for rows.Next() {
		m, err := scanChannelMeeting(rows)
		if err != nil {
			return nil, err
		}

		meetings = append(meetings, *m)
		ids = append(ids, m.ID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	invitees, err := c.getMeetingInvitees(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range meetings {
		meetings[i].Invitees = invitees[meetings[i].ID]
		meetings[i].SetStatus()
	}

	return meetings, nil
}

func (c *channelsRepository) GetActiveMeetings(ctx context.Context, channelID string) ([]model.ChannelMeeting, error) {
	return c.getChannelMeetingsWhere(ctx, "channel_id = ? AND started_at IS NOT NULL AND ended_at IS NULL", channelID)
}

// An invitation alone grants visibility; channel membership is not otherwise required.
func (c *channelsRepository) GetMeetingsForUser(ctx context.Context, userID string) ([]model.ChannelMeeting, error) {
	return c.getChannelMeetingsWhereOrdered(ctx,
		`ended_at IS NULL
		AND channel_id NOT IN (SELECT id FROM channels WHERE type = 'D' OR deleted_at <> 0)
		AND (
			host_id = ?
			OR id IN (SELECT meeting_id FROM channel_meeting_invitees WHERE user_id = ?)
			OR id IN (
				SELECT cmg.meeting_id FROM channel_meeting_groups cmg
				JOIN group_members gm ON gm.group_id = cmg.group_id
				JOIN user_groups ug ON ug.id = cmg.group_id AND ug.deleted_at = 0
				WHERE gm.user_id = ?
			)
			OR (
				NOT EXISTS (SELECT 1 FROM channel_meeting_invitees i WHERE i.meeting_id = channel_meetings.id)
				AND NOT EXISTS (SELECT 1 FROM channel_meeting_groups g WHERE g.meeting_id = channel_meetings.id)
				AND channel_id IN (SELECT channel_id FROM channel_members WHERE user_id = ?)
			)
		)`,
		"COALESCE(scheduled_at, started_at, created_at) ASC",
		userID, userID, userID, userID)
}

func (c *channelsRepository) GetScheduledMeetings(ctx context.Context, channelID string) ([]model.ChannelMeeting, error) {
	return c.getChannelMeetingsWhere(ctx, "channel_id = ? AND started_at IS NULL AND ended_at IS NULL", channelID)
}

func (c *channelsRepository) StartMeeting(ctx context.Context, id string, startedAt int64) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_meetings SET started_at = ? WHERE id = ? AND started_at IS NULL", startedAt, id)
	return err
}

func (c *channelsRepository) EndMeeting(ctx context.Context, id string, endedAt int64) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_meetings SET ended_at = ? WHERE id = ? AND ended_at IS NULL", endedAt, id)
	return err
}

func (c *channelsRepository) UpdateMeetingHost(ctx context.Context, id, hostID string) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_meetings SET host_id = ? WHERE id = ?", hostID, id)
	return err
}

func (c *channelsRepository) UpdateMeetingInviteOnly(ctx context.Context, id string, inviteOnly bool) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_meetings SET invite_only = ? WHERE id = ?", inviteOnly, id)
	return err
}

// Bumps the iCalendar sequence atomically with the edit so the resent ICS can carry it.
func (c *channelsRepository) UpdateMeeting(ctx context.Context, meeting model.ChannelMeeting) (int, error) {
	tx, err := c.Db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	var guestPassword any
	if meeting.GuestPassword != "" {
		guestPassword = meeting.GuestPassword
	}

	if _, err := tx.ExecContext(ctx, `UPDATE channel_meetings
		SET title = ?, guest_password = ?, duration_minutes = ?, timezone = ?, scheduled_at = ?, sequence = sequence + 1
		WHERE id = ?`,
		meeting.Title, guestPassword, meeting.DurationMinutes, meeting.Timezone, meeting.ScheduledAt, meeting.ID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM channel_meeting_invitees WHERE meeting_id = ?", meeting.ID); err != nil {
		return 0, err
	}

	for _, userID := range meeting.Invitees {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO channel_meeting_invitees (meeting_id, user_id) VALUES (?, ?)",
			meeting.ID, userID); err != nil {
			return 0, err
		}
	}

	var sequence int
	if err := tx.QueryRowContext(ctx, "SELECT sequence FROM channel_meetings WHERE id = ?", meeting.ID).Scan(&sequence); err != nil {
		return 0, err
	}

	return sequence, tx.Commit()
}

// Bumping the iCalendar sequence lets a cancellation ICS supersede the last invite the client saw.
func (c *channelsRepository) IncrementMeetingSequence(ctx context.Context, id string) (int, error) {
	if _, err := c.Db.ExecContext(ctx, "UPDATE channel_meetings SET sequence = sequence + 1 WHERE id = ?", id); err != nil {
		return 0, err
	}

	var sequence int
	if err := c.Db.QueryRowContext(ctx, "SELECT sequence FROM channel_meetings WHERE id = ?", id).Scan(&sequence); err != nil {
		return 0, err
	}

	return sequence, nil
}

func (c *channelsRepository) CreateVideoGuestLink(ctx context.Context, links []model.ChannelGuestLink) error {
	tx, err := c.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, link := range links {
		_, err := tx.ExecContext(ctx, `INSERT INTO channel_guest_links
			(id, channel_id, meeting_id, created_by, invited_email, expires_at, used_at, revoked_at, created_at)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			link.ID, link.ChannelID, link.MeetingID, link.CreatedBy, link.InvitedEmail,
			link.ExpiresAt, link.UsedAt, link.RevokedAt, link.CreatedAt)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (c *channelsRepository) GetVideoGuestLink(ctx context.Context, id string) (*model.ChannelGuestLink, error) {
	var link model.ChannelGuestLink

	err := c.Db.QueryRowContext(ctx, `SELECT id, channel_id, meeting_id, created_by, invited_email, expires_at, used_at, revoked_at, created_at
		FROM channel_guest_links WHERE id = ?`, id).Scan(
		&link.ID, &link.ChannelID, &link.MeetingID, &link.CreatedBy, &link.InvitedEmail,
		&link.ExpiresAt, &link.UsedAt, &link.RevokedAt, &link.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (c *channelsRepository) TouchVideoGuestLink(ctx context.Context, id string, usedAt int64) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_guest_links SET used_at = ? WHERE id = ?", usedAt, id)
	return err
}

func (c *channelsRepository) RevokeVideoGuestLink(ctx context.Context, id string, revokedAt int64) error {
	_, err := c.Db.ExecContext(ctx, "UPDATE channel_guest_links SET revoked_at = ? WHERE id = ?", revokedAt, id)
	return err
}

func (c *channelsRepository) RevokeMeetingGuestLinks(ctx context.Context, meetingID string, revokedAt int64) error {
	_, err := c.Db.ExecContext(ctx,
		"UPDATE channel_guest_links SET revoked_at = ? WHERE meeting_id = ? AND revoked_at IS NULL",
		revokedAt, meetingID)
	return err
}

func (c *channelsRepository) GetActiveMeetingGuestLinks(ctx context.Context, meetingID string, now int64) ([]model.ChannelGuestLink, error) {
	rows, err := c.Db.QueryContext(ctx, `SELECT id, channel_id, meeting_id, created_by, invited_email, expires_at, used_at, revoked_at, created_at
		FROM channel_guest_links
		WHERE meeting_id = ? AND revoked_at IS NULL AND expires_at > ?
		ORDER BY created_at DESC`, meetingID, now)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	links := []model.ChannelGuestLink{}
	for rows.Next() {
		var l model.ChannelGuestLink
		if err := rows.Scan(&l.ID, &l.ChannelID, &l.MeetingID, &l.CreatedBy, &l.InvitedEmail,
			&l.ExpiresAt, &l.UsedAt, &l.RevokedAt, &l.CreatedAt); err != nil {
			return nil, err
		}

		links = append(links, l)
	}

	return links, rows.Err()
}

// The channel row goes last so a failure part way through leaves it in place
// and the job can be run again.
func (c *channelsRepository) DeletePermanently(ctx context.Context, channelID string) error {
	tx, err := c.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, table := range []string{"channel_members", "channel_groups"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE channel_id=?", channelID); err != nil {
			return err
		}
	}

	// channel_guest_links and channel_meetings cascade from this row, and
	// channel_meeting_invitees cascades from the meetings.
	if _, err := tx.ExecContext(ctx, "DELETE FROM channels WHERE id=?", channelID); err != nil {
		return err
	}

	return tx.Commit()
}
