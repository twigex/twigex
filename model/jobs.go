// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "encoding/json"

const (
	JobTypeMoveFiles     = "move_files"
	JobTypeLDAPSync      = "ldap_sync"
	JobTypeCleanupJobs   = "cleanup_jobs"
	JobTypeDeleteChannel = "delete_channel"

	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusFailed    = "failed"
	JobStatusCompleted = "completed"
	JobStatusCancelled = "cancelled"
	JobStatusRefused   = "refused"
)

func IsJobFinished(status string) bool {
	switch status {
	case JobStatusCompleted, JobStatusFailed, JobStatusCancelled, JobStatusRefused:
		return true
	}

	return false
}

type Job struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`

	UserID string `json:"user_id"`

	Payload json.RawMessage `json:"payload"`

	Progress int `json:"progress"`

	Error *string `json:"error,omitempty"`

	CreatedAt  int64 `json:"created_at"`
	StartedAt  int64 `json:"started_at"`
	FinishedAt int64 `json:"finished_at"`
	UpdatedAt  int64 `json:"updated_at"`

	AcknowledgedAt *int64 `json:"acknowledged_at,omitempty"`
}

func (r *Job) GetPayload() (*MoveFilesJobPayload, error) {
	var payload MoveFilesJobPayload

	err := json.Unmarshal(r.Payload, &payload)
	if err != nil {
		return nil, err
	}

	return &payload, nil
}

func (r *Job) LDAPSyncPayload() (*LDAPSyncJobPayload, error) {
	var payload LDAPSyncJobPayload

	if len(r.Payload) == 0 {
		return &payload, nil
	}

	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

func (r *Job) GetAllFileIDs() ([]string, error) {
	payload, err := r.GetPayload()
	if err != nil {
		return nil, err
	}

	allFiles := make([]string, 0)
	for _, item := range payload.RootIDS {
		allFiles = append(allFiles, item.ID)
	}

	for _, item := range payload.SubfileIDs {
		allFiles = append(allFiles, item.ID)
	}

	//deduped IDs
	uniqueIDs := make(map[string]struct{})
	for _, id := range allFiles {
		uniqueIDs[id] = struct{}{}
	}

	var result []string
	for id := range uniqueIDs {
		result = append(result, id)
	}

	return result, nil
}

type LDAPSyncJobPayload struct {
	Deactivated int `json:"deactivated"`
	Reactivated int `json:"reactivated"`
	Updated     int `json:"updated"`
	Failed      int `json:"failed"`
}

type MoveFilesJobPayload struct {
	UserID     string     `json:"user_id"`
	RootIDS    []MoveItem `json:"root_ids"`
	SubfileIDs []MoveItem `json:"subfile_ids"`
	Target     string     `json:"target"`
	TotalBytes int64      `json:"total_bytes"`
}

type DeleteChannelPayload struct {
	ChannelID string `json:"channel_id"`
}

func (r *Job) GetDeleteChannelPayload() (*DeleteChannelPayload, error) {
	var payload DeleteChannelPayload
	if err := json.Unmarshal(r.Payload, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
