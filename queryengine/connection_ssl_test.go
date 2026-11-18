// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"
)

func mysqlSettings() ConnectionSettings {
	return ConnectionSettings{
		DatabaseType: DatabaseMySQL, Host: "db", Port: "3306",
		User: "u", Password: "p", Database: "shop",
	}
}

func postgresSettings() ConnectionSettings {
	return ConnectionSettings{
		DatabaseType: DatabasePostgres, Host: "db", Port: "5432",
		User: "u", Password: "p", Database: "shop",
	}
}

func testCACertificate(t *testing.T) string {
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

func TestSSLModeDefaultsToOff(t *testing.T) {
	config, err := tlsConfig(mysqlSettings())
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config != nil {
		t.Error("an unset mode should mean no encryption")
	}
}

func TestSSLDisableNegotiatesNothing(t *testing.T) {
	settings := mysqlSettings()
	settings.SSLMode = SSLModeDisable

	config, _ := tlsConfig(settings)
	if config != nil {
		t.Error("disable should not negotiate encryption")
	}
}

func TestSSLRequireEncryptsWithoutChecking(t *testing.T) {
	settings := mysqlSettings()
	settings.SSLMode = SSLModeRequire

	config, err := tlsConfig(settings)
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config == nil {
		t.Fatal("require should encrypt")
	}

	if !config.InsecureSkipVerify {
		t.Error("require does not check the server")
	}

	if config.VerifyConnection != nil {
		t.Error("require should not check a chain either")
	}
}

func TestSSLVerifyCAChecksTheChainButNotTheName(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyCA
	settings.CACertificate = testCACertificate(t)

	config, err := tlsConfig(settings)
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config.RootCAs == nil {
		t.Error("the given certificate should be trusted")
	}

	if config.ServerName != "" {
		t.Error("verify-ca should not check the host name")
	}

	if config.VerifyConnection == nil {
		t.Fatal("verify-ca must still check the chain")
	}
}

// Without a CA certificate, any public authority's certificate for any name
// would pass a check of the chain alone.
func TestSSLVerifyCAWithoutACertificateChecksTheName(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyCA

	config, err := tlsConfig(settings)
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config.InsecureSkipVerify {
		t.Error("verify-ca without a certificate must not skip Go's own check")
	}

	if config.ServerName != settings.Host {
		t.Errorf("want the name %q checked, got %q", settings.Host, config.ServerName)
	}
}

func TestSSLVerifyFullChecksTheName(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyFull
	settings.CACertificate = testCACertificate(t)

	config, err := tlsConfig(settings)
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config.ServerName != "db" {
		t.Errorf("want the host as the expected name, got %q", config.ServerName)
	}

	if config.InsecureSkipVerify {
		t.Error("verify-full must not skip checking")
	}
}

func TestSSLVerifyWithoutACertificateTrustsTheMachinesAuthorities(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyFull

	config, err := tlsConfig(settings)
	if err != nil {
		t.Fatalf("tlsConfig: %v", err)
	}

	if config.RootCAs != nil {
		t.Error("with no certificate given, this machine's authorities should be used")
	}
}

func TestSSLACertificateThatIsNotOneIsRefused(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyFull
	settings.CACertificate = "not a certificate"

	if _, err := tlsConfig(settings); err == nil {
		t.Error("want an error rather than silently trusting nothing")
	}
}

func TestSSLChainCheckRefusesAServerThatSendsNoCertificate(t *testing.T) {
	verify := verifyChainOnly(nil)

	if err := verify(tls.ConnectionState{}); err == nil {
		t.Error("want an error when no certificate is sent")
	}
}

// tlsServer answers handshakes with a self-signed certificate for
// commonName, and returns its address and that certificate in PEM.
func tlsServer(t *testing.T, commonName string) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key: %v", err)
	}

	template := x509.Certificate{
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

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("certificate: %v", err)
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
	})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	t.Cleanup(func() {
		listener.Close()
	})

	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}

			_ = connection.(*tls.Conn).Handshake()
			connection.Close()
		}
	}()

	certificate := string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))

	return listener.Addr().String(), certificate
}

// handshake connects to the server the way a saved connection with these
// settings would.
func handshake(t *testing.T, address string, mode SSLMode, caCertificate string) error {
	t.Helper()

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		t.Fatalf("address: %v", err)
	}

	settings := ConnectionSettings{
		DatabaseType:  DatabasePostgres,
		Host:          host,
		Port:          port,
		SSLMode:       mode,
		CACertificate: caCertificate,
	}

	config, err := tlsConfig(settings)
	if err != nil {
		return err
	}

	client, err := tls.Dial("tcp", address, config)
	if err != nil {
		return err
	}
	defer client.Close()

	return client.Handshake()
}

