// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type PasswordReset struct {
	ID        string
	UserID    string
	Token     string
	CreatedAt int64
	ExpiresAt int64
}
