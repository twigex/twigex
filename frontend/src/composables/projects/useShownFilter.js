// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";

// Whether a task still matches the filter shown, which the server decides as
// it did for the page: options and links arrive as objects, and date filters
// such as Today are worked out there. Without a filter, or when the check
// fails, the task is kept.
export function useShownFilter() {
    const workspaceStore = useWorkspaceStore();
    const userStore = useUserStore();

    async function matchesShownFilter(workspaceId, tableId, taskId) {
        const groups = workspaceStore.getSavedGroups || [];
        const flatFilters = workspaceStore.getSavedFlatFilters || [];

        if (!workspaceStore.getFilterActive || (groups.length === 0 && flatFilters.length === 0))
            return true;

        try {
            const res = await workspaceService.taskMatchesFilter(workspaceId, tableId, taskId, {
                groups,
                flatFilters,
                timezone: userStore.getTimezone,
            });

            return res.data?.matches !== false;
        } catch {
            return true;
        }
    }

    return { matchesShownFilter };
}
