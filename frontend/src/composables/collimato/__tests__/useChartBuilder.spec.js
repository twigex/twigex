// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { ref } from "vue";
import { useChartBuilder } from "@/composables/collimato/useChartBuilder.js";

const measure = (name) => ({ name });
const dimension = (name, type) => ({ name, type });

function setup(chartType = "bar") {
    return { builder: useChartBuilder(chartType) };
}

describe("useChartBuilder", () => {
    it("starts empty", () => {
        const { builder } = setup();

        expect(builder.measures.value).toEqual([]);
        expect(builder.dimensions.value).toEqual([]);
        expect(builder.selectedModel.value).toBeNull();
        expect(builder.query.value.measures).toEqual([]);
    });

    it("gives each instance its own state", () => {
        const a = setup().builder;
        const b = setup().builder;

        a.measures.value.push(measure("orders.count"));

        expect(a.query.value.measures).toEqual(["orders.count"]);
        expect(b.query.value.measures).toEqual([]);
    });

    it("derives the query from current state", () => {
        const { builder } = setup();

        builder.addMeasure(measure("orders.count"));
        builder.addDimension(dimension("orders.status"));
        builder.setOrder("orders.count", "desc");

        expect(builder.query.value).toMatchObject({
            measures: ["orders.count"],
            dimensions: ["orders.status"],
            order: [["orders.count", "desc"]],
        });
    });

    it("re-derives the query when a child mutates the array through a prop alias", () => {
        const { builder } = setup();

        const asProp = builder.measures.value;
        const childRef = ref(asProp);

        childRef.value.push(measure("orders.count"));

        expect(builder.measures.value).toHaveLength(1);
        expect(builder.query.value.measures).toEqual(["orders.count"]);

        childRef.value.splice(0, 1);

        expect(builder.query.value.measures).toEqual([]);
    });

    it("re-derives the query when the chart type changes", () => {
        const { builder } = setup("bar");

        builder.setTime({
            dimension: { name: "orders.created_at" },
            dateRange: "this month",
            granularity: "day",
        });

        expect(builder.query.value.timeDimensions).toEqual([]);

        builder.setChartType("line");

        expect(builder.query.value.timeDimensions).toHaveLength(1);
    });

    describe("setTime", () => {
        it("merges a patch and leaves the other keys alone", () => {
            const { builder } = setup("line");

            builder.setTime({
                dimension: { name: "orders.created_at" },
                dateRange: "this month",
                granularity: "day",
            });

            builder.setTime({ granularity: "week" });

            expect(builder.time.value).toEqual({
                dimension: { name: "orders.created_at" },
                dateRange: "this month",
                granularity: "week",
                startTime: "",
                endTime: "",
            });
        });

        it("feeds a named range straight into the query", () => {
            const { builder } = setup("line");

            builder.setTime({
                dimension: { name: "orders.created_at" },
                dateRange: "this month",
                granularity: "day",
            });

            expect(builder.query.value.timeDimensions[0].dateRange).toBe("this month");
        });

        it("builds the range pair only in the query, not in state", () => {
            const { builder } = setup("line");

            builder.setTime({
                dimension: { name: "orders.created_at" },
                dateRange: "custom",
                granularity: "day",
                startTime: "2026-01-01",
                endTime: "2026-01-31",
            });

            expect(builder.time.value.dateRange).toBe("custom");
            expect(builder.query.value.timeDimensions[0].dateRange).toEqual([
                "2026-01-01",
                "2026-01-31",
            ]);
        });

        it("clears back to an empty range", () => {
            const { builder } = setup("line");

            builder.setTime({
                dimension: { name: "orders.created_at" },
                dateRange: "this month",
                granularity: "day",
            });
            builder.setTime({
                dimension: null,
                dateRange: "",
                granularity: "",
                startTime: "",
                endTime: "",
            });

            expect(builder.query.value.timeDimensions).toEqual([]);
        });
    });

    describe("measures and dimensions", () => {
        it("tracks a new field in orders as unsorted", () => {
            const { builder } = setup();

            builder.addMeasure(measure("orders.count"));

            expect(builder.orders.value).toEqual([{ name: "orders.count", direction: "none" }]);
            expect(builder.query.value.order).toEqual([]);
        });

        it("refuses a duplicate and reports it", () => {
            const { builder } = setup();

            expect(builder.addMeasure(measure("orders.count"))).toBe(true);
            expect(builder.addMeasure(measure("orders.count"))).toBe(false);
            expect(builder.measures.value).toHaveLength(1);
            expect(builder.orders.value).toHaveLength(1);
        });

        it("refuses a duplicate dimension too", () => {
            const { builder } = setup();

            expect(builder.addDimension(dimension("orders.status"))).toBe(true);
            expect(builder.addDimension(dimension("orders.status"))).toBe(false);
            expect(builder.dimensions.value).toHaveLength(1);
        });

        it("drops the order entry when a field is removed", () => {
            const { builder } = setup();

            builder.addMeasure(measure("orders.count"));
            builder.addDimension(dimension("orders.status"));
            builder.setOrder("orders.count", "desc");

            builder.removeMeasure(measure("orders.count"));

            expect(builder.measures.value).toEqual([]);
            expect(builder.orders.value).toEqual([{ name: "orders.status", direction: "none" }]);
            expect(builder.query.value.order).toEqual([]);
        });

        it("never writes to the field objects it was given", () => {
            const { builder } = setup();
            const field = measure("orders.count");
            const snapshot = JSON.stringify(field);

            builder.addMeasure(field);
            builder.setOrder("orders.count", "desc");
            builder.reorderOrders([...builder.orders.value].reverse());

            expect(JSON.stringify(field)).toBe(snapshot);
        });
    });

    describe("orders", () => {
        it("applies a direction by name", () => {
            const { builder } = setup();

            builder.addMeasure(measure("orders.count"));
            builder.setOrder("orders.count", "asc");

            expect(builder.query.value.order).toEqual([["orders.count", "asc"]]);
        });

        it("ignores a direction for an untracked field", () => {
            const { builder } = setup();

            builder.setOrder("orders.missing", "asc");

            expect(builder.orders.value).toEqual([]);
        });

        it("keeps the sequence a reorder supplies", () => {
            const { builder } = setup();

            builder.addMeasure(measure("orders.count"));
            builder.addDimension(dimension("orders.status"));
            builder.setOrder("orders.count", "asc");
            builder.setOrder("orders.status", "desc");

            builder.reorderOrders([...builder.orders.value].reverse());

            expect(builder.query.value.order).toEqual([
                ["orders.status", "desc"],
                ["orders.count", "asc"],
            ]);
        });
    });

    describe("filters", () => {
        const filter = (member, value) => ({
            member,
            operator: "equals",
            value,
        });

        it("adds, replaces by member, and removes by identity", () => {
            const { builder } = setup();
            const first = filter("orders.status", ["new"]);

            builder.addFilter(first);
            expect(builder.query.value.filters[0].values).toEqual(["new"]);

            builder.updateFilter("orders.status", filter("orders.status", ["done"]));
            expect(builder.query.value.filters[0].values).toEqual(["done"]);
            expect(builder.filters.value).toHaveLength(1);

            builder.removeFilter(builder.filters.value[0]);
            expect(builder.filters.value).toEqual([]);
        });

        it("leaves the list alone when the member to replace is gone", () => {
            const { builder } = setup();

            builder.addFilter(filter("orders.status", ["new"]));

            builder.updateFilter("orders.missing", filter("orders.missing", ["x"]));

            expect(builder.filters.value).toHaveLength(1);
            expect(builder.filters.value[0].member).toBe("orders.status");
        });
    });

    describe("reset semantics", () => {
        it("clears the query when a different model is chosen", () => {
            const { builder } = setup("line");
            const model = {
                table: "orders",
                join: 0,
                measures: [],
                dimensions: [],
            };

            builder.dataModels.value = [model];

            builder.addMeasure(measure("orders.count"));
            builder.addFilter({
                member: "orders.status",
                operator: "equals",
                value: ["new"],
            });
            builder.setTime({
                dimension: { name: "orders.created_at" },
                dateRange: "this month",
                granularity: "day",
            });

            builder.selectModel(model);

            expect(builder.selectedModel.value.table).toBe("orders");
            expect(builder.query.value).toMatchObject({
                measures: [],
                dimensions: [],
                order: [],
                filters: [],
                timeDimensions: [],
            });
        });

        it("drops the appearance settings when the chart type changes", () => {
            const { builder } = setup("bar");

            builder.chartOptions.value = { legend: false, stack: true };

            builder.setChartType("pie");

            expect(builder.chartType.value).toBe("pie");
            expect(builder.chartOptions.value).toEqual({});
        });

        it("keeps the appearance settings when only the model changes", () => {
            const { builder } = setup("bar");

            builder.chartOptions.value = { legend: false };

            builder.selectModel({
                table: "orders",
                join: 0,
                measures: [],
                dimensions: [],
            });

            expect(builder.chartOptions.value).toEqual({ legend: false });
        });
    });

    describe("loadQuery", () => {
        function withModels(chartType = "line") {
            const { builder } = setup(chartType);

            builder.dataModels.value = [
                {
                    table: "orders",
                    join: 0,
                    measures: [measure("orders.count")],
                    dimensions: [
                        dimension("orders.status", "string"),
                        dimension("orders.created_at", "time"),
                    ],
                },
            ];

            return { builder };
        }

        it("restores a saved query and round-trips it", () => {
            const { builder } = withModels();
            const saved = {
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                order: [["orders.count", "desc"]],
                filters: [
                    {
                        member: "orders.status",
                        operator: "equals",
                        values: ["new"],
                    },
                ],
                timeDimensions: [
                    {
                        dimension: "orders.created_at",
                        dateRange: "this month",
                        granularity: "day",
                    },
                ],
            };

            builder.loadQuery(saved, "orders");

            expect(builder.query.value).toMatchObject({
                measures: saved.measures,
                dimensions: saved.dimensions,
                order: saved.order,
                filters: saved.filters,
                timeDimensions: saved.timeDimensions,
            });
        });

        it("tracks saved fields that carried no sort direction", () => {
            const { builder } = withModels();

            builder.loadQuery({
                measures: ["orders.count"],
                dimensions: [],
                order: [],
                filters: [],
                timeDimensions: [],
            });

            expect(builder.orders.value).toEqual([{ name: "orders.count", direction: "none" }]);
        });

        it("skips fields the models no longer expose", () => {
            const { builder } = withModels();

            builder.loadQuery({
                measures: ["orders.count", "orders.deleted"],
                dimensions: [],
                order: [],
                filters: [],
                timeDimensions: [],
            });

            expect(builder.query.value.measures).toEqual(["orders.count"]);
        });

        it("restores a custom range", () => {
            const { builder } = withModels();

            builder.loadQuery({
                measures: [],
                dimensions: [],
                order: [],
                filters: [],
                timeDimensions: [
                    {
                        dimension: "orders.created_at",
                        dateRange: ["2026-01-01", "2026-01-31"],
                        granularity: "week",
                    },
                ],
            });

            expect(builder.time.value.dateRange).toBe("custom");
            expect(builder.query.value.timeDimensions[0].dateRange).toEqual([
                "2026-01-01",
                "2026-01-31",
            ]);
        });
    });

    it("collects time dimensions across all models", () => {
        const { builder } = setup();

        builder.dataModels.value = [
            {
                table: "orders",
                join: 0,
                measures: [],
                dimensions: [
                    dimension("orders.created_at", "time"),
                    dimension("orders.status", "string"),
                ],
            },
            {
                table: "events",
                join: 0,
                measures: [],
                dimensions: [dimension("events.at", "time")],
            },
        ];

        expect(builder.timeDimensionFields.value.map((d) => d.name)).toEqual([
            "orders.created_at",
            "events.at",
        ]);
    });
});

