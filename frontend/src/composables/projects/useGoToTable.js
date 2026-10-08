// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";

// Opens a table of the workspace shown at its main grid view, and reports
// whether it did: a table that is not among the workspace's, or has no grid
// view yet, is not opened.
export function useGoToTable() {
    const route = useRoute();
    const router = useRouter();
    const workspaceStore = useWorkspaceStore();

    function goToTable(tableId) {
        const tables = workspaceStore.getWorkspaceTables;

        if (!Array.isArray(tables)) return false;

        const fid = tables.find((table) => table.id === tableId)?.options?.[0]?.id;

        if (!fid) return false;

        router.push({
            name: "grid-view",
            params: { id: route.params.id, tid: tableId, fid },
        });

        return true;
    }

    return { goToTable };
}
