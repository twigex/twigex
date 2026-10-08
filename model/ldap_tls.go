// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strings"
)

var ErrLDAPCACertificateUnusable = errors.New("the configured LDAP CA certificate could not be parsed")

// TLSConfig builds the settings both the login and the connection test dial with,
// so the two cannot drift apart. ServerName comes from the configured host rather
// than the URL, which keeps verification working when the host is an alias.
func (s *LDAPSettings) TLSConfig() (*tls.Config, error) {
	target := s.DialTarget()

	cfg := &tls.Config{
		InsecureSkipVerify: derefBool(s.InsecureSkipVerify),
		ServerName:         target.Host,
	}

	if target.Security == LDAPSecurityNone {
		return cfg, nil
	}

	pool, err := s.rootCAs()
	if err != nil {
		return nil, err
	}

	cfg.RootCAs = pool

	return cfg, nil
}

// rootCAs returns nil when nothing is configured, which leaves verification on the
// system trust store.
func (s *LDAPSettings) rootCAs() (*x509.CertPool, error) {
	raw := strings.TrimSpace(deref(s.CACertificate))
	if raw == "" {
		return nil, nil
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(raw)) {
		return nil, ErrLDAPCACertificateUnusable
	}

	return pool, nil
}

func derefBool(b *bool) bool {
	if b == nil {
		return false
	}

	return *b
}