describe("measure limit", () => {
    it("lets a bar chart take as many measures as it likes", () => {
        const { builder } = setup("bar");

        expect(builder.measureLimit.value).toBeNull();
        expect(builder.addMeasure(measure("orders.count"))).toBe(true);
        expect(builder.addMeasure(measure("orders.revenue"))).toBe(true);
        expect(builder.measures.value).toHaveLength(2);
    });

    it("stops a pie chart at one measure", () => {
        const { builder } = setup("pie");

        expect(builder.measureLimit.value).toBe(1);
        expect(builder.addMeasure(measure("orders.count"))).toBe(true);
        expect(builder.addMeasure(measure("orders.revenue"))).toBe(false);
        expect(builder.query.value.measures).toEqual(["orders.count"]);
    });

    it("drops the extra measures when switching to a chart that takes one", () => {
        const { builder } = setup("bar");

        builder.addMeasure(measure("orders.count"));
        builder.addMeasure(measure("orders.revenue"));
        builder.setChartType("pie");

        expect(builder.query.value.measures).toEqual(["orders.count"]);
    });

    it("drops the order of the measures it removed", () => {
        const { builder } = setup("bar");

        builder.addMeasure(measure("orders.count"));
        builder.addMeasure(measure("orders.revenue"));
        builder.setOrder("orders.revenue", "desc");
        builder.setChartType("pie");

        expect(builder.orders.value.map((o) => o.name)).toEqual(["orders.count"]);
        expect(builder.query.value.order).not.toHaveProperty("orders.revenue");
    });
});
