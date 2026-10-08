// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import (
	"regexp"
	"strings"
)

const UsernameMaxLength = 64

var usernamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]*[a-z0-9])?$`)

func NormalizeUsername(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func ValidUsername(s string) bool {
	if s == "" || len(s) > UsernameMaxLength {
		return false
	}

	return usernamePattern.MatchString(s)
}

// all and here are broadcast keywords in resolveHandles, and unknown-user is
// what a mention renders as when nobody holds the id. A user answering to any
// of them cannot be told apart from the thing it collides with.
var reservedUsernames = map[string]struct{}{
	"all":          {},
	"here":         {},
	"unknown-user": {},
}

// ReservedUsername is separate from ValidUsername so a caller can say which of
// the two a name failed, since only one of them is worth explaining.
func ReservedUsername(s string) bool {
	_, ok := reservedUsernames[NormalizeUsername(s)]
	return ok
}

type User struct {
	ID            string         `json:"id"`
	Email         string         `json:"email"`
	Role          string         `json:"role"`
	Password      string         `json:"-"`
	AuthService   string         `json:"auth_service"`
	AuthData      string         `json:"-"`
	Username      string         `json:"username"`
	Name          string         `json:"name"`
	LastName      string         `json:"lastname"`
	StorageLimit  int64          `json:"storage_limit"`
	Photo         string         `json:"photo"`
	DriveID       string         `json:"-"`
	Timezone      map[string]any `json:"timezone"`
	MfaActive     bool           `json:"mfaActive"`
	MfaSecret     string         `json:"-"`
	CreatedAt     int64          `json:"-"`
	UpdatedAt     int64          `json:"-"`
	DeactivatedAt int64          `json:"deactivated_at"`
}

type UsersPage struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
}

type SharedUsers struct {
	ID          string      `json:"id"`
	Owner       bool        `json:"owner"`
	Name        string      `json:"name"`
	LastName    string      `json:"lastname"`
	Email       string      `json:"email"`
	Photo       string      `json:"photo"`
	Expiration  int64       `json:"expiration"`
	AccessLevel AccessLevel `json:"accessLevel"`
	FileID      int         `json:"-"`
	Inherited   bool        `json:"inherited"`
	GrantedBy   string      `json:"grantedBy,omitempty"`
}

type UserPatch struct {
	ID           string
	Email        string
	Username     string
	Role         string
	Name         string
	LastName     string
	StorageLimit int64 `json:"storage_limit"`
	Photo        string
}

type UserProfileRequest struct {
	Name                 string `json:"name"`
	LastName             string `json:"lastname"`
	Email                string `json:"email"`
	ClockDisplay         bool   `json:"clockDisplay"`
	AutomaticTimezone    string `json:"automaticTimezone"`
	ManualTimezone       string `json:"manualTimezone"`
	UseAutomaticTimezone bool   `json:"useAutomaticTimezone"`
	Language             string `json:"language"`
}

type NewUser struct {
	ID            string
	Email         string
	Role          string
	Password      string
	AuthService   string
	Username      string
	Name          string
	LastName      string
	Timezone      string
	ClockDisplay  string
	StorageLimit  int64 `json:"storage_limit"`
	Photo         string
	DeactivatedAt int64
	AuthData      string
}

type UserStatus struct {
	UserID       string `json:"user_id"`
	Status       string `json:"status"`
	LastActivity int64  `json:"last_activity"`
	UserDefined  bool   `json:"user_defined"`
}
