// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    filterApplies,
    filtersForChart,
    queryFilterFor,
    withFilterApplied,
    withFilterRemoved,
} from "@/utils/collimato/dashboardFilters.js";

const filter = (overrides = {}) => ({
    id: "f1",
    column: "orders.status",
    operator: "equals",
    values: ["new"],
    apply_to: [],
    ...overrides,
});

describe("filterApplies", () => {
    it("applies to every chart when no panels are named", () => {
        expect(filterApplies(filter(), "any")).toBe(true);
    });

    it("applies to a chart it names", () => {
        expect(filterApplies(filter({ apply_to: ["a", "b"] }), "b")).toBe(true);
    });

    it("does not apply to a chart it leaves out", () => {
        expect(filterApplies(filter({ apply_to: ["a"] }), "b")).toBe(false);
    });

    it("treats a missing scope as every chart", () => {
        expect(filterApplies({}, "a")).toBe(true);
    });
});

describe("filtersForChart", () => {
    it("keeps only the filters a chart is in scope for", () => {
        const all = [
            filter({ id: "everywhere" }),
            filter({ id: "mine", apply_to: ["chart1"] }),
            filter({ id: "theirs", apply_to: ["chart2"] }),
        ];

        expect(filtersForChart(all, "chart1").map((f) => f.id)).toEqual(["everywhere", "mine"]);
    });

    it("copes with no filters at all", () => {
        expect(filtersForChart(undefined, "a")).toEqual([]);
    });
});

describe("queryFilterFor", () => {
    it("carries the column, operator and values", () => {
        expect(queryFilterFor(filter(), ["new", "done"])).toEqual({
            member: "orders.status",
            operator: "equals",
            values: ["new", "done"],
        });
    });

    it("falls back to equals when a filter has no operator", () => {
        expect(queryFilterFor(filter({ operator: "" }), ["new"]).operator).toBe("equals");
    });
});

describe("withFilterApplied", () => {
    it("adds the filter when the query has none for that column", () => {
        const result = withFilterApplied([], filter(), ["new"]);

        expect(result).toEqual([{ member: "orders.status", operator: "equals", values: ["new"] }]);
    });

    it("replaces an existing filter on the same column", () => {
        const existing = [
            { member: "orders.status", operator: "equals", values: ["old"] },
            { member: "orders.region", operator: "equals", values: ["eu"] },
        ];

        const result = withFilterApplied(existing, filter(), ["new"]);

        expect(result).toHaveLength(2);
        expect(result.find((f) => f.member === "orders.status").values).toEqual(["new"]);
    });

    it("removes the filter when no values are chosen", () => {
        const existing = [{ member: "orders.status", operator: "equals", values: ["old"] }];

        expect(withFilterApplied(existing, filter(), [])).toEqual([]);
    });

    it("leaves other columns alone", () => {
        const existing = [{ member: "orders.region", operator: "equals", values: ["eu"] }];

        expect(withFilterApplied(existing, filter(), [])).toEqual(existing);
    });

    it("returns a new list rather than editing the one given", () => {
        const existing = [];

        withFilterApplied(existing, filter(), ["new"]);

        expect(existing).toEqual([]);
    });
});

describe("withFilterRemoved", () => {
    it("drops the filter's column", () => {
        const existing = [
            { member: "orders.status", operator: "equals", values: ["new"] },
            { member: "orders.region", operator: "equals", values: ["eu"] },
        ];

        expect(withFilterRemoved(existing, filter())).toEqual([existing[1]]);
    });

    it("does nothing when the column is not filtered", () => {
        expect(withFilterRemoved([], filter())).toEqual([]);
    });
});

describe("deleting a filter", () => {
    it("removes its constraint whatever values it now holds", () => {
        const existing = [{ member: "orders.status", operator: "equals", values: ["old"] }];

        expect(withFilterRemoved(existing, filter({ values: ["new"] }))).toEqual([]);
    });

    it("leaves a filter on a different column alone", () => {
        const existing = [
            { member: "orders.region", operator: "equals", values: ["eu"] },
            { member: "orders.status", operator: "equals", values: ["new"] },
        ];

        expect(withFilterRemoved(existing, filter())).toEqual([existing[0]]);
    });

    it("removes every entry on the column, not just the first", () => {
        const existing = [
            { member: "orders.status", operator: "equals", values: ["new"] },
            { member: "orders.status", operator: "contains", values: ["x"] },
        ];

        expect(withFilterRemoved(existing, filter())).toEqual([]);
    });
});
