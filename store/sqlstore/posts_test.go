// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Deleting a post must take its reactions and attachments with it, or every
// deletion leaves rows pointing at a post that no longer exists.
// Runs against real MySQL (see testhelper_test.go).

import (
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func seedPost(t *testing.T, repo *postRepository, channelID, userID string) string {
	t.Helper()

	id := model.NewID()
	now := time.Now().UnixMilli()
	_, err := repo.Db.Exec(
		`INSERT INTO posts (id, user_id, channel_id, reply_id, message, type, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, '', 'hello', 'post', ?, ?, 0)`,
		id, userID, channelID, now, now,
	)
	if err != nil {
		t.Fatalf("seedPost: %v", err)
	}

	return id
}

func seedReaction(t *testing.T, repo *postRepository, channelID, postID, userID, reaction string) {
	t.Helper()

	now := time.Now().UnixMilli()
	_, err := repo.Db.Exec(
		`INSERT INTO post_reactions (id, channel_id, post_id, user_id, reaction, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
		model.NewID(), channelID, postID, userID, reaction, now, now,
	)
	if err != nil {
		t.Fatalf("seedReaction: %v", err)
	}
}

func seedAttachment(t *testing.T, repo *postRepository, channelID, postID, userID string) {
	t.Helper()

	now := time.Now().UnixMilli()
	_, err := repo.Db.Exec(
		`INSERT INTO post_attachments
		 (id, channel_id, post_id, user_id, name, size, mime_type, width, height, url, provider, kind, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, 'f.png', 1, 'image/png', 1, 1, '', '', 'image', ?, ?, 0)`,
		model.NewID(), channelID, postID, userID, now, now,
	)
	if err != nil {
		t.Fatalf("seedAttachment: %v", err)
	}
}

func countFor(t *testing.T, repo *postRepository, table, postID string) int {
	t.Helper()

	var n int
	if err := repo.Db.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE post_id=?", postID).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}

	return n
}

func TestDeletePostRemovesReactionsAndAttachments(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")

	repo := &postRepository{Db: db}
	ch, user := model.NewID(), model.NewID()

	post := seedPost(t, repo, ch, user)
	seedReaction(t, repo, ch, post, user, "+1")
	seedReaction(t, repo, ch, post, model.NewID(), "tada")
	seedAttachment(t, repo, ch, post, user)

	if err := repo.Delete(post); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	for _, table := range []string{"posts", "post_reactions", "post_attachments"} {
		col := "post_id"
		if table == "posts" {
			col = "id"
		}
		var n int
		if err := db.QueryRow("SELECT COUNT(*) FROM "+table+" WHERE "+col+"=?", post).Scan(&n); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("%s still holds %d rows for the deleted post", table, n)
		}
	}
}

// Another post's rows must survive, or the delete is matching on the wrong key.
func TestDeletePostLeavesOtherPostsAlone(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")

	repo := &postRepository{Db: db}
	ch, user := model.NewID(), model.NewID()

	doomed := seedPost(t, repo, ch, user)
	keep := seedPost(t, repo, ch, user)
	seedReaction(t, repo, ch, doomed, user, "+1")
	seedReaction(t, repo, ch, keep, user, "+1")
	seedAttachment(t, repo, ch, keep, user)

	if err := repo.Delete(doomed); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if got := countFor(t, repo, "post_reactions", keep); got != 1 {
		t.Errorf("the surviving post has %d reactions, want 1", got)
	}
	if got := countFor(t, repo, "post_attachments", keep); got != 1 {
		t.Errorf("the surviving post has %d attachments, want 1", got)
	}
}

func TestUpdatePostKeepsAttachmentDimensions(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_attachments")

	repo := &postRepository{Db: db}
	ch, user := model.NewID(), model.NewID()
	post := seedPost(t, repo, ch, user)

	attachment := model.PostFileAttachment{
		ID: model.NewID(), ChannelID: ch, PostID: post, UserID: user,
		Name: "photo.png", MimeType: "image/png", Width: 1920, Height: 1080,
	}
	if _, err := repo.Update(post, "edited", nil, []model.PostFileAttachment{attachment}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetAttachmentByID(attachment.ID)
	if err != nil {
		t.Fatalf("GetAttachmentByID: %v", err)
	}
	if got.Width != 1920 || got.Height != 1080 {
		t.Fatalf("dimensions = %dx%d, want 1920x1080", got.Width, got.Height)
	}
}
