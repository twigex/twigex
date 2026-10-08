// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Shared fixtures for the channel-membership cascade suite.
//
// Cascade tests for the hybrid channel-membership model. The behavior under
// test lives in SQL (NOT EXISTS clauses, multi-table DELETEs, transactional
// demote-vs-delete logic) so these run against real MySQL. See
// testhelper_test.go for connection setup.
//
// Tables touched: channels, channel_members, channel_groups, user_groups,
// group_members. cleanTables is called at the start of every test.

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

// resetCascadeTables wipes every table the cascade suite uses. Called at the
// top of each test so order-independence is preserved across the package.
func resetCascadeTables(t *testing.T) {
	cleanTables(t,
		"channel_members",
		"channel_groups",
		"group_members",
		"user_groups",
		"channels",
	)
}

// seedChannel inserts a minimal public channel and returns its ID.
func seedChannel(t *testing.T, db *sql.DB) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().UnixMilli()
	_, err := db.Exec(
		`INSERT INTO channels (id, type, display_name, name, header, description,
		    last_post, msg_count, created_by, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, '', '', 0, 0, ?, ?, ?, 0)`,
		id, model.ChannelTypePublic, "test", id, model.NewID(), now, now,
	)
	if err != nil {
		t.Fatalf("seedChannel: %v", err)
	}
	return id
}

// seedGroup inserts a user_groups row. If deleted is true, deleted_at is set.
func seedGroup(t *testing.T, db *sql.DB, deleted bool) string {
	t.Helper()
	id := model.NewID()
	now := time.Now().Unix()
	deletedAt := int64(0)
	if deleted {
		deletedAt = now
	}
	_, err := db.Exec(
		`INSERT INTO user_groups (id, name, description, owner_id, created_at, updated_at, deleted_at)
		 VALUES (?, ?, '', ?, ?, ?, ?)`,
		id, "test-"+id[:8], model.NewID(), now, now, deletedAt,
	)
	if err != nil {
		t.Fatalf("seedGroup: %v", err)
	}
	return id
}

func seedGroupMember(t *testing.T, db *sql.DB, groupID, userID string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO group_members (id, group_id, user_id, role, joined_at)
		 VALUES (?, ?, ?, 'member', ?)`,
		model.NewID(), groupID, userID, time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("seedGroupMember: %v", err)
	}
}

func seedChannelGroup(t *testing.T, db *sql.DB, channelID, groupID string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO channel_groups (id, channel_id, group_id, added_by, added_at)
		 VALUES (?, ?, ?, ?, ?)`,
		model.NewID(), channelID, groupID, model.NewID(), time.Now().Unix(),
	)
	if err != nil {
		t.Fatalf("seedChannelGroup: %v", err)
	}
}

// seedChannelMember inserts a channel_members row with the given is_direct
// flag and lets the caller observe per-user state (msg_count, notify_props)
// that the cascade is supposed to preserve. Returns the row's generated ID
// so tests can later verify it survived a promote (if the promote was an
// UPDATE, and not a DELETE+INSERT).
// updated_at is backdated 1 second so tests asserting "promote bumped
// updated_at" don't race with millisecond timestamps.
func seedChannelMember(t *testing.T, db *sql.DB, channelID, userID string, isDirect bool, msgCount, lastViewedAt int64) string {
	t.Helper()
	props, _ := json.Marshal(model.ChannelNotifyProps{})
	id := model.NewID()
	backdated := time.Now().UnixMilli() - 1000
	_, err := db.Exec(
		`INSERT INTO channel_members (id, channel_id, user_id, role, notify_props,
		    msg_count, mention_count, last_viewed_at, updated_at, date_joined, is_direct)
		 VALUES (?, ?, ?, 'member', ?, ?, 0, ?, ?, ?, ?)`,
		id, channelID, userID, props,
		msgCount, lastViewedAt, backdated, backdated, isDirect,
	)
	if err != nil {
		t.Fatalf("seedChannelMember: %v", err)
	}
	return id
}

// getMember fetches a channel_members row or nil if not present.
func getMember(t *testing.T, db *sql.DB, channelID, userID string) *model.ChannelMember {
	t.Helper()
	row := db.QueryRow(
		`SELECT id, channel_id, user_id, role, msg_count, mention_count,
		        last_viewed_at, updated_at, date_joined, is_direct
		 FROM channel_members WHERE channel_id = ? AND user_id = ?`,
		channelID, userID,
	)
	var m model.ChannelMember
	err := row.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.Role,
		&m.MsgCount, &m.MentionCount, &m.LastViewedAt, &m.UpdatedAt,
		&m.DateJoined, &m.IsDirect)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		t.Fatalf("getMember: %v", err)
	}
	return &m
}

func countMembers(t *testing.T, db *sql.DB, channelID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM channel_members WHERE channel_id = ?`, channelID,
	).Scan(&n); err != nil {
		t.Fatalf("countMembers: %v", err)
	}
	return n
}

// countChannelGroups returns the number of channel_groups rows for a channel.
func countChannelGroups(t *testing.T, db *sql.DB, channelID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM channel_groups WHERE channel_id = ?`, channelID,
	).Scan(&n); err != nil {
		t.Fatalf("countChannelGroups: %v", err)
	}
	return n
}

// newDirectMember builds a ChannelMember suitable for CreateOrPromoteMembers.
func newDirectMember(channelID, userID string) model.ChannelMember {
	now := time.Now().UnixMilli()
	return model.ChannelMember{
		ID:           model.NewID(),
		ChannelID:    channelID,
		UserID:       userID,
		Role:         model.ChannelRoleMember,
		NotifyProps:  model.ChannelNotifyProps{},
		MsgCount:     0,
		MentionCount: 0,
		LastViewedAt: 0,
		UpdatedAt:    now,
		DateJoined:   now,
	}
}
