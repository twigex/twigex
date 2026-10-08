// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeInviteOnlyStore struct {
	store.ChannelStore
	groups  []model.ChannelMeetingGroup
	updates []bool
}

func (f *fakeInviteOnlyStore) GetMeetingGroups(context.Context, string) ([]model.ChannelMeetingGroup, error) {
	return f.groups, nil
}

func (f *fakeInviteOnlyStore) UpdateMeetingInviteOnly(_ context.Context, _ string, inviteOnly bool) error {
	f.updates = append(f.updates, inviteOnly)
	return nil
}

func TestSyncMeetingInviteOnly(t *testing.T) {
	cases := []struct {
		name     string
		was      bool
		invitees []string
		groups   []model.ChannelMeetingGroup
		want     bool
		writes   int
	}{
		{"open meeting gets an invitee", false, []string{"u1"}, nil, true, 1},
		{"open meeting gets a group", false, nil, []model.ChannelMeetingGroup{{}}, true, 1},
		{"last invitee and group removed", true, nil, nil, false, 1},
		{"still invite only", true, []string{"u1"}, nil, true, 0},
		{"still open", false, nil, nil, false, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			channels := &fakeInviteOnlyStore{groups: tc.groups}
			a := &App{Store: store.Store{Channels: channels}}
			meeting := &model.ChannelMeeting{ID: "m1", InviteOnly: tc.was, Invitees: tc.invitees}

			if err := a.syncMeetingInviteOnly(context.Background(), meeting); err != nil {
				t.Fatalf("syncMeetingInviteOnly: %v", err)
			}

			if meeting.InviteOnly != tc.want {
				t.Errorf("want InviteOnly %v, got %v", tc.want, meeting.InviteOnly)
			}

			if len(channels.updates) != tc.writes {
				t.Errorf("want %d stored updates, got %d", tc.writes, len(channels.updates))
			}
		})
	}
}
