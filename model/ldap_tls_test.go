// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func testCAPEM(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Twigex Test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func settingsWithCA(ca string) *LDAPSettings {
	s := &LDAPSettings{}
	s.SetDefaults()
	s.Host = NewString("dir.example.com")
	s.ConnectionSecurity = NewString(LDAPSecurityTLS)
	s.CACertificate = NewString(ca)
	return s
}

// No certificate means the system trust store, which x509 selects on a nil pool.
func TestTLSConfigWithoutCAUsesSystemRoots(t *testing.T) {
	cfg, err := settingsWithCA("").TLSConfig()
	if err != nil {
		t.Fatalf("TLSConfig: %v", err)
	}
	if cfg.RootCAs != nil {
		t.Error("an unset certificate must leave RootCAs nil")
	}
	if cfg.ServerName != "dir.example.com" {
		t.Errorf("ServerName = %q", cfg.ServerName)
	}
}

func TestTLSConfigAcceptsInlinePEM(t *testing.T) {
	cfg, err := settingsWithCA(testCAPEM(t)).TLSConfig()
	if err != nil {
		t.Fatalf("TLSConfig: %v", err)
	}
	if cfg.RootCAs == nil {
		t.Fatal("an inline certificate must produce a pool")
	}
	if len(cfg.RootCAs.Subjects()) != 1 { //nolint:staticcheck // reading our own pool
		t.Error("the pool should hold exactly the configured certificate")
	}
}

// Silently falling back to the system roots would present a working connection
// that is not verified against the certificate the administrator supplied.
func TestTLSConfigRejectsUnusableCertificate(t *testing.T) {
	for _, in := range []string{
		"-----BEGIN CERTIFICATE-----\nnot base64\n-----END CERTIFICATE-----",
		"/etc/ssl/certs/ca.pem",
		"not a certificate at all",
	} {
		if _, err := settingsWithCA(in).TLSConfig(); err == nil {
			t.Errorf("%q should not be accepted as a certificate", in)
		}
	}
}

func TestTLSConfigCarriesInsecureSkipVerify(t *testing.T) {
	s := settingsWithCA("")
	s.InsecureSkipVerify = NewBool(true)

	cfg, err := s.TLSConfig()
	if err != nil {
		t.Fatalf("TLSConfig: %v", err)
	}
	if !cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify must reach the dialer")
	}
}
