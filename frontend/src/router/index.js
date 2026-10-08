// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { createRouter, createWebHistory } from "vue-router";
import { usePermissionsStore } from "@/store/permissions";

const sectionGuards = {
    files: "view_files",
    chat: "view_channels",
    collimato: "view_collimato",
    projects: "view_projects",
};

import { useUserStore } from "@/store/user";
import { hasSessionCookie } from "@/utils/utils";
import MainLayout from "@/views/MainLayout.vue";
import FilesMainView from "@/views/Files/FilesMainView.vue";
import FilesView from "@/views/Files/FilesView.vue";
import RecentView from "@/views/Files/RecentView.vue";
import FavoritesView from "@/views/Files/FavoritesView.vue";
import SharedView from "@/views/Files/SharedView.vue";
import DeletedView from "@/views/Files/DeletedView.vue";
import LoginView from "@/views/Login/LoginView.vue";
import SettingsView from "@/views/Settings/SettingsView.vue";
import ProfileView from "@/views/Settings/ProfileView.vue";
import PasswordView from "@/views/Settings/PasswordView.vue";
import NotificationsView from "@/views/Settings/NotificationsView.vue";
import UserSecurityView from "@/views/Settings/UserSecurityView.vue";
import ChannelsView from "@/views/Settings/ChannelsView.vue";
import ChannelFormView from "@/views/Settings/ChannelFormView.vue";
import ManagedWorkspacesView from "@/views/Settings/ManagedWorkspacesView.vue";
import ManagedWorkspaceView from "@/views/Settings/ManagedWorkspaceView.vue";
import ManagedProjectsView from "@/views/Settings/ManagedProjectsView.vue";
import ManagedProjectView from "@/views/Settings/ManagedProjectView.vue";
import UsersView from "@/views/Settings/UsersView.vue";
import UserFormView from "@/views/Settings/UserFormView.vue";
import GroupsView from "@/views/Settings/GroupsView.vue";
import GroupFormView from "@/views/Settings/GroupFormView.vue";
import SMTPView from "@/views/Settings/SMTPView.vue";
import SecurityView from "@/views/Settings/SecurityView.vue";
import MetadataView from "@/views/Settings/MetadataView.vue";
import LicenseView from "@/views/Settings/LicenseView.vue";
import LanguageView from "@/views/Settings/LanguageView.vue";
import AuthorizationView from "@/views/Settings/AuthorizationView.vue";
import OfficeView from "@/views/Settings/OfficeView.vue";
import StorageView from "@/views/Settings/StorageView.vue";
import RolesView from "@/views/Settings/RolesView.vue";
import RoleEditView from "@/views/Settings/RoleEditView.vue";
import CollimatoMainView from "@/views/Collimato/CollimatoMainView.vue";
import CollimatoWorkspacesView from "@/views/Collimato/WorkspacesView.vue";
import CollimatoDashboardsListView from "@/views/Collimato/DashboardsListView.vue";
import CollimatoDashboardView from "@/views/Collimato/DashboardView.vue";
import CollimatoConnectionView from "@/views/Collimato/ConnectionView.vue";
import CollimatoNewConnectionView from "@/views/Collimato/NewConnectionView.vue";
import CollimatoChartsView from "@/views/Collimato/ChartsView.vue";
import CollimatoNewChartView from "@/views/Collimato/NewChartView.vue";
import CollimatoDataView from "@/views/Collimato/DataView.vue";
import CollimatoWorkspaceView from "@/views/Collimato/WorkspaceView.vue";
import CollimatoWorkspaceSetupView from "@/views/Collimato/WorkspaceSetup/WorkspaceSetupView.vue";
import CollimatoUserView from "@/views/Collimato/UserView.vue";
import CollimatoRoleView from "@/views/Collimato/RoleView.vue";
import MessageView from "@/views/Chat/MessageView.vue";
import MeetingView from "@/views/Chat/MeetingView.vue";
import GuestVideoView from "@/views/Guest/GuestVideoView.vue";
import ChatRoom from "@/views/Chat/ChatRoom.vue";
import PasswordResetView from "@/views/Login/PasswordResetView.vue";
import PasswordForgotView from "@/views/Login/PasswordForgotView.vue";
import ProjectsView from "@/views/Projects/ProjectsView.vue";
import KanbanView from "@/components/Projects/Kanban/KanbanView.vue";
import GridView from "@/components/Projects/Grid/GridView.vue";
import ProjectsHomeView from "@/views/Projects/ProjectsHomeView.vue";
import WorkspaceSettingsView from "@/views/Projects/WorkspaceSettingsView.vue";
import AssignedToMe from "../components/Projects/Reports/AssignedToMe.vue";
import ProjectFolder from "@/views/Projects/ProjectFolderView.vue";
import DetailedTaskReport from "@/components/Projects/Reports/DetailedTaskReport.vue";
import CreateRole from "@/components/Projects/Members/CreateRole.vue";
import GanttView from "@/components/Projects/GanttView.vue";
import CalendarView from "@/components/Projects/CalendarView.vue";

