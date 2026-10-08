// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	SystemAdminRoleId = "system_admin"
	SystemUserRoleId  = "system_user"

	CollimatoWorkspaceAdminRoleId = "workspace_admin"
	CollimatoWorkspaceUserRoleId  = "workspace_user"
)

type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	BuiltIn     bool     `json:"built_in"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}

type RolePatch struct {
	Name        *string   `json:"name"`
	DisplayName *string   `json:"display_name"`
	Description *string   `json:"description"`
	Permissions *[]string `json:"permissions"`
}

func (r *Role) Patch(patch RolePatch) {
	if patch.Name != nil {
		r.Name = *patch.Name
	}

	if patch.DisplayName != nil {
		r.DisplayName = *patch.DisplayName
	}

	if patch.Description != nil {
		r.Description = *patch.Description
	}

	if patch.Permissions != nil {
		r.Permissions = *patch.Permissions
	}
}

type CollimatoRole struct {
	ID                string                      `json:"id"`
	WorkspaceID       string                      `json:"workspace_id"`
	Name              string                      `json:"name"`
	DisplayName       string                      `json:"display_name"`
	Description       string                      `json:"description"`
	Permissions       []string                    `json:"permissions"`
	TablePermissions  []string                    `json:"table_permissions"`
	ColumnPermissions []CollimatoColumnPermission `json:"column_permissions"`
	RowPermissions    []CollimatoRowPermission    `json:"row_permissions"`
	AutoUpdate        bool                        `json:"auto_update"`
	CreatedAt         int64                       `json:"created_at"`
	UpdatedAt         int64                       `json:"updated_at"`
}

type CollimatoRolePatch struct {
	Name              *string                      `json:"name"`
	DisplayName       *string                      `json:"display_name"`
	Description       *string                      `json:"description"`
	Permissions       *[]string                    `json:"permissions"`
	TablePermissions  *[]string                    `json:"table_permissions"`
	ColumnPermissions *[]CollimatoColumnPermission `json:"column_permissions"`
	RowPermissions    *[]CollimatoRowPermission    `json:"row_permissions"`
}

func (c *CollimatoRole) Patch(patch *CollimatoRolePatch) {
	if patch.Name != nil {
		c.Name = *patch.Name
	}

	if patch.DisplayName != nil {
		c.DisplayName = *patch.DisplayName
	}

	if patch.Description != nil {
		c.Description = *patch.Description
	}

	if patch.Permissions != nil {
		c.Permissions = *patch.Permissions
	}

	if patch.TablePermissions != nil {
		c.TablePermissions = *patch.TablePermissions
	}

	if patch.ColumnPermissions != nil {
		c.ColumnPermissions = *patch.ColumnPermissions
	}

	if patch.RowPermissions != nil {
		c.RowPermissions = *patch.RowPermissions
	}
}

type CollimatoColumnPermission struct {
	Table string `json:"table"`
	Field string `json:"field"`
}

type CollimatoRowPermission struct {
	Table    string   `json:"table"`
	Member   string   `json:"member"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

func MakeDefaultRoles() map[string]*Role {
	roles := make(map[string]*Role)

	roles[SystemUserRoleId] = &Role{
		Name:        "system_user",
		DisplayName: "User",
		Description: "Default user role",
		BuiltIn:     true,
		Permissions: []string{
			FilePermissions.PermissionViewFiles.Id,
			FilePermissions.PermissionCreateFiles.Id,
			FilePermissions.PermissionDownloadFiles.Id,
			FilePermissions.PermissionUploadFiles.Id,
			FilePermissions.PermissionShareFiles.Id,
			ChatPermissions.PermissionViewChannels.Id,
			ChatPermissions.PermissionCreateChannel.Id,
			ChatPermissions.PermissionSendMessage.Id,
			ChatPermissions.PermissionUploadChatFiles.Id,
			CollimatoSectionPermissions.PermissionViewCollimato.Id,
			ProjectSectionPermissions.PermissionViewProjects.Id,
			ProjectSectionPermissions.PermissionCreateProject.Id,
		},
	}

	roles[SystemAdminRoleId] = &Role{
		Name:        "system_admin",
		DisplayName: "Administrator",
		Description: "Full system access — bypasses all permission checks",
		BuiltIn:     true,
		Permissions: []string{},
	}

	return roles
}

