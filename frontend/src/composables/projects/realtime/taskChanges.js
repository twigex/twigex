// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useShownFilter } from "@/composables/projects/useShownFilter";
import { normalizeTaskRow, clearReferences, refreshEmbedded } from "@/utils/projects/rows";

// taskChanges applies changes to a table's tasks: rows created, changed or
// deleted, on the grid page shown and as cards on the board.
export function taskChanges(ctx) {
    const {
        route,
        workspaceStore,
        userStore,
        getSelectedItem,
        tableData,
        tableHeaders,
        taskLists,
        selectedView,
        workspaceTables,
    } = ctx;

    function addTaskToKanbanView(processedTask) {
        const view = selectedView.value;

        if (!view || view.view_type !== "kanban") return;

        try {
            const orderConfig =
                typeof view.order === "string" ? JSON.parse(view.order) : view.order;
            const kanbanField = orderConfig.section;

            if (!kanbanField) return;

            const taskKanbanValue = processedTask[kanbanField];
            const kanbanValueId = taskKanbanValue?.id || taskKanbanValue;

            let targetColumn = taskLists.value.find(
                (column) => String(column.id) === String(kanbanValueId),
            );

            if (!targetColumn && (kanbanValueId === null || kanbanValueId === undefined)) {
                targetColumn = taskLists.value.find((column) => String(column.id) === "0");
            }

            if (!targetColumn && taskLists.value.length > 0) {
                targetColumn = taskLists.value[0];
            }

            if (targetColumn) {
                // Check if task already exists in this column
                const taskExists = targetColumn.tasks.some(
                    (t) => String(t.id) === String(processedTask.id),
                );

                if (!taskExists) {
                    targetColumn.tasks.push(processedTask);
                    targetColumn.totalCount = (targetColumn.totalCount || 0) + 1;

                    if (Array.isArray(targetColumn.order)) {
                        targetColumn.order.push({ id: processedTask.id });
                    }
                }
            }
        } catch (error) {
            console.error("Error adding task to Kanban view:", error);
        }
    }

    function kanbanColumnId(task) {
        const view = selectedView.value;

        if (!view || view.view_type !== "kanban") return null;
        try {
            const order = typeof view.order === "string" ? JSON.parse(view.order) : view.order;
            const value = order?.section ? task[order.section] : null;

            return value && typeof value === "object" ? value.id : value;
        } catch {
            return null;
        }
    }

    function placeKanbanCard(task, created) {
        const view = selectedView.value;

        if (!view || view.view_type !== "kanban") return;
        const target = String(kanbanColumnId(task) ?? "0");
        let found = false;

        for (const column of taskLists.value) {
            const idx = (column.tasks || []).findIndex((t) => String(t.id) === String(task.id));

            if (idx === -1) continue;
            if (String(column.id) === target) {
                column.tasks.splice(idx, 1, { ...column.tasks[idx], ...task });

                return;
            }

            column.tasks.splice(idx, 1);
            column.totalCount = Math.max(0, (column.totalCount || 0) - 1);
            if (Array.isArray(column.order)) {
                column.order = column.order.filter((o) => String(o.id) !== String(task.id));
            }

            found = true;
            break;
        }

        if (found || created) addTaskToKanbanView(task);
    }

    function removeKanbanCard(taskId) {
        for (const column of taskLists.value) {
            if (Array.isArray(column.tasks)) {
                const before = column.tasks.length;

                column.tasks = column.tasks.filter((t) => String(t.id) !== String(taskId));
                if (column.tasks.length < before) {
                    column.totalCount = Math.max(0, (column.totalCount || 0) - 1);
                }
            }

            if (Array.isArray(column.order)) {
                column.order = column.order.filter((o) => String(o.id) !== String(taskId));
            }
        }
    }

    // Only option tables are cached with all their rows. A task table's cached
    // copy has none, so a change to it is not added there: that would collect
    // every task changed during the session into a list that is no table's rows.
    function patchCachedOptions(tableId, patch) {
        const id = String(tableId);
        const table = workspaceTables.value.find((t) => String(t.id) === id);

        if (!table?.single_select || !Array.isArray(table.data_base)) return;
        const rows = patch(table.data_base);

        if (rows === table.data_base) return;
        workspaceTables.value = workspaceTables.value.map((t) =>
            String(t.id) === id ? { ...t, data_base: rows } : t,
        );
    }

    // The grid shows one page of tasks. Someone else's new task joins it only on
    // the first page, or when it is a subtask of a task the page shows; on any
    // other page it would not be where the next load puts it.
    function belongsOnShownPage(task) {
        const parent = task.parent_task_id ? String(task.parent_task_id) : "";

        if (parent && parent !== String(task.id)) {
            return tableData.value.some((t) => String(t.id) === parent);
        }

        return route.name !== "grid-view" || workspaceStore.gridPage === 1;
    }

    // The change this tab made comes back as its own and is skipped, so the task
    // from the New task dialog is placed here. A board places it itself, since it
    // shows a card's options by name and keeps where the card sits.
    function onTaskCreated({ item, headers, open }) {
        if (selectedView.value?.view_type === "kanban") {
            workspaceStore.announceCreatedTask(item);
        } else {
            applyTaskRow(route.params.tid, item, false, true);
        }

        if (open) {
            workspaceStore.setFullViewHeaders(headers || tableHeaders.value);
            workspaceStore.setTableID(route.params.tid);
            workspaceStore.setSelectedItem(item);
            workspaceStore.setFullTask(true);
        }
    }

    // The tab that made a change already shows it, so its own changes only
    // refresh the cached option tables. The grid holds one page, so an update
    // touches rows already shown, a new task is added, and one that has come to
    // match the filter is read with the page.
    function applyTaskRow(tableId, row, own, created) {
        const table = workspaceTables.value.find((t) => String(t.id) === String(tableId));
        const task = table?.headers?.length ? normalizeTaskRow(row, table.headers) : row;

        // An option table's status type and colour are not among its headers, so
        // they are kept from the row, or the cached options would keep old ones.
        for (const key of ["status_type", "color"]) {
            if (key in row && !(key in task)) task[key] = row[key];
        }

        // A teammate may assign someone this tab has not loaded yet.
        const userIds = (table?.headers || [])
            .filter(
                (h) =>
                    ["assignee", "default_assignee"].includes(h.header_usage) ||
                    h.name === "created_by",
            )
            .map((h) => task[h.name])
            .filter((id) => typeof id === "string" && id);

        if (userIds.length) userStore.ensureUsers(userIds);

        patchCachedOptions(tableId, (rows) => {
            const i = rows.findIndex((r) => String(r.id) === String(task.id));

            return i === -1
                ? rows.concat(task)
                : rows.map((r, j) => (j === i ? { ...r, ...task } : r));
        });
        if (own) return;

        const selected = getSelectedItem.value;

        if (String(tableId) !== String(route.params.tid)) {
            tableData.value = refreshEmbedded(
                tableData.value,
                tableHeaders.value,
                tableId,
                task,
                workspaceTables.value,
            );
            if (selected) {
                const [refreshed] = refreshEmbedded(
                    [selected],
                    tableHeaders.value,
                    tableId,
                    task,
                    workspaceTables.value,
                );

                if (refreshed !== selected) workspaceStore.setSelectedItem(refreshed);
            }

            return;
        }

        if (selected && String(selected.id) === String(task.id)) {
            workspaceStore.setSelectedItem({ ...selected, ...task });
        }

        applyShownTask(tableId, task, created);
    }

    const shownFilter = useShownFilter();
    const matchesShownFilter = (tableId, taskId) =>
        shownFilter.matchesShownFilter(route.params.id, tableId, taskId);

    // Changes to a task are applied once its latest filter check answers,
    // together, so an earlier answer cannot undo a later change.
    const pendingShownTasks = new Map();

    async function applyShownTask(tableId, changed, isNew) {
        const key = `${tableId}:${changed.id}`;
        const pending = pendingShownTasks.get(key);
        const entry = {
            task: { ...pending?.task, ...changed },
            created: Boolean(pending?.created || isNew),
        };

        pendingShownTasks.set(key, entry);

        const matches = await matchesShownFilter(tableId, changed.id);

        if (pendingShownTasks.get(key) !== entry) return;
        pendingShownTasks.delete(key);
        if (String(route.params.tid) !== String(tableId)) return;

        const { task, created } = entry;

        const i = tableData.value.findIndex((t) => String(t.id) === String(task.id));
        const merged = i === -1 ? task : { ...tableData.value[i], ...task };

        if (!matches) {
            if (i !== -1) tableData.value = tableData.value.filter((_, j) => j !== i);
            removeKanbanCard(task.id);

            return;
        }

        if (i !== -1) {
            tableData.value = tableData.value.map((t, j) => (j === i ? merged : t));
        } else if (created && belongsOnShownPage(merged)) {
            tableData.value = [...tableData.value, merged];
        } else if (!created && route.name === "grid-view" && belongsOnShownPage(merged)) {
            workspaceStore.bumpGridReloadToken();
        }

        placeKanbanCard(merged, created);
    }

    function removeTaskRow(tableId, taskId, isOption, own) {
        patchCachedOptions(tableId, (rows) => {
            const next = rows.filter((r) => String(r.id) !== String(taskId));

            return next.length === rows.length ? rows : next;
        });
        if (own) return;

        let rows = tableData.value;

        if (String(tableId) === String(route.params.tid)) {
            rows = rows.filter((t) => String(t.id) !== String(taskId));
            removeKanbanCard(taskId);
        }

        tableData.value = clearReferences(rows, taskId, { isOption });
    }

    return {
        addTaskToKanbanView,
        kanbanColumnId,
        placeKanbanCard,
        removeKanbanCard,
        patchCachedOptions,
        belongsOnShownPage,
        onTaskCreated,
        applyTaskRow,
        applyShownTask,
        removeTaskRow,
    };
}
