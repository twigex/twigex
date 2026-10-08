// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useWorkspaceRolesStore = defineStore("workspaceRoles", {
    state: () => {
        return {
            user: null,
            layers: [],
        };
    },

    getters: {
        userPermissions: (state) => {
            return state.user?.permissions || [];
        },

        hasPermission: (state) => (permissionName) => {
            return state.userPermissions.includes(permissionName);
        },

        roleNames: (state) => (state.user?.role ?? "").split(" ").filter(Boolean),
        isWorkspaceAdmin: (state) => state.roleNames.includes("admin"),

        hasPermissionToCreateTask: (state) => state.hasPermission("create_task"),
        hasPermissionToCreateRoles: (state) => state.hasPermission("create_roles"),
        hasPermissionToUpdateRoles: (state) => state.hasPermission("update_roles"),
        hasPermissionToDeleteRoles: (state) => state.hasPermission("delete_roles"),
        hasPermissionToUpdateWorkspaceFolder: (state) =>
            state.hasPermission("update_workspace_folder"),
        hasPermissionToDeleteWorkspaceFolder: (state) =>
            state.hasPermission("delete_workspace_folder"),
        hasPermissionToDeleteWorkspaceMember: (state) =>
            state.hasPermission("delete_workspace_member"),
        hasPermissionToCreateWorkspaceView: (state) => state.hasPermission("create_workspace_view"),
        hasPermissionToUpdateWorkspaceView: (state) => state.hasPermission("update_workspace_view"),
        hasPermissionToUpdateWorkspaceTable: (state) =>
            state.hasPermission("update_workspace_table"),
        hasPermissionToDeleteTableView: (state) => state.hasPermission("delete_table_view"),
        hasPermissionToDeleteWorkspaceTable: (state) =>
            state.hasPermission("delete_workspace_table"),
        hasPermissionToAddMemberToWorkspace: (state) =>
            state.hasPermission("add_member_to_workspace"),

        hasPermissionToShowAssignedTasksOnly: (state) => {
            return state.userPermissions.includes("show_assigned_tasks_only");
        },

        // backward-compat: no views_* permissions at all → all view types allowed
        hasViewPermission: (state) => (viewType) => {
            const permissions = state.userPermissions;
            const hasAny = permissions.some((p) => p.startsWith("views_"));

            if (!hasAny) return true;

            return permissions.includes("views_" + viewType);
        },

        perTableMode: (state) => state.user?.per_table_mode || false,

        // canRowAction is whether the user may do action to the rows of
        // tableId: by the table's own permissions in per-table mode, by the
        // workspace role otherwise, and always while no role is loaded.
        canRowAction: (state) => (tableId, action) => {
            const permissions = state.userPermissions;

            if (permissions.length === 0) return true;

            return state.perTableMode
                ? state.hasTablePermission(tableId, action)
                : permissions.includes(action);
        },

        // backward-compat: per_table_mode false → all table actions allowed
        hasTablePermission: (state) => (tableId, action) => {
            if (!state.user?.per_table_mode) return true;
            const tablePerms = state.user?.table_permissions || [];
            const tableEntries = tablePerms.filter((p) => p.table_id === tableId);

            if (tableEntries.length === 0) {
                // Custom visibility mode: if any table has explicit view/hidden entries,
                // tables with no entries are hidden by default (mirrors backend visibilityCustomMode).
                const visibilityCustomMode = tablePerms.some(
                    (p) => p.action === "view" || p.action === "hidden",
                );

                return !visibilityCustomMode;
            }

            // Table with "hidden" sentinel is explicitly hidden, block all actions.
            if (tableEntries.some((p) => p.action === "hidden")) return false;

            return tableEntries.some((p) => p.action === action);
        },
    },

    actions: {
        setUser(user) {
            this.user = user;
        },
    },
});
