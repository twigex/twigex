// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "time"

type File struct {
	ID            string              `json:"id"`
	Owner         string              `json:"owner"`
	Parent        string              `json:"parentDir"`
	Storage       string              `json:"storage"`
	Name          string              `json:"-"`
	DisplayName   string              `json:"name"`
	OldName       string              `json:"-"` // used only internally
	Shared        bool                `json:"shared"`
	Children      []interface{}       `json:"children"`
	Size          int64               `json:"size"`
	Type          string              `json:"type"`
	Metadata      []FileMetadataEntry `json:"metadata"`
	IsFolder      bool                `json:"isFolder"`
	Favourite     bool                `json:"favourite"`
	Expiration    int64               `json:"expiration"`
	AccessLevel   AccessLevel         `json:"accessLevel"`
	Created       int64               `json:"created"`
	Modified      int64               `json:"modified"`
	DeletedAt     int64               `json:"deleted_at"`
	Trash         bool                `json:"-"`
	LockOwner     *string             `json:"-"`
	LockExpiresAt *int64              `json:"-"`
	Version       int64               `json:"-"`
}

type SharedLinks struct {
	ID                string `json:"id"`
	FullName          string `json:"name"`
	ShareTime         int    `json:"shareTime"`
	Expiration        int64  `json:"expiration"`
	PasswordProtected bool   `json:"passwordProtected"`
	AllowView         bool   `json:"allowView"`
	AllowDownload     bool   `json:"allowDownload"`
	AllowUpload       bool   `json:"allowUpload"`
	AllowEdit         bool   `json:"allowEdit"`
	MaxDownloads      int    `json:"maxDownloads"`
	Downloaded        int    `json:"downloaded"`
	Message           string `json:"message"`
}

type FileDetails struct {
	ID           string        `json:"id"`
	Type         string        `json:"type"`
	Owner        string        `json:"owner"`
	OwnerEmail   string        `json:"ownerEmail"`
	AccessLevel  AccessLevel   `json:"accessLevel"` // the requesting user's effective level
	Modified     int64         `json:"modified"`
	Created      int64         `json:"created"`
	SharedUsers  []SharedUsers `json:"sharedUsers"`
	SharedGroups []SharedGroup `json:"sharedGroups"`
	SharedLinks  []SharedLinks `json:"sharedLinks"`
}

type SharedFile struct {
	ID          string      `json:"id"`
	FileID      string      `json:"file_id"`
	Parent      string      `json:"parentDir"`
	Initiator   string      `json:"initiator"`
	ShareType   int         `json:"share_type"`
	ShareWith   string      `json:"share_with"`
	Favourite   bool        `json:"favourite"`
	Expiration  int64       `json:"expiration"`
	AccessLevel AccessLevel `json:"accessLevel"`
	TimeShared  int64       `json:"time_shared"`
	CreatedAt   int64       `json:"created_at"`
	UpdatedAt   int64       `json:"updated_at"`
}

// used by api in file creation handlers
type FileCrateRequest struct {
	DocName   string
	ID        string
	Type      string
	Extension string
}

type FileMove struct {
	Immediate bool            `json:"immediate"`
	Payload   FileMovePayload `json:"payload"`
}

type FileMovePayload struct {
	UserID     string     `json:"user_id"`
	RootIDs    []MoveItem `json:"root_ids"`
	SubfileIDs []MoveItem `json:"subfile_ids"`
	Target     string     `json:"target"`
}

type MoveItem struct {
	ID       string `json:"id"`
	IsFolder bool   `json:"is_folder"`
	Target   string `json:"target"`
	Source   string `json:"source"`
}

type MoveFileResult struct {
	Job      *Job      `json:"job,omitempty"`
	FileMove *FileMove `json:"file_move,omitempty"`
}

func (f *File) IsOwner(userID string) bool {
	return f.Owner == userID
}

func (f *File) IsLocked() bool {
	if f.LockOwner == nil {
		return false
	}

	if f.LockExpiresAt == nil {
		return false
	}

	return time.Now().Unix() <= *f.LockExpiresAt
}

func (f *File) Restore() {
	f.Trash = false
}
