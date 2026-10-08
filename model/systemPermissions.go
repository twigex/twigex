// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package model

const (
	PermissionScopeSystem    = "system_scope"
	PermissionScopeCollimato = "collimato_scope"
)

type SystemFilePermissions struct {
	PermissionViewFiles     *SystemPermission
	PermissionCreateFiles   *SystemPermission
	PermissionDownloadFiles *SystemPermission
	PermissionUploadFiles   *SystemPermission
	PermissionShareFiles    *SystemPermission
	PermissionChangeLicense *SystemPermission
}

type SystemAdminPermissions struct {
	PermissionCreateUsers     *SystemPermission
	PermissionDeleteUsers     *SystemPermission
	PermissionEditUsers       *SystemPermission
	PermissionManageRoles     *SystemPermission
	PermissionManageChannels  *SystemPermission
	PermissionArchiveChannels *SystemPermission
	PermissionDeleteChannels  *SystemPermission
}

var AdminPermissions SystemAdminPermissions

var FilePermissions SystemFilePermissions

type SystemChatPermissions struct {
	PermissionViewChannels    *SystemPermission
	PermissionCreateChannel   *SystemPermission
	PermissionSendMessage     *SystemPermission
	PermissionUploadChatFiles *SystemPermission
}

var ChatPermissions SystemChatPermissions

type SystemCollimatoPermissions struct {
	PermissionViewCollimato            *SystemPermission
	PermissionCreateCollimatoWorkspace *SystemPermission
	PermissionManageCollimato          *SystemPermission
}

var CollimatoSectionPermissions SystemCollimatoPermissions

type SystemProjectPermissions struct {
	PermissionViewProjects   *SystemPermission
	PermissionCreateProject  *SystemPermission
	PermissionManageProjects *SystemPermission
}

var ProjectSectionPermissions SystemProjectPermissions

type PermissionSection struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Permissions []*SystemPermission `json:"permissions"`
}

func GetSystemPermissionSections() []PermissionSection {
	return []PermissionSection{
		{
			ID:   "files",
			Name: "Files",
			Permissions: []*SystemPermission{
				FilePermissions.PermissionViewFiles,
				FilePermissions.PermissionCreateFiles,
				FilePermissions.PermissionUploadFiles,
				FilePermissions.PermissionDownloadFiles,
				FilePermissions.PermissionShareFiles,
			},
		},
		{
			ID:   "chat",
			Name: "Chat",
			Permissions: []*SystemPermission{
				ChatPermissions.PermissionViewChannels,
				ChatPermissions.PermissionCreateChannel,
				ChatPermissions.PermissionSendMessage,
				ChatPermissions.PermissionUploadChatFiles,
			},
		},
		{
			ID:   "administration",
			Name: "Administration",
			Permissions: []*SystemPermission{
				AdminPermissions.PermissionCreateUsers,
				AdminPermissions.PermissionDeleteUsers,
				AdminPermissions.PermissionEditUsers,
				AdminPermissions.PermissionManageRoles,
				AdminPermissions.PermissionManageChannels,
				AdminPermissions.PermissionArchiveChannels,
				AdminPermissions.PermissionDeleteChannels,
			},
		},
		{
			ID:   "collimato",
			Name: "Collimato",
			Permissions: []*SystemPermission{
				CollimatoSectionPermissions.PermissionViewCollimato,
				CollimatoSectionPermissions.PermissionCreateCollimatoWorkspace,
				CollimatoSectionPermissions.PermissionManageCollimato,
			},
		},
		{
			ID:   "projects",
			Name: "Projects",
			Permissions: []*SystemPermission{
				ProjectSectionPermissions.PermissionViewProjects,
				ProjectSectionPermissions.PermissionCreateProject,
				ProjectSectionPermissions.PermissionManageProjects,
			},
		},
	}
}

