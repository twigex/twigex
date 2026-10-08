// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func TestRowPlaceholders(t *testing.T) {
	cases := []struct {
		rows, columns int
		want          string
	}{
		{1, 1, "(?)"},
		{1, 2, "(?,?)"},
		{3, 2, "(?,?),(?,?),(?,?)"},
		{2, 4, "(?,?,?,?),(?,?,?,?)"},
		{0, 4, ""},
		{4, 0, ""},
	}

	for _, c := range cases {
		if got := rowPlaceholders(c.rows, c.columns); got != c.want {
			t.Errorf("rowPlaceholders(%d, %d) = %q, want %q", c.rows, c.columns, got, c.want)
		}
	}
}

// It used to insert one row per call outside a transaction, so a failure part
// way through left some of them written.
func TestAddMeetingInviteesWritesEveryRow(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_meeting_invitees", "channel_meetings", "channels")
	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	// A real meeting is needed: without it INSERT IGNORE turns the foreign key
	// violation into a warning and writes nothing.
	channel, meeting := model.NewID(), model.NewID()
	now := time.Now().Unix()

	if _, err := db.Exec(
		`INSERT INTO channels (id, type, display_name, name, header, description,
		 last_post, msg_count, created_by, created_at, updated_at, deleted_at)
		 VALUES (?, 'O', 'Team', ?, '', '', 0, 0, 'u1', ?, ?, 0)`,
		channel, "team-"+channel[:8], now, now); err != nil {
		t.Fatalf("seed channel: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO channel_meetings (id, channel_id, host_id, created_by, title, created_at)
		 VALUES (?, ?, 'u1', 'u1', 'Standup', ?)`,
		meeting, channel, now); err != nil {
		t.Fatalf("seed meeting: %v", err)
	}

	users := []string{model.NewID(), model.NewID(), model.NewID()}

	if err := repo.AddMeetingInvitees(ctx, meeting, users); err != nil {
		t.Fatalf("AddMeetingInvitees: %v", err)
	}
	if n := countInvitees(t, db, meeting); n != len(users) {
		t.Errorf("got %d invitees, want %d", n, len(users))
	}

	if err := repo.AddMeetingInvitees(ctx, meeting, users); err != nil {
		t.Fatalf("second call: %v", err)
	}
	if n := countInvitees(t, db, meeting); n != len(users) {
		t.Errorf("after repeating, got %d invitees, want %d", n, len(users))
	}

	if err := repo.AddMeetingInvitees(ctx, meeting, nil); err != nil {
		t.Errorf("empty list: %v", err)
	}
}

func countInvitees(t *testing.T, db *sql.DB, meeting string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM channel_meeting_invitees WHERE meeting_id=?", meeting).Scan(&n); err != nil {
		t.Fatalf("count invitees: %v", err)
	}

	return n
}

func TestPreferencesUpdateWritesEveryRow(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "preferences")
	repo := &preferencesRepository{Db: db}

	user := model.NewID()
	prefs := []model.Preference{
		{UserID: user, Category: "chat", Name: "sound", Value: "on"},
		{UserID: user, Category: "chat", Name: "desktop", Value: "off"},
		{UserID: user, Category: "files", Name: "view", Value: "grid"},
	}

	if err := repo.Update(prefs); err != nil {
		t.Fatalf("Update: %v", err)
	}

	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM preferences WHERE user_id=?", user).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != len(prefs) {
		t.Fatalf("got %d preferences, want %d", n, len(prefs))
	}

	prefs[0].Value = "off"
	if err := repo.Update(prefs); err != nil {
		t.Fatalf("second Update: %v", err)
	}

	var value string
	if err := db.QueryRow(
		"SELECT value FROM preferences WHERE user_id=? AND category='chat' AND name='sound'",
		user).Scan(&value); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if value != "off" {
		t.Errorf("value is %q, want the updated one", value)
	}

	if err := db.QueryRow("SELECT COUNT(*) FROM preferences WHERE user_id=?", user).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != len(prefs) {
		t.Errorf("repeating added rows: got %d, want %d", n, len(prefs))
	}

	if err := repo.Update(nil); err != nil {
		t.Errorf("empty list: %v", err)
	}
}

func TestBatchedInsertsWriteEveryRow(t *testing.T) {
	db := requireDB(t)

	t.Run("CreateAttachment", func(t *testing.T) {
		cleanTables(t, "post_attachments")
		repo := &postRepository{Db: db}

		post, channel := model.NewID(), model.NewID()
		files := []model.PostFileAttachment{
			{ID: model.NewID(), ChannelID: channel, PostID: post, UserID: "u1", Name: "a.png", Kind: "image"},
			{ID: model.NewID(), ChannelID: channel, PostID: post, UserID: "u1", Name: "b.png", Kind: "image"},
			{ID: model.NewID(), ChannelID: channel, PostID: post, UserID: "u1", Name: "c.pdf", Kind: "file"},
		}

		if _, err := repo.CreateAttachment(files); err != nil {
			t.Fatalf("CreateAttachment: %v", err)
		}
		if n := countRows(t, db, "SELECT COUNT(*) FROM post_attachments WHERE post_id=?", post); n != len(files) {
			t.Errorf("got %d attachments, want %d", n, len(files))
		}

		if _, err := repo.CreateAttachment(nil); err != nil {
			t.Errorf("empty list: %v", err)
		}
	})

	t.Run("AddMembers", func(t *testing.T) {
		cleanTables(t, "group_members")
		repo := &groupRepository{Db: db}

		group := model.NewID()
		users := []string{model.NewID(), model.NewID(), model.NewID()}

		if err := repo.AddMembers(t.Context(), group, users, ""); err != nil {
			t.Fatalf("AddMembers: %v", err)
		}
		if n := countRows(t, db, "SELECT COUNT(*) FROM group_members WHERE group_id=?", group); n != len(users) {
			t.Errorf("got %d members, want %d", n, len(users))
		}

		if err := repo.AddMembers(t.Context(), group, users, "admin"); err != nil {
			t.Fatalf("second AddMembers: %v", err)
		}
		if n := countRows(t, db, "SELECT COUNT(*) FROM group_members WHERE group_id=?", group); n != len(users) {
			t.Errorf("repeating added rows: got %d, want %d", n, len(users))
		}
		if n := countRows(t, db,
			"SELECT COUNT(*) FROM group_members WHERE group_id=? AND role='admin'", group); n != len(users) {
			t.Errorf("roles were not updated: %d of %d are admin", n, len(users))
		}

		if err := repo.AddMembers(t.Context(), group, nil, ""); err != nil {
			t.Errorf("empty list: %v", err)
		}
	})

	t.Run("AddWorkspaceUsers", func(t *testing.T) {
		cleanTables(t, "collimato_workspace_users")
		repo := &collimatoRepository{Db: db}

		ws := model.NewID()
		users := []string{model.NewID(), model.NewID()}

		if err := repo.AddWorkspaceUsers(ws, users, []string{"viewer"}); err != nil {
			t.Fatalf("AddWorkspaceUsers: %v", err)
		}
		if n := countRows(t, db,
			"SELECT COUNT(*) FROM collimato_workspace_users WHERE workspace_id=?", ws); n != len(users) {
			t.Errorf("got %d users, want %d", n, len(users))
		}

		if err := repo.AddWorkspaceUsers(ws, nil, nil); err != nil {
			t.Errorf("empty list: %v", err)
		}
	})

	t.Run("RemoveUsersFromSharedFile", func(t *testing.T) {
		cleanTables(t, "file_share")
		repo := &fileRepository{Db: db}

		file := model.NewID()
		kept, removed := model.NewID(), []string{model.NewID(), model.NewID()}

		for _, u := range append([]string{kept}, removed...) {
			seedShareRow(t, file, model.SHARE_TYPE_USER, u, 0, model.AccessEditor)
		}

		if err := repo.RemoveUsersFromSharedFile(file, removed); err != nil {
			t.Fatalf("RemoveUsersFromSharedFile: %v", err)
		}
		if n := countRows(t, db, "SELECT COUNT(*) FROM file_share WHERE file_id=?", file); n != 1 {
			t.Errorf("got %d shares left, want the one that was not named", n)
		}
		if n := countRows(t, db,
			"SELECT COUNT(*) FROM file_share WHERE file_id=? AND share_with=?", file, kept); n != 1 {
			t.Error("removed a user that was not asked for")
		}

		if err := repo.RemoveUsersFromSharedFile(file, nil); err != nil {
			t.Errorf("empty list: %v", err)
		}
	})
}

func countRows(t *testing.T, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}

	return n
}
