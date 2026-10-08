// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Attaching groups to a channel and materializing the members they imply.

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestMaterializeGroupsInChannel_AddsAbsentMembers(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA, userB := model.NewID(), model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedGroupMember(t, db, gr, userB)
	seedChannelGroup(t, db, ch, gr)

	if err := repo.MaterializeGroups(ctx, ch, []string{gr}); err != nil {
		t.Fatalf("MaterializeGroups: %v", err)
	}

	if got := countMembers(t, db, ch); got != 2 {
		t.Errorf("expected 2 members, got %d", got)
	}
	for _, uid := range []string{userA, userB} {
		m := getMember(t, db, ch, uid)
		if m == nil {
			t.Errorf("user %s not materialized", uid)
			continue
		}
		if m.IsDirect {
			t.Errorf("user %s should be is_direct=false, got true", uid)
		}
	}
}

func TestMaterializeGroupsInChannel_LeavesExistingDirectMemberUntouched(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA, userB := model.NewID(), model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedGroupMember(t, db, gr, userB)
	seedChannelGroup(t, db, ch, gr)

	// userA is already a direct member with established per-user state.
	const aMsgCount, aLastViewed = int64(42), int64(1234567890)
	seedChannelMember(t, db, ch, userA, true, aMsgCount, aLastViewed)

	if err := repo.MaterializeGroups(ctx, ch, []string{gr}); err != nil {
		t.Fatalf("MaterializeGroups: %v", err)
	}

	if got := countMembers(t, db, ch); got != 2 {
		t.Errorf("expected 2 members (A direct + B materialized), got %d", got)
	}
	a := getMember(t, db, ch, userA)
	if !a.IsDirect {
		t.Error("userA should still be is_direct=true")
	}
	if a.MsgCount != aMsgCount || a.LastViewedAt != aLastViewed {
		t.Errorf("userA state corrupted: got msg_count=%d last_viewed=%d",
			a.MsgCount, a.LastViewedAt)
	}
	b := getMember(t, db, ch, userB)
	if b.IsDirect {
		t.Error("userB should be materialized (is_direct=false)")
	}
}

func TestMaterializeGroupsInChannel_Idempotent(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedGroupMember(t, db, gr, userA)
	seedChannelGroup(t, db, ch, gr)

	for i := 0; i < 3; i++ {
		if err := repo.MaterializeGroups(ctx, ch, []string{gr}); err != nil {
			t.Fatalf("MaterializeGroups #%d: %v", i, err)
		}
	}
	if got := countMembers(t, db, ch); got != 1 {
		t.Errorf("expected 1 row after repeated materialize, got %d", got)
	}
}

// TestMaterializeGroupsInChannel_MultiGroupOverlappingMembers verifies the
// cross-group union: a user who belongs to two of the attached groups is
// materialized exactly once
func TestMaterializeGroupsInChannel_MultiGroupOverlappingMembers(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)
	userOnly1, userOnly2, userBoth := model.NewID(), model.NewID(), model.NewID()

	seedGroupMember(t, db, g1, userOnly1)
	seedGroupMember(t, db, g1, userBoth)
	seedGroupMember(t, db, g2, userOnly2)
	seedGroupMember(t, db, g2, userBoth)
	seedChannelGroup(t, db, ch, g1)
	seedChannelGroup(t, db, ch, g2)

	if err := repo.MaterializeGroups(ctx, ch, []string{g1, g2}); err != nil {
		t.Fatalf("MaterializeGroups: %v", err)
	}

	// 3 distinct users, no duplicate for userBoth.
	if got := countMembers(t, db, ch); got != 3 {
		t.Errorf("expected 3 distinct members, got %d", got)
	}
	for _, uid := range []string{userOnly1, userOnly2, userBoth} {
		m := getMember(t, db, ch, uid)
		if m == nil {
			t.Errorf("user %s not materialized", uid)
			continue
		}
		if m.IsDirect {
			t.Errorf("user %s should be is_direct=false", uid)
		}
	}
}

