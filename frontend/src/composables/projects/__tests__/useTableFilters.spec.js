// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach, vi } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import { useTableFilters } from "@/composables/projects/useTableFilters";
import workspaceService from "@/services/workspaceService";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";

vi.mock("@/services/workspaceService", () => ({
    default: { getTableStatusTypes: vi.fn() },
}));

beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
    useUserStore().user = { id: "me" };
    workspaceService.getTableStatusTypes.mockResolvedValue({
        data: { done_ids: ["done", "closed"] },
    });
});

describe("loadTablePreset", () => {
    it("leaves out the table's Done and Closed statuses for open tasks", async () => {
        const { loadTablePreset } = useTableFilters();

        const open = await loadTablePreset("open_tasks", "ws", "t1");

        expect(workspaceService.getTableStatusTypes).toHaveBeenCalledWith("ws", "t1");
        expect(open.flatFilters.map((f) => [f.operator, f.value])).toEqual([
            ["is_not", "done"],
            ["is_not", "closed"],
        ]);
        expect(open.active).toBe(true);
    });

    it("adds the overdue check for late tasks and needs no request for the others", async () => {
        const { loadTablePreset } = useTableFilters();

        const late = await loadTablePreset("late_tasks", "ws", "t1");

        expect(late.flatFilters.at(-1)).toMatchObject({ field: "due_date", value: "Overdue" });

        workspaceService.getTableStatusTypes.mockClear();
        expect(await loadTablePreset("my_tasks", "ws", "t1")).toMatchObject({
            active: true,
            flatFilters: [{ field: "assignee", value: "me" }],
        });
        expect(await loadTablePreset("all_tasks", "ws", "t1")).toMatchObject({
            active: false,
            flatFilters: [],
        });
        expect(workspaceService.getTableStatusTypes).not.toHaveBeenCalled();
    });

    it("is not active for open tasks when the table has no Done or Closed status", async () => {
        workspaceService.getTableStatusTypes.mockResolvedValue({ data: { done_ids: [] } });
        const { loadTablePreset } = useTableFilters();

        expect((await loadTablePreset("open_tasks", "ws", "t1")).active).toBe(false);
    });

    it("writes nothing to the store until it is stored", async () => {
        const store = useWorkspaceStore();
        const { loadTablePreset, storeTablePreset } = useTableFilters();
        const version = store.fastFilterVersion;

        const preset = await loadTablePreset("open_tasks", "ws", "t1");

        expect(store.fastFilterVersion).toBe(version);
        expect(store.getFilterActive).toBe(false);

        storeTablePreset(preset);
        expect(store.getFilterActive).toBe(true);
        expect(store.getDefaultFlatFilters).toEqual(preset.flatFilters);
        expect(store.fastFilterOption).toBe("open_tasks");
        expect(store.fastFilterVersion).toBe(version + 1);
    });
});
