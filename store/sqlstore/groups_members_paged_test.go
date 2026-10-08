// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestGroupMembersPagedAndCounted(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members", "users")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	group := seedGroup(t, db, false)
	other := seedGroup(t, db, false)

	anna := seedNamedUser(t, db, "Anna", "Berzina", "anna@example.com", "anna", 0)
	andris := seedNamedUser(t, db, "Andris", "Kalnins", "andris@example.com", "andris", 0)
	maris := seedNamedUser(t, db, "Maris", "Ozols", "maris@example.com", "maris", 0)

	seedGroupMember(t, db, group, anna)
	seedGroupMember(t, db, group, andris)
	seedGroupMember(t, db, group, maris)
	seedGroupMember(t, db, other, anna)

	total, err := repo.CountMembers(ctx, group, "")
	if err != nil {
		t.Fatalf("CountMembers: %v", err)
	}

	if total != 3 {
		t.Errorf("expected 3 members, got %d", total)
	}

	page, err := repo.GetMembersPaged(ctx, group, "", model.Sort{}, 2, 2)
	if err != nil {
		t.Fatalf("GetMembersPaged: %v", err)
	}

	if len(page) != 1 || page[0].UserID != maris {
		t.Errorf("expected the second page to hold only Maris, got %v", page)
	}

	matching, err := repo.CountMembers(ctx, group, "an")
	if err != nil {
		t.Fatalf("CountMembers with query: %v", err)
	}

	found, err := repo.GetMembersPaged(ctx, group, "an", model.Sort{}, 10, 0)
	if err != nil {
		t.Fatalf("GetMembersPaged with query: %v", err)
	}

	if matching != 2 || len(found) != 2 {
		t.Errorf("expected the search to count and return 2 members, got %d and %d", matching, len(found))
	}
}

func TestGroupMembersPagingIsStableForSameNames(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members", "users")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	group := seedGroup(t, db, false)

	for i := range 25 {
		id := seedNamedUser(t, db, "Anna", "Berzina", fmt.Sprintf("anna%02d@example.com", i), fmt.Sprintf("anna%02d", i), 0)
		seedGroupMember(t, db, group, id)
	}

	seen := map[string]int{}

	for offset := 0; offset < 25; offset += 7 {
		page, err := repo.GetMembersPaged(ctx, group, "", model.Sort{}, 7, offset)
		if err != nil {
			t.Fatalf("GetMembersPaged offset %d: %v", offset, err)
		}

		for _, m := range page {
			seen[m.UserID]++
		}
	}

	if len(seen) != 25 {
		t.Errorf("expected 25 distinct members across the pages, got %d", len(seen))
	}

	for id, n := range seen {
		if n != 1 {
			t.Errorf("member %s appeared %d times", id, n)
		}
	}
}

func TestGroupMembersPagedFollowTheRequestedSort(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members", "users")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	group := seedGroup(t, db, false)

	for _, u := range []struct{ name, email string }{
		{"Anna", "c@example.com"},
		{"Maris", "a@example.com"},
		{"Liga", "b@example.com"},
	} {
		id := seedNamedUser(t, db, u.name, "Test", u.email, u.name, 0)
		seedGroupMember(t, db, group, id)
	}

	cases := []struct {
		sort model.Sort
		want []string
	}{
		{model.Sort{}, []string{"Anna", "Liga", "Maris"}},
		{model.Sort{Key: "name", Desc: true}, []string{"Maris", "Liga", "Anna"}},
		{model.Sort{Key: "email"}, []string{"Maris", "Liga", "Anna"}},
	}

	for _, c := range cases {
		members, err := repo.GetMembersPaged(ctx, group, "", c.sort, 10, 0)
		if err != nil {
			t.Fatalf("GetMembersPaged(%+v): %v", c.sort, err)
		}

		got := make([]string, len(members))
		for i, m := range members {
			got[i] = m.UserInfo.Name
		}

		if fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("sort %+v: got %v, want %v", c.sort, got, c.want)
		}
	}
}
