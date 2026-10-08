// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package mfa

import (
	"bytes"
	"encoding/base32"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateSecretProducesAScannableQRCode(t *testing.T) {
	secret, qr, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret); err != nil {
		t.Errorf("secret is not valid unpadded base32: %v", err)
	}

	if !bytes.HasPrefix(qr, []byte("\x89PNG\r\n\x1a\n")) {
		t.Error("qr code is not a PNG")
	}
}

func TestGenerateSecretIsUnique(t *testing.T) {
	first, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Error("two enrolments produced the same secret")
	}
}

func TestValidateTokenAcceptsTheCurrentCode(t *testing.T) {
	secret, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	ok, err := ValidateToken(secret, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("the current code was rejected")
	}
}

// Surrounding whitespace comes from users pasting the code out of an app.
func TestValidateTokenTrimsWhitespace(t *testing.T) {
	secret, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	ok, err := ValidateToken(secret, "  "+code+" \n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("a padded code was rejected")
	}
}

// A wrong code is a failed attempt, not a fault: returning an error here would
// surface a typo to the user as a server error.
func TestValidateTokenRejectsAWrongCodeWithoutError(t *testing.T) {
	secret, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	wrong := "000000"
	if code == wrong {
		wrong = "111111"
	}

	ok, err := ValidateToken(secret, wrong)
	if err != nil {
		t.Fatalf("a wrong code must not error: %v", err)
	}
	if ok {
		t.Error("a wrong code was accepted")
	}
}

func TestValidateTokenRejectsCodeFromAnotherSecret(t *testing.T) {
	secret, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	other, _, err := GenerateSecret("other@example.com")
	if err != nil {
		t.Fatal(err)
	}

	code, err := totp.GenerateCode(other, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	ok, _ := ValidateToken(secret, code)
	if ok {
		t.Error("a code minted from another secret was accepted")
	}
}

func TestValidateTokenRejectsAnExpiredCode(t *testing.T) {
	secret, _, err := GenerateSecret("user@example.com")
	if err != nil {
		t.Fatal(err)
	}

	// Two periods back sits outside the one-step skew either side.
	stale, err := totp.GenerateCode(secret, time.Now().UTC().Add(-90*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	ok, _ := ValidateToken(secret, stale)
	if ok {
		t.Error("a code from two periods ago was accepted")
	}
}

// Secrets already in the database were written with padded base32 by the
// previous implementation. They have to keep validating, or every enrolled user
// is locked out by the upgrade.
func TestValidateTokenAcceptsPaddedLegacySecrets(t *testing.T) {
	raw := bytes.Repeat([]byte{0x2a}, secretBytes)
	legacy := base32.StdEncoding.EncodeToString(raw)

	code, err := totp.GenerateCode(legacy, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}

	ok, err := ValidateToken(legacy, code)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("a secret stored by the previous implementation was rejected")
	}
}
