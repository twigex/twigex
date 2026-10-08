// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Seeding the participant cache stands in for the LiveKit lookup, which cannot
// be faked here, and a cache miss would reach a nil client rather than pass
// silently.

func appWithRoomParticipants(room string, identities ...string) *App {
	a := &App{LiveKit: &liveKitRooms{}}
	seedRoomParticipants(a, room, identities...)
	return a
}

func seedRoomParticipants(a *App, room string, identities ...string) {
	set := make(map[string]struct{}, len(identities))
	for _, id := range identities {
		set[id] = struct{}{}
	}
	entry := a.LiveKit.roomFor(room)
	entry.identities = set
	entry.loaded = true
	entry.fetchedAt = time.Now()

}

func TestRoomFor_EvictsRoomsPastRetention(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1")
	a.LiveKit.roomFor("meeting-1").touchedAt = time.Now().Add(-2 * roomParticipantRetention)

	a.LiveKit.roomFor("meeting-2")

	if _, ok := a.LiveKit.rooms["meeting-1"]; ok {
		t.Fatal("expected the stale room to be evicted")
	}
}

// The cleanup goes through every entry while other rooms are being refilled,
// so the two must not share a field. Passes without -race either way.
func TestRoomFor_CleanupDoesNotRaceWithRefill(t *testing.T) {
	a := &App{LiveKit: &liveKitRooms{}}

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			entry := a.LiveKit.roomFor(fmt.Sprintf("meeting-%d", i%3))

			entry.mu.Lock()
			defer entry.mu.Unlock()
			entry.identities = map[string]struct{}{"user-1": {}}
			entry.loaded = true
			entry.fetchedAt = time.Now()
		}(i)
	}
	wg.Wait()
}
