// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/http"
	"testing"
)

// ResolveMeetingParticipant is the gate on the guest-facing photo endpoint: a
// guest may reach only someone who is in the meeting with them.

func TestResolveMeetingParticipant_ResolvesMemberInRoom(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1", "guest-abc")

	userID, appErr := a.ResolveMeetingParticipant(context.Background(), "meeting-1", "user-1")
	if appErr != nil {
		t.Fatalf("expected user-1 to resolve, got %v", appErr)
	}
	if userID != "user-1" {
		t.Fatalf("expected user-1, got %q", userID)
	}
}

func TestResolveMeetingParticipant_RejectsMemberNotInRoom(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1")

	// A real user who is not in this meeting: resolving them would turn one
	// guest link into a read of any member's photo.
	_, appErr := a.ResolveMeetingParticipant(context.Background(), "meeting-1", "user-2")
	requireAppErr(t, appErr, "user.not_found", http.StatusNotFound)
}

func TestResolveMeetingParticipant_DoesNotCrossMeetings(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1")
	seedRoomParticipants(a, "meeting-2", "user-2")

	// An identity learned in another meeting, reused against this link.
	_, appErr := a.ResolveMeetingParticipant(context.Background(), "meeting-1", "user-2")
	requireAppErr(t, appErr, "user.not_found", http.StatusNotFound)
}

func TestResolveMeetingParticipant_RejectsGuestIdentity(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1", "guest-abc")

	_, appErr := a.ResolveMeetingParticipant(context.Background(), "meeting-1", "guest-abc")
	requireAppErr(t, appErr, "user.not_found", http.StatusNotFound)
}

func TestResolveMeetingParticipant_RejectsEmptyIdentity(t *testing.T) {
	a := appWithRoomParticipants("meeting-1", "user-1")

	_, appErr := a.ResolveMeetingParticipant(context.Background(), "meeting-1", "")
	requireAppErr(t, appErr, "user.not_found", http.StatusNotFound)
}
