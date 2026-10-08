// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"net"
	"strconv"
	"strings"
)

type LDAPDialTarget struct {
	URL      string
	Host     string
	Port     int
	Security string
}

// DialTarget builds the address to connect to. The scheme is always derived from
// Security, so a mismatch between the two is not representable.
func (s *LDAPSettings) DialTarget() LDAPDialTarget {
	security := strings.ToLower(strings.TrimSpace(deref(s.ConnectionSecurity)))
	if security == "" {
		security = LDAPSecurityNone
	}

	host := strings.TrimSpace(deref(s.Host))

	port := derefInt(s.Port)
	if port == 0 {
		port = 389
		if security == LDAPSecurityTLS {
			port = 636
		}
	}

	scheme := "ldap"
	if security == LDAPSecurityTLS {
		scheme = "ldaps"
	}

	return LDAPDialTarget{
		URL:      scheme + "://" + net.JoinHostPort(host, strconv.Itoa(port)),
		Host:     host,
		Port:     port,
		Security: security,
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}

	return *i
}
