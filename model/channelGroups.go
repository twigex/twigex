// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

// ChannelGroup links a channel to a group. Effective channel membership is
// the union of direct channel_members rows and (for each ChannelGroup) the
// group's current members, resolved live, not snapshotted.
type ChannelGroup struct {
	ID          string `json:"id"`
	ChannelID   string `json:"channel_id"`
	GroupID     string `json:"group_id"`
	AddedBy     string `json:"added_by"`
	AddedAt     int64  `json:"added_at"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
}

type AddChannelGroupsRequest struct {
	GroupIDs []string `json:"group_ids"`
}

// ChannelMeetingGroup links a meeting to a group. Access is resolved live:
// any current member of an invited group may join the meeting.
type ChannelMeetingGroup struct {
	ID          string `json:"id"`
	MeetingID   string `json:"meeting_id"`
	GroupID     string `json:"group_id"`
	AddedBy     string `json:"added_by"`
	AddedAt     int64  `json:"added_at"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
}

type AddMeetingGroupsRequest struct {
	GroupIDs []string `json:"group_ids"`
}
