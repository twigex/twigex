// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func countWhere(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := testDB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

func TestAddMemberKeepsExistingMembership(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "workspace_members")
	repo := &workspaceRepository{Db: db}

	mustExec(t, `INSERT INTO workspace_members (id, user_id, workspace_id, role, date_joined) VALUES ('admin-row', 'anna', 'w1', 'admin', 1)`)

	got, err := repo.AddMember([]model.WorkspaceMember{
		{ID: model.NewID(), UserID: "anna", WorkspaceID: "w1", Role: "user", DateJoined: 2},
		{ID: model.NewID(), UserID: "bob", WorkspaceID: "w1", Role: "user", DateJoined: 2},
		{ID: model.NewID(), UserID: "bob", WorkspaceID: "w1", Role: "user", DateJoined: 2},
	})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}

	if n := countWhere(t, `SELECT COUNT(*) FROM workspace_members WHERE workspace_id = 'w1'`); n != 2 {
		t.Fatalf("got %d member rows, want 2", n)
	}
	if len(got) != 2 {
		t.Fatalf("returned %d members, want 2", len(got))
	}
	if got[0].ID != "admin-row" || got[0].Role != "admin" {
		t.Errorf("anna = %+v, want the existing admin row", got[0])
	}
	if got[1].UserID != "bob" || got[1].Role != "user" {
		t.Errorf("bob = %+v, want a new user row", got[1])
	}
}

func TestAddFavouriteTwiceKeepsOneRow(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "favorites")
	repo := &fileRepository{Db: db}

	for i := 0; i < 2; i++ {
		if err := repo.AddFavourite("anna", "file-1"); err != nil {
			t.Fatalf("AddFavourite: %v", err)
		}
	}

	if n := countWhere(t, `SELECT COUNT(*) FROM favorites WHERE user_id = 'anna'`); n != 1 {
		t.Fatalf("got %d favourite rows, want 1", n)
	}
}

func TestUpdateGridSortKeepsOneRowPerView(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "workspace_sort")
	repo := &workspaceRepository{Db: db}

	for _, settings := range []string{`[{"field":"a"}]`, `[{"field":"b"}]`} {
		if _, err := repo.UpdateGridSort("w1", "t1", "v1", settings, model.NewID()); err != nil {
			t.Fatalf("UpdateGridSort: %v", err)
		}
	}

	if n := countWhere(t, `SELECT COUNT(*) FROM workspace_sort WHERE view_id = 'v1'`); n != 1 {
		t.Fatalf("got %d sort rows, want 1", n)
	}
	sorts, err := repo.sortsByView("w1", []string{"v1"})
	if err != nil {
		t.Fatalf("sortsByView: %v", err)
	}
	if got := sorts[tableViewKey{"t1", "v1"}]; len(got) != 1 || got[0]["field"] != "b" {
		t.Fatalf("sort = %v, want the latest settings", got)
	}
}
