// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

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

func TestNoErrorHasNoReason(t *testing.T) {
	if got := Failure(nil); got != "" {
		t.Errorf("want no reason, got %q", got)
	}
}

func TestSomethingElseStaysUnknown(t *testing.T) {
	if got := Failure(errors.New("the disk is on fire")); got != FailureUnknown {
		t.Errorf("want %q, got %q", FailureUnknown, got)
	}
}

func TestMysqlCodesAreNamed(t *testing.T) {
	cases := map[uint16]string{
		1045: FailureCredentials,
		1044: FailureCredentials,
		1698: FailureCredentials,
		1049: FailureDatabase,
		1146: FailureUnknown,
	}

	for number, want := range cases {
		err := fmt.Errorf("connecting: %w", &mysql.MySQLError{Number: number})

		if got := Failure(err); got != want {
			t.Errorf("mysql %d: want %q, got %q", number, want, got)
		}
	}
}

func TestPostgresCodesAreNamed(t *testing.T) {
	cases := map[string]string{
		"28P01": FailureCredentials,
		"28000": FailureCredentials,
		"3D000": FailureDatabase,
		"42P01": FailureUnknown,
	}

	for code, want := range cases {
		err := fmt.Errorf("connecting: %w", errors.Join(&pgconn.PgError{Code: code}))

		if got := Failure(err); got != want {
			t.Errorf("postgres %s: want %q, got %q", code, want, got)
		}
	}
}

func TestACertificateThatCannotBeReadIsItsOwnReason(t *testing.T) {
	conn := postgresConn()
	conn.SSLMode = SSLModeVerifyFull
	conn.Config = map[string]string{CACertKey: "not a certificate"}

	_, err := conn.tlsConfig()

	if got := Failure(err); got != FailureCACert {
		t.Errorf("want %q, got %q", FailureCACert, got)
	}
}

func TestAnUntrustedServerIsACertificateProblem(t *testing.T) {
	addr, _ := tlsServer(t, "the-db")

	err := handshake(t, addr, SSLModeVerifyCA, "")
	if err == nil {
		t.Fatal("expected the handshake to fail")
	}

	if got := Failure(err); got != FailureCertificate {
		t.Errorf("want %q, got %q: %v", FailureCertificate, got, err)
	}
}

func TestAWrongNameIsNotBlamedOnTheAuthority(t *testing.T) {
	addr, ca := tlsServer(t, "the-db")

	err := handshake(t, addr, SSLModeVerifyFull, ca)
	if err == nil {
		t.Fatal("expected the handshake to fail")
	}

	if got := Failure(err); got != FailureHostname {
		t.Errorf("want %q, got %q: %v", FailureHostname, got, err)
	}
}

func TestAHostnameProblemBeatsTheWrapperItArrivesIn(t *testing.T) {
	wrapped := fmt.Errorf("connecting: %w", &tls.CertificateVerificationError{
		Err: x509.HostnameError{Host: "localhost"},
	})

	if got := Failure(wrapped); got != FailureHostname {
		t.Errorf("want %q, got %q", FailureHostname, got)
	}
}

func TestAServerSendingNothingIsACertificateProblem(t *testing.T) {
	if got := Failure(errNoPeerCertificate); got != FailureCertificate {
		t.Errorf("want %q, got %q", FailureCertificate, got)
	}
}

func TestAClosedPortIsUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	host, port, _ := net.SplitHostPort(addr)
	conn := Connection{Type: Postgre, Host: host, Port: port, User: "u", Database: "shop"}

	db, err := conn.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	if got := Failure(db.Ping()); got != FailureUnreachable {
		t.Errorf("want %q, got %q: %v", FailureUnreachable, got, db.Ping())
	}
}

func TestATimeoutIsUnreachable(t *testing.T) {
	if got := Failure(fmt.Errorf("dialing: %w", context.DeadlineExceeded)); got != FailureUnreachable {
		t.Errorf("want %q, got %q", FailureUnreachable, got)
	}
}

func TestACertificateProblemBeatsLookingLikeANetworkOne(t *testing.T) {
	err := fmt.Errorf("read tcp: %w", errors.Join(
		&net.OpError{Op: "read", Err: errors.New("connection reset")},
		x509.UnknownAuthorityError{},
	))

	if got := Failure(err); got != FailureCertificate {
		t.Errorf("want %q, got %q", FailureCertificate, got)
	}
}
