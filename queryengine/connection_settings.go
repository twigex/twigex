// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"fmt"
	"strconv"
	"strings"
)

// ConnectionSettings is everything needed to reach one database. The host app
// keeps them, with the password decrypted only for opening.
type ConnectionSettings struct {
	DatabaseType DatabaseType
	Host         string
	Port         string
	Database     string
	User         string
	Password     string
	SSLMode      SSLMode

	// CACertificate is the certificate, in PEM, that the database's
	// certificate must be signed by in the verifying SSL modes. Empty means
	// the authorities this machine trusts.
	CACertificate string
}

// SSLMode says whether a connection is encrypted and what is checked. The
// names are the ones Postgres's own tools use.
type SSLMode string

const (
	// SSLModeDisable does not encrypt.
	SSLModeDisable SSLMode = "disable"

	// SSLModeRequire encrypts but checks nothing about the server.
	SSLModeRequire SSLMode = "require"

	// SSLModeVerifyCA encrypts and checks that the server's certificate is
	// signed by the CA certificate, but not the name on it. With no CA
	// certificate it is the same as SSLModeVerifyFull.
	SSLModeVerifyCA SSLMode = "verify-ca"

	// SSLModeVerifyFull also checks that the certificate is for the host.
	SSLModeVerifyFull SSLMode = "verify-full"
)

func (settings ConnectionSettings) sslMode() SSLMode {
	if settings.SSLMode == "" {
		return SSLModeDisable
	}

	return settings.SSLMode
}

// portNumber reads the port, which must be a whole number up to 65535.
func (settings ConnectionSettings) portNumber() (uint16, error) {
	port, err := strconv.ParseUint(settings.Port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("query engine: bad port %q", settings.Port)
	}

	return uint16(port), nil
}

// checkHost refuses a host that is a file path. Postgres's driver reads one as
// a socket on this machine, which would reach a database the host app runs
// beside rather than one on the network.
func (settings ConnectionSettings) checkHost() error {
	if strings.HasPrefix(settings.Host, "/") {
		return fmt.Errorf("query engine: host %q is a file path, not a server", settings.Host)
	}

	return nil
}
