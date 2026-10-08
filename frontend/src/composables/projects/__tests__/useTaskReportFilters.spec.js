// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useTaskReportFilters } from "@/composables/projects/useTaskReportFilters";
import workspaceService from "@/services/workspaceService";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";

vi.mock("@/services/workspaceService", () => ({
    default: { getStatusList: vi.fn() },
}));

const statusList = {
    data: {
        flat_list: [
            { name: "Done", status_type: "Done" },
            { name: "Doing", status_type: "In progress" },
        ],
        name_to_ids: { Done: ["d1", "d2"], Doing: ["g1"] },
    },
};

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    useUserStore().user = { id: "me" };
});

describe("applyPreset", () => {
    it("leaves out every table's Done status by name for open tasks", async () => {
        workspaceService.getStatusList.mockResolvedValue(statusList);
        const store = useWorkspaceStore();

        expect(await useTaskReportFilters().applyPreset("open_tasks")).toBe(true);

        expect(store.getSavedFlatFilters).toEqual([
            {
                field: "status",
                operator: "is_not",
                value: "d1",
                values: ["d1", "d2"],
                operatorBetween: "AND",
            },
        ]);
        expect(store.fastFilterOption).toBe("open_tasks");
    });

    it("drops an answer that arrives after a later choice", async () => {
        let answer;

        workspaceService.getStatusList.mockReturnValue(
            new Promise((resolve) => (answer = resolve)),
        );
        const store = useWorkspaceStore();
        const { applyPreset } = useTaskReportFilters();

        const open = applyPreset("open_tasks");

        expect(await applyPreset("my_tasks")).toBe(true);
        answer(statusList);

        expect(await open).toBe(false);
        expect(store.fastFilterOption).toBe("my_tasks");
        expect(store.getSavedFlatFilters).toEqual([
            { field: "assignee", operator: "is", value: "me", operatorBetween: "AND" },
        ]);
    });

    it("drops a pending preset when a saved filter is chosen", async () => {
        let answer;

        workspaceService.getStatusList.mockReturnValue(
            new Promise((resolve) => (answer = resolve)),
        );
        const store = useWorkspaceStore();
        const { applyPreset, applySavedFilter } = useTaskReportFilters();

        const open = applyPreset("open_tasks");

        applySavedFilter({
            name: "Mine",
            filters: { flatFilters: [{ field: "name", value: "x" }], groups: [] },
        });
        answer(statusList);

        expect(await open).toBe(false);
        expect(store.fastFilterOption).toBe("saved_filter");
    });
});
