// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func seedPostAt(t *testing.T, repo *postRepository, channelID, userID, replyID, message string, createdAt int64) string {
	t.Helper()

	id := model.NewID()
	_, err := repo.Db.Exec(
		`INSERT INTO posts (id, user_id, channel_id, reply_id, message, type, created_at, updated_at, deleted_at)
		 VALUES (?, ?, ?, ?, ?, 'post', ?, ?, 0)`,
		id, userID, channelID, replyID, message, createdAt, createdAt,
	)
	if err != nil {
		t.Fatalf("seedPostAt: %v", err)
	}

	return id
}

func postByMessage(posts []model.Post, message string) *model.Post {
	for i := range posts {
		if posts[i].Message == message {
			return &posts[i]
		}
	}
	return nil
}

// Written against the per-post version first: it is only evidence the batched
// one behaves the same if it passed before the rewrite.
func TestGetAllForChannelContract(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")
	repo := &postRepository{Db: db}

	const channel, user, other = "channel-1", "user-1", "user-2"
	base := time.Now().UnixMilli()

	root := seedPostAt(t, repo, channel, user, "", "root", base)
	seedAttachment(t, repo, channel, root, user)
	seedAttachment(t, repo, channel, root, user)
	seedReaction(t, repo, channel, root, user, ":tada:")

	seedPostAt(t, repo, channel, user, "", "plain", base+1)

	answer := seedPostAt(t, repo, channel, other, root, "answer", base+2)
	seedReaction(t, repo, channel, answer, user, ":+1:")
	seedReaction(t, repo, channel, answer, other, ":eyes:")

	seedPostAt(t, repo, "channel-2", user, "", "elsewhere", base+3)

	got, err := repo.GetAllForChannel(t.Context(), channel, "", "")
	if err != nil {
		t.Fatalf("GetAllForChannel: %v", err)
	}

	if len(got.Posts) != 3 {
		t.Fatalf("got %d posts, want 3 (channel-2 must be excluded)", len(got.Posts))
	}

	if got.Posts[0].Message != "answer" || got.Posts[2].Message != "root" {
		t.Errorf("order is %s,%s,%s, want answer,plain,root",
			got.Posts[0].Message, got.Posts[1].Message, got.Posts[2].Message)
	}

	t.Run("attachments land on their own post", func(t *testing.T) {
		if n := len(postByMessage(got.Posts, "root").Metadata.Files); n != 2 {
			t.Errorf("root has %d attachments, want 2", n)
		}
		if n := len(postByMessage(got.Posts, "plain").Metadata.Files); n != 0 {
			t.Errorf("plain has %d attachments, want 0", n)
		}
		if n := len(postByMessage(got.Posts, "answer").Metadata.Files); n != 0 {
			t.Errorf("answer has %d attachments, want 0", n)
		}
	})

	t.Run("reactions land on their own post", func(t *testing.T) {
		if n := len(postByMessage(got.Posts, "root").Reactions); n != 1 {
			t.Errorf("root has %d reactions, want 1", n)
		}
		if n := len(postByMessage(got.Posts, "answer").Reactions); n != 2 {
			t.Errorf("answer has %d reactions, want 2", n)
		}
		if n := len(postByMessage(got.Posts, "plain").Reactions); n != 0 {
			t.Errorf("plain has %d reactions, want 0", n)
		}
	})

	t.Run("a reply carries the post it answers, with that post's attachments", func(t *testing.T) {
		a := postByMessage(got.Posts, "answer")
		if a.Reply == nil {
			t.Fatal("answer has no Reply")
		}
		if a.Reply.ID != root {
			t.Errorf("Reply.ID is %s, want the root post", a.Reply.ID)
		}
		if n := len(a.Reply.Metadata.Files); n != 2 {
			t.Errorf("Reply carries %d attachments, want the root post's 2", n)
		}
	})

	t.Run("a post with no reply has none", func(t *testing.T) {
		if p := postByMessage(got.Posts, "plain"); p.Reply != nil {
			t.Errorf("plain carries a Reply it never had: %+v", p.Reply)
		}
	})

	t.Run("ReplyPosts holds the answered posts", func(t *testing.T) {
		if len(got.ReplyPosts) != 1 {
			t.Fatalf("got %d reply posts, want 1", len(got.ReplyPosts))
		}
		if got.ReplyPosts[0].ID != root {
			t.Errorf("ReplyPosts[0] is %s, want the root post", got.ReplyPosts[0].ID)
		}
	})

	t.Run("an empty channel reads as an empty list", func(t *testing.T) {
		empty, err := repo.GetAllForChannel(t.Context(), "channel-empty", "", "")
		if err != nil {
			t.Fatalf("GetAllForChannel: %v", err)
		}
		if len(empty.Posts) != 0 || len(empty.ReplyPosts) != 0 {
			t.Errorf("got %d posts and %d replies, want none", len(empty.Posts), len(empty.ReplyPosts))
		}
	})
}

