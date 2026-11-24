// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestConnectionFailureOfNoErrorIsNone(t *testing.T) {
	if got := ConnectionFailureOf(nil); got != "" {
		t.Errorf("want no reason, got %q", got)
	}
}

func TestConnectionFailureOfSomethingElseIsUnknown(t *testing.T) {
	if got := ConnectionFailureOf(errors.New("the disk is on fire")); got != ConnectionFailureUnknown {
		t.Errorf("want %q, got %q", ConnectionFailureUnknown, got)
	}
}

func TestConnectionFailureMySQLCodesAreNamed(t *testing.T) {
	cases := map[uint16]ConnectionFailure{
		1045: ConnectionFailureCredentials,
		1044: ConnectionFailureCredentials,
		1698: ConnectionFailureCredentials,
		1049: ConnectionFailureDatabase,
		1146: ConnectionFailureUnknown,
	}

	for number, want := range cases {
		err := fmt.Errorf("connecting: %w", &mysql.MySQLError{Number: number})

		if got := ConnectionFailureOf(err); got != want {
			t.Errorf("MySQL %d: want %q, got %q", number, want, got)
		}
	}
}

func TestConnectionFailurePostgresCodesAreNamed(t *testing.T) {
	cases := map[string]ConnectionFailure{
		"28P01": ConnectionFailureCredentials,
		"28000": ConnectionFailureCredentials,
		"3D000": ConnectionFailureDatabase,
		"42P01": ConnectionFailureUnknown,
	}

	for code, want := range cases {
		err := fmt.Errorf("connecting: %w", errors.Join(&pgconn.PgError{Code: code}))

		if got := ConnectionFailureOf(err); got != want {
			t.Errorf("Postgres %s: want %q, got %q", code, want, got)
		}
	}
}

func TestConnectionFailureACertificateThatCannotBeReadIsItsOwnReason(t *testing.T) {
	settings := postgresSettings()
	settings.SSLMode = SSLModeVerifyFull
	settings.CACertificate = "not a certificate"

	_, err := tlsConfig(settings)

	if got := ConnectionFailureOf(err); got != ConnectionFailureCACertificate {
		t.Errorf("want %q, got %q", ConnectionFailureCACertificate, got)
	}
}

func TestConnectionFailureAnUntrustedServerIsACertificateProblem(t *testing.T) {
	address, _ := tlsServer(t, "the-db")
	_, otherCA := tlsServer(t, "somewhere-else")

	err := handshake(t, address, SSLModeVerifyCA, otherCA)
	if err == nil {
		t.Fatal("want the handshake to fail")
	}

	if got := ConnectionFailureOf(err); got != ConnectionFailureCertificate {
		t.Errorf("want %q, got %q: %v", ConnectionFailureCertificate, got, err)
	}
}

func TestConnectionFailureAWrongNameIsNotBlamedOnTheAuthority(t *testing.T) {
	address, ca := tlsServer(t, "the-db")

	err := handshake(t, address, SSLModeVerifyFull, ca)
	if err == nil {
		t.Fatal("want the handshake to fail")
	}

	if got := ConnectionFailureOf(err); got != ConnectionFailureHostname {
		t.Errorf("want %q, got %q: %v", ConnectionFailureHostname, got, err)
	}
}

func TestConnectionFailureAHostnameProblemBeatsTheErrorAroundIt(t *testing.T) {
	wrapped := fmt.Errorf("connecting: %w", &tls.CertificateVerificationError{
		Err: x509.HostnameError{Host: "localhost"},
	})

	if got := ConnectionFailureOf(wrapped); got != ConnectionFailureHostname {
		t.Errorf("want %q, got %q", ConnectionFailureHostname, got)
	}
}

func TestConnectionFailureAServerSendingNothingIsACertificateProblem(t *testing.T) {
	if got := ConnectionFailureOf(errNoPeerCertificate); got != ConnectionFailureCertificate {
		t.Errorf("want %q, got %q", ConnectionFailureCertificate, got)
	}
}

func TestConnectionFailureAClosedPortIsUnreachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	address := listener.Addr().String()
	listener.Close()

	host, port, _ := net.SplitHostPort(address)

	for _, databaseType := range []DatabaseType{DatabaseMySQL, DatabasePostgres} {
		settings := ConnectionSettings{DatabaseType: databaseType, Host: host, Port: port, User: "u", Database: "shop"}

		err := NewConnector(ConnectorOptions{}).Check(context.Background(), settings)

		if got := ConnectionFailureOf(err); got != ConnectionFailureUnreachable {
			t.Errorf("%s: want %q, got %q: %v", databaseType, ConnectionFailureUnreachable, got, err)
		}
	}
}

func TestConnectionFailureATimeoutIsUnreachable(t *testing.T) {
	err := fmt.Errorf("dialing: %w", context.DeadlineExceeded)

	if got := ConnectionFailureOf(err); got != ConnectionFailureUnreachable {
		t.Errorf("want %q, got %q", ConnectionFailureUnreachable, got)
	}
}

func TestConnectionFailureACertificateProblemBeatsLookingLikeANetworkOne(t *testing.T) {
	err := fmt.Errorf("read tcp: %w", errors.Join(
		&net.OpError{Op: "read", Err: errors.New("connection reset")},
		x509.UnknownAuthorityError{},
	))

	if got := ConnectionFailureOf(err); got != ConnectionFailureCertificate {
		t.Errorf("want %q, got %q", ConnectionFailureCertificate, got)
	}
}
