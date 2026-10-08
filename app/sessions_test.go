// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import "testing"

func TestSessionHandleHidesTheAccessToken(t *testing.T) {
	const token = "58b11234-dccb-436f-9fd8-87480371abd9"

	handle := sessionHandle(token)
	if handle == token {
		t.Fatal("handle must not be the access token")
	}
	if handle != sessionHandle(token) {
		t.Error("handle must be stable for the same token")
	}
	if handle == sessionHandle(token+"x") {
		t.Error("different tokens must not share a handle")
	}
}
