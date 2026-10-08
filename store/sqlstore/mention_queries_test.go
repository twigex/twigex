// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for the mention/notification store queries added for chat @mentions:
// resolving handles to members, the composer's channel-scoped member search,
// and the mention-badge counters. Runs against real MySQL (see
// testhelper_test.go). Reuses seedNamedUser/seedChannel/seedChannelMember.

import (
	"context"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
)

func containsUserID(users []model.User, id string) bool {
	for _, u := range users {
		if u.ID == id {
			return true
		}
	}
	return false
}

func TestFindUsersByUsernames(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	ctx := context.Background()

	alice := seedNamedUser(t, db, "Alice", "A", "alice@e.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "B", "bob@e.com", "bob", 0)
	seedNamedUser(t, db, "Gone", "G", "gone@e.com", "gone", time.Now().Unix())

	got, err := repo.GetByUsernames(ctx, []string{"alice", "bob"})
	if err != nil {
		t.Fatalf("GetByUsernames: %v", err)
	}
	if len(got) != 2 || !containsUserID(got, alice) || !containsUserID(got, bob) {
		t.Fatalf("exact match: got %d, want alice+bob", len(got))
	}

	// Unknown handles resolve to nothing rather than erroring.
	got, _ = repo.GetByUsernames(ctx, []string{"alice", "nobody"})
	if len(got) != 1 || !containsUserID(got, alice) {
		t.Fatalf("unknown handle: got %d, want just alice", len(got))
	}

	// Match is exact, not a substring: "alic" must not match "alice".
	got, _ = repo.GetByUsernames(ctx, []string{"alic"})
	if len(got) != 0 {
		t.Fatalf("substring leaked: got %d, want 0", len(got))
	}

	// Deactivated users are never mentionable.
	got, _ = repo.GetByUsernames(ctx, []string{"gone"})
	if len(got) != 0 {
		t.Fatalf("deactivated leaked: got %d, want 0", len(got))
	}

	got, _ = repo.GetByUsernames(ctx, nil)
	if len(got) != 0 {
		t.Fatalf("empty input: got %d, want 0", len(got))
	}
}

func TestFindChannelMembersByUsernames(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_members", "channels", "users")

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	alice := seedNamedUser(t, db, "Alice", "A", "alice@e.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "B", "bob@e.com", "bob", 0)
	carol := seedNamedUser(t, db, "Carol", "C", "carol@e.com", "carol", 0)

	seedChannelMember(t, db, ch, alice, true, 0, 0) // direct
	seedChannelMember(t, db, ch, bob, false, 0, 0)  // materialized via a group
	_ = carol                                       // seeded but not a channel member

	// Resolves direct and group members, excludes a non-member even though the
	// handle names a real user.
	got, err := repo.GetMembersByUsernames(ctx, ch, []string{"alice", "bob", "carol"})
	if err != nil {
		t.Fatalf("GetMembersByUsernames: %v", err)
	}
	if len(got) != 2 || !containsUserID(got, alice) || !containsUserID(got, bob) {
		t.Fatalf("got %d, want alice+bob (carol excluded)", len(got))
	}

	got, _ = repo.GetMembersByUsernames(ctx, ch, nil)
	if len(got) != 0 {
		t.Fatalf("empty input: got %d, want 0", len(got))
	}
}

// Mirrors TestFindChannelMembersByUsernames: both forms must filter alike, or
// they would notify differently.
func TestFindChannelMembersByIDs(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_members", "channels", "users")

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	alice := seedNamedUser(t, db, "Alice", "A", "alice@e.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "B", "bob@e.com", "bob", 0)
	carol := seedNamedUser(t, db, "Carol", "C", "carol@e.com", "carol", 0)
	gone := seedNamedUser(t, db, "Gone", "G", "gone@e.com", "gone", time.Now().Unix())

	seedChannelMember(t, db, ch, alice, true, 0, 0) // direct
	seedChannelMember(t, db, ch, bob, false, 0, 0)  // materialized via a group
	seedChannelMember(t, db, ch, gone, true, 0, 0)
	_ = carol // seeded but not a channel member

	got, err := repo.GetMembersByIDs(ctx, ch, []string{alice, bob, carol})
	if err != nil {
		t.Fatalf("GetMembersByIDs: %v", err)
	}
	if len(got) != 2 || !containsUserID(got, alice) || !containsUserID(got, bob) {
		t.Fatalf("got %d, want alice+bob (carol excluded)", len(got))
	}

	got, _ = repo.GetMembersByIDs(ctx, ch, []string{gone})
	if len(got) != 0 {
		t.Fatalf("deactivated leaked: got %d, want 0", len(got))
	}

	got, err = repo.GetMembersByIDs(ctx, ch, []string{model.NewID()})
	if err != nil {
		t.Fatalf("unknown id: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("unknown id: got %d, want 0", len(got))
	}

	got, _ = repo.GetMembersByIDs(ctx, ch, nil)
	if len(got) != 0 {
		t.Fatalf("empty input: got %d, want 0", len(got))
	}
}

func TestSearchChannelMembers(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_members", "channels", "users")

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	alice := seedNamedUser(t, db, "Alice", "A", "alice@e.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "B", "bob@e.com", "bob", 0)
	carol := seedNamedUser(t, db, "Carol", "C", "carol@e.com", "carol", 0)

	seedChannelMember(t, db, ch, alice, true, 0, 0)
	seedChannelMember(t, db, ch, bob, false, 0, 0)
	seedChannelMember(t, db, ch, carol, true, 0, 0)

	got, err := repo.SearchMembers(ctx, ch, "ali", 20)
	if err != nil {
		t.Fatalf("SearchMembers: %v", err)
	}
	if len(got) != 1 || !containsUserID(got, alice) {
		t.Fatalf("prefix search: got %d, want alice", len(got))
	}

	// Empty query lists members up to the limit.
	got, _ = repo.SearchMembers(ctx, ch, "", 20)
	if len(got) != 3 {
		t.Fatalf("empty query: got %d, want 3", len(got))
	}

	got, _ = repo.SearchMembers(ctx, ch, "", 2)
	if len(got) != 2 {
		t.Fatalf("limit not honored: got %d, want 2", len(got))
	}

	// A non-member matching the query is not returned.
	seedNamedUser(t, db, "Dave", "D", "dave@e.com", "dave", 0)
	got, _ = repo.SearchMembers(ctx, ch, "dave", 20)
	if len(got) != 0 {
		t.Fatalf("non-member leaked: got %d, want 0", len(got))
	}
}

func TestIncrementMentionCount(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_members", "channels", "users")

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	alice := seedNamedUser(t, db, "Alice", "A", "alice@e.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "B", "bob@e.com", "bob", 0)
	carol := seedNamedUser(t, db, "Carol", "C", "carol@e.com", "carol", 0)
	seedChannelMember(t, db, ch, alice, true, 0, 0)
	seedChannelMember(t, db, ch, bob, true, 0, 0)
	seedChannelMember(t, db, ch, carol, true, 0, 0)

	if err := repo.IncrementMentionCount(ctx, ch, []string{alice, bob}); err != nil {
		t.Fatalf("IncrementMentionCount: %v", err)
	}
	if getMember(t, db, ch, alice).MentionCount != 1 {
		t.Fatalf("alice badge = %d, want 1", getMember(t, db, ch, alice).MentionCount)
	}
	if getMember(t, db, ch, bob).MentionCount != 1 {
		t.Fatalf("bob badge = %d, want 1", getMember(t, db, ch, bob).MentionCount)
	}
	if getMember(t, db, ch, carol).MentionCount != 0 {
		t.Fatalf("carol badge = %d, want 0 (not mentioned)", getMember(t, db, ch, carol).MentionCount)
	}

	// Counts accumulate across posts.
	if err := repo.IncrementMentionCount(ctx, ch, []string{alice}); err != nil {
		t.Fatalf("IncrementMentionCount: %v", err)
	}
	if getMember(t, db, ch, alice).MentionCount != 2 {
		t.Fatalf("alice badge = %d, want 2", getMember(t, db, ch, alice).MentionCount)
	}

	// Empty id list is a no-op, not an error.
	if err := repo.IncrementMentionCount(ctx, ch, nil); err != nil {
		t.Fatalf("empty ids: %v", err)
	}
}

func TestIncrementAllMentionCounts(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channel_members", "channels", "users")

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	author := seedNamedUser(t, db, "Author", "A", "author@e.com", "author", 0)
	u1 := seedNamedUser(t, db, "One", "O", "one@e.com", "one", 0)
	u2 := seedNamedUser(t, db, "Two", "T", "two@e.com", "two", 0)
	seedChannelMember(t, db, ch, author, true, 0, 0)
	seedChannelMember(t, db, ch, u1, true, 0, 0)
	seedChannelMember(t, db, ch, u2, false, 0, 0)

	// @all/@here raises everyone except the author in one statement.
	if err := repo.IncrementAllMentionCounts(ctx, ch, author); err != nil {
		t.Fatalf("IncrementAllMentionCounts: %v", err)
	}
	if getMember(t, db, ch, u1).MentionCount != 1 || getMember(t, db, ch, u2).MentionCount != 1 {
		t.Fatalf("members not badged: u1=%d u2=%d, want 1/1",
			getMember(t, db, ch, u1).MentionCount, getMember(t, db, ch, u2).MentionCount)
	}
	if getMember(t, db, ch, author).MentionCount != 0 {
		t.Fatalf("author badged = %d, want 0", getMember(t, db, ch, author).MentionCount)
	}
}
