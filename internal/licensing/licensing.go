// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package licensing

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/twigex/twigex/model"
)

func VerifyLicense(bytes string) (*model.License, error) {
	block, _ := pem.Decode(model.PB) // Public key in PEM format
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the public key")
	}

	publicKeyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %s", err.Error())
	}

	rsaPublic, ok := publicKeyInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	token, err := parser.Parse(bytes, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return rsaPublic, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %s", err.Error())
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		licenseData, err := json.Marshal(claims["license"])
		if err != nil {
			return nil, fmt.Errorf("failed to marshal license claims: %s", err.Error())
		}

		license := &model.License{}
		if err := json.Unmarshal(licenseData, license); err != nil {
			return nil, fmt.Errorf("failed to unmarshal license data: %s", err.Error())
		}

		return license, nil
	}

	return nil, fmt.Errorf("invalid token")
}
