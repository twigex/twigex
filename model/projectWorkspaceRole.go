// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	ProjectWorkspaceAdminRoleID  = "admin"
	ProjectWorkspaceUserRoleID   = "user"
	ProjectWorkspaceViewerRoleID = "viewer"
	ProjectWorkspaceEditorRoleID = "editor"
)

// Note: viewer and editor are legacy role names kept as constants for the
// doRemoveViewerEditorRolesMigration cleanup. New workspaces only get admin and user.

type ProjectWorkspaceRole struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	DisplayName  string   `json:"display_name"`
	Description  string   `json:"description"`
	Permissions  []string `json:"permissions"`
	WorkspaceID  string   `json:"workspace_id"`
	PerTableMode bool     `json:"per_table_mode"`
	CreatedAt    int64    `json:"created_at"`
	UpdatedAt    int64    `json:"updated_at"`
}

type ProjectWorkspaceRolePatch struct {
	DisplayName  *string   `json:"display_name"`
	Description  *string   `json:"description"`
	Permissions  *[]string `json:"permissions"`
	PerTableMode *bool     `json:"per_table_mode"`
}

func (p *ProjectWorkspaceRole) Patch(patch *ProjectWorkspaceRolePatch) {
	if patch.DisplayName != nil {
		p.DisplayName = *patch.DisplayName
	}

	if patch.Description != nil {
		p.Description = *patch.Description
	}

	if patch.Permissions != nil {
		p.Permissions = *patch.Permissions
	}

	if patch.PerTableMode != nil {
		p.PerTableMode = *patch.PerTableMode
	}
}

func MakeDefaultProjectWorkspaceRoles() map[string]*ProjectWorkspaceRole {
	roles := make(map[string]*ProjectWorkspaceRole)

	roles[ProjectWorkspaceUserRoleID] = &ProjectWorkspaceRole{
		Name:        ProjectWorkspaceUserRoleID,
		DisplayName: "authentication.roles.user.name",
		Description: "authentication.roles.user.description",
		Permissions: []string{
			PermissionCreateTask,
			PermissionUpdateTask,
			PermissionUpdateWorkspaceView,
			PermissionEditFields,
			PermissionViewGrid,
			PermissionViewKanban,
			PermissionViewCalendar,
			PermissionViewGantt,
			PermissionViewForm,
		},
	}

	roles[ProjectWorkspaceAdminRoleID] = &ProjectWorkspaceRole{
		Name:        ProjectWorkspaceAdminRoleID,
		DisplayName: "authentication.roles.admin.name",
		Description: "authentication.roles.admin.description",
		Permissions: []string{
			PermissionCreateTask,
			PermissionUpdateTask,
			PermissionDeleteTask,
			PermissionCreateRoles,
			PermissionUpdateRoles,
			PermissionDeleteRoles,
			PermissionUpdateWorkspaceFolder,
			PermissionDeleteWorkspaceFolder,
			PermissionDeleteWorkspaceMember,
			PermissionCreateWorkspaceView,
			PermissionUpdateWorkspaceView,
			PermissionCreateSingleSelectField,
			PermissionEditFields,
			PermissionCreateFields,
			PermissionUpdateWorkspaceTable,
			PermissionUpdateWorkspace,
			PermissionDeleteTableView,
			PermissionDeleteWorkspaceTable,
			PermissionAddMemberToWorkspace,
			PermissionCreateTable,
			PermissionViewGrid,
			PermissionViewKanban,
			PermissionViewCalendar,
			PermissionViewGantt,
			PermissionViewForm,
		},
	}

	return roles
}
