// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package safehttp

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestIsDisallowedIP(t *testing.T) {
	cases := []struct {
		ip      string
		blocked bool
	}{
		// Public, allowed.
		{"8.8.8.8", false},
		{"1.1.1.1", false},
		{"93.184.216.34", false},
		{"2606:4700:4700::1111", false},
		{"100.63.255.255", false}, // just below CGNAT
		{"100.128.0.1", false},    // just above CGNAT

		// Non-public, blocked.
		{"127.0.0.1", true},
		{"::1", true},
		{"10.0.0.1", true},
		{"192.168.1.1", true},
		{"172.16.0.1", true},
		{"169.254.169.254", true}, // cloud metadata
		{"0.0.0.0", true},
		{"0.1.2.3", true},         // 0.0.0.0/8 "this network"
		{"255.255.255.255", true}, // broadcast
		{"100.64.0.1", true},      // CGNAT low
		{"100.127.0.1", true},     // CGNAT high
		{"fd00::1", true},         // IPv6 ULA
		{"fe80::1", true},         // IPv6 link-local
		{"ff02::1", true},         // IPv6 multicast

		// IPv6 transition/embedding forms carrying a private v4, blocked.
		{"::ffff:127.0.0.1", true}, // IPv4-mapped loopback
		{"::127.0.0.1", true},      // deprecated IPv4-compatible loopback
		{"64:ff9b::7f00:1", true},  // NAT64 127.0.0.1
		{"2002:7f00:1::", true},    // 6to4 127.0.0.1
		{"64:ff9b::a00:1", true},   // NAT64 10.0.0.1

		// RFC 8215 local NAT64 prefix, blocked as a whole so even a public-looking
		// embedded v4 is refused.
		{"64:ff9b:1::7f00:1", true}, // local NAT64, embeds 127.0.0.1
		{"64:ff9b:1::1", true},      // anything in 64:ff9b:1::/48
		{"64:ff9b:1::808:808", true},

		// Transition encodings embedding a public v4, allowed.
		{"64:ff9b::808:808", false}, // NAT64 8.8.8.8
		{"2002:808:808::", false},   // 6to4 8.8.8.8
	}

	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		if ip == nil {
			t.Fatalf("bad test IP %q", c.ip)
		}
		if got := IsDisallowedIP(ip); got != c.blocked {
			t.Errorf("IsDisallowedIP(%s) = %v, want %v", c.ip, got, c.blocked)
		}
	}
}

// Loopback must be refused at connect time even when its port is allowed.
func TestClientBlocksLoopbackDial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	u, _ := url.Parse(srv.URL) // httptest listens on 127.0.0.1
	client := NewClient(Config{AllowedPorts: []string{u.Port()}})

	_, err := client.Get(srv.URL)
	if err == nil {
		t.Fatal("expected loopback dial to be blocked")
	}
	if !errors.Is(err, ErrBlocked) {
		t.Errorf("expected ErrBlocked, got %v", err)
	}
}

func TestIsOwnIP(t *testing.T) {
	// Loopback is always bound to an interface; 8.8.8.8 never is. This also
	// exercises the Equal-based match across 4-byte/16-byte forms.
	if !isOwnIP(net.ParseIP("127.0.0.1")) {
		t.Error("expected 127.0.0.1 to be an own (loopback) address")
	}
	if isOwnIP(net.ParseIP("8.8.8.8")) {
		t.Error("8.8.8.8 should not be an own address")
	}
}

// A disallowed port is rejected before the IP is even considered.
func TestClientBlocksDisallowedPort(t *testing.T) {
	client := NewClient(Config{}) // defaults to 80/443 only
	if _, err := client.Get("http://93.184.216.34:8080/"); err == nil {
		t.Fatal("expected non-80/443 port to be blocked")
	}
}
