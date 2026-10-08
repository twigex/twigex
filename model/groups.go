// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const GroupRoleMember = "member"

type Group struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	DeletedAt   int64  `json:"deleted_at,omitempty"`
	MemberCount int    `json:"member_count,omitempty"`
	// Roles are the system role names granted to every member of this group.
	// Membership is additive on top of each member's own user.Role.
	Roles []string `json:"roles"`
}

type GroupMember struct {
	ID       string `json:"id"`
	GroupID  string `json:"group_id"`
	UserID   string `json:"user_id"`
	Role     string `json:"role"`
	JoinedAt int64  `json:"joined_at"`
	UserInfo *User  `json:"user_info,omitempty"`
}

type GroupsPage struct {
	Items []Group `json:"items"`
	Total int     `json:"total"`
}

type GroupMembersPage struct {
	Items []GroupMember `json:"items"`
	Total int           `json:"total"`
}

type CreateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"member_ids,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}

type UpdateGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roles       []string `json:"roles,omitempty"`
}

type AddGroupMembersRequest struct {
	UserIDs []string `json:"user_ids"`
}
