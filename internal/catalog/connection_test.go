// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func mysqlConn() Connection {
	return Connection{
		Type: Mysql, Host: "db", Port: "3306",
		User: "u", Password: "p", Database: "shop",
	}
}

func postgresConn() Connection {
	return Connection{
		Type: Postgre, Host: "db", Port: "5432",
		User: "u", Password: "p", Database: "shop",
	}
}

func testCA(t *testing.T) string {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"twigex test"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate: %v", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestModeDefaultsToOff(t *testing.T) {
	conn := mysqlConn()

	cfg, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if cfg != nil {
		t.Error("an unset mode should mean no encryption")
	}
}

func TestDisableNegotiatesNothing(t *testing.T) {
	conn := mysqlConn()
	conn.SSLMode = SSLModeDisable

	cfg, _ := conn.tlsConfig()
	if cfg != nil {
		t.Error("disable should not negotiate tls")
	}
}

func TestRequireEncryptsWithoutChecking(t *testing.T) {
	conn := mysqlConn()
	conn.SSLMode = SSLModeRequire

	cfg, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if cfg == nil {
		t.Fatal("require should negotiate tls")
	}
	if !cfg.InsecureSkipVerify {
		t.Error("require does not verify the server")
	}
	if cfg.VerifyConnection != nil {
		t.Error("require should not verify a chain either")
	}
}

func TestVerifyCAChecksTheChainButNotTheName(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyCA
	conn.Config = map[string]string{CACertKey: testCA(t)}

	cfg, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if cfg.RootCAs == nil {
		t.Error("the supplied certificate should be trusted")
	}
	if cfg.ServerName != "" {
		t.Error("verify-ca should not check the host name")
	}
	if cfg.VerifyConnection == nil {
		t.Fatal("verify-ca must still verify the chain")
	}
}

func TestVerifyFullChecksTheName(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyFull
	conn.Config = map[string]string{CACertKey: testCA(t)}

	cfg, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if cfg.ServerName != "db" {
		t.Errorf("want the host as the expected name, got %q", cfg.ServerName)
	}
	if cfg.InsecureSkipVerify {
		t.Error("verify-full must not skip verification")
	}
}

func TestVerifyWithoutACertificateTrustsTheHost(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyFull

	cfg, err := conn.tlsConfig()
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}
	if cfg.RootCAs != nil {
		t.Error("with no certificate given, the host's own roots should be used")
	}
}

func TestACertificateThatIsNotOneIsRejected(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyFull
	conn.Config = map[string]string{CACertKey: "not a certificate"}

	if _, err := conn.tlsConfig(); err == nil {
		t.Error("expected an error rather than silently trusting nothing")
	}
}

func TestChainOnlyRefusesAServerThatSendsNoCertificate(t *testing.T) {
	verify := chainOnly(nil)

	if err := verify(tls.ConnectionState{}); err == nil {
		t.Error("expected an error when no certificate is presented")
	}
}

// tlsServer answers one handshake with a self-signed certificate naming
// commonName, and returns its address plus that certificate as PEM.
func tlsServer(t *testing.T, commonName string) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key: %v", err)
	}

	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		DNSNames:              []string{commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate: %v", err)
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	return ln.Addr().String(), string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

// handshake dials the server under the mode a stored connection would use.
func handshake(t *testing.T, addr, mode, caCert string) error {
	t.Helper()

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("addr: %v", err)
	}

	conn := Connection{Type: Postgre, Host: host, Port: port, SSLMode: mode}
	if caCert != "" {
		conn.Config = map[string]string{CACertKey: caCert}
	}

	cfg, err := conn.tlsConfig()
	if err != nil {
		return err
	}

	client, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return err
	}
	defer client.Close()

	return client.Handshake()
}

func TestVerifyCAAcceptsACertificateIssuedToAnotherName(t *testing.T) {
	addr, ca := tlsServer(t, "the-db")

	if err := handshake(t, addr, SSLModeVerifyCA, ca); err != nil {
		t.Errorf("verify-ca should not care about the name: %v", err)
	}
}

func TestVerifyFullRejectsTheSameCertificate(t *testing.T) {
	addr, ca := tlsServer(t, "the-db")

	if err := handshake(t, addr, SSLModeVerifyFull, ca); err == nil {
		t.Error("verify-full must reject a certificate issued to another name")
	}
}

func TestVerifyCARejectsAnAuthorityItWasNotGiven(t *testing.T) {
	addr, _ := tlsServer(t, "the-db")
	_, otherCA := tlsServer(t, "somewhere-else")

	if err := handshake(t, addr, SSLModeVerifyCA, otherCA); err == nil {
		t.Error("verify-ca must reject a chain it cannot anchor")
	}
}

func TestVerifyCAWithNoCertificateRejectsASelfSignedServer(t *testing.T) {
	addr, _ := tlsServer(t, "the-db")

	if err := handshake(t, addr, SSLModeVerifyCA, ""); err == nil {
		t.Error("falling back to the system roots must not trust a self-signed server")
	}
}

func TestRequireAcceptsAServerItCannotVerifyAtAll(t *testing.T) {
	addr, _ := tlsServer(t, "the-db")

	if err := handshake(t, addr, SSLModeRequire, ""); err != nil {
		t.Errorf("require should connect regardless: %v", err)
	}
}

