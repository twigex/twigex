// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

// TablePermission is one row in workspace_table_permissions.
type TablePermission struct {
	ID          string `json:"id"`
	RoleID      string `json:"role_id"`
	WorkspaceID string `json:"workspace_id"`
	TableID     string `json:"table_id"`
	Action      string `json:"action"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

const (
	PermissionCreateTask              = "create_task"
	PermissionUpdateTask              = "update_task"
	PermissionDeleteTask              = "delete_task"
	PermissionCreateRoles             = "create_roles"
	PermissionUpdateRoles             = "update_roles"
	PermissionDeleteRoles             = "delete_roles"
	PermissionUpdateWorkspaceFolder   = "update_workspace_folder"
	PermissionDeleteWorkspaceFolder   = "delete_workspace_folder"
	PermissionDeleteWorkspaceMember   = "delete_workspace_member"
	PermissionCreateWorkspaceView     = "create_workspace_view"
	PermissionUpdateWorkspaceView     = "update_workspace_view"
	PermissionCreateSingleSelectField = "create_single_select_field"
	PermissionUpdateWorkspaceTable    = "update_workspace_table"
	PermissionUpdateWorkspace         = "update_workspace"
	PermissionDeleteTableView         = "delete_table_view"
	PermissionDeleteWorkspaceTable    = "delete_workspace_table"
	PermissionAddMemberToWorkspace    = "add_member_to_workspace"
	PermissionCreateTable             = "create_table"
	PermissionShowAssignedTasksOnly   = "show_assigned_tasks_only"
	PermissionCreateFields            = "create_fields"
	PermissionEditFields              = "edit_fields"

	// View-type permissions use the format "views_{type}".
	PermissionViewGrid     = "views_grid"
	PermissionViewKanban   = "views_kanban"
	PermissionViewCalendar = "views_calendar"
	PermissionViewGantt    = "views_gantt"
	PermissionViewForm     = "views_form"
)
