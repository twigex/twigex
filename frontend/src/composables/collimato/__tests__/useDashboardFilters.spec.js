// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach } from "vitest";
import { ref } from "vue";
import { createPinia, setActivePinia } from "pinia";
import { useDashboardFilters } from "@/composables/collimato/useDashboardFilters.js";

vi.mock("@/services/collimatoService.js", () => ({
    default: {
        updateDashboardFilter: vi.fn().mockResolvedValue({ data: {} }),
    },
}));

vi.mock("vue-router", () => ({
    useRoute: () => ({ params: { workspaceId: "ws1", id: "d1" } }),
}));

describe("updateFilter", () => {
    beforeEach(() => {
        setActivePinia(createPinia());
    });

    it("re-runs only non-map charts", () => {
        const charts = ref([
            {
                id: "bar",
                chart_type: "bar",
                data: { query: { filters: [] } },
            },
            {
                id: "map",
                chart_type: "map",
                configuration: { layers: [] },
            },
        ]);

        const loadData = vi.fn();

        const { updateFilter } = useDashboardFilters({
            charts,
            pollLoad: vi.fn(),
            loadData,
        });

        updateFilter({
            id: "f1",
            filter: {
                column: "orders.status",
                operator: "equals",
                values: ["paid"],
                apply_to: [],
            },
        });

        expect(loadData).toHaveBeenCalledTimes(1);
        expect(loadData).toHaveBeenCalledWith(charts.value[0], 0);
        expect(charts.value[0].data.query.filters).toEqual([
            { member: "orders.status", operator: "equals", values: ["paid"] },
        ]);
    });
});
