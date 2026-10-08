// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Removal cascades: demote where another group path remains, delete where none does.

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestRemoveUserFromChannel_DemotesWhenGroupPathExists(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)

	const msgCount, lastViewed = int64(7), int64(111)
	seedChannelMember(t, db, ch, userA, true, msgCount, lastViewed)

	if err := repo.RemoveUser(ch, userA); err != nil {
		t.Fatalf("RemoveUser: %v", err)
	}

	m := getMember(t, db, ch, userA)
	if m == nil {
		t.Fatal("row should survive as materialized; got nil")
	}
	if m.IsDirect {
		t.Error("expected is_direct=false after demote")
	}

	if m.MsgCount != msgCount || m.LastViewedAt != lastViewed {
		t.Errorf("state not preserved during demote: msg_count=%d last_viewed=%d",
			m.MsgCount, m.LastViewedAt)
	}
}

func TestRemoveUserFromChannel_DeletesWhenNoGroupPath(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}

	ch := seedChannel(t, db)
	userA := model.NewID()
	seedChannelMember(t, db, ch, userA, true, 0, 0)

	if err := repo.RemoveUser(ch, userA); err != nil {
		t.Fatalf("RemoveUser: %v", err)
	}
	if got := countMembers(t, db, ch); got != 0 {
		t.Errorf("expected row deleted, got %d rows", got)
	}
}

func TestDematerializeGroupFromChannel_KeepsUsersWithAnotherGroupPath(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, g1, userA)
	seedGroupMember(t, db, g2, userA)
	seedChannelGroup(t, db, ch, g1)
	seedChannelGroup(t, db, ch, g2)
	seedChannelMember(t, db, ch, userA, false, 0, 0)

	// Detach g1; the channel_groups row must be gone before the cascade runs,
	// per the contract documented on DematerializeGroup.
	if err := repo.RemoveGroup(ctx, ch, g1); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	if err := repo.DematerializeGroup(ctx, ch, g1); err != nil {
		t.Fatalf("DematerializeGroup: %v", err)
	}

	m := getMember(t, db, ch, userA)
	if m == nil {
		t.Error("userA should still be in channel via g2; got removed")
	}
}

func TestDematerializeGroupFromChannel_RemovesUsersWithNoOtherPath(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)
	seedChannelMember(t, db, ch, userA, false, 0, 0)

	if err := repo.RemoveGroup(ctx, ch, gr); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	if err := repo.DematerializeGroup(ctx, ch, gr); err != nil {
		t.Fatalf("DematerializeGroup: %v", err)
	}
	if got := countMembers(t, db, ch); got != 0 {
		t.Errorf("expected userA removed, got %d rows", got)
	}
}

func TestDematerializeGroupFromChannel_LeavesDirectRowsAlone(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)
	// userA is direct in the channel, and also in the group, but their
	// direct membership is the path that matters.
	seedChannelMember(t, db, ch, userA, true, 0, 0)

	if err := repo.RemoveGroup(ctx, ch, gr); err != nil {
		t.Fatalf("RemoveGroup: %v", err)
	}
	if err := repo.DematerializeGroup(ctx, ch, gr); err != nil {
		t.Fatalf("DematerializeGroup: %v", err)
	}
	m := getMember(t, db, ch, userA)
	if m == nil || !m.IsDirect {
		t.Errorf("direct row should be untouched; got %+v", m)
	}
}

func TestRemoveGroupFromAllChannels_RemovesMaterializedRows(t *testing.T) {
	// Group attached to two channels; deleting the group should drop
	// materialized rows where the deleted group was the only path, while
	// leaving alone any user reachable via another attached group.
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	c1, c2 := seedChannel(t, db), seedChannel(t, db)
	gDel, gOther := seedGroup(t, db, false), seedGroup(t, db, false)
	userA, userB := model.NewID(), model.NewID()

	// userA is in both groups; userB is only in gDel.
	seedGroupMember(t, db, gDel, userA)
	seedGroupMember(t, db, gDel, userB)
	seedGroupMember(t, db, gOther, userA)

	// gDel attached to both channels; gOther only attached to c1.
	seedChannelGroup(t, db, c1, gDel)
	seedChannelGroup(t, db, c1, gOther)
	seedChannelGroup(t, db, c2, gDel)

	// Materialized rows for both users in both channels (only ones gDel reaches).
	seedChannelMember(t, db, c1, userA, false, 0, 0)
	seedChannelMember(t, db, c1, userB, false, 0, 0)
	seedChannelMember(t, db, c2, userA, false, 0, 0)
	seedChannelMember(t, db, c2, userB, false, 0, 0)

	if err := repo.RemoveGroupFromAllChannels(ctx, gDel); err != nil {
		t.Fatalf("RemoveGroupFromAllChannels: %v", err)
	}

	// c1: userA still has gOther path: stays. userB has no other path: gone.
	if m := getMember(t, db, c1, userA); m == nil {
		t.Error("c1/userA should remain (via gOther)")
	}
	if m := getMember(t, db, c1, userB); m != nil {
		t.Error("c1/userB should be removed (no other path)")
	}
	// c2: neither user has a remaining path.
	if got := countMembers(t, db, c2); got != 0 {
		t.Errorf("c2 should be empty, got %d rows", got)
	}
}