const routes = [
    {
        path: "/",
        name: "main",
        component: MainLayout,
        meta: {
            requireAuth: true,
        },
        redirect: () => {
            const permissionsStore = usePermissionsStore();

            if (permissionsStore.loaded) {
                const fallback = Object.entries(sectionGuards).find(([, perm]) =>
                    permissionsStore.permissions.includes(perm),
                );

                if (fallback) return { name: fallback[0] };
            }

            return { name: "files" };
        },
        children: [
            {
                path: "files",
                name: "files",
                meta: {
                    requireAuth: true,
                },
                component: FilesMainView,
                children: [
                    {
                        path: "recents",
                        name: "recents",
                        meta: {
                            requireAuth: true,
                        },
                        component: RecentView,
                    },
                    {
                        path: "favorites",
                        name: "favorites",
                        meta: {
                            requireAuth: true,
                        },
                        component: FavoritesView,
                    },
                    {
                        path: "shared",
                        name: "shared",
                        meta: {
                            requireAuth: true,
                        },
                        component: SharedView,
                    },
                    {
                        path: "deleted",
                        name: "deleted",
                        meta: {
                            requireAuth: true,
                        },
                        component: DeletedView,
                    },
                    {
                        path: ":id", //file id.
                        name: "file",
                        meta: {
                            requireAuth: true,
                        },
                        component: FilesView,
                    },
                ],
            },
            {
                path: "collimato",
                name: "collimato",
                meta: {
                    requireAuth: true,
                },
                component: CollimatoMainView,
                redirect: {
                    name: "workspaces",
                },
                children: [
                    {
                        path: "workspaces",
                        name: "workspaces",
                        meta: {
                            requireAuth: true,
                        },
                        component: CollimatoWorkspacesView,
                    },
                    {
                        path: "workspaces/setup",
                        name: "workspace-setup",
                        meta: {
                            requireAuth: true,
                        },
                        component: CollimatoWorkspaceSetupView,
                        children: [],
                    },
                    {
                        path: "workspaces/:workspaceId",
                        name: "collimato-workspace",
                        meta: {
                            requireAuth: true,
                        },
                        component: CollimatoWorkspaceView,
                        redirect: {
                            name: "dashboards",
                        },
                        children: [
                            {
                                path: "dashboards",
                                name: "dashboards",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_dashboards",
                                },
                                component: CollimatoDashboardsListView,
                            },
                            {
                                path: "dashboards/:id",
                                name: "dashboard",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_dashboards",
                                },
                                component: CollimatoDashboardView,
                            },

                            {
                                path: "charts",
                                name: "charts",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_charts",
                                },
                                component: CollimatoChartsView,
                            },
                            {
                                path: "charts/new",
                                name: "new-chart",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: ["view_charts", "create_charts"],
                                },
                                component: CollimatoNewChartView,
                            },
                            {
                                path: "charts/edit/:id",
                                name: "edit-chart",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_charts",
                                },
                                component: CollimatoNewChartView,
                            },
                            {
                                path: "data",
                                name: "data",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_datamodels",
                                },
                                component: CollimatoDataView,
                            },
                            {
                                path: "connections",
                                name: "connections",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_connections",
                                },
                                component: CollimatoConnectionView,
                            },
                            {
                                path: "connections/new",
                                name: "new-connection",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "create_connections",
                                },
                                component: CollimatoNewConnectionView,
                            },
                            {
                                path: "connections/edit/:id",
                                name: "edit-connection",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "edit_connections",
                                },
                                component: CollimatoNewConnectionView,
                            },
                            {
                                path: "users",
                                name: "collimato-users",
                                meta: {
                                    requireAuth: true,
                                },
                                component: CollimatoUserView,
                            },
                            {
                                path: "roles/new",
                                name: "collimato-role",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "create_roles",
                                },
                                component: CollimatoRoleView,
                            },
                            {
                                path: "roles/:id",
                                name: "collimato-role-edit",
                                meta: {
                                    requireAuth: true,
                                    collimatoPermission: "view_roles",
                                },
                                component: CollimatoRoleView,
                            },
                        ],
                    },
                ],
            },

            {
                path: "chat/:chatId?",
                name: "chat",
                redirect: {
                    name: "messages",
                },
                meta: {
                    requireAuth: true,
                },
                children: [
                    {
                        path: "messages",
                        name: "messages",
                        meta: { requireAuth: true },
                        component: MessageView,
                    },
                ],
                component: ChatRoom,
            },

            {
                path: "meeting/:id",
                name: "meeting",
                meta: { requireAuth: true },
                component: MeetingView,
            },

            {
                path: "settings",
                name: "settings",
                meta: {
                    requireAuth: true,
                },
                component: SettingsView,
                children: [
                    {
                        path: "profile",
                        name: "profile",
                        meta: {
                            requireAuth: true,
                        },
                        component: ProfileView,
                    },
                    {
                        path: "password",
                        name: "password",
                        meta: {
                            requireAuth: true,
                        },
                        component: PasswordView,
                    },
                    {
                        path: "notifications",
                        name: "notifications",
                        meta: {
                            requireAuth: true,
                        },
                        component: NotificationsView,
                    },
                    {
                        path: "user-security",
                        name: "user-security",
                        meta: {
                            requireAuth: true,
                        },
                        component: UserSecurityView,
                    },
                    {
                        path: "users",
                        name: "users",
                        meta: { requireAuth: true },
                        component: UsersView,
                    },
                    {
                        path: "users/new",
                        name: "new-user",
                        meta: { requireAuth: true },
                        component: UserFormView,
                    },
                    {
                        path: "users/edit/:id",
                        name: "edit-user",
                        meta: { requireAuth: true },
                        component: UserFormView,
                    },
                    {
                        path: "groups",
                        name: "groups",
                        meta: { requireAuth: true, requireAdmin: true },
                        component: GroupsView,
                    },
                    {
                        path: "groups/new",
                        name: "new-group",
                        meta: { requireAuth: true, requireAdmin: true },
                        component: GroupFormView,
                    },
                    {
                        path: "groups/edit/:id",
                        name: "edit-group",
                        meta: { requireAuth: true, requireAdmin: true },
                        component: GroupFormView,
                    },
                    {
                        path: "smtp",
                        name: "smtp",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: SMTPView,
                    },
                    {
                        path: "security",
                        name: "security",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: SecurityView,
                    },

                    {
                        path: "authorization",
                        name: "authorization",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: AuthorizationView,
                    },
                    {
                        path: "metadata",
                        name: "metadata",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: MetadataView,
                    },
                    {
                        path: "office",
                        name: "office-settings",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: OfficeView,
                    },
                    {
                        path: "storage",
                        name: "storage-settings",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: StorageView,
                    },
                    {
                        path: "roles",
                        name: "roles",
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_roles",
                        },
                        component: RolesView,
                    },
                    {
                        path: "roles/new",
                        name: "new-role",
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_roles",
                        },
                        component: RoleEditView,
                    },
                    {
                        path: "roles/:id",
                        name: "edit-role",
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_roles",
                        },
                        component: RoleEditView,
                    },
                    {
                        path: "license",
                        name: "license",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: LicenseView,
                    },
                    {
                        path: "language",
                        name: "language",
                        meta: {
                            requireAuth: true,
                            requireAdmin: true,
                        },
                        component: LanguageView,
                    },
                    {
                        path: "channels",
                        name: "channels",
                        redirect: { name: "channel-list" },
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_channels",
                        },
                        children: [
                            {
                                path: "",
                                name: "channel-list",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_channels",
                                },
                                component: ChannelsView,
                            },
                            {
                                path: ":id",
                                name: "channel",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_channels",
                                },
                                component: ChannelFormView,
                            },
                        ],
                    },
                    {
                        path: "collimato",
                        name: "collimato-settings",
                        redirect: { name: "collimato-settings-list" },
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_collimato",
                        },
                        children: [
                            {
                                path: "",
                                name: "collimato-settings-list",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_collimato",
                                },
                                component: ManagedWorkspacesView,
                            },
                            {
                                path: ":id",
                                name: "collimato-settings-workspace",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_collimato",
                                },
                                component: ManagedWorkspaceView,
                            },
                        ],
                    },
                    {
                        path: "projects",
                        name: "projects-settings",
                        redirect: { name: "projects-settings-list" },
                        meta: {
                            requireAuth: true,
                            requirePermission: "manage_projects",
                        },
                        children: [
                            {
                                path: "",
                                name: "projects-settings-list",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_projects",
                                },
                                component: ManagedProjectsView,
                            },
                            {
                                path: ":id",
                                name: "projects-settings-workspace",
                                meta: {
                                    requireAuth: true,
                                    requirePermission: "manage_projects",
                                },
                                component: ManagedProjectView,
                            },
                        ],
                    },
                ],
            },

            {
                path: "projects",
                name: "projects",
                redirect: { name: "projects-home" },
                meta: {
                    requireAuth: true,
                },
                component: ProjectsView,
                children: [
                    {
                        path: "",
                        name: "projects-home",
                        meta: {
                            requireAuth: true,
                        },
                        component: ProjectsHomeView,
                    },
                    {
                        path: ":id/settings/:tab?",
                        name: "workspace-settings",
                        meta: {
                            requireAuth: true,
                        },
                        component: WorkspaceSettingsView,
                    },
                    {
                        path: ":id/project-folder/:fid?",
                        name: "project-folder",
                        meta: {
                            requireAuth: true,
                        },
                        component: ProjectFolder,
                    },
                    {
                        path: ":id/kanban/:tid/view/:fid",
                        name: "kanban-view",
                        meta: {
                            requireAuth: true,
                        },
                        component: KanbanView,
                    },
                    {
                        path: ":id/grid/:tid/view/:fid?",
                        name: "grid-view",
                        meta: {
                            requireAuth: true,
                        },
                        component: GridView,
                    },
                    {
                        path: ":id/roles/create",
                        name: "create-role",
                        meta: {
                            requireAuth: true,
                        },
                        component: CreateRole,
                    },
                    {
                        path: ":id/role/update/:rid",
                        name: "update-role",
                        meta: {
                            requireAuth: true,
                        },
                        component: CreateRole,
                    },
                    {
                        path: "workspace-members/:workspaceId?",
                        name: "workspace-members",
                        redirect: (to) =>
                            to.params.workspaceId
                                ? {
                                      name: "workspace-settings",
                                      params: { id: to.params.workspaceId, tab: "members" },
                                  }
                                : { name: "projects-home" },
                    },
                    {
                        path: ":id/gantt/:tid/view/:fid",
                        name: "gantt-view",
                        meta: {
                            requireAuth: true,
                        },
                        component: GanttView,
                    },
                    {
                        path: ":id/calendar/:tid/view/:fid",
                        name: "calendar-view",
                        meta: {
                            requireAuth: true,
                        },
                        component: CalendarView,
                    },
                    {
                        path: "assigned-to-me",
                        name: "assigned-to-me",
                        meta: {
                            requireAuth: true,
                        },
                        component: AssignedToMe,
                    },
                    {
                        path: "detailed-task-report",
                        name: "detailed-task-report",
                        meta: {
                            requireAuth: true,
                        },
                        component: DetailedTaskReport,
                    },
                ],
            },
        ],
    },
    {
        path: "/guest/video/:token",
        name: "guest-video",
        component: GuestVideoView,
        meta: {
            requireAuth: false,
        },
    },
    {
        path: "/share/:token",
        name: "public-share",
        component: () => import("@/views/Public/PublicShareView.vue"),
        meta: {
            requireAuth: false,
        },
    },
    {
        path: "/login",
        name: "login",
        component: LoginView,
    },
    {
        path: "/password/reset",
        name: "password-forgot",
        component: PasswordForgotView,
        meta: {
            requireAuth: false,
        },
    },
    {
        component: PasswordResetView,
        name: "password-reset",
        path: "/password/reset/:token",
        meta: {
            requireAuth: false,
        },
    },
    {
        path: "/:catchAll(.*)",
        redirect: { name: "main" },
    },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach((to, from, next) => {
    const hasSession = hasSessionCookie();
    const userStore = useUserStore();

    if (to.meta.requireAuth && !hasSession) {
        return next({ name: "login", query: { redirect: to.fullPath } });
    }

    if (to.name === "login" && hasSession) {
        return next({ path: "/" });
    }

    if (to.meta.requireAdmin && userStore.user?.role !== "system_admin") {
        return next({ name: "profile" });
    }

    // Only enforce permission-based redirects once permissions have been loaded.
    // Before that, the section's main view shows the access-denied wall as a fallback.
    const permissionsStore = usePermissionsStore();

    if (permissionsStore.loaded) {
        if (
            to.meta.requirePermission &&
            !permissionsStore.permissions.includes(to.meta.requirePermission)
        ) {
            return next({ name: "profile" });
        }

        const guardedMatch = to.matched.find((r) => sectionGuards[r.name]);

        if (
            guardedMatch &&
            !permissionsStore.permissions.includes(sectionGuards[guardedMatch.name])
        ) {
            const fallbackSection = Object.entries(sectionGuards).find(([, perm]) =>
                permissionsStore.permissions.includes(perm),
            );

            return next({
                name: fallbackSection ? fallbackSection[0] : "profile",
            });
        }
    }

    next();
});

export default router;
