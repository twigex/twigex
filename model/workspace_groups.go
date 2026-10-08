// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

type WorkspaceGroup struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	GroupID     string   `json:"group_id"`
	Roles       []string `json:"roles"`
	AddedBy     string   `json:"added_by"`
	AddedAt     int64    `json:"added_at"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	MemberCount int      `json:"member_count"`
}

type AddProjectGroupsRequest struct {
	GroupIDs []string `json:"group_ids"`
	Roles    []string `json:"roles"`
}

type UpdateProjectGroupRolesRequest struct {
	Roles []string `json:"roles"`
}
