// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { onBeforeUnmount, ref, toValue, watch } from "vue";
import workspaceService from "@/services/workspaceService";
import { onProjectChange, onReconnect } from "@/js/websocket";
import { useLatestRequest } from "@/composables/useLatestRequest";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const TASK_CHANGES = new Set(["task_created", "task_updated", "task_deleted"]);

// The subtasks of the task the panel shows, read from the server: each view
// loads a different part of the table, a page, a board or a date range, so
// none of them can be trusted to hold them. They are read again when one of
// them, or a new one, changes.
export function useTaskSubtasks(workspaceId, tableId, taskId) {
    const userStore = useUserStore();
    const subtasks = ref([]);
    const load = useLatestRequest();

    async function refresh() {
        const ids = {
            workspace_id: toValue(workspaceId),
            table_id: toValue(tableId),
            task_id: toValue(taskId),
        };

        if (!ids.workspace_id || !ids.table_id || !ids.task_id) {
            load.cancel();
            subtasks.value = [];

            return;
        }

        const request = load.start();

        try {
            const response = await workspaceService.getSubtasks(ids, request.signal);

            if (!request.isCurrent()) return;
            subtasks.value = Array.isArray(response.data) ? response.data : [];

            const assignees = subtasks.value.map((subtask) => subtask.assignee).filter(Boolean);

            if (assignees.length) userStore.ensureUsers(assignees);
        } catch (error) {
            if (request.isCurrent()) useAlertStore().showError(extractErrorMessage(error));
        }
    }

    function concernsShownTask(change) {
        if (!TASK_CHANGES.has(change?.type)) return false;
        if (String(change.table_id) !== String(toValue(tableId))) return false;

        const row = change.task || change.row;
        const id = String(row?.id ?? change.row_id ?? "");

        return (
            String(row?.parent_task_id ?? "") === String(toValue(taskId)) ||
            subtasks.value.some((subtask) => String(subtask.id) === id)
        );
    }

    const stopChanges = onProjectChange((change) => {
        if (concernsShownTask(change)) refresh();
    });
    const stopReconnect = onReconnect(refresh);

    onBeforeUnmount(() => {
        stopChanges();
        stopReconnect();
    });

    watch(() => [toValue(workspaceId), toValue(tableId), toValue(taskId)], refresh, {
        immediate: true,
    });

    return { subtasks };
}
