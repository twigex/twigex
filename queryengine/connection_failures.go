// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
)

// ConnectionFailure says why a database could not be reached, in words a
// person can act on. The host app turns it into text in its languages.
type ConnectionFailure string

const (
	ConnectionFailureCredentials   ConnectionFailure = "credentials"
	ConnectionFailureDatabase      ConnectionFailure = "database"
	ConnectionFailureCertificate   ConnectionFailure = "certificate"
	ConnectionFailureHostname      ConnectionFailure = "hostname"
	ConnectionFailureCACertificate ConnectionFailure = "ca_cert"
	ConnectionFailureUnreachable   ConnectionFailure = "unreachable"
	ConnectionFailureUnknown       ConnectionFailure = "unknown"
)

// ConnectionFailureOf names the reason behind an error from opening,
// checking or using a connection. No error has no reason.
func ConnectionFailureOf(err error) ConnectionFailure {
	if err == nil {
		return ""
	}

	if errors.Is(err, errBadCACertificate) {
		return ConnectionFailureCACertificate
	}

	if isHostnameFailure(err) {
		return ConnectionFailureHostname
	}

	if isCertificateFailure(err) {
		return ConnectionFailureCertificate
	}

	// Each driver knows only its own errors, so the order does not matter.
	for _, database := range supportedDatabases {
		failure := database.connectionFailure(err)
		if failure != "" {
			return failure
		}
	}

	if isUnreachable(err) {
		return ConnectionFailureUnreachable
	}

	return ConnectionFailureUnknown
}

// isHostnameFailure is checked before isCertificateFailure: a wrong name
// arrives wrapped in a certificate error, but the CA is not what is wrong.
func isHostnameFailure(err error) bool {
	var hostnameError x509.HostnameError

	return errors.As(err, &hostnameError)
}

func isCertificateFailure(err error) bool {
	var verificationError *tls.CertificateVerificationError
	if errors.As(err, &verificationError) {
		return true
	}

	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return true
	}

	var invalidCertificate x509.CertificateInvalidError
	if errors.As(err, &invalidCertificate) {
		return true
	}

	return errors.Is(err, errNoPeerCertificate)
}

func isUnreachable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var networkError net.Error

	return errors.As(err, &networkError)
}
