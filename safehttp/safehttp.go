// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Package safehttp provides an SSRF-hardened HTTP client: its dialer validates
// the resolved destination IP at connect time, so one check covers the URL,
// every redirect, and DNS rebinding. Use it only for untrusted URLs; it blocks
// private addresses, so it is wrong for admin-configured internal integrations.
package safehttp

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"syscall"
	"time"
)

// ErrBlocked is returned when a connection targets a non-public address.
var ErrBlocked = errors.New("safehttp: blocked non-public address")

// Config tunes a safe client; zero values fall back to safe defaults.
type Config struct {
	Timeout      time.Duration
	DialTimeout  time.Duration
	MaxRedirects int
	AllowedPorts []string
	// AllowOwnIPs permits connecting to this host's own interface addresses.
	// Blocking them (the default) closes the connect-back-to-own-public-IP case
	// the pure IP classifier can't catch.
	AllowOwnIPs bool
}

func (c Config) withDefaults() Config {
	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}

	if c.DialTimeout == 0 {
		c.DialTimeout = 5 * time.Second
	}

	if c.MaxRedirects == 0 {
		c.MaxRedirects = 5
	}

	if len(c.AllowedPorts) == 0 {
		c.AllowedPorts = []string{"80", "443"}
	}

	return c
}

func NewClient(cfg Config) *http.Client {
	cfg = cfg.withDefaults()
	return &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: cfg.DialTimeout,
				Control: dialControl(cfg),
			}).DialContext,
			TLSHandshakeTimeout:   cfg.DialTimeout,
			ResponseHeaderTimeout: cfg.Timeout,
			DisableKeepAlives:     true,
		},
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= cfg.MaxRedirects {
				return errors.New("safehttp: too many redirects")
			}

			return nil
		},
	}
}

func dialControl(cfg Config) func(string, string, syscall.RawConn) error {
	return func(network, address string, _ syscall.RawConn) error {
		if network != "tcp4" && network != "tcp6" {
			return fmt.Errorf("safehttp: blocked network %q", network)
		}

		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return err
		}

		if !slices.Contains(cfg.AllowedPorts, port) {
			return fmt.Errorf("safehttp: blocked port %q", port)
		}

		ip := net.ParseIP(host)
		if ip == nil {
			return ErrBlocked
		}

		if IsDisallowedIP(ip) {
			return fmt.Errorf("%w: %s", ErrBlocked, ip)
		}

		if !cfg.AllowOwnIPs && isOwnIP(ip) {
			return fmt.Errorf("%w: own address %s", ErrBlocked, ip)
		}

		return nil
	}
}

func IsDisallowedIP(ip net.IP) bool {
	if isDisallowedRaw(ip) {
		return true
	}
	// Also block the v4 embedded in a NAT64/6to4/IPv4-compatible v6. Checking the
	// embedded address only adds blocks, so it can't unblock a special v6 like ::1.
	if v4 := embeddedV4(ip); v4 != nil && isDisallowedRaw(v4) {
		return true
	}

	return false
}

func isDisallowedRaw(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsInterfaceLocalMulticast() {
		return true
	}

	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 0: // 0.0.0.0/8
			return true
		case v4[0] == 100 && v4[1]&0xc0 == 64: // CGNAT 100.64.0.0/10, missed by IsPrivate
			return true
		case v4[0] == 255 && v4[1] == 255 && v4[2] == 255 && v4[3] == 255: // broadcast
			return true
		}

		return false
	}
	// RFC 8215 local NAT64 64:ff9b:1::/48: block the whole prefix. The embedded
	// v4 offset varies with the deployed prefix length, and no public host lives
	// here, so extraction would be both wrong and pointless.
	if ip16 := ip.To16(); ip16 != nil && bytes.Equal(ip16[:6], nat64LocalPrefix) {
		return true
	}

	return false
}

var (
	nat64Prefix      = []byte{0x00, 0x64, 0xff, 0x9b, 0, 0, 0, 0, 0, 0, 0, 0} // 64:ff9b::/96
	nat64LocalPrefix = []byte{0x00, 0x64, 0xff, 0x9b, 0x00, 0x01}             // 64:ff9b:1::/48
	v4CompatPrefix   = make([]byte, 12)                                       // ::/96
)

func embeddedV4(ip net.IP) net.IP {
	ip16 := ip.To16()
	if ip16 == nil || ip.To4() != nil {
		return nil
	}

	if bytes.Equal(ip16[:12], nat64Prefix) {
		return net.IPv4(ip16[12], ip16[13], ip16[14], ip16[15])
	}

	if ip16[0] == 0x20 && ip16[1] == 0x02 { // 6to4: v4 in bytes 2..5
		return net.IPv4(ip16[2], ip16[3], ip16[4], ip16[5])
	}

	if bytes.Equal(ip16[:12], v4CompatPrefix) { // ::a.b.c.d; ::1 and :: caught by the raw check
		return net.IPv4(ip16[12], ip16[13], ip16[14], ip16[15])
	}

	return nil
}

// isOwnIP reports whether ip is bound to a local interface. Queried fresh, not
// cached: a host's addresses change over uptime and a stale set would stop
// blocking a new own IP. On enumeration error it fails open, briefly leaving the
// own-public-IP case unguarded rather than breaking every fetch.
func isOwnIP(ip net.IP) bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false
	}

	for _, a := range addrs {
		// Equal handles the 4-byte vs 16-byte forms; bytes.Equal would miss
		// an interface's ::ffff:1.2.3.4 against a resolved 1.2.3.4.
		if ipn, ok := a.(*net.IPNet); ok && ipn.IP.Equal(ip) {
			return true
		}
	}

	return false
}
