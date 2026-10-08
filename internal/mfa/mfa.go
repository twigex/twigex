// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Package mfa issues and verifies time-based one-time passwords (RFC 6238)
// for use as a second authentication factor.
package mfa

import (
	"bytes"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	issuer = "Twigex"

	// RFC 4226 recommends a 160-bit shared secret.
	secretBytes = 20

	qrPixels = 256
)

var validateOpts = totp.ValidateOpts{
	Period:    30,
	Skew:      1,
	Digits:    otp.DigitsSix,
	Algorithm: otp.AlgorithmSHA1,
}

func GenerateSecret(userEmail string) (string, []byte, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: userEmail,
		Period:      validateOpts.Period,
		SecretSize:  secretBytes,
		Digits:      validateOpts.Digits,
		Algorithm:   validateOpts.Algorithm,
	})
	if err != nil {
		return "", nil, fmt.Errorf("generating totp key: %w", err)
	}

	img, err := key.Image(qrPixels, qrPixels)
	if err != nil {
		return "", nil, fmt.Errorf("rendering qr code: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", nil, fmt.Errorf("encoding qr code: %w", err)
	}

	return key.Secret(), buf.Bytes(), nil
}

func ValidateToken(secret, token string) (bool, error) {
	ok, err := totp.ValidateCustom(strings.TrimSpace(token), secret, time.Now().UTC(), validateOpts)
	if err != nil {
		return false, fmt.Errorf("validating totp token: %w", err)
	}

	return ok, nil
}
