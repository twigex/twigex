// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"strings"
	"testing"
)

func TestIsValidReactionAcceptsPickerShortcodes(t *testing.T) {
	for _, s := range []string{":smile:", ":+1:", ":-1:", ":thumbsup:", ":flag-lv:", ":u5272:"} {
		if !IsValidReaction(s) {
			t.Errorf("%q should be valid", s)
		}
	}
}

func TestIsValidReactionRejectsMarkup(t *testing.T) {
	for _, s := range []string{
		"",
		"smile",
		":smile",
		"smile:",
		"::",
		":sm ile:",
		":SMILE:",
		"<img src=x onerror=alert(1)>",
		":smile:<script>alert(1)</script>",
		":" + strings.Repeat("a", 65) + ":",
	} {
		if IsValidReaction(s) {
			t.Errorf("%q should be rejected", s)
		}
	}
}

func TestValidPendingPostIDKeepsAnIDWithinTheLimit(t *testing.T) {
	id := strings.Repeat("a", MaxPendingPostIDLength)

	if got := (NewPost{PendingPostID: id}).ValidPendingPostID(); got != id {
		t.Errorf("ValidPendingPostID() = %q, want %q", got, id)
	}
}

func TestValidPendingPostIDDropsAnIDOverTheLimit(t *testing.T) {
	id := strings.Repeat("a", MaxPendingPostIDLength+1)

	if got := (NewPost{PendingPostID: id}).ValidPendingPostID(); got != "" {
		t.Errorf("ValidPendingPostID() = %q, want empty", got)
	}
}
