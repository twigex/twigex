// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"testing"
)

func TestEncodeLDAPIDKeepsTextIdentifiersReadable(t *testing.T) {
	uuid := "3b319c96-358a-1041-8481-bdf53f6d2a8a"
	if got := EncodeLDAPID([]byte(uuid)); got != uuid {
		t.Errorf("entryUUID must pass through unchanged: got %q", got)
	}
}

func TestEncodeLDAPIDHexEncodesBinaryIdentifiers(t *testing.T) {
	guid := []byte{0xa4, 0xf2, 0xb8, 0xc1, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb}

	if got := EncodeLDAPID(guid); got != "a4f2b8c100112233445566778899aabb" {
		t.Errorf("objectGUID: got %q", got)
	}
}

func TestEncodeLDAPIDEmptyIsEmpty(t *testing.T) {
	if got := EncodeLDAPID(nil); got != "" {
		t.Errorf("want empty, got %q", got)
	}
}
