// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestGroupsPagedAndCounted(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	for range 5 {
		seedGroup(t, db, false)
	}

	seedGroup(t, db, true)

	total, err := repo.CountAll(ctx, "")
	if err != nil {
		t.Fatalf("CountAll: %v", err)
	}

	if total != 5 {
		t.Errorf("expected 5 live groups, got %d", total)
	}

	first, err := repo.GetAllPaged(ctx, "", model.Sort{}, 2, 0)
	if err != nil {
		t.Fatalf("GetAllPaged: %v", err)
	}

	last, err := repo.GetAllPaged(ctx, "", model.Sort{}, 2, 4)
	if err != nil {
		t.Fatalf("GetAllPaged: %v", err)
	}

	if len(first) != 2 || len(last) != 1 {
		t.Errorf("expected pages of 2 and 1, got %d and %d", len(first), len(last))
	}

	if first[0].Name > first[1].Name {
		t.Errorf("expected groups ordered by name, got %q before %q", first[0].Name, first[1].Name)
	}

	name := first[0].Name

	matching, err := repo.CountAll(ctx, name)
	if err != nil {
		t.Fatalf("CountAll with query: %v", err)
	}

	found, err := repo.GetAllPaged(ctx, name, model.Sort{}, 10, 0)
	if err != nil {
		t.Fatalf("GetAllPaged with query: %v", err)
	}

	if matching != 1 || len(found) != 1 || found[0].Name != name {
		t.Errorf("expected the search to find only %q, got %d and %v", name, matching, found)
	}
}

func TestSearchGroupsForUserShowsOnlyTheirGroups(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	user := model.NewID()

	joined := seedGroup(t, db, false)
	owned := seedGroup(t, db, false)
	stranger := seedGroup(t, db, false)
	deleted := seedGroup(t, db, true)

	seedGroupMember(t, db, joined, user)
	seedGroupMember(t, db, deleted, user)
	seedGroupMember(t, db, stranger, model.NewID())

	if _, err := db.Exec(`UPDATE user_groups SET owner_id = ? WHERE id = ?`, user, owned); err != nil {
		t.Fatalf("set owner: %v", err)
	}

	found, err := repo.SearchForUser(ctx, user, "", 20)
	if err != nil {
		t.Fatalf("SearchForUser: %v", err)
	}

	got := map[string]bool{}
	for _, g := range found {
		got[g.ID] = true
	}

	if len(found) != 2 || !got[joined] || !got[owned] {
		t.Errorf("expected only the joined and owned groups, got %v", found)
	}

	byName, err := repo.SearchForUser(ctx, user, found[0].Name, 20)
	if err != nil {
		t.Fatalf("SearchForUser with query: %v", err)
	}

	if len(byName) != 1 || byName[0].ID != found[0].ID {
		t.Errorf("expected the search to narrow to %s, got %v", found[0].ID, byName)
	}

	limited, err := repo.SearchForUser(ctx, user, "", 1)
	if err != nil {
		t.Fatalf("SearchForUser with limit: %v", err)
	}

	if len(limited) != 1 {
		t.Errorf("expected the limit to cap results at 1, got %d", len(limited))
	}
}

func TestGroupsPagedSortByMemberCount(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	small := seedGroup(t, db, false)
	large := seedGroup(t, db, false)
	empty := seedGroup(t, db, false)

	seedGroupMember(t, db, small, model.NewID())

	for range 3 {
		seedGroupMember(t, db, large, model.NewID())
	}

	groups, err := repo.GetAllPaged(ctx, "", model.Sort{Key: "member_count", Desc: true}, 10, 0)
	if err != nil {
		t.Fatalf("GetAllPaged: %v", err)
	}

	if len(groups) != 3 || groups[0].ID != large || groups[1].ID != small || groups[2].ID != empty {
		t.Errorf("expected large, small, empty, got %v", groups)
	}
}
