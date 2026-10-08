// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Tests for UpdateUsername, which follows a rename in an external directory.
// Runs against real MySQL (see testhelper_test.go).

import (
	"testing"
)

func TestUpdateUsername_Renames(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "users")

	repo := userRepository{Db: db}
	id := seedNamedUser(t, db, "Jane", "Doe", "jane@example.com", "jane", 0)

	ok, err := repo.UpdateUsername(id, "jsmith")
	if err != nil {
		t.Fatalf("UpdateUsername: %v", err)
	}
	if !ok {
		t.Fatal("UpdateUsername returned false without an error")
	}

	found, err := repo.GetByUsername("jsmith")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if found == nil || found.ID != id {
		t.Fatal("the new username does not resolve to the same account")
	}

	stale, err := repo.GetByUsername("jane")
	if err != nil {
		t.Fatalf("GetByUsername: %v", err)
	}
	if stale != nil {
		t.Error("the old username still resolves")
	}
}
