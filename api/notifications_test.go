// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSameOrigin(t *testing.T) {
	cases := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{"own origin", "twigex.example.com", "https://twigex.example.com", true},
		{"own origin, other scheme", "twigex.example.com", "http://twigex.example.com", true},
		{"host as a prefix of the origin", "twigex.example.com", "https://twigex.example.com.evil.net", false},
		{"host as a suffix of the origin", "twigex.example.com", "https://evil.twigex.example.com", false},
		{"host inside a path", "twigex.example.com", "https://evil.net/twigex.example.com", false},
		{"another host entirely", "twigex.example.com", "https://evil.net", false},
		{"no origin at all", "twigex.example.com", "", false},
		{"unparsable origin", "twigex.example.com", "://", false},
		{"vite dev server on loopback", "localhost:8065", "http://localhost:5173", true},
		{"vite dev origin against a public host", "twigex.example.com", "http://localhost:5173", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/notifications/ws", nil)
			r.Host = c.host
			if c.origin != "" {
				r.Header.Set("Origin", c.origin)
			}

			if got := sameOrigin(r); got != c.want {
				t.Errorf("want %t, got %t", c.want, got)
			}
		})
	}
}

func TestIsLoopback(t *testing.T) {
	cases := map[string]bool{
		"localhost:8065":     true,
		"localhost":          true,
		"127.0.0.1:8065":     true,
		"[::1]:8065":         true,
		"twigex.example.com": false,
		"localhost.evil.net": false,
		"10.0.0.1:8065":      false,
	}

	for host, want := range cases {
		if got := isLoopback(host); got != want {
			t.Errorf("%s: want %t, got %t", host, want, got)
		}
	}
}
