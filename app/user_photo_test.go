// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import "testing"

// etagMatches decides whether a photo request is answered with a 304 or the
// whole file, so a tag that fails to match when it should goes unnoticed: the
// response is still correct, just needlessly large.

func TestEtagMatches(t *testing.T) {
	cases := []struct {
		name        string
		ifNoneMatch string
		etag        string
		want        bool
	}{
		{"no header", "", `"abc"`, false},
		{"exact", `"abc"`, `"abc"`, true},
		{"different tag", `"abc"`, `"def"`, false},
		{"wildcard", "*", `"abc"`, true},
		{"one of several", `"def", "abc"`, `"abc"`, true},
		{"none of several", `"def", "ghi"`, `"abc"`, false},
		{"weak request tag", `W/"abc"`, `"abc"`, true},
		{"weak stored tag", `"abc"`, `W/"abc"`, true},
		{"unquoted never matches", "abc", `"abc"`, false},
		{"surrounding spaces", `  "abc"  `, `"abc"`, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := etagMatches(c.ifNoneMatch, c.etag); got != c.want {
				t.Fatalf("etagMatches(%q, %q) = %v, want %v", c.ifNoneMatch, c.etag, got, c.want)
			}
		})
	}
}