func TestMysqlNamesItsTLSConfig(t *testing.T) {
	conn := mysqlConn()
	conn.SSLMode = SSLModeVerifyFull

	name, err := registerMysqlTLS(&conn)
	if err != nil {
		t.Fatalf("registerMysqlTLS: %v", err)
	}
	if name == "" || name == "false" {
		t.Errorf("a verifying connection needs a named config, got %q", name)
	}
}

func TestMysqlSaysFalseWhenThereIsNothingToNegotiate(t *testing.T) {
	conn := mysqlConn()
	conn.SSLMode = SSLModeDisable

	name, err := registerMysqlTLS(&conn)
	if err != nil {
		t.Fatalf("registerMysqlTLS: %v", err)
	}
	if name != "false" {
		t.Errorf("want the driver told not to use tls, got %q", name)
	}
}

func TestTheSameSettingsReuseOneNamedConfig(t *testing.T) {
	first := mysqlConn()
	first.SSLMode = SSLModeVerifyFull
	second := first

	a, _ := registerMysqlTLS(&first)
	b, _ := registerMysqlTLS(&second)

	if a != b {
		t.Errorf("expected one entry for identical settings, got %q and %q", a, b)
	}
}

func TestADifferentCertificateGetsItsOwnConfig(t *testing.T) {
	withCert := mysqlConn()
	withCert.SSLMode = SSLModeVerifyFull
	withCert.Config = map[string]string{CACertKey: testCA(t)}

	without := mysqlConn()
	without.SSLMode = SSLModeVerifyFull

	a, _ := registerMysqlTLS(&withCert)
	b, _ := registerMysqlTLS(&without)

	if a == b {
		t.Error("a changed certificate must not reuse the old config")
	}
}

func TestOpenRefusesAnUnknownStore(t *testing.T) {
	conn := Connection{Type: "bigquery"}

	if _, err := conn.Open(); err == nil {
		t.Error("expected an error for an unsupported store")
	}
}

func TestPostgresOpensWithoutWritingACertificateToDisk(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyFull
	conn.Config = map[string]string{CACertKey: testCA(t)}

	db, err := conn.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	files, err := filepath.Glob(filepath.Join(os.TempDir(), "cospace-ca-*.pem"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected no certificate on disk, found %v", files)
	}
}

var awkward = []string{
	"p@ss/word",
	"with space",
	"quote'd",
	"back\\slash",
	"tcp(evil:3306)/other",
	"x host=evil.example.com",
}

func TestMysqlKeepsTheCredentialsItWasGiven(t *testing.T) {
	for _, password := range awkward {
		conn := mysqlConn()
		conn.Password = password

		cfg, err := conn.mysqlConfig()
		if err != nil {
			t.Errorf("password %q: %v", password, err)
			continue
		}

		// Round trip through the driver's own parser: what it writes it must read back.
		parsed, err := mysql.ParseDSN(cfg.FormatDSN())
		if err != nil {
			t.Errorf("password %q: unparseable dsn: %v", password, err)
			continue
		}

		if parsed.Passwd != password {
			t.Errorf("password %q came back as %q", password, parsed.Passwd)
		}
		if parsed.Addr != "db:3306" {
			t.Errorf("password %q moved the address to %q", password, parsed.Addr)
		}
		if parsed.DBName != "shop" {
			t.Errorf("password %q changed the database to %q", password, parsed.DBName)
		}
		if parsed.User != "u" {
			t.Errorf("password %q changed the user to %q", password, parsed.User)
		}
	}
}

func TestPostgresKeepsTheCredentialsItWasGiven(t *testing.T) {
	for _, password := range awkward {
		conn := postgresConn()
		conn.Password = password

		cfg, err := conn.postgresConfig()
		if err != nil {
			t.Errorf("password %q: %v", password, err)
			continue
		}

		if cfg.Password != password {
			t.Errorf("password %q came back as %q", password, cfg.Password)
		}
		if cfg.Host != "db" {
			t.Errorf("password %q moved the host to %q", password, cfg.Host)
		}
		if cfg.Port != 5432 {
			t.Errorf("password %q moved the port to %d", password, cfg.Port)
		}
		if cfg.Database != "shop" {
			t.Errorf("password %q changed the database to %q", password, cfg.Database)
		}
	}
}

func TestABadPortIsRejected(t *testing.T) {
	conn := postgresConn()
	conn.Port = "not a port"

	if _, err := conn.Open(); err == nil {
		t.Error("expected an error rather than a connection to port 0")
	}
}

func TestMysqlCarriesTheTLSConfigNameIntoItsDSN(t *testing.T) {
	conn := mysqlConn()
	conn.SSLMode = SSLModeVerifyFull

	cfg, err := conn.mysqlConfig()
	if err != nil {
		t.Fatalf("mysqlConfig: %v", err)
	}

	parsed, err := mysql.ParseDSN(cfg.FormatDSN())
	if err != nil {
		t.Fatalf("unparseable dsn: %v", err)
	}
	if parsed.TLSConfig == "" || parsed.TLSConfig == "false" {
		t.Errorf("expected a registered config in the dsn, got %q", parsed.TLSConfig)
	}
}
