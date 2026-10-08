// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type ExternalShare struct {
	ID                string      `json:"id"`
	Owner             string      `json:"owner"`
	FileID            string      `json:"fileID"`
	PasswordProtected bool        `json:"passwordProtected"`
	Password          string      `json:"-"`
	Source            string      `json:"-"`
	Parent            string      `json:"parent"`
	ShareToken        string      `json:"shareToken"`
	ShareTime         int         `json:"shareTime"`
	Message           string      `json:"message"`
	Expiration        int64       `json:"expiration"`
	UserInfo          interface{} `json:"userInfo"`
	AllowView         bool        `json:"allowView"`
	AllowDownload     bool        `json:"allowDownload"`
	AllowUpload       bool        `json:"allowUpload"`
	AllowEdit         bool        `json:"allowEdit"`
	MaxDownloads      int         `json:"maxDownloads"`
	Downloaded        int         `json:"downloaded"`
	Accessed          int         `json:"accessed"`
	LastAccessedAt    int64       `json:"lastAccessedAt"`
	Active            bool        `json:"-"`
	CreatedAt         int64       `json:"-"`
	UpdatedAt         int64       `json:"-"`
}

type LinkUpdate struct {
	Token             string
	Expiration        int64
	PasswordProtected bool
	Password          string
	Message           string
	AllowView         bool
	AllowDownload     bool
	AllowUpload       bool
	AllowEdit         bool
	MaxDownloads      int
}

type Request struct {
	Item              string
	Expiration        int64
	PasswordProtected bool
	Password          string
	Folder            bool
	Message           string
	Parent            string
	Source            string
	AllowView         bool
	AllowDownload     bool
	AllowUpload       bool
	AllowEdit         bool
	MaxDownloads      int
}

type PublicFile struct {
	ID            string
	Name          string
	Size          int64
	Type          string
	IsFolder      bool
	Message       string
	Children      []PublicFile
	User          UserInfo
	AllowView     bool
	AllowDownload bool
	AllowUpload   bool
	AllowEdit     bool
}

type UserInfo struct {
	Name     string
	LastName string
}

type ShareResponse struct {
	SharedUsers  []SharedUsers `json:"shared_users"`
	SharedGroups []SharedGroup `json:"shared_groups,omitempty"`
}

type SharedGroup struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	MemberCount int         `json:"member_count"`
	Expiration  int64       `json:"expiration"`
	AccessLevel AccessLevel `json:"access_level"`
	Inherited   bool        `json:"inherited"`
	GrantedBy   string      `json:"granted_by,omitempty"`
}
