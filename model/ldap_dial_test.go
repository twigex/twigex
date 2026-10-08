// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "testing"

func settings(host string, port int, security string) *LDAPSettings {
	return &LDAPSettings{
		Host:               NewString(host),
		Port:               NewInt(port),
		ConnectionSecurity: NewString(security),
	}
}

func TestDialTargetSchemeAlwaysMatchesSecurity(t *testing.T) {
	cases := []struct {
		security string
		wantURL  string
	}{
		{LDAPSecurityNone, "ldap://dir.example.com:389"},
		{LDAPSecurityStartTLS, "ldap://dir.example.com:389"},
		{LDAPSecurityTLS, "ldaps://dir.example.com:389"},
	}

	for _, c := range cases {
		got := settings("dir.example.com", 389, c.security).DialTarget()
		if got.URL != c.wantURL {
			t.Errorf("%s: want %s, got %s", c.security, c.wantURL, got.URL)
		}
		if got.Security != c.security {
			t.Errorf("%s: security want %s, got %s", c.security, c.security, got.Security)
		}
	}
}

func TestDialTargetPortDefaultsBySecurity(t *testing.T) {
	if got := settings("dir.example.com", 0, LDAPSecurityTLS).DialTarget(); got.URL != "ldaps://dir.example.com:636" {
		t.Errorf("tls default port: got %s", got.URL)
	}
	if got := settings("dir.example.com", 0, LDAPSecurityStartTLS).DialTarget(); got.URL != "ldap://dir.example.com:389" {
		t.Errorf("starttls default port: got %s", got.URL)
	}
	if got := settings("dir.example.com", 0, "").DialTarget(); got.URL != "ldap://dir.example.com:389" {
		t.Errorf("unset security default port: got %s", got.URL)
	}
}

func TestDialTargetUnsetSecurityIsNone(t *testing.T) {
	if got := settings("dir.example.com", 389, "").DialTarget(); got.Security != LDAPSecurityNone {
		t.Errorf("want none, got %s", got.Security)
	}
}

func TestDialTargetHandlesIPv6(t *testing.T) {
	if got := settings("::1", 389, LDAPSecurityNone).DialTarget(); got.URL != "ldap://[::1]:389" {
		t.Errorf("ipv6: got %s", got.URL)
	}
}

func TestDialTargetToleratesNilFields(t *testing.T) {
	if got := (&LDAPSettings{}).DialTarget(); got.URL != "ldap://:389" {
		t.Errorf("nil fields: got %s", got.URL)
	}
}
