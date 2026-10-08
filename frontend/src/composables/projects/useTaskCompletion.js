// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

// The question TaskCompletionDialog shows, one at a time for the whole page.
const question = ref(null);
const statusTypes = new Map();

function ask(details) {
    return new Promise((resolve) => {
        question.value = { ...details, resolve };
    });
}

function answer(choice) {
    const current = question.value;

    question.value = null;
    current?.resolve(choice);
}

// A table's statuses: which close a task, and the first that marks one done.
function statusesOf(workspaceId, tableId) {
    if (!statusTypes.has(tableId)) {
        const loading = workspaceService
            .getTableStatusTypes(workspaceId, tableId)
            .then(({ data }) => ({
                closing: new Set((data?.done_ids || []).map(String)),
                done: (data?.options || []).find((option) => option.status_type === "Done") || null,
            }));

        loading.catch(() => statusTypes.delete(tableId));
        statusTypes.set(tableId, loading);
    }

    return statusTypes.get(tableId);
}

// Closing a task can close what belongs with it: its open subtasks, and its
// parent once nothing below the parent is open. Every place a status is
// changed asks here first and finishes here after saving.
export function useTaskCompletion() {
    const workspaceStore = useWorkspaceStore();

    // Resolves to null when the change is cancelled, or to what to do once it
    // is saved.
    async function beforeStatusChange({ workspaceId, tableId, taskId, statusId }) {
        let statuses;

        try {
            statuses = await statusesOf(workspaceId, tableId);
        } catch {
            return { afterSave: async () => {} };
        }

        if (!statuses.closing.has(String(statusId))) return { afterSave: async () => {} };

        const ids = { workspace_id: workspaceId, table_id: tableId };
        let choice = "only";

        try {
            const { data } = await workspaceService.getTaskCompletion({ ...ids, task_id: taskId });

            if (data?.open_subtasks > 0)
                choice = await ask({ kind: "subtasks", count: data.open_subtasks });
        } catch {
            choice = "only";
        }

        if (choice === "cancel") return null;

        return {
            afterSave: async () => {
                try {
                    if (choice === "all") {
                        await workspaceService.completeSubtasks({
                            ...ids,
                            task_id: taskId,
                            status: statusId,
                        });
                        workspaceStore.bumpGridReloadToken();
                    }

                    await offerParent(ids, taskId, statuses);
                } catch (error) {
                    useAlertStore().showError(extractErrorMessage(error));
                }
            },
        };
    }

    async function offerParent(ids, taskId, statuses) {
        if (!statuses.done) return;

        const { data } = await workspaceService.getTaskCompletion({ ...ids, task_id: taskId });
        const parent = data?.parent;

        if (!parent?.open || parent.open_subtasks > 0) return;

        if ((await ask({ kind: "parent", name: parent.name })) !== "complete") return;

        await workspaceService.updateTask({
            ...ids,
            task_id: parent.id,
            field: "status",
            value: statuses.done.id,
        });
        workspaceStore.bumpGridReloadToken();
        await offerParent(ids, parent.id, statuses);
    }

    return { question, answer, beforeStatusChange };
}
