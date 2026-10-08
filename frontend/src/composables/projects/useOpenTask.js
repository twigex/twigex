// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

// Opens a task in the task panel, read from the server. The panel shows the
// fields the server sends with it unless headers, or a promise of them, are
// given. Resolves to the task, or to null once a failure has been shown.
export function useOpenTask() {
    const workspaceStore = useWorkspaceStore();

    async function openTask({ workspaceId, tableId, taskId, headers }) {
        try {
            const [response, given] = await Promise.all([
                workspaceService.getItemForTableByID({
                    workspace_id: workspaceId,
                    table_id: tableId,
                    task_id: taskId,
                }),
                headers,
            ]);

            workspaceStore.setSelectedItem(response.data.item);
            workspaceStore.setFullViewHeaders(given ?? response.data.headers);
            workspaceStore.setTableID(tableId);
            workspaceStore.setFullTask(true);

            return response.data.item;
        } catch (error) {
            useAlertStore().showError(extractErrorMessage(error));

            return null;
        }
    }

    return { openTask };
}
