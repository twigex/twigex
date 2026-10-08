// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData, drillDownFilters } from "@/utils/collimato/chartTypes.js";

const result = (dimensions, measures, rows, titles) => ({
    query: { measures, dimensions, timeDimensions: [] },
    annotation: {
        measures: Object.fromEntries(measures.map((m) => [m, { title: titles[m] }])),
    },
    data: rows,
});

describe("drillDownFilters", () => {
    it("reads one dimension off a bar label", () => {
        const query = {
            dimensions: ["w.weather"],
            measures: ["w.count"],
            timeDimensions: [],
        };

        expect(drillDownFilters("bar", {}, query, "rain")).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("splits a label built from two dimensions", () => {
        const query = {
            dimensions: ["w.location", "w.weather"],
            measures: ["w.count"],
            timeDimensions: [],
        };

        expect(drillDownFilters("bar", {}, query, "Seattle, rain")).toEqual([
            { member: "w.location", operator: "equals", values: ["Seattle"] },
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("drops the measure a series name carries", () => {
        const query = {
            dimensions: ["w.weather"],
            measures: ["w.count", "w.total"],
            timeDimensions: [{ dimension: "w.created", granularity: "day" }],
        };

        expect(drillDownFilters("line", {}, query, "rain, Count")).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("keeps an axis label whole, since it never carries the measure", () => {
        const query = {
            dimensions: ["w.weather"],
            measures: ["w.count", "w.total"],
            timeDimensions: [],
        };

        expect(drillDownFilters("bar", {}, query, "rain")).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("follows the chosen x-axis field", () => {
        const query = {
            dimensions: ["w.location", "w.weather"],
            measures: ["w.count"],
            timeDimensions: [],
        };
        const config = { xAxisField: "w.location" };

        expect(drillDownFilters("bar", config, query, "Seattle")).toEqual([
            { member: "w.location", operator: "equals", values: ["Seattle"] },
        ]);
    });

    it("uses the series dimensions for a line chart", () => {
        const query = {
            dimensions: ["w.created", "w.weather"],
            measures: ["w.count"],
            timeDimensions: [{ dimension: "w.created", granularity: "day" }],
        };

        expect(drillDownFilters("line", {}, query, "rain")).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("uses only the first dimension for a pie", () => {
        const query = {
            dimensions: ["w.weather", "w.location"],
            measures: ["w.count"],
            timeDimensions: [],
        };

        expect(drillDownFilters("pie", {}, query, "rain")).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("returns nothing when there are no dimensions to drill on", () => {
        const query = {
            dimensions: [],
            measures: ["w.count"],
            timeDimensions: [],
        };

        expect(drillDownFilters("bar", {}, query, "Count")).toEqual([]);
    });
});

describe("drillDownFilters against real labels", () => {
    it("matches the label a bar chart actually renders", () => {
        const data = result(
            ["w.location", "w.weather"],
            ["w.count"],
            [
                { "w.location": "Seattle", "w.weather": "rain", "w.count": 5 },
                { "w.location": "Oslo", "w.weather": "sun", "w.count": 2 },
            ],
            { "w.count": "Count" },
        );

        const label = transformData("bar", data, {}).xAxis.data[0];

        expect(drillDownFilters("bar", {}, data.query, label)).toEqual([
            { member: "w.location", operator: "equals", values: ["Seattle"] },
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("matches a legend entry when a dimension is the breakdown", () => {
        const data = result(
            ["w.location", "w.weather"],
            ["w.count"],
            [{ "w.location": "Seattle", "w.weather": "rain", "w.count": 5 }],
            { "w.count": "Count" },
        );
        const config = { xAxisField: "w.location" };

        const label = transformData("bar", data, config).xAxis.data[0];

        expect(drillDownFilters("bar", config, data.query, label)).toEqual([
            { member: "w.location", operator: "equals", values: ["Seattle"] },
        ]);
    });

    it("matches a label carrying the measure name", () => {
        const data = result(
            ["w.weather"],
            ["w.count", "w.total"],
            [{ "w.weather": "rain", "w.count": 5, "w.total": 9 }],
            { "w.count": "Count", "w.total": "Total" },
        );

        const label = transformData("bar", data, {}).xAxis.data[0];

        expect(label).toBe("rain");
        expect(drillDownFilters("bar", {}, data.query, label)).toEqual([
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });
});
