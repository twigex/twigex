// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"encoding/json"
	"strconv"
)

type UploadSessionStatus string

const (
	UploadStatusPending   UploadSessionStatus = "pending"
	UploadStatusUploading UploadSessionStatus = "uploading"
	UploadStatusCompleted UploadSessionStatus = "completed"
	UploadStatusFailed    UploadSessionStatus = "failed"
	UploadStatusExpired   UploadSessionStatus = "expired"
)

type UploadSession struct {
	ID       string `json:"id" db:"id"`
	UploadID string `json:"upload_id" db:"upload_id"`
	UserID   string `json:"user_id" db:"user_id"`
	Storage  string `json:"storage" db:"storage"`

	ContextType string `json:"context_type" db:"context_type"`
	ContextID   string `json:"context_id,omitempty" db:"context_id"`

	FileName string `json:"file_name" db:"file_name"`
	MimeType string `json:"mime_type" db:"mime_type"`

	TotalSize     int64 `json:"total_size" db:"total_size"`
	UploadedSize  int64 `json:"uploaded_size" db:"uploaded_size"`
	UploadedParts int   `json:"uploaded_parts" db:"uploaded_parts"`

	Status      UploadSessionStatus `json:"status" db:"status"`
	ErrorReason *string             `json:"error_reason,omitempty" db:"error_reason"`

	PartsJSON       *string `json:"parts_json,omitempty" db:"parts_json"`
	BackendMetadata *string `json:"backend_metadata,omitempty" db:"backend_metadata"`

	CreatedAt int64 `json:"created_at" db:"created_at"`
	UpdatedAt int64 `json:"updated_at" db:"updated_at"`
	ExpiresAt int64 `json:"expires_at,omitempty" db:"expires_at"`
}

func (s *UploadSession) GetParts() (map[int]string, error) {
	if s.PartsJSON == nil || *s.PartsJSON == "" {
		return make(map[int]string), nil
	}

	var parts map[int]string
	err := json.Unmarshal([]byte(*s.PartsJSON), &parts)
	if err != nil {
		var stringParts map[string]string
		if err2 := json.Unmarshal([]byte(*s.PartsJSON), &stringParts); err2 == nil {
			parts = make(map[int]string)
			for k, v := range stringParts {
				if partNum, err := strconv.Atoi(k); err == nil {
					parts[partNum] = v
				}
			}

			return parts, nil
		}

		return nil, err
	}

	return parts, nil
}

func (s *UploadSession) SetParts(parts map[int]string) error {
	if parts == nil {
		empty := ""
		s.PartsJSON = &empty
		return nil
	}

	data, err := json.Marshal(parts)
	if err != nil {
		return err
	}

	jsonStr := string(data)
	s.PartsJSON = &jsonStr
	return nil
}
