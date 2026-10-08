// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, onBeforeUnmount } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";
import { toCommentRow } from "@/utils/projects/taskComment";
import { onProjectChange, onReconnect } from "@/js/websocket";
import { taskChanges } from "./realtime/taskChanges";
import { fieldChanges } from "./realtime/fieldChanges";
import { boardChanges } from "./realtime/boardChanges";
import { viewChanges } from "./realtime/viewChanges";
import { navigationChanges } from "./realtime/navigationChanges";

// Applies the changes other tabs make to the workspace that is open, as they
// arrive on the notification socket.
export function useProjectRealtime() {
    const route = useRoute();
    const router = useRouter();
    const workspaceStore = useWorkspaceStore();
    const userStore = useUserStore();
    const activityDeps = {
        getUser: userStore.getUserById,
    };

    const getSelectedItem = computed({
        get() {
            return workspaceStore.getSelectedItem;
        },
        set(value) {
            workspaceStore.setSelectedItem(value);
        },
    });

    const setFullViewHeaders = computed({
        get() {
            return workspaceStore.getFullViewHeaders;
        },
        set(value) {
            workspaceStore.setFullViewHeaders(value);
        },
    });

    const workspaces = computed({
        get() {
            return workspaceStore.getWorkspaces || []; // Ensure it's always an array
        },
        set(value) {
            workspaceStore.setWorkspaces(value);
        },
    });

    const tableData = computed({
        get() {
            return workspaceStore.getTableData;
        },
        set(value) {
            workspaceStore.setTableData(value);
        },
    });

    const realoadUI = computed({
        get() {
            return workspaceStore.getWorkspacesReady;
        },
        set(value) {
            workspaceStore.setWorkspacesReady(value);
        },
    });

    const columnWidths = computed({
        get() {
            return workspaceStore.getColumnsWidths;
        },
        set(value) {
            workspaceStore.setColumnsWidths(value);
        },
    });

    const tableHeaders = computed({
        get() {
            return workspaceStore.getTableHeaders;
        },
        set(value) {
            workspaceStore.setTableHeaders(value);
        },
    });

    const typesCreated = computed({
        get() {
            return workspaceStore.getViewTypes;
        },
        set(value) {
            workspaceStore.setViewTypes(value);
        },
    });

    const taskLists = computed({
        get() {
            return workspaceStore.getKanbanTaskList || []; // Ensure it's always an array
        },
        set(value) {
            workspaceStore.setKanbanTaskList(value);
        },
    });

    const foldersStore = computed({
        get() {
            return workspaceStore.getWorkspaceFolders;
        },
        set(value) {
            workspaceStore.setWorkspaceFolders(value);
        },
    });

    const titles = computed({
        get() {
            return workspaceStore.getKanbanTitles || []; // Ensure it's always an array
        },
        set(value) {
            workspaceStore.setKanbanTitles(value);
        },
    });

    const selectedView = computed({
        get() {
            return workspaceStore.getSelectedViewData;
        },
        set(value) {
            workspaceStore.setSelectedViewData(value);
        },
    });

    const workspaceTables = computed({
        get() {
            return workspaceStore.getWorkspaceTables;
        },
        set(value) {
            workspaceStore.setWorkspaceTables(value);
        },
    });

    const loaded = computed({
        get() {
            return workspaceStore.getIsLoaded;
        },
        set(value) {
            workspaceStore.setIsLoaded(value);
        },
    });

    const ctx = {
        route,
        router,
        workspaceStore,
        userStore,
        getSelectedItem,
        setFullViewHeaders,
        workspaces,
        tableData,
        realoadUI,
        columnWidths,
        tableHeaders,
        typesCreated,
        taskLists,
        foldersStore,
        titles,
        selectedView,
        workspaceTables,
        loaded,
    };
    const { onTaskCreated, applyTaskRow, removeTaskRow, removeKanbanCard } = taskChanges(ctx);
    const {
        newFieldHeader,
        applyFieldFormula,
        onNewField,
        onDisplayNameUpdate,
        onHeaderOrderUpdate,
        onDeleteSingleField,
        onDeleteField,
        onGridHeaderUpdate,
    } = fieldChanges({ ...ctx, removeTaskRow });
    const { onNewKanbanSection, onKanbanSectionRename, onKanbanLayout } = boardChanges(ctx);
    const { onCreateGridView, onFilterSaved, onDeleteView, onViewVisibility, onViewCreated } =
        viewChanges(ctx);
    const {
        onDeleteFolder,
        onDeleteTable,
        onCreateFolder,
        onUpdateTableName,
        onMoveItem,
        onUpdateFolderName,
        onUpdateTables,
        onDeleteWorkspace,
    } = navigationChanges(ctx);

    function handleProjectChange(change) {
        if (!change || route.params.id !== change.workspace_id) return;
        const own =
            !!change.origin_client_id && change.origin_client_id === workspaceStore.getConnectionID;

        const relayedHandler = RELAYED_HANDLERS[change.type];

        if (relayedHandler) {
            const relayed = { ...change, client_id: change.origin_client_id || "" };

            if (change.type === "NEW_FIELD") relayed.fieldData = newFieldHeader(change);
            relayedHandler(relayed);

            return;
        }

        switch (change.type) {
            case "task_created":
            case "task_updated": {
                const row = change.task || change.row;

                if (row) applyTaskRow(change.table_id, row, own, change.type === "task_created");
                break;
            }

            case "task_deleted": {
                const id = change.task?.id || change.row_id;

                if (id) removeTaskRow(change.table_id, id, !!change.row_id, own);
                break;
            }

            case "field_formula_updated":
                if (!own) applyFieldFormula(change.table_id, change.field_name, change.formula);
                break;
            case "task_unassigned":
                if (route.params.tid === change.table_id) {
                    tableData.value = tableData.value.filter(
                        (t) => String(t.id) !== String(change.task_id),
                    );
                    removeKanbanCard(change.task_id);
                }

                break;
            case "comment_added":
                if (
                    !own &&
                    route.params.tid === change.table_id &&
                    change.task_id === getSelectedItem.value?.id
                ) {
                    workspaceStore.appendActivity(toCommentRow(change.comment, activityDeps));
                }

                break;
        }
    }

    // These changes arrive in the shape their handlers expect, with the tab that
    // made them as client_id.
    const RELAYED_HANDLERS = {
        NEW_FIELD: onNewField,
        CREATE_GRID_VIEW: onCreateGridView,
        DISPLAY_NAME_UPDATE: onDisplayNameUpdate,
        UPDATE_GRD_HEADER_ORDER: onHeaderOrderUpdate,
        FILTER_CREATED: onFilterSaved,
        FILTER_UPDATED: onFilterSaved,
        DELETE_VIEW: onDeleteView,
        VIEW_VISIBILITY: onViewVisibility,
        DELETE_FOLDER: onDeleteFolder,
        DELETE_SINGLE_FIELD: onDeleteSingleField,
        DELETE_FIELD: onDeleteField,
        DELETE_TABLE: onDeleteTable,
        VIEW_CREATED: onViewCreated,
        NEW_KANBAN_SECTION_ADDED: onNewKanbanSection,
        UPDATE_SINGLE_SELECT_HEADER_NAME: onKanbanSectionRename,
        GRID_HEADER_UPDATE: onGridHeaderUpdate,
        CREATE_FOLDER: onCreateFolder,
        UPDATE_TABLE_NAME: onUpdateTableName,
        MOVE_ITEM: onMoveItem,
        UPDATE_FOLDER_NAME: onUpdateFolderName,
        UPDATE_TABLES: onUpdateTables,
        DELETE_WORKSPACE: onDeleteWorkspace,
        KANBAN_COLUMNS: onKanbanLayout,
        KANBAN_CARD_MOVED: onKanbanLayout,
    };

    const stopProjectChanges = onProjectChange(handleProjectChange);
    const stopReconnect = onReconnect(() => workspaceStore.bumpGridReloadToken());

    onBeforeUnmount(() => {
        stopProjectChanges();
        stopReconnect();
    });

    return { onTaskCreated };
}
