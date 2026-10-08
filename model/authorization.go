// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "time"

const (
	SessionCookieToken    = "tw_session"
	SessionCookieCsrf     = "tw_csrf"
	SessionCookieLoggedIn = "tw_logged_in"
)

type Credentials struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

type SessionDetails struct {
	AccessToken  string `json:"key"`
	UserID       string `json:"user_id"`
	IP           string `json:"ip"`
	DeviceID     string `json:"device_id"`
	IsOauth      bool   `json:"is_oauth"`
	Authorized   bool   `json:"authorized"`
	Browser      string `json:"browser"`
	CSRF         string `json:"csrf"`
	Created      int64  `json:"created"`
	LastActivity int64  `json:"last_activity"`
	Expires      int64  `json:"expires"`
}

// SessionInfo is the client-facing view: the access and CSRF tokens on
// SessionDetails must not leave the server.
type SessionInfo struct {
	ID           string `json:"id"`
	IP           string `json:"ip"`
	DeviceID     string `json:"device_id"`
	Browser      string `json:"browser"`
	Created      int64  `json:"created"`
	LastActivity int64  `json:"last_activity"`
	Expires      int64  `json:"expires"`
}

func (s *SessionDetails) IsAuthorized() bool {
	return s.Authorized
}

// Checks if user session is expired
func (s *SessionDetails) IsExpired() bool {
	return time.Now().Unix() > s.Expires
}
