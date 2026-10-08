// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/store"
)

// A ring fans out to every session a user has open, so whatever resolves it must
// reach the callee's other sessions too or they keep ringing.

const (
	callChannelID = "channel-dm"
	callerID      = "user-caller"
	calleeID      = "user-callee"
)

type fakeCallChannelStore struct {
	store.ChannelStore
	members []model.ChannelMember
}

func (f *fakeCallChannelStore) Get(id string) (*model.Channel, error) {
	return &model.Channel{ID: id, Type: model.ChannelTypeDirect}, nil
}

func (f *fakeCallChannelStore) IsMember(_, _ string) (bool, error) { return true, nil }

func (f *fakeCallChannelStore) GetMembers(_ string) ([]model.ChannelMember, error) {
	return f.members, nil
}

func (f *fakeCallChannelStore) CreateMeeting(_ context.Context, _ model.ChannelMeeting) error {
	return nil
}

func callApp(sessions map[string]int) (*App, map[string][]*Client) {
	hub := &Hub{clients: make(map[*Client]bool)}
	clients := make(map[string][]*Client, len(sessions))
	for user, count := range sessions {
		for range count {
			c := &Client{hub: hub, user: user, send: make(chan []byte, 16)}
			hub.clients[c] = true
			clients[user] = append(clients[user], c)
		}
	}
	a := &App{
		Server: Server{NotificationHub: hub},
		Store: store.Store{
			Channels: &fakeCallChannelStore{members: []model.ChannelMember{
				{UserID: callerID, ChannelID: callChannelID},
				{UserID: calleeID, ChannelID: callChannelID},
			}},
		},
	}
	return a, clients
}

// Notifications are dispatched in goroutines, so wait for the socket to go quiet.
func drainEvents(t *testing.T, c *Client) []string {
	t.Helper()
	events := []string{}
	for {
		select {
		case raw := <-c.send:
			var ev model.WebsocketEvent
			if err := json.Unmarshal(raw, &ev); err != nil {
				t.Fatalf("unmarshal websocket event: %v", err)
			}
			events = append(events, ev.Event)
		case <-time.After(200 * time.Millisecond):
			return events
		}
	}
}

func requireEvents(t *testing.T, c *Client, label string, want ...string) {
	t.Helper()
	got := drainEvents(t, c)
	if len(got) != len(want) {
		t.Fatalf("%s: expected events %v, got %v", label, want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s: expected events %v, got %v", label, want, got)
		}
	}
}

func TestStartDirectCall_RingsEveryCalleeSession(t *testing.T) {
	a, clients := callApp(map[string]int{callerID: 2, calleeID: 2})

	result, appErr := a.StartDirectCall(context.Background(), model.User{ID: callerID}, callChannelID)
	if appErr != nil {
		t.Fatalf("StartDirectCall: %v", appErr)
	}
	if result.Status != model.DirectCallRinging {
		t.Fatalf("expected status %q, got %q", model.DirectCallRinging, result.Status)
	}

	for _, c := range clients[calleeID] {
		requireEvents(t, c, "callee tab", model.NOTIFICATION_CHANNEL_INCOMING_CALL)
	}
	for _, c := range clients[callerID] {
		requireEvents(t, c, "caller tab")
	}
}

func TestAnswerDirectCall_ClearsRingInCalleesOtherSessions(t *testing.T) {
	a, clients := callApp(map[string]int{callerID: 2, calleeID: 2})

	if _, appErr := a.StartDirectCall(context.Background(), model.User{ID: callerID}, callChannelID); appErr != nil {
		t.Fatalf("StartDirectCall: %v", appErr)
	}
	for _, c := range clients[calleeID] {
		drainEvents(t, c)
	}
	for _, c := range clients[callerID] {
		drainEvents(t, c)
	}

	result, appErr := a.AnswerDirectCall(context.Background(), model.User{ID: calleeID}, callChannelID)
	if appErr != nil {
		t.Fatalf("AnswerDirectCall: %v", appErr)
	}
	if result.Status != model.DirectCallConnected {
		t.Fatalf("expected status %q, got %q", model.DirectCallConnected, result.Status)
	}

	for _, c := range clients[calleeID] {
		requireEvents(t, c, "callee tab", model.NOTIFICATION_CHANNEL_CALL_HANDLED)
	}
	for _, c := range clients[callerID] {
		requireEvents(t, c, "caller tab", model.NOTIFICATION_CHANNEL_CALL_CONNECTED)
	}
}

func TestDeclineDirectCall_ClearsRingInCalleesOtherSessions(t *testing.T) {
	a, clients := callApp(map[string]int{callerID: 1, calleeID: 2})

	if _, appErr := a.StartDirectCall(context.Background(), model.User{ID: callerID}, callChannelID); appErr != nil {
		t.Fatalf("StartDirectCall: %v", appErr)
	}
	for _, c := range clients[calleeID] {
		drainEvents(t, c)
	}
	drainEvents(t, clients[callerID][0])

	if appErr := a.DeclineDirectCall(context.Background(), model.User{ID: calleeID}, callChannelID); appErr != nil {
		t.Fatalf("DeclineDirectCall: %v", appErr)
	}

	for _, c := range clients[calleeID] {
		requireEvents(t, c, "callee tab", model.NOTIFICATION_CHANNEL_CALL_HANDLED)
	}
	requireEvents(t, clients[callerID][0], "caller tab", model.NOTIFICATION_CHANNEL_CALL_DECLINED)
}

// Glare: the callee presses Call instead of Accept, which becomes an answer.
func TestStartDirectCall_GlareClearsRingInAnswerersOtherSessions(t *testing.T) {
	a, clients := callApp(map[string]int{callerID: 1, calleeID: 2})

	if _, appErr := a.StartDirectCall(context.Background(), model.User{ID: callerID}, callChannelID); appErr != nil {
		t.Fatalf("StartDirectCall: %v", appErr)
	}
	for _, c := range clients[calleeID] {
		drainEvents(t, c)
	}
	drainEvents(t, clients[callerID][0])

	result, appErr := a.StartDirectCall(context.Background(), model.User{ID: calleeID}, callChannelID)
	if appErr != nil {
		t.Fatalf("StartDirectCall (glare): %v", appErr)
	}
	if result.Status != model.DirectCallConnected {
		t.Fatalf("expected status %q, got %q", model.DirectCallConnected, result.Status)
	}

	for _, c := range clients[calleeID] {
		requireEvents(t, c, "callee tab", model.NOTIFICATION_CHANNEL_CALL_HANDLED)
	}
	requireEvents(t, clients[callerID][0], "caller tab", model.NOTIFICATION_CHANNEL_CALL_CONNECTED)
}

func TestExpireDirectCall_ClearsRingInEverySession(t *testing.T) {
	a, clients := callApp(map[string]int{callerID: 2, calleeID: 2})

	if _, appErr := a.StartDirectCall(context.Background(), model.User{ID: callerID}, callChannelID); appErr != nil {
		t.Fatalf("StartDirectCall: %v", appErr)
	}
	for _, c := range clients[calleeID] {
		drainEvents(t, c)
	}
	for _, c := range clients[callerID] {
		drainEvents(t, c)
	}

	a.expireDirectCall(callChannelID, callerID)

	for _, c := range clients[calleeID] {
		requireEvents(t, c, "callee tab", model.NOTIFICATION_CHANNEL_CALL_CANCELLED)
	}
	for _, c := range clients[callerID] {
		requireEvents(t, c, "caller tab", model.NOTIFICATION_CHANNEL_CALL_CANCELLED)
	}
}
