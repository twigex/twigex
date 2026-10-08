// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	FailureCredentials = "credentials"
	FailureDatabase    = "database"
	FailureCertificate = "certificate"
	FailureHostname    = "hostname"
	FailureCACert      = "ca_cert"
	FailureUnreachable = "unreachable"
	FailureUnknown     = "unknown"
)

func Failure(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, errBadCACert) {
		return FailureCACert
	}

	if isHostname(err) {
		return FailureHostname
	}

	if isCertificate(err) {
		return FailureCertificate
	}

	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		switch myErr.Number {
		case 1044, 1045, 1698:
			return FailureCredentials
		case 1049:
			return FailureDatabase
		}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "28000", "28P01":
			return FailureCredentials
		case "3D000":
			return FailureDatabase
		}
	}

	if isUnreachable(err) {
		return FailureUnreachable
	}

	return FailureUnknown
}

// isHostname is checked before isCertificate: a name mismatch arrives wrapped
// in a CertificateVerificationError, and the CA is not the setting at fault.
func isHostname(err error) bool {
	var hostname x509.HostnameError

	return errors.As(err, &hostname)
}

func isCertificate(err error) bool {
	var verification *tls.CertificateVerificationError
	if errors.As(err, &verification) {
		return true
	}

	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return true
	}

	var invalid x509.CertificateInvalidError
	if errors.As(err, &invalid) {
		return true
	}

	return errors.Is(err, errNoPeerCertificate)
}

func isUnreachable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var netErr net.Error

	return errors.As(err, &netErr)
}
