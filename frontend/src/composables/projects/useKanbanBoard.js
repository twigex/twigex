// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, nextTick, onBeforeUnmount, ref, toRaw, watch } from "vue";
import { useRoute } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import workspaceService from "@/services/workspaceService";
import { extractErrorMessage } from "@/utils/errors";
import { appliedFlatFilters } from "@/utils/projects/filterCombine";
import { fieldLabel, normalizeTaskRow, optionLookup, withOptions } from "@/utils/projects/rows";
import { tableDisplayNames } from "@/utils/projects/gridView";
import { workspaceTables as tablesOf } from "@/utils/projects/tree";

export const KANBAN_PAGE = 50; // render 50 cards initially, user loads more on demand

// Loads the board the route names, a page of cards per column, and loads it
// again when the grid reload token moves. A column loads its next page when
// asked.
export function useKanbanBoard() {
    const route = useRoute();
    const workspaceStore = useWorkspaceStore();
    const userStore = useUserStore();

    const kanbanLoading = ref(false);
    const singleSelectValue = ref(null);
    const subtaskMap = ref({});

    let kanbanActive = true;
    // A column's "load more" belongs to the board it was asked for, so it is
    // cancelled together with that board.
    let boardRequest = new AbortController();

    onBeforeUnmount(() => {
        kanbanActive = false;
        boardRequest.abort();
    });

    const taskLists = computed({
        get() {
            return workspaceStore.getKanbanTaskList || [];
        },
        set(value) {
            workspaceStore.setKanbanTaskList(value);
        },
    });

    const titles = computed({
        get() {
            return workspaceStore.getKanbanTitles || [];
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

    const tableHeaders = computed({
        get() {
            return workspaceStore.getTableHeaders;
        },
        set(value) {
            workspaceStore.setTableHeaders(value);
        },
    });

    const tableNameLookup = computed({
        get() {
            return workspaceStore.getTableNameLookup;
        },
        set(value) {
            workspaceStore.setTableNameLookup(value);
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

    const workspaceTables = computed({
        get() {
            return workspaceStore.getWorkspaceTables;
        },
        set(value) {
            workspaceStore.setWorkspaceTables(value);
        },
    });

    function addSubtasks(subtasks) {
        for (const sub of subtasks || []) {
            const parentId = sub.parent_task_id;

            if (!parentId) continue;
            const list = subtaskMap.value[parentId] || (subtaskMap.value[parentId] = []);

            if (!list.some((s) => s.id === sub.id)) list.push(sub);
        }
    }

    function activeFilters() {
        const groups = toRaw(workspaceStore.getSavedGroups || []);
        const flatFilters = appliedFlatFilters(toRaw(workspaceStore.getSavedFlatFilters));

        return { groups, flatFilters, timezone: userStore.getTimezone };
    }

    function updateTitles() {
        if (Array.isArray(taskLists.value)) {
            titles.value = taskLists.value
                .sort((a, b) => (a.title === "Unassigned" ? -1 : b.title === "Unassigned" ? 1 : 0))
                .map((list) => ({
                    id: list.id,
                    title: list.title,
                    display_name: list.display_name,
                    color: list.color,
                    width: list.width,
                }));
        }
    }

    function enrichHeaders(headers, allTables) {
        // Build table lookup once instead of O(tables) find per header.
        const tableById = new Map(allTables.map((t) => [t.id, t]));

        return headers.map((header) => {
            const enrichedHeader = { ...header };

            if (header.single_select && header.linked_id) {
                const linkedTable = tableById.get(header.linked_id);

                enrichedHeader.options =
                    linkedTable && Array.isArray(linkedTable.data_base)
                        ? linkedTable.data_base
                        : [];
            }

            return enrichedHeader;
        });
    }

    async function loadMoreTasks(colId) {
        const list = taskLists.value.find((l) => l.id === colId);

        if (!list || list.loadingMore) return;
        list.loadingMore = true;
        let next;
        const { signal } = boardRequest;

        try {
            const res = await workspaceService.getKanbanData(
                route.params.id,
                route.params.tid,
                {
                    view_id: route.params.fid,
                    column: colId,
                    after: list.next,
                    limit: KANBAN_PAGE,
                    filters: activeFilters(),
                },
                signal,
            );

            if (signal.aborted) return;
            const page = res.data?.columns?.[0];
            const shown = new Set(list.tasks.map((t) => String(t.id)));

            next = (page?.tasks || []).filter((t) => !shown.has(String(t.id)));
            if (page) {
                list.totalCount = page.total;
                list.next = page.next || "";
            }

            addSubtasks(res.data?.subtasks);

            const userFields = tableHeaders.value
                .filter(
                    (h) => h.header_usage === "assignee" || h.header_usage === "default_assignee",
                )
                .map((h) => h.name);
            const userIds = [
                ...new Set(next.flatMap((task) => userFields.map((f) => task[f])).filter(Boolean)),
            ];

            if (userIds.length) userStore.ensureUsers(userIds);
        } catch (error) {
            if (!signal.aborted) useAlertStore().showError(extractErrorMessage(error));

            return;
        } finally {
            list.loadingMore = false;
        }

        const options = optionLookup(tableHeaders.value, workspaceTables.value);
        const enriched = next.map((task) => withOptions(task, options));

        list.tasks.push(...enriched);
        list.order = [...(list.order || []), ...enriched.map((t) => ({ id: t.id }))];
    }

    let boardLoad = 0;

    watch(
        [
            () => route.params.id,
            () => route.params.tid,
            () => route.params.fid,
            () => workspaceStore.getGridReloadToken,
        ],
        async ([id, tid, fid]) => {
            const load = ++boardLoad;

            boardRequest.abort();
            boardRequest = new AbortController();
            kanbanLoading.value = true;
            try {
                const [navRes, kanbanRes] = await Promise.all([
                    workspaceStore.fetchWorkspaceNav(id),
                    workspaceService.getKanbanData(
                        id,
                        tid,
                        {
                            view_id: fid,
                            limit: KANBAN_PAGE,
                            filters: activeFilters(),
                        },
                        boardRequest.signal,
                    ),
                ]);

                if (!kanbanActive || load !== boardLoad) return;

                const allTables = tablesOf(navRes.data);

                workspaceTables.value = allTables;
                tableNameLookup.value = tableDisplayNames(allTables);
                const foundTable = allTables.find((t) => String(t.id) === String(tid));

                if (!foundTable) {
                    kanbanLoading.value = false;

                    return;
                }

                const foundView =
                    foundTable.views?.find((v) => v.id === fid) ||
                    foundTable.options?.find((v) => v.id === fid);

                if (!foundView || foundView.view_type !== "kanban") {
                    kanbanLoading.value = false;

                    return;
                }

                const viewOrder = foundView.order ? JSON.parse(foundView.order) : { fields: [] };
                const board = kanbanRes.data || {};
                const columnPages = new Map((board.columns || []).map((c) => [String(c.id), c]));
                const allTasks = [
                    ...(board.columns || []).flatMap((c) => c.tasks || []),
                    ...(board.subtasks || []),
                ];

                subtaskMap.value = {};
                addSubtasks(board.subtasks);

                if (!Array.isArray(viewOrder.fieldsVisible)) {
                    viewOrder.fieldsVisible = [];
                }

                const visibilityMap = new Map(
                    viewOrder.fieldsVisible.map((field) => [field.name, field.visible]),
                );

                let orderFields = [];
                const rawOrder = foundTable.options?.[0]?.order;

                if (typeof rawOrder === "string") {
                    orderFields = JSON.parse(rawOrder);
                }

                // Pre-build lookup so each header doesn't scan the full orderFields array.
                const orderFieldMap = new Map(orderFields.map((f) => [f.name, f]));

                const enriched = enrichHeaders(foundTable.headers, allTables);

                tableHeaders.value = enriched
                    .filter((header) => header.name !== "id")
                    .map((header) => {
                        const visibleState = visibilityMap.has(header.name)
                            ? visibilityMap.get(header.name)
                            : header.name === "name";

                        const displayFromOrder = orderFieldMap.get(header.name)?.display_name;

                        return {
                            ...header,
                            visible: visibleState,
                            display_name: fieldLabel(header, displayFromOrder),
                        };
                    });

                tableData.value = (foundTable.data_base || []).map((item) =>
                    normalizeTaskRow(item, foundTable.headers || []),
                );

                const userFieldNames = tableHeaders.value
                    .filter(
                        (h) =>
                            h.header_usage === "assignee" || h.header_usage === "default_assignee",
                    )
                    .map((h) => h.name);
                const allUserIds = new Set([
                    ...tableData.value.map((r) => r.assignee).filter(Boolean),
                    ...allTasks.flatMap((t) => userFieldNames.map((f) => t[f]).filter(Boolean)),
                ]);

                if (allUserIds.size) userStore.ensureUsers([...allUserIds]);

                const options = optionLookup(tableHeaders.value, allTables);

                const newTaskLists = viewOrder.fields.map((field) => {
                    const page = columnPages.get(String(field.id));
                    const tasks = (page?.tasks || []).map((task) => withOptions(task, options));

                    return {
                        id: field.id,
                        title: field.name || "Unnamed",
                        display_name: field.display_name || field.name || "Unnamed",
                        color: field.color || "#FFFFFF",
                        width: field.width || 300,
                        visible: field.visible ?? true,
                        order: tasks.map((t) => ({ id: t.id })),
                        totalCount: page?.total || 0,
                        next: page?.next || "",
                        tasks,
                    };
                });

                taskLists.value = [...newTaskLists];
                updateTitles();
                selectedView.value = foundView;
                singleSelectValue.value = viewOrder.section;

                await nextTick();
                kanbanLoading.value = false;
            } catch (error) {
                if (!kanbanActive || load !== boardLoad) return;
                kanbanLoading.value = false;
                useAlertStore().showError(extractErrorMessage(error));
            }
        },
        { immediate: true },
    );

    return { kanbanLoading, singleSelectValue, subtaskMap, loadMoreTasks };
}
