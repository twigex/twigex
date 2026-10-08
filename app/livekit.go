// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/twigex/twigex/model"
	"github.com/twigex/twigex/tlog"
)

type liveKitRooms struct {
	*lksdk.RoomServiceClient

	mu    sync.Mutex
	rooms map[string]*roomParticipants
}

// touchedAt is guarded by liveKitRooms.mu, everything below it by mu, so the
// cleanup never has to read an entry that a lookup is still filling in.
type roomParticipants struct {
	touchedAt time.Time

	mu         sync.Mutex
	identities map[string]struct{}
	loaded     bool
	fetchedAt  time.Time
}

const (
	// A guest asks for every member's photo at once on join, which without this
	// would cost a round trip to LiveKit per tile.
	roomParticipantTTL = 3 * time.Second

	// Nothing removes a room on its own, so drop the ones untouched for long
	// enough that their meeting is over.
	roomParticipantRetention = time.Minute

	// The room's lock is held across the lookup, so a call with no timeout
	// would block every later reader of that room rather than just itself.
	roomParticipantTimeout = 10 * time.Second
)

func (l *liveKitRooms) roomFor(room string) *roomParticipants {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.rooms == nil {
		l.rooms = make(map[string]*roomParticipants)
	}

	for id, entry := range l.rooms {
		if time.Since(entry.touchedAt) > roomParticipantRetention {
			delete(l.rooms, id)
		}
	}

	entry, ok := l.rooms[room]
	if !ok {
		entry = &roomParticipants{}
		l.rooms[room] = entry
	}

	entry.touchedAt = time.Now()

	return entry
}

// Present reports whether an identity is in the room, from a listing that can
// be up to roomParticipantTTL out of date.
func (l *liveKitRooms) Present(ctx context.Context, room, identity string) (bool, *model.AppError) {
	entry := l.roomFor(room)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if appErr := l.refreshLocked(ctx, entry, room); appErr != nil {
		return false, appErr
	}

	_, ok := entry.identities[identity]

	return ok, nil
}

// Count uses the same listing as Present.
func (l *liveKitRooms) Count(ctx context.Context, room string) (int, *model.AppError) {
	entry := l.roomFor(room)

	entry.mu.Lock()
	defer entry.mu.Unlock()

	if appErr := l.refreshLocked(ctx, entry, room); appErr != nil {
		return 0, appErr
	}

	return len(entry.identities), nil
}

// refreshLocked expects entry.mu, and keeps it across the call so that several
// requests for one room share a single lookup instead of each making its own.
func (l *liveKitRooms) refreshLocked(ctx context.Context, entry *roomParticipants, room string) *model.AppError {
	if entry.loaded && time.Since(entry.fetchedAt) < roomParticipantTTL {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, roomParticipantTimeout)
	defer cancel()

	resp, err := l.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: room})
	if err != nil {
		tlog.Errorw("Failed to retrieve video participants", "room", room, "error", err)
		return model.NewAppError("channel.video_participants_failed", http.StatusInternalServerError)
	}

	identities := make(map[string]struct{}, len(resp.Participants))
	for _, p := range resp.Participants {
		identities[p.Identity] = struct{}{}
	}

	entry.identities = identities
	entry.loaded = true
	entry.fetchedAt = time.Now()

	return nil
}

// longestPresentMember skips the cache: promoting someone who has already left
// would leave the meeting without a host, and promotions are too rare to be
// worth saving a lookup on.
func (a *App) longestPresentMember(ctx context.Context, meetingID, excludeUserID string) string {
	resp, err := a.LiveKit.ListParticipants(ctx, &livekit.ListParticipantsRequest{Room: meetingID})
	if err != nil {
		tlog.Errorw("Failed to list participants for host auto-promote", "meeting_id", meetingID, "error", err)
		return ""
	}

	best := ""
	var bestJoined int64
	for _, p := range resp.Participants {
		if p.Identity == excludeUserID || strings.HasPrefix(p.Identity, "guest-") {
			continue
		}

		if best == "" || p.JoinedAt < bestJoined {
			best = p.Identity
			bestJoined = p.JoinedAt
		}
	}

	return best
}

func (a *App) mintVideoToken(room, identity, name string, validFor time.Duration) (string, *model.AppError) {
	canPublish := true
	canSubscribe := true
	at := auth.NewAccessToken(*a.ConfigStore.Config.ChannelSettings.Key, *a.ConfigStore.Config.ChannelSettings.Secret)
	grant := &auth.VideoGrant{
		RoomJoin:     true,
		Room:         room,
		CanPublish:   &canPublish,
		CanSubscribe: &canSubscribe,
	}
	at.AddGrant(grant).SetIdentity(identity).SetName(name).SetValidFor(validFor)
	s, err := at.ToJWT()
	if err != nil {
		tlog.Errorw("Failed to mint video token", "room", room, "error", err)
		return "", model.NewAppError("channel.video_token_failed", http.StatusInternalServerError)
	}

	return s, nil
}

func (a *App) videoParticipantCount(room string) (int, *model.AppError) {
	return a.LiveKit.Count(context.Background(), room)
}

func (a *App) NormalizeLiveKitHost(raw string, proto model.Protocol) string {
	host := strings.TrimSpace(raw)
	host = strings.TrimRight(host, "/")

	insecure := strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "ws://")

	// Remove existing scheme if present
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "wss://")
	host = strings.TrimPrefix(host, "ws://")

	switch proto {
	case model.ProtocolHTTPS:
		if insecure {
			return "http://" + host
		}

		return "https://" + host
	case model.ProtocolWSS:
		if insecure {
			return "ws://" + host
		}

		return "wss://" + host
	default:
		return host
	}
}
