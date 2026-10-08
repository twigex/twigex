// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { buildQuery, joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";

const measure = (name) => ({ name });
const dimension = (name) => ({ name });

describe("buildQuery", () => {
    it("maps selected fields to their names", () => {
        const query = buildQuery({
            measures: [measure("orders.count")],
            dimensions: [dimension("orders.status")],
        });

        expect(query.measures).toEqual(["orders.count"]);
        expect(query.dimensions).toEqual(["orders.status"]);
        expect(query.limit).toBe(MAX_QUERY_LIMIT);
    });

    it("omits fields whose order is none, keeping sequence and direction", () => {
        const query = buildQuery({
            orders: [
                { name: "orders.created_at", direction: "desc" },
                { name: "orders.count", direction: "none" },
                { name: "orders.status", direction: "asc" },
            ],
        });

        expect(query.order).toEqual([
            ["orders.created_at", "desc"],
            ["orders.status", "asc"],
        ]);
    });

    describe("filters", () => {
        it("wraps a scalar value in an array", () => {
            const query = buildQuery({
                filters: [{ member: "orders.status", operator: "equals", value: "new" }],
            });

            expect(query.filters).toEqual([
                {
                    member: "orders.status",
                    operator: "equals",
                    values: ["new"],
                },
            ]);
        });

        it("merges entries sharing a member and operator", () => {
            const query = buildQuery({
                filters: [
                    { member: "orders.status", operator: "equals", value: "new" },
                    { member: "orders.status", operator: "equals", value: "done" },
                    { member: "orders.status", operator: "not_equals", value: "void" },
                ],
            });

            expect(query.filters).toEqual([
                {
                    member: "orders.status",
                    operator: "equals",
                    values: ["new", "done"],
                },
                {
                    member: "orders.status",
                    operator: "not_equals",
                    values: ["void"],
                },
            ]);
        });

        it("does not mutate the source filters when merging", () => {
            const filters = [
                { member: "orders.status", operator: "equals", value: ["new"] },
                { member: "orders.status", operator: "equals", value: ["done"] },
            ];

            buildQuery({ filters });
            const second = buildQuery({ filters });

            expect(filters[0].value).toEqual(["new"]);
            expect(second.filters[0].values).toEqual(["new", "done"]);
        });
    });

    describe("time dimensions", () => {
        const time = {
            dimension: { name: "orders.created_at" },
            dateRange: "this month",
            granularity: "day",
        };

        it.each(["line", "time_bar"])("attaches the range for %s", (chartType) => {
            const query = buildQuery({ time, chartType });

            expect(query.timeDimensions).toEqual([
                {
                    dimension: "orders.created_at",
                    dateRange: "this month",
                    granularity: "day",
                },
            ]);
        });

        it.each(["bar", "pie", "big_number"])("ignores the range for %s", (chartType) => {
            expect(buildQuery({ time, chartType }).timeDimensions).toEqual([]);
        });

        it("expands a custom range into a start and end pair", () => {
            const query = buildQuery({
                chartType: "line",
                time: {
                    ...time,
                    dateRange: "custom",
                    startTime: "2026-01-01",
                    endTime: "2026-01-31",
                },
            });

            expect(query.timeDimensions[0].dateRange).toEqual(["2026-01-01", "2026-01-31"]);
        });

        it("sends a null granularity when none is chosen", () => {
            const query = buildQuery({
                chartType: "line",
                time: { ...time, granularity: "" },
            });

            expect(query.timeDimensions[0].granularity).toBeNull();
        });

        it("attaches nothing when no time dimension is chosen", () => {
            const query = buildQuery({
                chartType: "line",
                time: { dimension: null, dateRange: "this month" },
            });

            expect(query.timeDimensions).toEqual([]);
        });
    });
});

describe("joinableDataModels", () => {
    const orders = { table: "orders", join: 1 };
    const customers = { table: "customers", join: 1 };
    const events = { table: "events", join: 2 };
    const standalone = { table: "standalone", join: 0 };
    const all = [orders, customers, events, standalone];

    it("returns nothing when no model is selected", () => {
        expect(joinableDataModels(all, null)).toEqual([]);
    });

    it("returns the selected model plus others in its join group", () => {
        expect(joinableDataModels(all, orders)).toEqual([orders, customers]);
    });

    it("returns only the selected model when it joins nothing", () => {
        expect(joinableDataModels(all, standalone)).toEqual([standalone]);
    });

    it("puts the selected model first and never repeats it", () => {
        const result = joinableDataModels(all, customers);

        expect(result[0]).toBe(customers);
        expect(result.filter((m) => m.table === "customers")).toHaveLength(1);
    });
});
