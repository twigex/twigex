// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http/httptest"
	"testing"
)

func TestIsPageNavigation(t *testing.T) {
	cases := []struct {
		name   string
		header string
		value  string
		want   bool
	}{
		{"browser opening the link", "Sec-Fetch-Mode", "navigate", true},
		{"browser asking for a page", "Accept", "text/html,application/xhtml+xml", true},
		{"the app's own request", "Accept", "application/json", false},
		{"no hints", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/public/tok/download", nil)
			if tc.header != "" {
				r.Header.Set(tc.header, tc.value)
			}

			if got := isPageNavigation(r); got != tc.want {
				t.Errorf("want %v, got %v", tc.want, got)
			}
		})
	}
}
