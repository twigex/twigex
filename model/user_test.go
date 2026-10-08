// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"strings"
	"testing"
)

func TestNormalizeUsernameFoldsAndTrims(t *testing.T) {
	cases := map[string]string{
		"Jane":       "jane",
		"  jane  ":   "jane",
		"JANE.DOE":   "jane.doe",
		"jane":       "jane",
		"\tjsmith\n": "jsmith",
	}

	for in, want := range cases {
		if got := NormalizeUsername(in); got != want {
			t.Errorf("NormalizeUsername(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestValidUsernameAcceptsDirectoryShapes(t *testing.T) {
	for _, s := range []string{"jane", "jane.doe", "jane-doe", "jane_doe", "j", "jd", "user123", "1jane"} {
		if !ValidUsername(s) {
			t.Errorf("%q should be valid", s)
		}
	}
}

func TestValidUsernameRejectsUnsafeValues(t *testing.T) {
	for _, s := range []string{
		"",
		"jane doe",
		"jane@example.com",
		`DOMAIN\jane`,
		"Jane",
		".jane",
		"-jane",
		"_jane",
		"jane\n",
		"jane/../root",
	} {
		if ValidUsername(s) {
			t.Errorf("%q should be rejected", s)
		}
	}
}

func TestValidUsernameEnforcesMaxLength(t *testing.T) {
	if !ValidUsername(strings.Repeat("a", UsernameMaxLength)) {
		t.Error("a username at the limit should be valid")
	}
	if ValidUsername(strings.Repeat("a", UsernameMaxLength+1)) {
		t.Error("a username over the limit should be rejected")
	}
}

// Normalising has to produce something ValidUsername accepts, or callers that
// normalise first would still be refused for a name that is merely uppercase.
func TestNormalizeThenValidateAcceptsMixedCaseInput(t *testing.T) {
	for _, s := range []string{"Jane", "JANE.DOE", "  JSmith  "} {
		if !ValidUsername(NormalizeUsername(s)) {
			t.Errorf("%q should be valid once normalised", s)
		}
	}
}

func TestReservedUsernameMatchesTheKeywordsItProtects(t *testing.T) {
	for _, s := range []string{"all", "here", "unknown-user"} {
		if !ReservedUsername(s) {
			t.Errorf("%q should be reserved", s)
		}
	}
}

func TestReservedUsernameIgnoresCaseAndSpace(t *testing.T) {
	for _, s := range []string{"ALL", "Here", "  all  ", "Unknown-User"} {
		if !ReservedUsername(s) {
			t.Errorf("%q should be reserved", s)
		}
	}
}

func TestReservedUsernameLeavesOrdinaryNamesAlone(t *testing.T) {
	for _, s := range []string{"", "alice", "allan", "hereford", "here.we.go", "unknown", "unknown-users"} {
		if ReservedUsername(s) {
			t.Errorf("%q should not be reserved", s)
		}
	}
}

// A reserved name is otherwise well formed, so nothing else would catch it.
func TestReservedUsernamesAreValidUsernames(t *testing.T) {
	for _, s := range []string{"all", "here", "unknown-user"} {
		if !ValidUsername(s) {
			t.Errorf("%q is expected to pass format validation", s)
		}
	}
}