// TestMaterializeGroupsInChannel_SkipsSoftDeletedGroupInBatch verifies that a
// soft deleted group mixed into the batch contributes no members, while the
// live group in the same batch still materializes.
func TestMaterializeGroupsInChannel_SkipsSoftDeletedGroupInBatch(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gLive, gDead := seedGroup(t, db, false), seedGroup(t, db, true)
	userLive, userDead := model.NewID(), model.NewID()
	seedGroupMember(t, db, gLive, userLive)
	seedGroupMember(t, db, gDead, userDead)
	seedChannelGroup(t, db, ch, gLive)
	seedChannelGroup(t, db, ch, gDead)

	if err := repo.MaterializeGroups(ctx, ch, []string{gLive, gDead}); err != nil {
		t.Fatalf("MaterializeGroups: %v", err)
	}

	if got := countMembers(t, db, ch); got != 1 {
		t.Errorf("expected only the live group's member, got %d rows", got)
	}
	if m := getMember(t, db, ch, userLive); m == nil {
		t.Error("live group's member should be materialized")
	}
	if m := getMember(t, db, ch, userDead); m != nil {
		t.Error("soft-deleted group's member must not be materialized")
	}
}

func TestAddGroupsToChannel_AttachesBatch(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	g1, g2, g3 := seedGroup(t, db, false), seedGroup(t, db, false), seedGroup(t, db, false)

	if err := repo.AddGroups(ctx, ch, []string{g1, g2, g3}, model.NewID()); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}
	if got := countChannelGroups(t, db, ch); got != 3 {
		t.Errorf("expected 3 channel_groups rows, got %d", got)
	}
}

// TestAddGroupsToChannel_Idempotent verifies a re-attach of an already-present
// group is a no-op (no duplicate row, no 1062 error)
func TestAddGroupsToChannel_Idempotent(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)
	adder := model.NewID()

	if err := repo.AddGroups(ctx, ch, []string{g1}, adder); err != nil {
		t.Fatalf("AddGroups first: %v", err)
	}
	// Re-attach g1 alongside a new g2; g1 must not duplicate.
	if err := repo.AddGroups(ctx, ch, []string{g1, g2}, adder); err != nil {
		t.Fatalf("AddGroups second: %v", err)
	}
	if got := countChannelGroups(t, db, ch); got != 2 {
		t.Errorf("expected 2 channel_groups rows after idempotent re-add, got %d", got)
	}
}

func TestAddAndMaterializeGroups_EndToEnd(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	g1, g2 := seedGroup(t, db, false), seedGroup(t, db, false)
	userA, userB, userShared := model.NewID(), model.NewID(), model.NewID()
	seedGroupMember(t, db, g1, userA)
	seedGroupMember(t, db, g1, userShared)
	seedGroupMember(t, db, g2, userB)
	seedGroupMember(t, db, g2, userShared)

	groupIDs := []string{g1, g2}
	if err := repo.AddGroups(ctx, ch, groupIDs, model.NewID()); err != nil {
		t.Fatalf("AddGroups: %v", err)
	}
	if err := repo.MaterializeGroups(ctx, ch, groupIDs); err != nil {
		t.Fatalf("MaterializeGroups: %v", err)
	}

	if got := countChannelGroups(t, db, ch); got != 2 {
		t.Errorf("expected 2 channel_groups rows, got %d", got)
	}
	// userA, userB, userShared: 3 distinct, userShared not doubled.
	if got := countMembers(t, db, ch); got != 3 {
		t.Errorf("expected 3 materialized members, got %d", got)
	}

	groups, err := repo.GetGroups(ctx, ch)
	if err != nil {
		t.Fatalf("GetGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 enriched groups, got %d", len(groups))
	}
	for _, cg := range groups {
		if cg.Name == "" {
			t.Errorf("group %s missing enriched name", cg.GroupID)
		}
		if cg.MemberCount != 2 {
			t.Errorf("group %s expected member_count 2, got %d", cg.GroupID, cg.MemberCount)
		}
	}
}

func TestMaterializeUserInGroupChannels_MultiUserMultiChannelCrossProduct(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	c1, c2, c3 := seedChannel(t, db), seedChannel(t, db), seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA, userB := model.NewID(), model.NewID()

	// Group attached to all three channels.
	seedChannelGroup(t, db, c1, gr)
	seedChannelGroup(t, db, c2, gr)
	seedChannelGroup(t, db, c3, gr)

	if err := repo.MaterializeUserInGroupChannels(ctx, []string{userA, userB}, gr); err != nil {
		t.Fatalf("MaterializeUserInGroupChannels: %v", err)
	}

	// Expect 6 materialized rows (2 users × 3 channels), all is_direct=false.
	for _, cid := range []string{c1, c2, c3} {
		if got := countMembers(t, db, cid); got != 2 {
			t.Errorf("channel %s: expected 2 members, got %d", cid, got)
		}
		for _, uid := range []string{userA, userB} {
			m := getMember(t, db, cid, uid)
			if m == nil {
				t.Errorf("missing row for channel %s, user %s", cid, uid)
				continue
			}
			if m.IsDirect {
				t.Errorf("channel %s, user %s: should be materialized (is_direct=false)", cid, uid)
			}
		}
	}
}

