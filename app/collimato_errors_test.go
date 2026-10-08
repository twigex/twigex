// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import "testing"

func TestPollingIsNotAnError(t *testing.T) {
	if got := CubeErrorID("Continue wait"); got != "" {
		t.Errorf("Continue wait must pass through untouched, got %q", got)
	}
}

func TestNoErrorIsNotAnError(t *testing.T) {
	if got := CubeErrorID(""); got != "" {
		t.Errorf("want no id, got %q", got)
	}
}

func TestNodesCertificateErrorsAreNamed(t *testing.T) {
	cases := []string{
		"Error: self-signed certificate; if the root CA is installed locally, try running Node.js with --use-system-ca",
		"Error: unable to verify the first certificate",
		"Error: unable to get local issuer certificate",
		"Error: certificate has expired",
		"Error: self signed certificate in certificate chain",
	}

	for _, message := range cases {
		if got := CubeErrorID(message); got != "connection.verification.certificate" {
			t.Errorf("%q: got %q", message, got)
		}
	}
}

// A name mismatch must not be blamed on the CA: the certificate is trusted,
// it just belongs to another name.
func TestANameMismatchIsItsOwnReason(t *testing.T) {
	cases := []string{
		"Hostname/IP does not match certificate's altnames: Host: db. is not in the cert's altnames",
		"Error: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs",
		"Error: Hostname/IP does not match certificate's altnames",
	}

	for _, message := range cases {
		if got := CubeErrorID(message); got != "connection.verification.hostname" {
			t.Errorf("%q: got %q", message, got)
		}
	}
}

func TestUnreachableIsNamed(t *testing.T) {
	for _, message := range []string{
		"Error: connect ETIMEDOUT",
		"Error: connect ECONNREFUSED 10.0.0.5:3306",
		"Error: getaddrinfo ENOTFOUND db.internal",
	} {
		if got := CubeErrorID(message); got != "connection.verification.unreachable" {
			t.Errorf("%q: got %q", message, got)
		}
	}
}

func TestCredentialsAndDatabaseAreNamed(t *testing.T) {
	if got := CubeErrorID("Error: Access denied for user 'cube'@'172.19.0.3'"); got != "connection.verification.credentials" {
		t.Errorf("credentials: got %q", got)
	}

	if got := CubeErrorID("Error: Unknown database 'nosuchdb'"); got != "connection.verification.database" {
		t.Errorf("database: got %q", got)
	}
}

func TestAnUnrecognisedErrorKeepsCubesOwnText(t *testing.T) {
	for _, message := range []string{
		"Error: Compile errors:\nprimary key for 'orders' is required when join is defined",
		"Error: Query timeout",
		"something nobody has seen before",
	} {
		if got := CubeErrorID(message); got != "" {
			t.Errorf("%q should not be rewritten, got %q", message, got)
		}
	}
}
