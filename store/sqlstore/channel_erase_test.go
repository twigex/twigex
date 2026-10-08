// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"testing"
)

// Every table that carries a channel_id. A new one added without being handled
// when erasing leaves rows behind, which is what this asserts against.
var channelChildTables = []string{
	"channel_members",
	"channel_groups",
	"channel_guest_links",
	"channel_meetings",
	"posts",
	"post_attachments",
	"post_reactions",
}

func seedChannelWithContent(t *testing.T, db *sql.DB, channelID, postID string) {
	t.Helper()

	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("seed %q: %v", q, err)
		}
	}

	exec(`INSERT INTO channels (id, type, display_name, name, msg_count, created_by, created_at, updated_at, deleted_at)
	      VALUES (?, 'O', 'Erase me', ?, 0, 'u1', 1, 1, 0)`, channelID, channelID)
	exec(`INSERT INTO channel_members (id, channel_id, user_id, notify_props, msg_count, mention_count, last_viewed_at, updated_at, date_joined)
	      VALUES (?, ?, 'u1', '{}', 0, 0, 0, 1, 1)`, channelID+"-m", channelID)
	exec(`INSERT INTO posts (id, channel_id, user_id, message, type, created_at, updated_at, deleted_at)
	      VALUES (?, ?, 'u1', 'hello', 'post', 1, 1, 0)`, postID, channelID)
	exec(`INSERT INTO post_attachments (id, channel_id, post_id, user_id, name, size, mime_type, width, height, created_at, updated_at, url, provider, kind)
	      VALUES (?, ?, ?, 'u1', 'a.png', 1, 'image/png', 1, 1, 1, 1, '', '', 'file')`, postID+"-a", channelID, postID)
	exec(`INSERT INTO post_reactions (id, post_id, channel_id, user_id, reaction, created_at, updated_at)
	      VALUES (?, ?, ?, 'u1', ':+1:', 1, 1)`, postID+"-r", postID, channelID)
	exec(`INSERT INTO channel_groups (id, channel_id, group_id, added_by, added_at)
	      VALUES (?, ?, 'g1', 'u1', 1)`, channelID+"-g", channelID)

	// These two cascade from the channel row rather than being deleted in code,
	// so seeding them is what proves the constraints still fire.
	exec(`INSERT INTO channel_meetings (id, channel_id, host_id, created_by, title, created_at)
	      VALUES (?, ?, 'u1', 'u1', 'Standup', 1)`, channelID+"-mt", channelID)
	exec(`INSERT INTO channel_guest_links (id, channel_id, meeting_id, created_by, expires_at, created_at)
	      VALUES (?, ?, ?, 'u1', 9999999999, 1)`, channelID+"-gl", channelID, channelID+"-mt")
}

func countChannelRows(t *testing.T, db *sql.DB, table, channelID string) int {
	t.Helper()

	var n int
	q := "SELECT COUNT(*) FROM " + table + " WHERE channel_id=?"
	if table == "channels" {
		q = "SELECT COUNT(*) FROM channels WHERE id=?"
	}
	if err := db.QueryRow(q, channelID).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}

	return n
}

func TestEraseLeavesNoRowsBehind(t *testing.T) {
	db := requireDB(t)

	const channelID = "erase-channel"
	const postID = "erase-post"

	seedChannelWithContent(t, db, channelID, postID)

	posts := &postRepository{Db: db}
	channels := &channelsRepository{Db: db}

	ids, err := posts.GetBatchForChannel(t.Context(), channelID, 500)
	if err != nil {
		t.Fatalf("GetBatchForChannel: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected the seeded post, got %d", len(ids))
	}

	attachments, err := posts.GetAttachmentsForMany(t.Context(), ids)
	if err != nil {
		t.Fatalf("GetAttachmentsForMany: %v", err)
	}
	if len(attachments) != 1 {
		t.Fatalf("attachments must be readable before deletion, got %d", len(attachments))
	}

	if err := posts.DeleteBatch(t.Context(), ids); err != nil {
		t.Fatalf("DeleteBatch: %v", err)
	}
	if err := channels.DeletePermanently(t.Context(), channelID); err != nil {
		t.Fatalf("DeletePermanently: %v", err)
	}

	for _, table := range append(channelChildTables, "channels") {
		if n := countChannelRows(t, db, table, channelID); n != 0 {
			t.Errorf("%s still holds %d row(s) for the deleted channel", table, n)
		}
	}
}

func TestPostBatchIsEmptyForAnEmptyChannel(t *testing.T) {
	db := requireDB(t)

	posts := &postRepository{Db: db}

	ids, err := posts.GetBatchForChannel(t.Context(), "no-such-channel", 500)
	if err != nil {
		t.Fatalf("GetBatchForChannel: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("expected no posts, got %d", len(ids))
	}
}
