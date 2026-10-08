// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import {
    parseViewSort,
    pickView,
    tableDisplayNames,
    viewFilterState,
    viewHeaders,
} from "../gridView";

const table = {
    headers: [
        { name: "name", display_name: "Name" },
        { name: "c1", display_name: "Budget" },
        { name: "c2", display_name: "Owner" },
    ],
    options: [{ id: "v1" }, { id: "v2" }],
    views: [{ id: "v3" }],
};

describe("pickView", () => {
    it("finds the view named among the table's views", () => {
        expect(pickView(table, "v2").id).toBe("v2");
        expect(pickView(table, "v3").id).toBe("v3");
    });

    it("falls back to the table's first view", () => {
        expect(pickView(table, "gone").id).toBe("v1");
        expect(pickView({}, "v1")).toBeNull();
    });
});

describe("viewHeaders", () => {
    it("names, shows and sizes the fields as the view saved them", () => {
        const order = [
            { name: "c1", display_name: "Cost", visible: false, width: "220" },
            { name: "name", visible: false },
        ];
        const { headers, widths } = viewHeaders(table, { order: JSON.stringify(order) });

        expect(headers.map((h) => [h.name, h.display_name, h.visible])).toEqual([
            ["name", "Name", true],
            ["c1", "Cost", false],
            ["c2", "Owner", true],
        ]);
        expect(widths).toEqual([150, 220, 150]);
    });

    it("puts the fields in the view's saved order", () => {
        const order = [{ name: "headerOrder", value: ["c2", "name", "gone", "c1"] }];
        const { headers } = viewHeaders(table, { order: JSON.stringify(order) });

        expect(headers.map((h) => h.name)).toEqual(["c2", "name", "c1"]);
    });

    it("keeps the headers as they are when the table sent none", () => {
        expect(viewHeaders({}, null)).toEqual({ headers: null, widths: [] });
    });
});

describe("tableDisplayNames", () => {
    it("uses a table's display name when it has one", () => {
        const names = tableDisplayNames([
            { id: "t1", name: "t1abc", display_name: { Valid: true, String: "Tasks" } },
            { id: "t2", name: "Bugs", display_name: { Valid: true, String: "  " } },
            { id: "t3", name: "Ideas" },
        ]);

        expect(names).toEqual({ t1: "Tasks", t2: "Bugs", t3: "Ideas" });
    });
});

describe("viewFilterState", () => {
    const flat = [{ field: "status", value: "open" }];

    it("is empty for a view without a filter", () => {
        expect(viewFilterState([], [], [])).toBeNull();
    });

    it("opens with the active saved filter that matches the view's", () => {
        const saved = { is_active: true, filters: { groups: [], flatFilters: flat } };
        const state = viewFilterState([saved], [], flat);

        expect(state.filter).toBe(saved);
        expect(state.built).toBeNull();
    });

    it("otherwise builds the view's own filter with options for every row", () => {
        const state = viewFilterState([], [{ filters: [{ field: "x" }] }], flat);

        expect(state.built.flatFilters[0].linkedOptions).toEqual([]);
        expect(state.built.groups[0].filters[0].linkedOptions).toEqual([]);
        expect(state.filter.filters).toBe(state.built);
    });
});

describe("parseViewSort", () => {
    it("reads a sort saved as text or as a list", () => {
        const sort = [{ field: "name", direction: "asc" }];

        expect(parseViewSort(JSON.stringify(sort))).toEqual(sort);
        expect(parseViewSort(sort)).toEqual(sort);
    });

    it("reads nothing from a missing or broken sort", () => {
        expect(parseViewSort(undefined)).toEqual([]);
        expect(parseViewSort("{broken")).toEqual([]);
        expect(parseViewSort('{"field":"name"}')).toEqual([]);
    });
});
