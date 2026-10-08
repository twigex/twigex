// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const collimatoService = {
    status() {
        return axios.get("/collimato/status");
    },

    getWorkspaceConnections(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/connections`);
    },

    getConnectionById(workspaceId, id) {
        return axios.get(`/collimato/workspaces/${workspaceId}/connections/${id}`);
    },

    createConnection(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/connections`, data);
    },

    deleteConnection(workspaceId, id) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/connections/${id}`);
    },

    updateConnection(workspaceId, id, data) {
        return axios.put(`/collimato/workspaces/${workspaceId}/connections/${id}`, data);
    },

    testConnection(workspaceId, id, data) {
        const path = id
            ? `/collimato/workspaces/${workspaceId}/connections/${id}/test`
            : `/collimato/workspaces/${workspaceId}/connections/test`;

        return axios.post(path, data);
    },

    createChart(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/charts`, data);
    },

    getCharts(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/charts`);
    },

    getChartById(workspaceId, id) {
        return axios.get(`/collimato/workspaces/${workspaceId}/charts/${id}`);
    },

    updateChart(workspaceId, id, data) {
        return axios.put(`/collimato/workspaces/${workspaceId}/charts/${id}`, data);
    },

    deleteChart(workspaceId, id) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/charts/${id}`);
    },

    updateDashboard(workspaceId, id, data) {
        return axios.put(`/collimato/workspaces/${workspaceId}/dashboards/${id}`, data);
    },

    getDashboards(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/dashboards`);
    },

    getDashboardById(workspaceId, id) {
        return axios.get(`/collimato/workspaces/${workspaceId}/dashboards/${id}`);
    },

    createDashboard(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/dashboards`, data);
    },

    createDashboardFilter(workspaceId, id, data) {
        return axios.post(
            `/collimato/workspaces/${workspaceId}/dashboards/${id}/filters/add`,
            data,
        );
    },

    updateDashboardFilter(workspaceId, id, filterId, data) {
        return axios.put(
            `/collimato/workspaces/${workspaceId}/dashboards/${id}/filters/${filterId}`,
            data,
        );
    },

    deleteDashboardFilter(workspaceId, id, filterId) {
        return axios.delete(
            `/collimato/workspaces/${workspaceId}/dashboards/${id}/filters/${filterId}`,
        );
    },

    updateDashboardCharts(workspaceId, id, data) {
        return axios.put(`/collimato/workspaces/${workspaceId}/dashboards/${id}/charts`, data);
    },

    deleteDashboard(workspaceId, id) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/dashboards/${id}`);
    },

    createWorkspace(data) {
        return axios.post("/collimato/workspaces", data);
    },

    getWorkspaceById(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}`);
    },

    deleteWorkspace(workspaceId) {
        return axios.delete(`/collimato/workspaces/${workspaceId}`);
    },

    finsihWorkspaceCreation(workspaceId) {
        return axios.post(`/collimato/workspaces/${workspaceId}/finish`);
    },

    getWorkspaces() {
        return axios.get("/collimato/workspaces");
    },

    addUsersToWorkspace(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/users`, data);
    },

    removeUserFromWorkspace(workspaceId, userId) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/users/${userId}`);
    },

    me(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/users/me`);
    },

    getWorkspaceUsers(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/users`);
    },

    getWorkspaceGroups(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/groups`);
    },

    addGroupsToWorkspace(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/groups`, data);
    },

    removeGroupFromWorkspace(workspaceId, groupId) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/groups/${groupId}`);
    },

    updateWorkspaceGroupRoles(workspaceId, groupId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/groups/${groupId}/roles`, data);
    },

    workspaceFiles(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/files`);
    },

    saveFile(workspaceId, data) {
        return axios.put(`/collimato/workspaces/${workspaceId}/files`, data);
    },

    deleteFile(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/files/delete`, data);
    },

    workspaceRoles(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/roles`);
    },

    workspaceRoleById(workspaceId, roleId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/roles/${roleId}`);
    },

    createWorkspaceRole(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/roles`, data);
    },

    updateWorkspaceRole(workspaceId, roleId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/roles/${roleId}/update`, data);
    },

    deleteWorkspaceRole(workspaceId, roleId) {
        return axios.delete(`/collimato/workspaces/${workspaceId}/roles/${roleId}`);
    },

    updateUserRoles(wokspaceId, userId, data) {
        return axios.post(`/collimato/workspaces/${wokspaceId}/users/${userId}/roles`, data);
    },

    newCubeFile(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/files`, data);
    },

    generate(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/files/generate`, data);
    },

    // build persists a structured cube/view model as a rendered YAML file.
    // data = { type: "cube" | "view", model: <Dataset> }
    build(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/files/build`, data);
    },

    // buildView persists a structured view model as a rendered YAML file.
    // data = { model: <View> }
    buildView(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/files/build-view`, data);
    },

    tables(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/tables`);
    },

    loadData(workspaceId, data, signal) {
        return axios.post(`/collimato/workspaces/${workspaceId}/data`, data, {
            signal,
        });
    },

    previewSql(workspaceId, data) {
        return axios.post(`/collimato/workspaces/${workspaceId}/sql`, data);
    },

    meta(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/meta`);
    },

    roleMeta(workspaceId) {
        return axios.get(`/collimato/workspaces/${workspaceId}/meta/roles`);
    },

    listAllWorkspaces() {
        return axios.get("/collimato/manage/workspaces");
    },

    getWorkspaceDetails(workspaceId) {
        return axios.get(`/collimato/manage/workspaces/${workspaceId}`);
    },

    updateManagedWorkspace(workspaceId, data) {
        return axios.put(`/collimato/manage/workspaces/${workspaceId}`, data);
    },

    deleteManagedWorkspace(workspaceId) {
        return axios.delete(`/collimato/manage/workspaces/${workspaceId}`);
    },

    joinManagedWorkspace(workspaceId) {
        return axios.post(`/collimato/manage/workspaces/${workspaceId}/join`);
    },

    addManagedWorkspaceUsers(workspaceId, userIds) {
        return axios.post(`/collimato/manage/workspaces/${workspaceId}/users`, { users: userIds });
    },

    removeManagedWorkspaceUser(workspaceId, userId) {
        return axios.delete(`/collimato/manage/workspaces/${workspaceId}/users/${userId}`);
    },

    updateManagedWorkspaceUserRoles(workspaceId, userId, roles) {
        return axios.post(`/collimato/manage/workspaces/${workspaceId}/users/${userId}/roles`, {
            roles,
        });
    },

    addManagedWorkspaceGroups(workspaceId, data) {
        return axios.post(`/collimato/manage/workspaces/${workspaceId}/groups`, data);
    },

    removeManagedWorkspaceGroup(workspaceId, groupId) {
        return axios.delete(`/collimato/manage/workspaces/${workspaceId}/groups/${groupId}`);
    },

    updateManagedWorkspaceGroupRoles(workspaceId, groupId, roles) {
        return axios.post(`/collimato/manage/workspaces/${workspaceId}/groups/${groupId}/roles`, {
            roles,
        });
    },
};

export default collimatoService;
