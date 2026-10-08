// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for UserStore.Search, the typeahead behind the user pickers. Runs against
// real MySQL (see testhelper_test.go).

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// seedNamedUser inserts a user with explicit name/lastname/email/username and
// deactivation state, against the applied 17-column users schema.
func seedNamedUser(t *testing.T, db *sql.DB, name, lastname, email, username string, deactivatedAt int64) string {
	t.Helper()
	return insertUser(t, db, userRow{
		name:          name,
		lastname:      lastname,
		email:         email,
		username:      username,
		deactivatedAt: deactivatedAt,
	})
}

func TestSearchUsers_MatchesNameEmailUsername(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	ctx := context.Background()

	alice := seedNamedUser(t, db, "Alice", "Anderson", "alice@acme.com", "alice", 0)
	bob := seedNamedUser(t, db, "Bob", "Brown", "bob@example.com", "bobby", 0)
	seedNamedUser(t, db, "Carol", "Clark", "carol@example.com", "carol", 0)

	// Match by first name.
	got, err := repo.Search(ctx, "alice", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 || got[0].ID != alice {
		t.Fatalf("by name: got %d users, want just alice", len(got))
	}

	// Match by email substring.
	got, err = repo.Search(ctx, "acme", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 || got[0].ID != alice {
		t.Fatalf("by email: got %d, want alice", len(got))
	}

	// Match by username substring.
	got, err = repo.Search(ctx, "bobb", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 || got[0].ID != bob {
		t.Fatalf("by username: got %d, want bob", len(got))
	}
}

func TestSearchUsers_ExcludesDeactivatedAndRespectsLimit(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	ctx := context.Background()

	seedNamedUser(t, db, "Active", "One", "a1@example.com", "active1", 0)
	seedNamedUser(t, db, "Active", "Two", "a2@example.com", "active2", 0)
	seedNamedUser(t, db, "Gone", "User", "gone@example.com", "gone", time.Now().Unix())

	// Deactivated user must not appear even on a matching query.
	got, err := repo.Search(ctx, "gone", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("deactivated user leaked: got %d", len(got))
	}

	// Limit is honored.
	got, err = repo.Search(ctx, "Active", 1)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("limit not honored: got %d, want 1", len(got))
	}
}

func TestSearchUsers_EmptyQueryReturnsActiveUpToLimit(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	ctx := context.Background()

	seedNamedUser(t, db, "A", "A", "a@example.com", "a", 0)
	seedNamedUser(t, db, "B", "B", "b@example.com", "b", 0)
	seedNamedUser(t, db, "Z", "Z", "z@example.com", "z", time.Now().Unix())

	got, err := repo.Search(ctx, "", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("empty query: got %d active users, want 2", len(got))
	}
}

func TestSearchUsers_EscapesLikeWildcards(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	ctx := context.Background()

	seedNamedUser(t, db, "Normal", "User", "normal@example.com", "normal", 0)

	// "%" is a LIKE wildcard; escaped it should match nothing literally.
	got, err := repo.Search(ctx, "%", 20)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("unescaped wildcard matched %d users, want 0", len(got))
	}
}
