// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	// File actions
	ActivityFileCreate       = "file_create"
	ActivityFileView         = "file_view"
	ActivityFileUpload       = "file_upload"
	ActivityFilePublicUpload = "file_public_upload"
	ActivityFilePublicEdit   = "file_public_edit"
	ActivityFileDelete       = "file_delete"
	ActivityFileRename       = "file_rename"
	ActivityFileRestore      = "file_restore"
	ActivityFileShare        = "file_share"
	ActivityFileMeta         = "file_meta"
	ActivityFileMoveFile     = "file_move"
	ActivityFileDownload     = "file_download"

	// Project actions
	ActivityProjectCreate  = "project_create"
	ActivityProjectDelete  = "project_delete"
	ActivityProjectDeleted = "project_deleted"
	ActivityProjectInvite  = "project_invite"

	// Task actions
	ActivityTaskStatusChanged = "task_status_changed"
	ActivityTaskCreated       = "task_created"
	ActivityTaskAssigned      = "task_assigned"
	ActivityTaskDeleted       = "task_deleted"
)

type Activity struct {
	ID             string
	App            string
	Type           string
	UserID         string
	UserName       string
	AffectedUser   string
	ItemID         string
	ParentID       string
	Parameters     map[string]interface{}
	ParametersJSON string `json:"-"`
	CreatedAt      int64
}
