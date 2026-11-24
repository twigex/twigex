// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

var (
	errBadCACertificate  = errors.New("query engine: no certificate found in the CA certificate")
	errNoPeerCertificate = errors.New("query engine: the server sent no certificate")
)

// tlsConfig is the encryption a connection uses, or nil for none, for the
// four modes of SSLMode.
func tlsConfig(settings ConnectionSettings) (*tls.Config, error) {
	mode := settings.sslMode()

	if mode == SSLModeDisable {
		return nil, nil
	}

	pool, err := certificatePool(settings.CACertificate)
	if err != nil {
		return nil, err
	}

	switch mode {
	case SSLModeRequire:
		return &tls.Config{InsecureSkipVerify: true}, nil

	case SSLModeVerifyCA:
		// Without a CA certificate of its own, a check of the chain alone
		// would pass any certificate a public authority issued, for any
		// name. The name is checked as well then, as Postgres's own tools do.
		if pool == nil {
			return &tls.Config{ServerName: settings.Host}, nil
		}

		// Go checks the name together with the chain, so the chain is
		// checked here instead, without the name.
		config := &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: true,
			VerifyConnection:   verifyChainOnly(pool),
		}

		return config, nil

	case SSLModeVerifyFull:
		return &tls.Config{RootCAs: pool, ServerName: settings.Host}, nil
	}

	return nil, fmt.Errorf("query engine: unknown SSL mode %q", mode)
}

// certificatePool is the CA certificate as something to check against, or
// nil for the authorities this machine trusts.
func certificatePool(pem string) (*x509.CertPool, error) {
	if pem == "" {
		return nil, nil
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(pem)) {
		return nil, errBadCACertificate
	}

	return pool, nil
}

func verifyChainOnly(pool *x509.CertPool) func(tls.ConnectionState) error {
	return func(state tls.ConnectionState) error {
		if len(state.PeerCertificates) == 0 {
			return errNoPeerCertificate
		}

		options := x509.VerifyOptions{Roots: pool, Intermediates: x509.NewCertPool()}
		for _, certificate := range state.PeerCertificates[1:] {
			options.Intermediates.AddCert(certificate)
		}

		_, err := state.PeerCertificates[0].Verify(options)

		return err
	}
}