type AppCollimatoPermissions struct {
	PermissionCreateConnections      *CollimatoPermission
	PermissionViewConnections        *CollimatoPermission
	PermissionEditConnections        *CollimatoPermission
	PermissionDeleteConnections      *CollimatoPermission
	PermissionCreateCharts           *CollimatoPermission
	PermissionViewCharts             *CollimatoPermission
	PermissionEditCharts             *CollimatoPermission
	PermissionDeleteCharts           *CollimatoPermission
	PermissionAddUsers               *CollimatoPermission
	PermissionDeleteUsers            *CollimatoPermission
	PermissionAssignRoles            *CollimatoPermission
	PermissionCreateRoles            *CollimatoPermission
	PermissionViewRoles              *CollimatoPermission
	PermissionEditRoles              *CollimatoPermission
	PermissionDeleteRoles            *CollimatoPermission
	PermissionCreateDataModels       *CollimatoPermission
	PermissionViewDataModels         *CollimatoPermission
	PermissionEditDataModels         *CollimatoPermission
	PermissionDeleteDataModels       *CollimatoPermission
	PermissionCreateDashboards       *CollimatoPermission
	PermissionViewDashboards         *CollimatoPermission
	PermissionEditDashboards         *CollimatoPermission
	PermissionDeleteDashboards       *CollimatoPermission
	PermissionCreateDashboardFilters *CollimatoPermission
	PermissionViewDashboardFilters   *CollimatoPermission
	PermissionEditDashboardFilters   *CollimatoPermission
	PermissionDeleteDashboardFilters *CollimatoPermission
}

var CollimatoPermissions AppCollimatoPermissions

type SystemPermission struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
}

func initializePermissions() {
	FilePermissions = SystemFilePermissions{
		PermissionViewFiles: &SystemPermission{
			"view_files",
			"View files",
			"Browse and view files",
			PermissionScopeSystem,
		},

		PermissionCreateFiles: &SystemPermission{
			"create_files",
			"Create files",
			"Create documents and folders",
			PermissionScopeSystem,
		},

		PermissionDownloadFiles: &SystemPermission{
			"download_files",
			"Download files",
			"Download files to local device",
			PermissionScopeSystem,
		},

		PermissionUploadFiles: &SystemPermission{
			"upload_files",
			"Upload files",
			"Upload files from local device",
			PermissionScopeSystem,
		},

		PermissionShareFiles: &SystemPermission{
			"share_files",
			"Share files",
			"Share files with other users",
			PermissionScopeSystem,
		},

		PermissionChangeLicense: &SystemPermission{
			"change_license",
			"Change license",
			"Manage system license",
			PermissionScopeSystem,
		},
	}

	ChatPermissions = SystemChatPermissions{
		PermissionViewChannels: &SystemPermission{
			"view_channels",
			"View channels",
			"Access the chat section and view channels",
			PermissionScopeSystem,
		},

		PermissionCreateChannel: &SystemPermission{
			"create_channel",
			"Create channels",
			"Create new chat channels",
			PermissionScopeSystem,
		},

		PermissionSendMessage: &SystemPermission{
			"send_message",
			"Send messages",
			"Post messages in channels",
			PermissionScopeSystem,
		},

		PermissionUploadChatFiles: &SystemPermission{
			"upload_chat_files",
			"Upload files in chat",
			"Upload file attachments in chat",
			PermissionScopeSystem,
		},
	}

	AdminPermissions = SystemAdminPermissions{
		PermissionCreateUsers: &SystemPermission{
			"create_users",
			"Create users",
			"Invite and create new user accounts",
			PermissionScopeSystem,
		},
		PermissionDeleteUsers: &SystemPermission{
			"delete_users",
			"Delete users",
			"Deactivate and delete user accounts",
			PermissionScopeSystem,
		},
		PermissionEditUsers: &SystemPermission{
			"edit_users",
			"Edit users",
			"Modify user details and reset passwords",
			PermissionScopeSystem,
		},
		PermissionManageChannels: &SystemPermission{
			"manage_channels",
			"Manage channels",
			"View and edit any channel",
			PermissionScopeSystem,
		},

		PermissionArchiveChannels: &SystemPermission{
			"archive_channels",
			"Archive channels",
			"Archive and restore any channel",
			PermissionScopeSystem,
		},

		PermissionDeleteChannels: &SystemPermission{
			"delete_channels",
			"Delete channels",
			"Permanently delete any channel",
			PermissionScopeSystem,
		},

		PermissionManageRoles: &SystemPermission{
			"manage_roles",
			"Manage roles",
			"Create, edit and delete system roles",
			PermissionScopeSystem,
		},
	}

	CollimatoSectionPermissions = SystemCollimatoPermissions{
		PermissionViewCollimato: &SystemPermission{
			"view_collimato",
			"View Collimato",
			"Access the Collimato analytics section",
			PermissionScopeSystem,
		},
		PermissionCreateCollimatoWorkspace: &SystemPermission{
			"create_collimato_workspace",
			"Create Collimato workspace",
			"Create new Collimato workspaces",
			PermissionScopeSystem,
		},
		PermissionManageCollimato: &SystemPermission{
			"manage_collimato",
			"Manage Collimato workspaces",
			"View and manage every Collimato workspace from Settings",
			PermissionScopeSystem,
		},
	}

	ProjectSectionPermissions = SystemProjectPermissions{
		PermissionViewProjects: &SystemPermission{
			"view_projects",
			"View projects",
			"Access the Projects section",
			PermissionScopeSystem,
		},
		PermissionCreateProject: &SystemPermission{
			"create_project",
			"Create project",
			"Create new project workspaces",
			PermissionScopeSystem,
		},
		PermissionManageProjects: &SystemPermission{
			"manage_projects",
			"Manage project workspaces",
			"View and manage every project workspace from Settings",
			PermissionScopeSystem,
		},
	}
}

