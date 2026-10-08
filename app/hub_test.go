// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"sync"
	"testing"

	"github.com/twigex/twigex/model"
)

func TestHubSurvivesReadsDuringRegistration(t *testing.T) {
	hub := newHub()
	a := &App{Server: Server{NotificationHub: hub}}
	hub.app = a

	event := model.WebsocketEvent{Event: model.NOTIFICATION_CHANNEL_PRESENCE, App: model.AppChat}
	absent := model.User{ID: "nobody"}

	concurrent := map[string]func() error{
		"register": func() error {
			hub.register <- &Client{hub: hub, user: "connected", send: make(chan []byte, 256)}
			return nil
		},
		"SendEvent":             func() error { return hub.SendEvent(event) },
		"sendPushNotifications": func() error { return a.sendPushNotifications(absent, event) },
		"connectedUsers":        func() error { hub.connectedUsers(); return nil },
	}

	var wg sync.WaitGroup

	for name, op := range concurrent {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 100 {
				if err := op(); err != nil {
					t.Error(name+":", err)
					return
				}
			}
		}()
	}

	wg.Wait()
}

func TestConnectedUsersReportsEachUserOnce(t *testing.T) {
	hub := &Hub{clients: map[*Client]bool{
		{user: "ada"}:   true,
		{user: "ada"}:   true,
		{user: "grace"}: true,
	}}

	if got := hub.connectedUsers(); len(got) != 2 {
		t.Errorf("got %v, want one entry each for ada and grace", got)
	}
}