func MakeDefaultCollimatoWorkspaceRoles() map[string]*CollimatoRole {
	roles := make(map[string]*CollimatoRole)

	roles[CollimatoWorkspaceUserRoleId] = &CollimatoRole{
		Name:        CollimatoWorkspaceUserRoleId,
		DisplayName: "authentication.roles.collimato_user.name",
		Description: "authentication.roles.collimato_user.description",
		Permissions: []string{
			CollimatoPermissions.PermissionCreateCharts.Id,
			CollimatoPermissions.PermissionViewCharts.Id,
			CollimatoPermissions.PermissionEditCharts.Id,
			CollimatoPermissions.PermissionDeleteCharts.Id,
			CollimatoPermissions.PermissionCreateDashboards.Id,
			CollimatoPermissions.PermissionViewDashboards.Id,
			CollimatoPermissions.PermissionEditDashboards.Id,
			CollimatoPermissions.PermissionDeleteDashboards.Id,
		},
		TablePermissions:  []string{},
		ColumnPermissions: []CollimatoColumnPermission{},
		RowPermissions:    []CollimatoRowPermission{},
		AutoUpdate:        false,
	}

	roles[CollimatoWorkspaceAdminRoleId] = &CollimatoRole{
		Name:        CollimatoWorkspaceAdminRoleId,
		DisplayName: "authentication.roles.collimato_admin.name",
		Description: "authentication.roles.collimato_admin.description",
		Permissions: []string{
			CollimatoPermissions.PermissionCreateConnections.Id,
			CollimatoPermissions.PermissionViewConnections.Id,
			CollimatoPermissions.PermissionEditConnections.Id,
			CollimatoPermissions.PermissionDeleteConnections.Id,
			CollimatoPermissions.PermissionCreateCharts.Id,
			CollimatoPermissions.PermissionViewCharts.Id,
			CollimatoPermissions.PermissionEditCharts.Id,
			CollimatoPermissions.PermissionDeleteCharts.Id,
			CollimatoPermissions.PermissionAddUsers.Id,
			CollimatoPermissions.PermissionDeleteUsers.Id,
			CollimatoPermissions.PermissionAssignRoles.Id,
			CollimatoPermissions.PermissionCreateRoles.Id,
			CollimatoPermissions.PermissionViewRoles.Id,
			CollimatoPermissions.PermissionEditRoles.Id,
			CollimatoPermissions.PermissionDeleteRoles.Id,
			CollimatoPermissions.PermissionCreateDataModels.Id,
			CollimatoPermissions.PermissionViewDataModels.Id,
			CollimatoPermissions.PermissionEditDataModels.Id,
			CollimatoPermissions.PermissionDeleteDataModels.Id,
			CollimatoPermissions.PermissionCreateDashboards.Id,
			CollimatoPermissions.PermissionViewDashboards.Id,
			CollimatoPermissions.PermissionEditDashboards.Id,
			CollimatoPermissions.PermissionDeleteDashboards.Id,
			CollimatoPermissions.PermissionViewDashboardFilters.Id,
			CollimatoPermissions.PermissionCreateDashboardFilters.Id,
			CollimatoPermissions.PermissionEditDashboardFilters.Id,
			CollimatoPermissions.PermissionDeleteDashboardFilters.Id,
		},
		TablePermissions:  []string{},
		ColumnPermissions: []CollimatoColumnPermission{},
		RowPermissions:    []CollimatoRowPermission{},
		AutoUpdate:        true,
	}

	return roles
}