type CollimatoPermission struct {
	Id    string `json:"id"`
	Scope string `json:"scope"`
}

func initializeCollimatoPermissions() {
	CollimatoPermissions = AppCollimatoPermissions{
		PermissionCreateConnections: &CollimatoPermission{
			"create_connections",
			PermissionScopeCollimato,
		},
		PermissionViewConnections: &CollimatoPermission{
			"view_connections",
			PermissionScopeCollimato,
		},
		PermissionEditConnections: &CollimatoPermission{
			"edit_connections",
			PermissionScopeCollimato,
		},
		PermissionDeleteConnections: &CollimatoPermission{
			"delete_connections",
			PermissionScopeCollimato,
		},
		PermissionCreateCharts: &CollimatoPermission{
			"create_charts",
			PermissionScopeCollimato,
		},
		PermissionViewCharts: &CollimatoPermission{
			"view_charts",
			PermissionScopeCollimato,
		},
		PermissionEditCharts: &CollimatoPermission{
			"edit_charts",
			PermissionScopeCollimato,
		},
		PermissionDeleteCharts: &CollimatoPermission{
			"delete_charts",
			PermissionScopeCollimato,
		},
		PermissionAddUsers: &CollimatoPermission{
			"add_users",
			PermissionScopeCollimato,
		},
		PermissionDeleteUsers: &CollimatoPermission{
			"delete_users",
			PermissionScopeCollimato,
		},
		PermissionAssignRoles: &CollimatoPermission{
			"assign_roles",
			PermissionScopeCollimato,
		},
		PermissionCreateRoles: &CollimatoPermission{
			"create_roles",
			PermissionScopeCollimato,
		},
		PermissionViewRoles: &CollimatoPermission{
			"view_roles",
			PermissionScopeCollimato,
		},
		PermissionEditRoles: &CollimatoPermission{
			"edit_roles",
			PermissionScopeCollimato,
		},
		PermissionDeleteRoles: &CollimatoPermission{
			"delete_roles",
			PermissionScopeCollimato,
		},
		PermissionCreateDataModels: &CollimatoPermission{
			"create_datamodels",
			PermissionScopeCollimato,
		},
		PermissionViewDataModels: &CollimatoPermission{
			"view_datamodels",
			PermissionScopeCollimato,
		},
		PermissionEditDataModels: &CollimatoPermission{
			"edit_datamodels",
			PermissionScopeCollimato,
		},
		PermissionDeleteDataModels: &CollimatoPermission{
			"delete_datamodels",
			PermissionScopeCollimato,
		},
		PermissionCreateDashboards: &CollimatoPermission{
			"create_dashboards",
			PermissionScopeCollimato,
		},
		PermissionViewDashboards: &CollimatoPermission{
			"view_dashboards",
			PermissionScopeCollimato,
		},
		PermissionEditDashboards: &CollimatoPermission{
			"edit_dashboards",
			PermissionScopeCollimato,
		},
		PermissionDeleteDashboards: &CollimatoPermission{
			"delete_dashboards",
			PermissionScopeCollimato,
		},
		PermissionCreateDashboardFilters: &CollimatoPermission{
			"create_dashboard_filters",
			PermissionScopeCollimato,
		},
		PermissionViewDashboardFilters: &CollimatoPermission{
			"view_dashboard_filters",
			PermissionScopeCollimato,
		},
		PermissionEditDashboardFilters: &CollimatoPermission{
			"edit_dashboard_filters",
			PermissionScopeCollimato,
		},
		PermissionDeleteDashboardFilters: &CollimatoPermission{
			"delete_dashboard_filters",
			PermissionScopeCollimato,
		},
	}
}

func init() {
	initializePermissions()
	initializeCollimatoPermissions()
}
