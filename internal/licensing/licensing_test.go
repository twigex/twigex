// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package licensing

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/model"
)

func TestVerifyLicenseRejectsBadSignatures(t *testing.T) {
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	claims := jwt.MapClaims{
		"license": map[string]any{"id": "forged"},
		"iat":     time.Now().Unix(),
		"nbf":     time.Now().Add(-time.Hour).Unix(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	wrongKey, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(otherKey)
	if err != nil {
		t.Fatal(err)
	}

	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	hmacSigned, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(model.PB)
	if err != nil {
		t.Fatal(err)
	}

	for name, token := range map[string]string{
		"signed with another key":             wrongKey,
		"alg none":                            unsigned,
		"hmac using the public key as secret": hmacSigned,
		"garbage":                             "not.a.token",
		"empty":                               "",
	} {
		if _, err := VerifyLicense(token); err == nil {
			t.Errorf("%s: expected rejection, got none", name)
		}
	}
}
