// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"sort"
	"strings"
	"testing"

	"github.com/twigex/twigex/model"
)

func memberUsersByChannel(t *testing.T, channels []model.Channel) map[string]string {
	t.Helper()
	got := make(map[string]string, len(channels))
	for _, ch := range channels {
		if ch.ChannelMembers == nil {
			t.Errorf("channel %s has nil members, want an empty list", ch.ID)
		}
		ids := make([]string, 0, len(ch.ChannelMembers))
		for _, m := range ch.ChannelMembers {
			if m.ChannelID != ch.ID || m.UserInfo == nil || m.UserInfo.ID != m.UserID {
				t.Errorf("channel %s got a member row for channel %s", ch.ID, m.ChannelID)
			}
			ids = append(ids, m.UserID)
		}
		sort.Strings(ids)
		got[ch.ID] = strings.Join(ids, ",")
	}
	return got
}

func sortedJoin(ids ...string) string {
	sort.Strings(ids)
	return strings.Join(ids, ",")
}

func TestChannelListsAttachEachChannelsMembers(t *testing.T) {
	db := requireDB(t)
	cleanTables(t, "channels", "channel_members", "users")
	repo := &channelsRepository{Db: db}

	me, bob, carol, dave := seedUser(t, db), seedUser(t, db), seedUser(t, db), seedUser(t, db)

	direct := seedChannel(t, db)
	seedChannelMember(t, db, direct, me, true, 0, 0)
	seedChannelMember(t, db, direct, bob, true, 0, 0)

	viaGroup := seedChannel(t, db)
	seedChannelMember(t, db, viaGroup, me, false, 0, 0)
	seedChannelMember(t, db, viaGroup, carol, true, 0, 0)
	seedChannelMember(t, db, viaGroup, dave, false, 0, 0)

	empty := seedChannel(t, db)

	deleted := seedChannel(t, db)
	seedChannelMember(t, db, deleted, me, true, 0, 0)
	mustExec(t, `UPDATE channels SET deleted_at = 1 WHERE id = ?`, deleted)

	mine, err := repo.GetAllForUser(me)
	if err != nil {
		t.Fatalf("GetAllForUser: %v", err)
	}
	wantMine := map[string]string{
		direct:   sortedJoin(me, bob),
		viaGroup: sortedJoin(me, carol),
	}
	if got := memberUsersByChannel(t, mine); len(got) != len(wantMine) || got[direct] != wantMine[direct] || got[viaGroup] != wantMine[viaGroup] {
		t.Errorf("GetAllForUser members = %v, want %v", got, wantMine)
	}

	public, err := repo.GetAllByType(model.ChannelTypePublic)
	if err != nil {
		t.Fatalf("GetAllByType: %v", err)
	}
	wantPublic := map[string]string{
		direct:   sortedJoin(me, bob),
		viaGroup: carol,
		empty:    "",
	}
	got := memberUsersByChannel(t, public)
	if len(got) != len(wantPublic) {
		t.Fatalf("GetAllByType returned %d channels, want %d", len(got), len(wantPublic))
	}
	for id, want := range wantPublic {
		if got[id] != want {
			t.Errorf("GetAllByType channel %s members = %q, want %q", id, got[id], want)
		}
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	wantAll := map[string]string{
		direct:   sortedJoin(me, bob),
		viaGroup: carol,
		empty:    "",
		deleted:  me,
	}
	gotAll := memberUsersByChannel(t, all)
	if len(gotAll) != len(wantAll) {
		t.Fatalf("GetAll returned %d channels, want %d", len(gotAll), len(wantAll))
	}
	for id, want := range wantAll {
		if gotAll[id] != want {
			t.Errorf("GetAll channel %s members = %q, want %q", id, gotAll[id], want)
		}
	}
}
