// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package catalog

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

const CACertKey = "ca_cert"

var (
	errBadCACert         = errors.New("catalog: no certificate found in ca_cert")
	errNoPeerCertificate = errors.New("catalog: server sent no certificate")
)

func certPool(pem string) (*x509.CertPool, error) {
	if pem == "" {
		return nil, nil
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM([]byte(pem)) {
		return nil, errBadCACert
	}

	return pool, nil
}

func chainOnly(pool *x509.CertPool) func(tls.ConnectionState) error {
	return func(state tls.ConnectionState) error {
		if len(state.PeerCertificates) == 0 {
			return errNoPeerCertificate
		}

		opts := x509.VerifyOptions{Roots: pool, Intermediates: x509.NewCertPool()}
		for _, cert := range state.PeerCertificates[1:] {
			opts.Intermediates.AddCert(cert)
		}

		_, err := state.PeerCertificates[0].Verify(opts)

		return err
	}
}

func (c *Connection) tlsConfig() (*tls.Config, error) {
	mode := c.mode()

	if mode == SSLModeDisable {
		return nil, nil
	}

	pool, err := certPool(c.Config[CACertKey])
	if err != nil {
		return nil, err
	}

	switch mode {
	case SSLModeRequire:
		return &tls.Config{InsecureSkipVerify: true}, nil
	case SSLModeVerifyCA:
		return &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: true,
			VerifyConnection:   chainOnly(pool),
		}, nil
	case SSLModeVerifyFull:
		return &tls.Config{RootCAs: pool, ServerName: c.Host}, nil
	}

	return nil, fmt.Errorf("catalog: unknown ssl mode %s", mode)
}

func registerMysqlTLS(c *Connection) (string, error) {
	cfg, err := c.tlsConfig()
	if err != nil {
		return "", err
	}

	if cfg == nil {
		return "false", nil
	}

	sum := sha256.Sum256([]byte(c.mode() + "\x00" + c.Host + "\x00" + c.Config[CACertKey]))
	name := "twigex-" + hex.EncodeToString(sum[:8])

	if err := mysql.RegisterTLSConfig(name, cfg); err != nil {
		return "", err
	}

	return name, nil
}
