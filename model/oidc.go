// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type OIDCUser struct {
	Subject       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	Username      string
	PictureUrl    string
}
