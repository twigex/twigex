// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestGroupMemberCounts(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "user_groups", "group_members")
	repo := &groupRepository{Db: db}
	ctx := context.Background()

	full := seedGroup(t, db, false)
	empty := seedGroup(t, db, false)
	member := model.NewID()
	seedGroupMember(t, db, full, member)
	seedGroupMember(t, db, full, model.NewID())
	seedGroupMember(t, db, full, model.NewID())

	want := map[string]int{full: 3, empty: 0}
	check := func(name string, groups []model.Group, wantLen int) {
		t.Helper()
		if len(groups) != wantLen {
			t.Fatalf("%s: got %d groups, want %d", name, len(groups), wantLen)
		}
		for _, g := range groups {
			if g.MemberCount != want[g.ID] {
				t.Errorf("%s: group %s count = %d, want %d", name, g.ID, g.MemberCount, want[g.ID])
			}
		}
	}

	for _, id := range []string{full, empty} {
		g, err := repo.Get(ctx, id)
		if err != nil || g == nil {
			t.Fatalf("Get %s: %v", id, err)
		}
		check("Get", []model.Group{*g}, 1)
	}

	byIDs, err := repo.GetByIDs(ctx, []string{full, empty})
	if err != nil {
		t.Fatalf("GetByIDs: %v", err)
	}
	check("GetByIDs", byIDs, 2)

	all, err := repo.GetAllPaged(ctx, "", model.Sort{}, 10, 0)
	if err != nil {
		t.Fatalf("GetAllPaged: %v", err)
	}
	check("GetAllPaged", all, 2)

	mine, err := repo.SearchForUser(ctx, member, "", 10)
	if err != nil {
		t.Fatalf("SearchForUser: %v", err)
	}
	check("SearchForUser", mine, 1)
}
