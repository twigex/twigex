// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	LocalStorage = 1
	S3Storage    = 2
)

type Storage struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	CreatedAt   int64  `json:"created_at"`

	// Local storage
	Directory string `json:"directory,omitempty"`

	// S3 storage. SecretKey is decrypted in memory, encrypted in DB.
	Endpoint  string `json:"endpoint,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
	Bucket    string `json:"bucket,omitempty"`
	SSL       bool   `json:"ssl,omitempty"`
	Primary   bool   `json:"is_primary"`
}

func (s *Storage) DriverName() string {
	switch s.Type {
	case S3Storage:
		return "s3"
	case LocalStorage:
		return "local"
	default:
		return "local"
	}
}