func TestRemoveGroupFromAllChannels_LeavesDirectRowsAlone(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)
	seedChannelMember(t, db, ch, userA, true, 0, 0) // direct

	if err := repo.RemoveGroupFromAllChannels(ctx, gr); err != nil {
		t.Fatalf("RemoveGroupFromAllChannels: %v", err)
	}
	m := getMember(t, db, ch, userA)
	if m == nil || !m.IsDirect {
		t.Errorf("direct row should survive group delete; got %+v", m)
	}
}

func TestDematerializeUserFromGroupChannels_RemovesWhereNoOtherPath(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	c1, c2 := seedChannel(t, db), seedChannel(t, db)
	gLeft, gKept := seedGroup(t, db, false), seedGroup(t, db, false)
	userA := model.NewID()

	// userA is in both groups; both groups are in c1, only gLeft in c2.
	seedGroupMember(t, db, gLeft, userA)
	seedGroupMember(t, db, gKept, userA)
	seedChannelGroup(t, db, c1, gLeft)
	seedChannelGroup(t, db, c1, gKept)
	seedChannelGroup(t, db, c2, gLeft)
	seedChannelMember(t, db, c1, userA, false, 0, 0)
	seedChannelMember(t, db, c2, userA, false, 0, 0)

	if _, err := db.Exec(`DELETE FROM group_members WHERE group_id = ? AND user_id = ?`,
		gLeft, userA); err != nil {
		t.Fatalf("seed: remove group_members: %v", err)
	}
	if err := repo.DematerializeUserFromGroupChannels(ctx, userA, gLeft); err != nil {
		t.Fatalf("DematerializeUserFromGroupChannels: %v", err)
	}

	// c1: userA still in gKept: stays. c2: no other path: gone.
	if m := getMember(t, db, c1, userA); m == nil {
		t.Error("c1/userA should remain (still in gKept)")
	}
	if m := getMember(t, db, c2, userA); m != nil {
		t.Error("c2/userA should be removed (no other path)")
	}
}

func TestCascadeWritesAreNoOpsOnEmptyInput(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)

	cases := map[string]struct {
		nilInput   func() error
		emptyInput func() error
	}{
		"MaterializeGroups": {
			func() error { return repo.MaterializeGroups(ctx, ch, nil) },
			func() error { return repo.MaterializeGroups(ctx, ch, []string{}) },
		},
		"AddGroups": {
			func() error { return repo.AddGroups(ctx, ch, nil, model.NewID()) },
			func() error { return repo.AddGroups(ctx, ch, []string{}, model.NewID()) },
		},
		"MaterializeUserInGroupChannels": {
			func() error { return repo.MaterializeUserInGroupChannels(ctx, nil, gr) },
			func() error { return repo.MaterializeUserInGroupChannels(ctx, []string{}, gr) },
		},
		"CreateOrPromoteMembers": {
			func() error { return repo.CreateOrPromoteMembers(ctx, nil) },
			func() error { return repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{}) },
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := c.nilInput(); err != nil {
				t.Errorf("nil input should be a no-op, got %v", err)
			}
			if err := c.emptyInput(); err != nil {
				t.Errorf("empty input should be a no-op, got %v", err)
			}
		})
	}

	if got := countChannelGroups(t, db, ch); got != 0 {
		t.Errorf("expected 0 channel_groups rows, got %d", got)
	}
	if got := countMembers(t, db, ch); got != 0 {
		t.Errorf("expected 0 channel_members rows, got %d", got)
	}
}

func TestRemoveUserFromChannel_DropsTheAdminRoleWhenDemoting(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)
	seedChannelMember(t, db, ch, userA, true, 0, 0)

	if err := repo.UpdateUserRole(ch, userA, model.ChannelRoleAdmin); err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}

	if err := repo.RemoveUser(ch, userA); err != nil {
		t.Fatalf("RemoveUser: %v", err)
	}

	m := getMember(t, db, ch, userA)
	if m == nil {
		t.Fatal("row should survive as materialized; got nil")
	}
	if m.Role != model.ChannelRoleMember {
		t.Errorf("a removed admin kept %q on a row the member list cannot show", m.Role)
	}
}