func TestGetReactionsForMany(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")
	repo := &postRepository{Db: db}

	const channel, user = "channel-1", "user-1"
	base := time.Now().UnixMilli()
	a := seedPostAt(t, repo, channel, user, "", "a", base)
	b := seedPostAt(t, repo, channel, user, "", "b", base+1)
	c := seedPostAt(t, repo, channel, user, "", "c", base+2)

	seedReaction(t, repo, channel, a, user, ":tada:")
	seedReaction(t, repo, channel, b, user, ":+1:")
	seedReaction(t, repo, channel, b, "user-2", ":eyes:")
	seedReaction(t, repo, channel, c, user, ":fire:")

	got, err := repo.GetReactionsForMany(t.Context(), []string{a, b})
	if err != nil {
		t.Fatalf("GetReactionsForMany: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d reactions, want 3 (c was not asked for)", len(got))
	}
	for _, r := range got {
		if r.PostID == c {
			t.Errorf("returned a reaction for a post that was not requested")
		}
	}

	empty, err := repo.GetReactionsForMany(t.Context(), nil)
	if err != nil {
		t.Fatalf("GetReactionsForMany(nil): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("want no rows for an empty request, got %d", len(empty))
	}
}

// Written against the version that queried per post, for the same reason.
func TestGetPostContract(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")
	repo := &postRepository{Db: db}

	const channel, user = "channel-1", "user-1"
	base := time.Now().UnixMilli()

	root := seedPostAt(t, repo, channel, user, "", "root", base)
	seedAttachment(t, repo, channel, root, user)

	answer := seedPostAt(t, repo, channel, user, root, "answer", base+1)
	seedAttachment(t, repo, channel, answer, user)
	seedAttachment(t, repo, channel, answer, user)
	seedReaction(t, repo, channel, answer, user, ":+1:")

	got, err := repo.Get(t.Context(), answer)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Message != "answer" {
		t.Errorf("got %q, want the answer post", got.Message)
	}
	if n := len(got.Metadata.Files); n != 2 {
		t.Errorf("got %d attachments, want 2", n)
	}
	if n := len(got.Reactions); n != 1 {
		t.Errorf("got %d reactions, want 1", n)
	}
	if got.Reply == nil {
		t.Fatal("answer carries no Reply")
	}
	if got.Reply.ID != root {
		t.Errorf("Reply is %s, want the root post", got.Reply.ID)
	}

	t.Run("a post with nothing attached", func(t *testing.T) {
		plain, err := repo.Get(t.Context(), root)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if plain.Reply != nil {
			t.Error("root carries a Reply it never had")
		}
		if len(plain.Reactions) != 0 {
			t.Errorf("root has %d reactions, want 0", len(plain.Reactions))
		}
		if len(plain.Metadata.Files) != 1 {
			t.Errorf("root has %d attachments, want 1", len(plain.Metadata.Files))
		}
	})

	t.Run("a missing post is reported as not found", func(t *testing.T) {
		if _, err := repo.Get(t.Context(), "no-such-post"); !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("got %v, want sql.ErrNoRows", err)
		}
	})
}

// The point of threading the request context down: a client that disconnects
// mid-read should not leave the remaining queries running.
func TestChannelReadStopsWhenTheCallerGivesUp(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "posts", "post_reactions", "post_attachments")
	repo := &postRepository{Db: db}

	base := time.Now().UnixMilli()
	for i := 0; i < 3; i++ {
		seedPostAt(t, repo, "channel-1", "user-1", "", "m", base+int64(i))
	}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := repo.GetAllForChannel(ctx, "channel-1", "", ""); !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}
