// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"net"
	"net/http"
	"strings"
)

func getIPAddress(r *http.Request, trustedHeaders, trustedProxies []string) string {
	peer := peerHost(r.RemoteAddr)

	nets := parseCIDRs(trustedProxies)
	if len(nets) == 0 || !ipInNets(peer, nets) {
		return peer
	}

	for _, header := range trustedHeaders {
		v := strings.TrimSpace(r.Header.Get(header))
		if v == "" {
			continue
		}

		if strings.EqualFold(header, "X-Forwarded-For") {
			ips := splitForwardedFor(v)
			for i := len(ips) - 1; i >= 0; i-- {
				if !ipInNets(ips[i], nets) {
					return ips[i]
				}
			}

			continue
		}

		return v
	}

	return peer
}

func peerHost(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}

	return host
}

func splitForwardedFor(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}

	return out
}

func parseCIDRs(entries []string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		if _, ipNet, err := net.ParseCIDR(entry); err == nil {
			nets = append(nets, ipNet)
			continue
		}

		if ip := net.ParseIP(entry); ip != nil {
			bits := 32
			if ip.To4() == nil {
				bits = 128
			}

			nets = append(nets, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		}
	}

	return nets
}

func ipInNets(ipStr string, nets []*net.IPNet) bool {
	ip := net.ParseIP(strings.TrimSpace(ipStr))
	if ip == nil {
		return false
	}

	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}

	return false
}
