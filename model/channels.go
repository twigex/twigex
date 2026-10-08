// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

import "strings"

type Protocol int

const (
	ProtocolHTTPS Protocol = iota
	ProtocolWSS
)

const (
	ChannelTypePublic  = "O"
	ChannelTypePrivate = "P"
	ChannelTypeDirect  = "D"
)

const (
	ChannelRoleAdmin  = "admin"
	ChannelRoleMember = "member"
)

type Channel struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	DisplayName    string          `json:"displayname"`
	Name           string          `json:"name"`
	Header         string          `json:"header"`
	Description    string          `json:"description"`
	ChannelMembers []ChannelMember `json:"channel_members"`
	LastPost       int64           `json:"last_post"` // timestamp in milliseconds
	MessageCount   int64           `json:"msg_count"`
	CreatedBy      string          `json:"created_by"`
	CreatedAt      int64           `json:"created"`
	UpdatedAt      int64           `json:"updated"`
	DeletedAt      int64           `json:"deleted"`
}

func (c *Channel) Patch(patch ChannelPatch) {
	if patch.DisplayName != nil {
		c.DisplayName = *patch.DisplayName
		c.Name = strings.ToLower(strings.ReplaceAll(*patch.DisplayName, " ", ""))
	}

	if patch.Header != nil {
		c.Header = *patch.Header
	}

	if patch.Description != nil {
		c.Description = *patch.Description
	}
}

type ChannelPatch struct {
	DisplayName *string `json:"displayname"`
	Name        *string `json:"name"`
	Header      *string `json:"header"`
	Description *string `json:"description"`
}

type ChannelMember struct {
	ID           string             `json:"id"`
	ChannelID    string             `json:"channel_id"`
	UserID       string             `json:"user_id"`
	Role         string             `json:"role"`
	NotifyProps  ChannelNotifyProps `json:"notify_props"`
	MsgCount     int64              `json:"msg_count"`
	MentionCount int64              `json:"mention_count"`
	LastViewedAt int64              `json:"last_viewed_at"` // timestamp in milliseconds
	UpdatedAt    int64              `json:"updated"`        // timestamp in milliseconds
	DateJoined   int64              `json:"date_joined"`    // timestamp in milliseconds
	IsDirect     bool               `json:"is_direct"`
	UserInfo     *User              `json:"user_info,omitempty"`
}

type ChannelNotifyProps struct{}

// ChannelMemberUser is the minimal user info needed to display channel members
// (e.g. the meeting invitee picker) without loading the full user directory.
type ChannelMemberUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Username string `json:"username"`
}

const (
	MeetingStatusScheduled = "scheduled"
	MeetingStatusActive    = "active"
	MeetingStatusEnded     = "ended"
)

type ChannelMeeting struct {
	ID              string `json:"id"`
	ChannelID       string `json:"channel_id"`
	HostID          string `json:"host_id"`
	CreatedBy       string `json:"created_by"`
	Title           string `json:"title"`
	GuestPassword   string `json:"-"`
	DurationMinutes int    `json:"duration_minutes"`
	Sequence        int    `json:"sequence"`
	Timezone        string `json:"timezone"`
	InviteOnly      bool   `json:"invite_only"`
	ScheduledAt     *int64 `json:"scheduled_at"`
	StartedAt       *int64 `json:"started_at"`
	EndedAt         *int64 `json:"ended_at"`
	CreatedAt       int64  `json:"created_at"`

	// Populated at read time, not stored.
	Participants     int      `json:"participants"`
	Status           string   `json:"status"`
	HasGuestPassword bool     `json:"has_guest_password"`
	Invitees         []string `json:"invitees"`
}

const (
	DirectCallRinging   = "ringing"
	DirectCallConnected = "connected"
	DirectCallEnded     = "ended"
)

// Meeting is set only when Status is connected; a ringing call has no room yet.
type DirectCallResult struct {
	Status  string          `json:"status"`
	Meeting *ChannelMeeting `json:"meeting,omitempty"`
}

// Status derives the lifecycle state from the meeting's timestamps.
func (m *ChannelMeeting) computeStatus() string {
	switch {
	case m.EndedAt != nil:
		return MeetingStatusEnded
	case m.StartedAt != nil:
		return MeetingStatusActive
	default:
		return MeetingStatusScheduled
	}
}

// SetStatus stores the derived lifecycle state and the masked guest-password
// flag, and ensures Invitees serializes as an array rather than null.
func (m *ChannelMeeting) SetStatus() {
	m.Status = m.computeStatus()
	m.HasGuestPassword = m.GuestPassword != ""
	if m.Invitees == nil {
		m.Invitees = []string{}
	}
}

type ChannelGuestLink struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	MeetingID    string `json:"meeting_id"`
	CreatedBy    string `json:"created_by"`
	InvitedEmail string `json:"invited_email"`
	ExpiresAt    int64  `json:"expires_at"`
	UsedAt       *int64 `json:"used_at"`
	RevokedAt    *int64 `json:"revoked_at"`
	CreatedAt    int64  `json:"created_at"`
}
