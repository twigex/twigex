// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	SHARE_TYPE_USER  = 1
	SHARE_TYPE_GROUP = 2
)

type FileShare struct {
	Users       []string
	Groups      []string
	FileID      string
	Expiration  int64
	AccessLevel AccessLevel
}

type FileSharePatch struct {
	ID          string
	FileID      string
	ShareType   int // SHARE_TYPE_USER or SHARE_TYPE_GROUP; 0 defaults to user
	AccessLevel AccessLevel
	Expiration  int64
}
