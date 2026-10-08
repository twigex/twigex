// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const workspaceService = {
    createWorkspace(data) {
        return axios.post(`/workspaces/create`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    createWorkspaceTable(data) {
        return axios.post(`/workspaces/${data.id}/table/create`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    uploadFiles(formData) {
        const workspaceId = formData.get("workspace_id");

        formData.delete("workspace_id");

        return axios.post(`/workspaces/${workspaceId}/files`, formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        });
    },

    updateWorkspaceFolder(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/folder/${data.folder_id}/update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    moveWorkspaceItem(data) {
        return axios.post(`/workspaces/${data.workspace_id}/move-item`, data);
    },

    getLinkingTables({ workspace_id, table_id, folder_id }) {
        const item = folder_id ? `folder/${folder_id}` : `table/${table_id}`;

        return axios.get(`/workspaces/${workspace_id}/${item}/linking-tables`);
    },

    deleteWorkspaceFolder(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/folder/${data.folder_id}/delete`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceMember(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/member/${data.member_id}/delete`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceTableField(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/field/${data.field_id}/delete`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceTableSingleField(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/single/${data.field_id}/delete`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceTableView(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/delete`,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceItem(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.item_id}/delete`,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceTable(data) {
        return axios.post(`/workspaces/${data.workspace_id}/table/${data.table_id}/delete`, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    deleteWorkspace(workspace_id) {
        return axios.post(`/workspaces/${workspace_id}/delete`, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    createNewField(data) {
        return axios.post(`/workspaces/${data.workspace_id}/table/${data.table_id}/create`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    updateFieldFormula(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/field-edit-formula`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    createNewView(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/create`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateView(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    moveKanbanCard(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/kanban/move`,
            { task_id: data.task_id, column: data.column, after_id: data.after_id },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateMemberRole(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/member/${data.user_id}/update-role`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateTableLink(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/update-link`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    createNewTask(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/create`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    addMemberToTask(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/assigne`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    createWorkspaceRole(workspaceId, data) {
        return axios.post(`/workspaces/${workspaceId}/member/role/create`, data);
    },

    updateWorkspaceRole(workspaceId, roleId, data) {
        return axios.post(`/workspaces/${workspaceId}/member/role/${roleId}/update`, data);
    },

    addMembers(data) {
        return axios.post(`/workspaces/${data.workspace_id}/member/add`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    addFieldValue(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/field-value/add`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateDisplayName(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/display-name-update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateSingleSelectHeaderName(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/single-select-name/${data.task_id}/update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateTask(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateColumnWidth(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/field-update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    setViewVisibility(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/set-visibility`,
            { is_public: data.is_public },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    getViewShares(data) {
        return axios.get(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/shares`,
        );
    },

    setViewShares(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/shares`,
            { user_ids: data.user_ids },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteFile(data) {
        return axios.delete(`/workspaces/${data.workspace_id}/file/${data.file_id}`, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    updateGridHeaderOrder(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/header/update`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateWorkspaceTable(data) {
        return axios.post(`/workspaces/${data.workspace_id}/table/${data.table_id}/update`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    saveFilter(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/save-filter`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateFilter(data) {
        return axios.put(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/update-filter`,
            {
                filter_id: data.filter_id,
                saved_filter_id: data.saved_filter_id,
                name: data.name,
                type: data.type,
                filters: data.filters,
                is_private: data.is_private,
                is_active: data.is_active,
            },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateWorkspace(data) {
        return axios.post(`/workspaces/${data.workspace_id}/update`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    deleteSavedFilter(data) {
        return axios.delete(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/saved-filter/${data.filter_id}`,
        );
    },

    updateFilterActiveStatus(data) {
        return axios.put(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/update-filter-active`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    saveGridSort(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/save-sort`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    getTasksByDateRange(workspaceId, tableId, from, to, filters = null, signal = null) {
        return axios.post(
            `/workspaces/${workspaceId}/table/${tableId}/tasks/date-range`,
            {
                from,
                to,
                ...(filters ? { filters } : {}),
            },
            signal ? { signal } : {},
        );
    },

    getTasksByCalendarRange(workspaceId, tableId, from, to, filters = null, signal) {
        return axios.post(
            `/workspaces/${workspaceId}/table/${tableId}/tasks/calendar-range`,
            {
                from,
                to,
                ...(filters ? { filters } : {}),
            },
            { signal },
        );
    },

    getFilteredTableData(data, signal) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/view/${data.view_id}/filter`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
                signal,
            },
        );
    },

    addTaskComment(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/add-comment`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    deleteWorkspaceRole(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/member/role/${data.role_id}/delete`,
            data,
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    createWorkspaceFolder(data) {
        return axios.post(`/workspaces/${data.id}/folder/create`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    getWorkspaceRoles(workspaceId) {
        return axios.get(`/workspaces/${workspaceId}/member/roles/get`);
    },

    getTablesForRoleEditor(workspaceId) {
        return axios.get(`/workspaces/${workspaceId}/tables/for-role-editor`);
    },

    getRoleTablePermissions(workspaceId, roleId) {
        return axios.get(`/workspaces/${workspaceId}/member/role/${roleId}/table-permissions`);
    },

    updateRoleTablePermissions(workspaceId, roleId, perms) {
        return axios.put(
            `/workspaces/${workspaceId}/member/role/${roleId}/table-permissions`,
            perms,
        );
    },

    getItemForTableByID(data) {
        return axios.get(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/item/${data.task_id}`,
        );
    },

    getSubtasks(data, signal) {
        return axios.get(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/item/${data.task_id}/subtasks`,
            { signal },
        );
    },

    getTaskCompletion(data) {
        return axios.get(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/item/${data.task_id}/completion`,
        );
    },

    completeSubtasks(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/item/${data.task_id}/complete-subtasks`,
            { status: data.status },
        );
    },

    taskMatchesFilter(workspaceId, tableId, taskId, filters) {
        return axios.post(
            `/workspaces/${workspaceId}/table/${tableId}/task/${taskId}/matches`,
            filters,
        );
    },

    getTaskPage(data) {
        return axios.post(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/page`,
            {
                limit: data.limit || 100,
                sort: data.sort || [],
                flat_filters: data.flat_filters || [],
                groups: data.groups || [],
            },
        );
    },

    getTaskComments(data) {
        return axios.get(
            `/workspaces/${data.workspace_id}/table/${data.table_id}/task/${data.task_id}/get-comments`,
        );
    },

    getAssignedToMe(params = {}, signal) {
        return axios.get(`/workspaces/assigned-to-me`, { params, signal });
    },

    searchWorkspaceMembers(workspaceId, q = "", limit = 20, offset = 0) {
        return axios.get(`/workspaces/${workspaceId}/members/search`, {
            params: { q, limit, offset },
        });
    },

    searchTaskReportMembers(workspaceIds = [], q = "", limit = 50) {
        return axios.get("/workspaces/all-workspace-tasks/members", {
            params: { workspace_ids: workspaceIds.join(","), q, limit },
        });
    },

    getWorkspaceGroups(workspaceId) {
        return axios.get(`/workspaces/${workspaceId}/groups`);
    },

    getKanbanData(workspaceId, tableId, body, signal) {
        return axios.post(`/workspaces/${workspaceId}/table/${tableId}/kanban-data`, body, {
            signal,
        });
    },

    addWorkspaceGroups(workspaceId, groupIds, roles) {
        return axios.post(`/workspaces/${workspaceId}/groups`, { group_ids: groupIds, roles });
    },

    removeWorkspaceGroup(workspaceId, groupId) {
        return axios.delete(`/workspaces/${workspaceId}/groups/${groupId}`);
    },

    updateWorkspaceGroupRoles(workspaceId, groupId, roles) {
        return axios.post(`/workspaces/${workspaceId}/groups/${groupId}/roles`, { roles });
    },

    getWorkspaceNav(id) {
        return axios.get(`/workspaces/get/${id}?nav=1`);
    },

    getWorkspaces() {
        return axios.get("/workspaces/get");
    },

    getWorkspaceTables(id, data, page, limit) {
        const params = page && limit ? { page, limit } : undefined;

        return axios.get(`/workspaces/${id}/tables/get/${data}`, { params });
    },

    getAllWorkspaceTasks(data) {
        return axios.post("/workspaces/all-workspace-tasks", data);
    },

    getStatusList(data) {
        return axios.post(`/workspaces/tasks/status-list`, data);
    },

    getTableStatusTypes(workspaceId, tableId) {
        return axios.get(`/workspaces/${workspaceId}/table/${tableId}/status-types`);
    },

    getLinkedRecordsLite(workspaceId, tableId, params = {}) {
        return axios.get(`/workspaces/${workspaceId}/table/${tableId}/rows-lite`, { params });
    },

    me(id) {
        return axios.get(`/workspaces/${id}/users/me`);
    },

    listAllManaged() {
        return axios.get("/workspaces/manage");
    },

    getManaged(id) {
        return axios.get(`/workspaces/manage/${id}`);
    },

    updateManaged(id, data) {
        return axios.put(`/workspaces/manage/${id}`, data);
    },

    deleteManaged(id) {
        return axios.delete(`/workspaces/manage/${id}`);
    },

    joinManaged(id) {
        return axios.post(`/workspaces/manage/${id}/join`);
    },

    addManagedMembers(id, userIds) {
        return axios.post(`/workspaces/manage/${id}/members`, { users: userIds });
    },

    removeManagedMember(id, userId) {
        return axios.delete(`/workspaces/manage/${id}/members/${userId}`);
    },

    updateManagedMemberRole(id, userId, role) {
        return axios.post(`/workspaces/manage/${id}/members/${userId}/role`, { role });
    },

    addManagedGroups(id, data) {
        return axios.post(`/workspaces/manage/${id}/groups`, data);
    },

    removeManagedGroup(id, groupId) {
        return axios.delete(`/workspaces/manage/${id}/groups/${groupId}`);
    },

    updateManagedGroupRoles(id, groupId, roles) {
        return axios.post(`/workspaces/manage/${id}/groups/${groupId}/roles`, { roles });
    },
};

export default workspaceService;
