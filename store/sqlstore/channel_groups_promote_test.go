// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// Direct membership: inserting it, and promoting a materialized row to direct.

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestInsertOrPromoteChannelMembers_FreshInsert(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	userA, userB := model.NewID(), model.NewID()

	err := repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{
		newDirectMember(ch, userA),
		newDirectMember(ch, userB),
	})
	if err != nil {
		t.Fatalf("CreateOrPromoteMembers: %v", err)
	}

	if got := countMembers(t, db, ch); got != 2 {
		t.Errorf("expected 2 rows, got %d", got)
	}
	for _, uid := range []string{userA, userB} {
		m := getMember(t, db, ch, uid)
		if m == nil || !m.IsDirect {
			t.Errorf("user %s should be inserted with is_direct=true", uid)
		}
	}
}

func TestInsertOrPromoteChannelMembers_PromotesMaterializedAndPreservesState(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	userA := model.NewID()
	const msgCount, lastViewed = int64(99), int64(987654321)
	seededID := seedChannelMember(t, db, ch, userA, false, msgCount, lastViewed)
	seededUpdatedAt := getMember(t, db, ch, userA).UpdatedAt

	if err := repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{
		newDirectMember(ch, userA),
	}); err != nil {
		t.Fatalf("CreateOrPromoteMembers: %v", err)
	}

	m := getMember(t, db, ch, userA)
	if m == nil {
		t.Fatal("row disappeared after promote")
	}
	if m.ID != seededID {
		t.Errorf("row ID changed during promote: want %s got %s — likely DELETE+INSERT instead of UPDATE",
			seededID, m.ID)
	}
	if !m.IsDirect {
		t.Error("expected is_direct=true after promote")
	}
	if m.MsgCount != msgCount {
		t.Errorf("msg_count changed: want %d got %d", msgCount, m.MsgCount)
	}
	if m.LastViewedAt != lastViewed {
		t.Errorf("last_viewed_at changed: want %d got %d", lastViewed, m.LastViewedAt)
	}
	if m.UpdatedAt <= seededUpdatedAt {
		t.Errorf("updated_at not bumped on promote: want >%d got %d", seededUpdatedAt, m.UpdatedAt)
	}
	if got := countMembers(t, db, ch); got != 1 {
		t.Errorf("promote must not create a second row, got %d total", got)
	}
}

func TestInsertOrPromoteChannelMembers_AlreadyDirectIsNoOp(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	userA := model.NewID()
	const msgCount, lastViewed = int64(50), int64(123)
	seededID := seedChannelMember(t, db, ch, userA, true, msgCount, lastViewed)
	seededUpdatedAt := getMember(t, db, ch, userA).UpdatedAt

	if err := repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{
		newDirectMember(ch, userA),
	}); err != nil {
		t.Fatalf("CreateOrPromoteMembers: %v", err)
	}

	m := getMember(t, db, ch, userA)
	if m == nil || !m.IsDirect {
		t.Fatal("row should still exist as direct")
	}
	if m.ID != seededID {
		t.Errorf("row ID changed: want %s got %s. Likely DELETE+INSERT instead of UPDATE",
			seededID, m.ID)
	}
	if m.MsgCount != msgCount || m.LastViewedAt != lastViewed {
		t.Errorf("already-direct row's state was clobbered: msg_count=%d last_viewed=%d",
			m.MsgCount, m.LastViewedAt)
	}
	if m.UpdatedAt <= seededUpdatedAt {
		t.Errorf("updated_at not bumped on already-direct re-add: want >%d got %d",
			seededUpdatedAt, m.UpdatedAt)
	}
}

func TestInsertOrPromoteChannelMembers_MixedBatch(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	ch := seedChannel(t, db)
	userFresh := model.NewID()
	userMaterialized := model.NewID()
	userDirect := model.NewID()

	const matMsgCount, matLastViewed = int64(7), int64(70)
	const dirMsgCount, dirLastViewed = int64(8), int64(80)
	seedChannelMember(t, db, ch, userMaterialized, false, matMsgCount, matLastViewed)
	seedChannelMember(t, db, ch, userDirect, true, dirMsgCount, dirLastViewed)

	err := repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{
		newDirectMember(ch, userFresh),
		newDirectMember(ch, userMaterialized),
		newDirectMember(ch, userDirect),
	})
	if err != nil {
		t.Fatalf("CreateOrPromoteMembers: %v", err)
	}

	if got := countMembers(t, db, ch); got != 3 {
		t.Errorf("expected 3 rows (one per user), got %d", got)
	}

	fresh := getMember(t, db, ch, userFresh)
	if fresh == nil || !fresh.IsDirect {
		t.Error("fresh user should be inserted as direct")
	}

	mat := getMember(t, db, ch, userMaterialized)
	if mat == nil || !mat.IsDirect {
		t.Error("materialized user should be promoted to direct")
	}
	if mat.MsgCount != matMsgCount || mat.LastViewedAt != matLastViewed {
		t.Errorf("promoted user's state lost: msg_count=%d last_viewed=%d",
			mat.MsgCount, mat.LastViewedAt)
	}

	dir := getMember(t, db, ch, userDirect)
	if dir == nil || !dir.IsDirect {
		t.Error("already-direct user should remain direct")
	}
	if dir.MsgCount != dirMsgCount || dir.LastViewedAt != dirLastViewed {
		t.Errorf("already-direct user's state clobbered: msg_count=%d last_viewed=%d",
			dir.MsgCount, dir.LastViewedAt)
	}
}

// TestInsertOrPromoteChannelMembers_CrossChannelBatch verifies the store
// contract is genuinely batch-per-batch, not batch-per-channel: members for
// different channels can be inserted/promoted in one statement. Current
// callers happen to pass single-channel batches, but the contract is broader.
func TestInsertOrPromoteChannelMembers_CrossChannelBatch(t *testing.T) {
	db := requireDB(t)
	resetCascadeTables(t)

	repo := &channelsRepository{Db: db}
	ctx := context.Background()

	c1, c2 := seedChannel(t, db), seedChannel(t, db)
	userA, userB := model.NewID(), model.NewID()

	err := repo.CreateOrPromoteMembers(ctx, []model.ChannelMember{
		newDirectMember(c1, userA),
		newDirectMember(c2, userA),
		newDirectMember(c1, userB),
	})
	if err != nil {
		t.Fatalf("CreateOrPromoteMembers: %v", err)
	}

	if got := countMembers(t, db, c1); got != 2 {
		t.Errorf("c1: expected 2 rows, got %d", got)
	}
	if got := countMembers(t, db, c2); got != 1 {
		t.Errorf("c2: expected 1 row, got %d", got)
	}
	if m := getMember(t, db, c1, userA); m == nil || !m.IsDirect {
		t.Error("c1/userA missing or not direct")
	}
	if m := getMember(t, db, c2, userA); m == nil || !m.IsDirect {
		t.Error("c2/userA missing or not direct")
	}
	if m := getMember(t, db, c1, userB); m == nil || !m.IsDirect {
		t.Error("c1/userB missing or not direct")
	}
	// userB was not inserted into c2.
	if m := getMember(t, db, c2, userB); m != nil {
		t.Error("c2/userB should not exist — batch must route by channel_id per row")
	}
}
