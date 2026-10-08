// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"testing"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

type fakeMeetingAudienceStore struct {
	store.ChannelStore
	members      []string
	groupMembers []string
}

func (f *fakeMeetingAudienceStore) GetMembers(string) ([]model.ChannelMember, error) {
	members := make([]model.ChannelMember, 0, len(f.members))
	for _, id := range f.members {
		members = append(members, model.ChannelMember{UserID: id})
	}

	return members, nil
}

func (f *fakeMeetingAudienceStore) GetMeetingGroupMemberIDs(context.Context, string) ([]string, error) {
	return f.groupMembers, nil
}

func meetingListsApp(channels *fakeMeetingAudienceStore, connected ...string) (*App, map[string]*Client) {
	hub := &Hub{clients: map[*Client]bool{}}
	clients := map[string]*Client{}
	for _, id := range connected {
		c := &Client{hub: hub, user: id, send: make(chan []byte, 4)}
		hub.clients[c] = true
		clients[id] = c
	}

	a := &App{Store: store.Store{Channels: channels}}
	a.Server.NotificationHub = hub

	return a, clients
}

func TestRefreshMeetingListsReachesEveryChannelMemberForOpenMeeting(t *testing.T) {
	channels := &fakeMeetingAudienceStore{members: []string{"host", "member", "other"}}
	a, clients := meetingListsApp(channels, "host", "member", "other", "outsider")

	a.RefreshMeetingLists(model.ChannelMeeting{ID: "m1", HostID: "host"}, "c1", nil)

	for _, id := range []string{"host", "member", "other"} {
		got := received(clients[id])
		if got == nil || got["notificationType"] != model.NOTIFICATION_CHANNEL_MEETINGS_REFRESH {
			t.Errorf("%s: expected a meetings_refresh event, got %v", id, got)
		}
	}

	if got := received(clients["outsider"]); got != nil {
		t.Errorf("outsider: expected nothing, got %v", got)
	}
}

func TestRefreshMeetingListsReachesOnlyInviteesForInviteOnlyMeeting(t *testing.T) {
	channels := &fakeMeetingAudienceStore{
		members:      []string{"host", "invitee", "grouped", "member"},
		groupMembers: []string{"grouped"},
	}
	a, clients := meetingListsApp(channels, "host", "invitee", "grouped", "member")

	meeting := model.ChannelMeeting{
		ID:         "m1",
		HostID:     "host",
		InviteOnly: true,
		Invitees:   []string{"invitee"},
	}
	a.RefreshMeetingLists(meeting, "c1", nil)

	for _, id := range []string{"host", "invitee", "grouped"} {
		if received(clients[id]) == nil {
			t.Errorf("%s: expected a meetings_refresh event", id)
		}
	}

	if got := received(clients["member"]); got != nil {
		t.Errorf("uninvited member: expected nothing, got %v", got)
	}
}

func TestRefreshMeetingListsReachesFormerAudienceOnce(t *testing.T) {
	channels := &fakeMeetingAudienceStore{members: []string{"host"}}
	a, clients := meetingListsApp(channels, "host", "removed")

	meeting := model.ChannelMeeting{ID: "m1", HostID: "host", InviteOnly: true}
	a.RefreshMeetingLists(meeting, "c1", []string{"removed", "host"})

	if received(clients["removed"]) == nil {
		t.Error("removed: expected a meetings_refresh event so the meeting leaves their list")
	}

	if received(clients["host"]) == nil {
		t.Fatal("host: expected a meetings_refresh event")
	}
	if got := received(clients["host"]); got != nil {
		t.Errorf("host: expected one event, got a second %v", got)
	}
}
