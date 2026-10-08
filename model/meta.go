// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type FileMetadata struct {
	ID        string      `json:"id"`
	FileID    string      `json:"file_id"`
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Fields    interface{} `json:"fields"`
	CreatedAt int64       `json:"created"`
	UpdatedAt int64       `json:"updated"`
	CreatedBy string      `json:"created_by"`
}

type FileMetadataEntry struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Type       string      `json:"type"`
	FileID     string      `json:"file_id"`
	MetadataID string      `json:"metadata_id"`
	Value      interface{} `json:"value"`
	Fields     interface{} `json:"fields"`
	CreatedAt  int64       `json:"created"`
	UpdatedAt  int64       `json:"updated"`
}
