// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import { t } from "@/i18n/index.js";
import workspaceService from "@/services/workspaceService";

// Requests for a workspace's navigation still in flight, by workspace id.
// Views that load together on one navigation share a single request; once it
// settles the next caller asks again, so nothing is served stale.
const navRequests = new Map();

export const useWorkspaceStore = defineStore("workspaces", {
    state: () => {
        return {
            workspace: {},
            workspaces: [],
            data: [],
            headers: [],
            connectionID: null,
            columnWidths: [],
            selectedView: "Grid View",
            viewTypes: [{ name: "Grid View" }],
            openWorkspaces: new Set(),
            kanbanTaskList: [],
            kanbanTitles: ["Uncategorized"],
            view: {},
            tables: [],
            showRightNavigation: false,
            showFullTask: false,
            testStore: [],
            selectedItem: {},
            dialogTableID: null,
            fullViewHeaders: [],
            tableID: null,
            currentWorkspace: null,
            tableNameLookup: {},
            workspaceFolders: [],
            workspacesUpdateID: 0,
            filterActive: false,
            savedFlatFilters: [
                {
                    field: "",
                    operator: "is",
                    value: "",
                    operatorBetween: "AND",
                },
            ],
            savedGroups: [],
            fieldName: "",
            sortOptions: [{ field: "", direction: "asc" }],
            gridReloadToken: 0,
            gridPage: 1,
            kanbanDisplayNameUpdateKey: 220,
            selectedWorkspaces: [],
            workspaceTablesForLinked: [],
            catchedDataForDialog: [],
            linkedFieldName: "",
            storedTableDataForNestedDialog: [],
            propsLinkedUpdate: false,
            sortActive: false,
            workspaceDetails: {},
            isTouchedChecbox: false,
            quickFilter: false,
            defaultFlatFilters: [],
            defaultGroupFilter: [],
            workspaceCreateDialogOpen: false,
            newTaskDialog: null,
            createdTask: null,
            // A field the task panel saved, for pages that keep their own
            // copy of the tasks: { id, field, value }.
            savedTaskField: null,
            isLoaded: false,
            workspaceRole: {},
            hasWorkspaces: false,
            notificationRedirect: false,
            filters: [],
            filter: {},
            statusOptions: [],
            activity: [],
            taskReportData: [],
            // Whether the task report has run and loaded; its toolbar shows
            // its controls only then, built from what the run applied.
            taskReportReady: false,
            fastFilterOption: "",
            fastFilterLabel: "",
            // Counts every setFastFilter, so the same preset applied again
            // is still seen as a change.
            fastFilterVersion: 0,
        };
    },

    getters: {
        getTaskReportData() {
            return this.taskReportData;
        },
        getFastFilterLabel() {
            return this.fastFilterLabel;
        },
        getStatusOptions() {
            return this.statusOptions;
        },
        getFilter() {
            return this.filter;
        },
        getFilters() {
            return this.filters;
        },
        getNotificationRedirect() {
            return this.notificationRedirect;
        },
        getHasWorkspaces() {
            return this.hasWorkspaces;
        },
        getIsLoaded() {
            return this.isLoaded;
        },
        getWorkspaceCreateDialogOpen() {
            return this.workspaceCreateDialogOpen;
        },
        getWorkspaceRole() {
            return this.workspaceRole;
        },
        getDefaultGroupFilters() {
            return this.defaultGroupFilter;
        },
        getDefaultFlatFilters() {
            return this.defaultFlatFilters;
        },
        getQuickFilterActive() {
            return this.quickFilter;
        },
        getIsTouchedCheckbox() {
            return this.isTouchedChecbox;
        },
        getWorkspaceDetails() {
            return this.workspaceDetails;
        },

        getSortActive() {
            return this.sortActive;
        },
        getPropsLinkedUpdate() {
            return this.propsLinkedUpdate;
        },
        getStoredTableDataForNestedDialog() {
            return this.storedTableDataForNestedDialog;
        },
        getLinkedFieldName() {
            return this.linkedFieldName;
        },
        getCatchedDataForDialog() {
            return this.catchedDataForDialog;
        },
        getWorkspaceTablesForLinked() {
            return this.workspaceTablesForLinked;
        },
        getSelectedWorkspaces() {
            return this.selectedWorkspaces;
        },
        getKanbanDisplayNameUpdateKey() {
            return this.kanbanDisplayNameUpdateKey;
        },
        getSortOptions() {
            return this.sortOptions;
        },
        getGridReloadToken() {
            return this.gridReloadToken;
        },
        getFieldName() {
            return this.fieldName;
        },
        getSavedFlatFilters() {
            return this.savedFlatFilters;
        },
        getSavedGroups() {
            return this.savedGroups;
        },
        getFilterActive() {
            return this.filterActive;
        },
        getWorkspacesReady() {
            return this.workspacesUpdateID;
        },
        getCurrentWorkspace() {
            return this.currentWorkspace;
        },
        getTableNameLookup() {
            return this.tableNameLookup;
        },
        tableNameOf: (state) => (tableId) =>
            state.tableNameLookup[tableId] ||
            t.value("projects.full_task_navigation.unknown_table"),
        getTableID() {
            return this.tableID;
        },
        getFullViewHeaders() {
            return this.fullViewHeaders;
        },

        getDialogTableID() {
            return this.dialogTableID;
        },

        getTestStore() {
            return this.testStore;
        },

        getWorkspaceFolders() {
            return this.workspaceFolders;
        },

        getWorkspace() {
            return this.workspace;
        },

        getSelectedItem() {
            return this.selectedItem;
        },

        getFullTask() {
            return this.showFullTask;
        },

        getRightNavigation() {
            return this.showRightNavigation;
        },

        getWorkspaceTables() {
            return this.tables;
        },

        getSelectedViewData() {
            return this.view;
        },

        getKanbanTitles() {
            return this.kanbanTitles;
        },
        getKanbanTaskList() {
            return this.kanbanTaskList;
        },

        getViewTypes() {
            return this.viewTypes;
        },

        getSelectedView() {
            return this.selectedView;
        },

        getColumnsWidths() {
            return this.columnWidths;
        },

        getTableHeaders() {
            return this.headers;
        },

        getTableData() {
            return this.data;
        },

        getWorkspaces() {
            return this.workspaces;
        },

        getConnectionID() {
            return this.connectionID;
        },
    },

    actions: {
        fetchWorkspaceNav(workspaceId) {
            const pending = navRequests.get(workspaceId);

            if (pending) return pending;
            const request = workspaceService
                .getWorkspaceNav(workspaceId)
                .finally(() => navRequests.delete(workspaceId));

            navRequests.set(workspaceId, request);

            return request;
        },
        setTaskReportData(data) {
            this.taskReportData = data;
        },
        setFastFilter(option, label) {
            this.fastFilterOption = option;
            this.fastFilterLabel = label;
            this.fastFilterVersion++;
        },
        setActivity(activity) {
            this.activity = activity;
        },
        appendActivity(row) {
            this.activity.push(row);
        },
        setStatusOptions(statusOptions) {
            this.statusOptions = statusOptions;
        },
        setFilter(filter) {
            this.filter = filter;
        },
        setFilters(filters) {
            this.filters = filters;
        },
        setNotificationRedirect(notificationRedirect) {
            this.notificationRedirect = notificationRedirect;
        },
        setHasWorkspaces(hasWorkspaces) {
            this.hasWorkspaces = hasWorkspaces;
        },
        setIsLoaded(loaded) {
            this.isLoaded = loaded;
        },
        setWorkspaceCreateDialogOpen(open) {
            this.workspaceCreateDialogOpen = open;
        },
        // preset may name the Kanban column the task is added in, as
        // { section, singleSelect, columnName }.
        openNewTaskDialog(preset = {}) {
            this.newTaskDialog = { ...preset };
        },
        closeNewTaskDialog() {
            this.newTaskDialog = null;
        },
        announceCreatedTask(task) {
            this.createdTask = task;
        },
        announceSavedTaskField(id, field, value) {
            this.savedTaskField = { id, field, value };
        },
        setWorkspaceRole(role) {
            this.workspaceRole = role;
        },
        setDefaultGroupFilters(defaultGroupFilter) {
            this.defaultGroupFilter = defaultGroupFilter;
        },
        setDefaultFlatFilters(defaultFlatFilters) {
            this.defaultFlatFilters = defaultFlatFilters;
        },
        setQuickFilterActive(quickFilter) {
            this.quickFilter = quickFilter;
        },
        setIsTouchedCheckbox(isTouchedChecbox) {
            this.isTouchedChecbox = isTouchedChecbox;
        },
        setWorkspaceDetails(workspaceDetails) {
            this.workspaceDetails = workspaceDetails;
        },

        setSortActive(sortActive) {
            this.sortActive = sortActive;
        },
        setPropsLinkedUpdate(propsLinkedUpdate) {
            this.propsLinkedUpdate = propsLinkedUpdate;
        },
        setStoredTableDataForNestedDialog(storedTableDataForNestedDialog) {
            this.storedTableDataForNestedDialog = storedTableDataForNestedDialog;
        },
        setLinkedFieldName(linkedFieldName) {
            this.linkedFieldName = linkedFieldName;
        },
        setCatchedDataForDialog(catchedDataForDialog) {
            this.catchedDataForDialog = catchedDataForDialog;
        },
        setWorkspaceTablesForLinked(workspaceTablesForLinked) {
            this.workspaceTablesForLinked = workspaceTablesForLinked;
        },
        setSelectedWorkspaces(selectedWorkspaces) {
            this.selectedWorkspaces = selectedWorkspaces;
        },
        setKanbanDisplayNameUpdateKey(kanbanDisplayNameUpdateKey) {
            this.kanbanDisplayNameUpdateKey = kanbanDisplayNameUpdateKey;
        },
        setSortOptions(sortOptions) {
            this.sortOptions = sortOptions;
        },
        bumpGridReloadToken() {
            this.gridReloadToken++;
        },
        setFieldName(fieldName) {
            this.fieldName = fieldName;
        },
        setSavedFlatFilters(savedFlatFilters) {
            this.savedFlatFilters = savedFlatFilters;
        },
        setSavedGroups(savedGroups) {
            this.savedGroups = savedGroups;
        },
        setFilterActive(filterActive) {
            this.filterActive = filterActive;
        },
        setWorkspacesReady(ready) {
            this.workspacesUpdateID = ready;
        },
        setCurrentWorkspace(workspace) {
            this.currentWorkspace = workspace;
        },
        setTableNameLookup(tableNameLookup) {
            this.tableNameLookup = tableNameLookup;
        },
        setTableID(tableID) {
            this.tableID = tableID;
        },

        setFullViewHeaders(fullViewHeaders) {
            this.fullViewHeaders = fullViewHeaders;
        },

        setDialogTableID(dialogTableID) {
            this.dialogTableID = dialogTableID;
        },

        setTestStore(testStore) {
            this.testStore = testStore;
        },

        setWorkspaceFolders(workspaceFolders) {
            this.workspaceFolders = workspaceFolders;
        },

        setSelectedItem(selectedItem) {
            this.selectedItem = { ...selectedItem };
        },

        setFullTask(showFullTask) {
            this.showFullTask = showFullTask;
        },

        setRightNavigation(showRightNavigation) {
            this.showRightNavigation = showRightNavigation;
        },

        setWorkspaceTables(tables) {
            this.tables = tables;
        },

        setSelectedViewData(view) {
            this.view = view;
        },

        setKanbanTitles(kanbanTitles) {
            this.kanbanTitles = kanbanTitles;
        },

        setKanbanTaskList(kanbanTaskList) {
            this.kanbanTaskList = kanbanTaskList;
        },

        toggleWorkspace(workspaceId) {
            if (this.openWorkspaces.has(workspaceId)) {
                this.openWorkspaces.delete(workspaceId);
            } else {
                this.openWorkspaces.add(workspaceId);
            }
        },
        isWorkspaceOpen(workspaceId) {
            return this.openWorkspaces.has(workspaceId);
        },

        setViewTypes(viewTypes) {
            this.viewTypes = viewTypes;
        },

        setSelectedView(view) {
            this.selectedView = view;
        },

        setTaskReportReady(ready) {
            this.taskReportReady = ready;
        },

        clearTableData() {
            this.data = [];
        },

        setTableHeaders(headers) {
            this.headers = headers;
        },

        setTableData(data) {
            this.data = data;
        },

        setWorkspace(workspace) {
            this.workspace = workspace;
        },

        setWorkspaces(workspaces) {
            this.workspaces = workspaces;
        },

        setConnectionID(connectionID) {
            this.connectionID = connectionID;
        },

        setColumnsWidths(columnWidths) {
            this.columnWidths = columnWidths;
        },
    },
});
