// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import { collimatoPermissions as p } from "@/constants/permissions";

function granted(state, permission) {
    return (state.user?.permissions ?? []).includes(permission);
}

export const useCollimatoStore = defineStore("collimato", {
    state: () => {
        return { user: null };
    },

    getters: {
        can: (state) => (permission) => granted(state, permission),
        canQueryData: (state) =>
            granted(state, p.PERMISSION_VIEW_CHARTS) ||
            granted(state, p.PERMISSION_VIEW_DASHBOARDS),
        roleNames: (state) => (state.user?.role ?? "").split(",").filter(Boolean),
        isWorkspaceAdmin() {
            return this.roleNames.includes("workspace_admin");
        },

        hasPermissionToCreateConnections: (state) =>
            granted(state, p.PERMISSION_CREATE_CONNECTIONS),
        hasPermissionToViewConnections: (state) => granted(state, p.PERMISSION_VIEW_CONNECTIONS),
        hasPermissionToEditConnections: (state) => granted(state, p.PERMISSION_EDIT_CONNECTIONS),
        hasPermissionToDeleteConnections: (state) =>
            granted(state, p.PERMISSION_DELETE_CONNECTIONS),

        hasPermissionToCreateCharts: (state) => granted(state, p.PERMISSION_CREATE_CHARTS),
        hasPermissionToViewCharts: (state) => granted(state, p.PERMISSION_VIEW_CHARTS),
        hasPermissionToEditCharts: (state) => granted(state, p.PERMISSION_EDIT_CHARTS),
        hasPermissionToDeleteCharts: (state) => granted(state, p.PERMISSION_DELETE_CHARTS),

        hasPermissionToAddUsers: (state) => granted(state, p.PERMISSION_ADD_USERS),
        hasPermissionToDeleteUsers: (state) => granted(state, p.PERMISSION_DELETE_USERS),
        hasPermissionToAssignRoles: (state) => granted(state, p.PERMISSION_ASSIGN_ROLES),

        hasPermissionToCreateRoles: (state) => granted(state, p.PERMISSION_CREATE_ROLES),
        hasPermissionToViewRoles: (state) => granted(state, p.PERMISSION_VIEW_ROLES),
        hasPermissionToEditRoles: (state) => granted(state, p.PERMISSION_EDIT_ROLES),
        hasPermissionToDeleteRoles: (state) => granted(state, p.PERMISSION_DELETE_ROLES),

        hasPermissionToCreateDataModels: (state) => granted(state, p.PERMISSION_CREATE_DATAMODELS),
        hasPermissionToViewDataModels: (state) => granted(state, p.PERMISSION_VIEW_DATAMODELS),
        hasPermissionToEditDataModels: (state) => granted(state, p.PERMISSION_EDIT_DATAMODELS),
        hasPermissionToDeleteDataModels: (state) => granted(state, p.PERMISSION_DELETE_DATAMODELS),

        hasPermissionToCreateDashboards: (state) => granted(state, p.PERMISSION_CREATE_DASHBOARDS),
        hasPermissionToViewDashboards: (state) => granted(state, p.PERMISSION_VIEW_DASHBOARDS),
        hasPermissionToEditDashboards: (state) => granted(state, p.PERMISSION_EDIT_DASHBOARDS),
        hasPermissionToDeleteDashboards: (state) => granted(state, p.PERMISSION_DELETE_DASHBOARDS),

        hasPermissionToCreateDashboardFilters: (state) =>
            granted(state, p.PERMISSION_CREATE_DASHBOARD_FILTERS),
        hasPermissionToViewDashboardFilters: (state) =>
            granted(state, p.PERMISSION_VIEW_DASHBOARD_FILTERS),
        hasPermissionToEditDashboardFilters: (state) =>
            granted(state, p.PERMISSION_EDIT_DASHBOARD_FILTERS),
        hasPermissionToDeleteDashboardFilters: (state) =>
            granted(state, p.PERMISSION_DELETE_DASHBOARD_FILTERS),
    },

    actions: {
        setUser(user) {
            this.user = user;
        },
    },
});