func TestMaterializeUserInGroupChannels_LeavesExistingRowsAlone(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userDirect, userMaterialized, userFresh := model.NewID(), model.NewID(), model.NewID()
	seedChannelGroup(t, db, ch, gr)

	// userDirect is already a direct member with per-user state.
	const dMsgCount, dLastViewed = int64(42), int64(420)
	directID := seedChannelMember(t, db, ch, userDirect, true, dMsgCount, dLastViewed)

	// userMaterialized already has a materialized row (via some other path).
	const mMsgCount, mLastViewed = int64(7), int64(70)
	materializedID := seedChannelMember(t, db, ch, userMaterialized, false, mMsgCount, mLastViewed)

	if err := repo.MaterializeUserInGroupChannels(ctx,
		[]string{userDirect, userMaterialized, userFresh}, gr); err != nil {
		t.Fatalf("MaterializeUserInGroupChannels: %v", err)
	}

	if got := countMembers(t, db, ch); got != 3 {
		t.Errorf("expected 3 rows (direct + materialized + fresh), got %d", got)
	}

	// userDirect's row must be the same row, unchanged.
	d := getMember(t, db, ch, userDirect)
	if d == nil || d.ID != directID {
		t.Errorf("userDirect's row ID changed: want %s got %+v", directID, d)
	}
	if !d.IsDirect {
		t.Error("userDirect should still be is_direct=true")
	}
	if d.MsgCount != dMsgCount || d.LastViewedAt != dLastViewed {
		t.Errorf("userDirect state corrupted: msg_count=%d last_viewed=%d",
			d.MsgCount, d.LastViewedAt)
	}

	// userMaterialized's row must be the same row, unchanged (still materialized,
	// state preserved). Materialize must NOT promote.
	m := getMember(t, db, ch, userMaterialized)
	if m == nil || m.ID != materializedID {
		t.Errorf("userMaterialized's row ID changed: want %s got %+v", materializedID, m)
	}
	if m.IsDirect {
		t.Error("userMaterialized should NOT be promoted to direct by materialize")
	}
	if m.MsgCount != mMsgCount || m.LastViewedAt != mLastViewed {
		t.Errorf("userMaterialized state corrupted: msg_count=%d last_viewed=%d",
			m.MsgCount, m.LastViewedAt)
	}

	// userFresh got a new materialized row.
	f := getMember(t, db, ch, userFresh)
	if f == nil || f.IsDirect {
		t.Error("userFresh should be inserted as materialized (is_direct=false)")
	}
}

func TestMaterializeUserInGroupChannels_SkipsSoftDeletedGroup(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, true) // deleted
	seedChannelGroup(t, db, ch, gr)

	if err := repo.MaterializeUserInGroupChannels(ctx,
		[]string{model.NewID(), model.NewID()}, gr); err != nil {
		t.Fatalf("MaterializeUserInGroupChannels: %v", err)
	}
	if got := countMembers(t, db, ch); got != 0 {
		t.Errorf("deleted group should not materialize users, got %d rows", got)
	}
}

func TestMaterializeUserInGroupChannels_NoAttachedChannels(t *testing.T) {
	// Group exists but isn't attached to any channel. Should be a no-op,
	// not an error.
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	gr := seedGroup(t, db, false)
	if err := repo.MaterializeUserInGroupChannels(ctx,
		[]string{model.NewID()}, gr); err != nil {
		t.Errorf("expected no-op for unattached group, got %v", err)
	}
}

func TestMaterializeUserInGroupChannels_Idempotent(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	gr := seedGroup(t, db, false)
	userA := model.NewID()
	seedChannelGroup(t, db, ch, gr)

	for i := 0; i < 3; i++ {
		if err := repo.MaterializeUserInGroupChannels(ctx, []string{userA}, gr); err != nil {
			t.Fatalf("MaterializeUserInGroupChannels #%d: %v", i, err)
		}
	}
	if got := countMembers(t, db, ch); got != 1 {
		t.Errorf("expected 1 row after repeated materialize, got %d", got)
	}
}