func TestSSLVerifyCAAcceptsACertificateForAnotherName(t *testing.T) {
	address, ca := tlsServer(t, "the-db")

	if err := handshake(t, address, SSLModeVerifyCA, ca); err != nil {
		t.Errorf("verify-ca should not care about the name: %v", err)
	}
}

func TestSSLVerifyFullRefusesTheSameCertificate(t *testing.T) {
	address, ca := tlsServer(t, "the-db")

	if err := handshake(t, address, SSLModeVerifyFull, ca); err == nil {
		t.Error("verify-full must refuse a certificate for another name")
	}
}

func TestSSLVerifyCARefusesAnAuthorityItWasNotGiven(t *testing.T) {
	address, _ := tlsServer(t, "the-db")
	_, otherCA := tlsServer(t, "somewhere-else")

	if err := handshake(t, address, SSLModeVerifyCA, otherCA); err == nil {
		t.Error("verify-ca must refuse a chain it cannot trace to the CA")
	}
}

func TestSSLVerifyCAWithoutACertificateRefusesASelfSignedServer(t *testing.T) {
	address, _ := tlsServer(t, "the-db")

	if err := handshake(t, address, SSLModeVerifyCA, ""); err == nil {
		t.Error("this machine's authorities must not trust a self-signed server")
	}
}

func TestSSLRequireAcceptsAServerItCannotCheckAtAll(t *testing.T) {
	address, _ := tlsServer(t, "the-db")

	if err := handshake(t, address, SSLModeRequire, ""); err != nil {
		t.Errorf("require should connect anyway: %v", err)
	}
}

func TestSSLMySQLNamesItsSettings(t *testing.T) {
	settings := mysqlSettings()
	settings.SSLMode = SSLModeVerifyFull

	name, err := registerMySQLTLS(settings)
	if err != nil {
		t.Fatalf("registerMySQLTLS: %v", err)
	}

	if name == "" || name == "false" {
		t.Errorf("an encrypted connection needs named settings, got %q", name)
	}
}

func TestSSLMySQLSaysFalseWhenNotEncrypting(t *testing.T) {
	settings := mysqlSettings()
	settings.SSLMode = SSLModeDisable

	name, err := registerMySQLTLS(settings)
	if err != nil {
		t.Fatalf("registerMySQLTLS: %v", err)
	}

	if name != "false" {
		t.Errorf("want the driver told not to encrypt, got %q", name)
	}
}

func TestSSLTheSameSettingsReuseOneName(t *testing.T) {
	first := mysqlSettings()
	first.SSLMode = SSLModeVerifyFull
	second := first

	firstName, _ := registerMySQLTLS(first)
	secondName, _ := registerMySQLTLS(second)

	if firstName != secondName {
		t.Errorf("want one name for the same settings, got %q and %q", firstName, secondName)
	}
}

func TestSSLADifferentCertificateGetsItsOwnName(t *testing.T) {
	withCertificate := mysqlSettings()
	withCertificate.SSLMode = SSLModeVerifyFull
	withCertificate.CACertificate = testCACertificate(t)

	without := mysqlSettings()
	without.SSLMode = SSLModeVerifyFull

	firstName, _ := registerMySQLTLS(withCertificate)
	secondName, _ := registerMySQLTLS(without)

	if firstName == secondName {
		t.Error("a changed certificate must not reuse the old settings")
	}
}

func mysqlTLSUseCount(name string) int {
	mysqlTLSUsers.mutex.Lock()
	defer mysqlTLSUsers.mutex.Unlock()

	return mysqlTLSUsers.count[name]
}

// Encryption settings stay registered while a pool uses them, and are
// removed when the last pool using them closes.
func TestSSLMySQLSettingsAreRemovedWhenTheirLastPoolCloses(t *testing.T) {
	settings := mysqlSettings()
	settings.Host = "pool-release-test"
	settings.SSLMode = SSLModeRequire

	name, err := registerMySQLTLS(settings)
	if err != nil {
		t.Fatalf("registerMySQLTLS: %v", err)
	}
	releaseMySQLTLS(name)

	connector := NewConnector(ConnectorOptions{})

	first, err := connector.Open(settings)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	second, err := connector.Open(settings)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if got := mysqlTLSUseCount(name); got != 2 {
		t.Errorf("two open pools: want 2 uses, got %d", got)
	}

	first.Close()

	if got := mysqlTLSUseCount(name); got != 1 {
		t.Errorf("one pool closed: want 1 use, got %d", got)
	}

	second.Close()

	if got := mysqlTLSUseCount(name); got != 0 {
		t.Errorf("both pools closed: want no uses, got %d", got)
	}
}
