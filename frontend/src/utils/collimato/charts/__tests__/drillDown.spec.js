// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData, applyConfiguration } from "@/utils/collimato/charts/chartDrawing.js";
import {
    drillDownValuesFor,
    drillDownFilters,
    filterMembers,
    filterValues,
    drillDownBreadcrumbs,
} from "@/utils/collimato/charts/drillDown.js";

const result = (dimensions, measures, rows, titles) => ({
    query: { measures, dimensions, timeDimensions: [] },
    annotation: {
        measures: Object.fromEntries(measures.map((m) => [m, { title: titles[m] }])),
    },
    data: rows,
});

describe("drillDownFilters", () => {
    it("builds one filter per member", () => {
        const clicked = { "w.location": ["Seattle"], "w.weather": ["rain"] };

        expect(drillDownFilters(clicked)).toEqual([
            { member: "w.location", operator: "equals", values: ["Seattle"] },
            { member: "w.weather", operator: "equals", values: ["rain"] },
        ]);
    });

    it("keeps a value containing the label separator whole", () => {
        expect(drillDownFilters({ "w.owner": ["Jane, Smith"] })).toEqual([
            { member: "w.owner", operator: "equals", values: ["Jane, Smith"] },
        ]);
    });

    it("sends values as text", () => {
        expect(drillDownFilters({ "w.year": [2024] })).toEqual([
            { member: "w.year", operator: "equals", values: ["2024"] },
        ]);
    });

    it("asks for unset rows when the value was NULL", () => {
        expect(drillDownFilters({ "w.weather": [null] })).toEqual([
            { member: "w.weather", operator: "notSet" },
        ]);
    });

    it("asks for the empty string when that is what the rows hold", () => {
        expect(drillDownFilters({ "w.weather": [""] })).toEqual([
            { member: "w.weather", operator: "equals", values: [""] },
        ]);
    });

    it("asks for both when NULL and the empty string share a label", () => {
        expect(drillDownFilters({ "w.weather": [null, ""] })).toEqual([
            {
                or: [
                    { member: "w.weather", operator: "notSet" },
                    { member: "w.weather", operator: "equals", values: [""] },
                ],
            },
        ]);
    });

    it("returns nothing when the click named no values", () => {
        expect(drillDownFilters({})).toEqual([]);
        expect(drillDownFilters(null)).toEqual([]);
    });
});

describe("drillDownValuesFor against real labels", () => {
    it("finds every axis dimension behind a bar label", () => {
        const data = result(
            ["w.owner", "w.weather"],
            ["w.count"],
            [
                { "w.owner": "Jane, Smith", "w.weather": "rain", "w.count": 5 },
                { "w.owner": "Oslo", "w.weather": "sun", "w.count": 2 },
            ],
            { "w.count": "Count" },
        );

        const option = transformData("bar", data, {});
        const name = option.xAxis.data[0];

        expect(name).toBe("Jane, Smith, rain");
        expect(drillDownValuesFor("bar", option, { name })).toEqual({
            "w.owner": ["Jane, Smith"],
            "w.weather": ["rain"],
        });
    });

    it("follows the chosen x-axis field", () => {
        const data = result(
            ["w.location", "w.weather"],
            ["w.count"],
            [{ "w.location": "Seattle", "w.weather": "rain", "w.count": 5 }],
            { "w.count": "Count" },
        );
        const config = { xAxisField: "w.location" };

        const option = transformData("bar", data, config);

        expect(drillDownValuesFor("bar", option, { name: "Seattle" })).toEqual({
            "w.location": ["Seattle"],
        });
    });

    it("uses the series dimensions for a line chart", () => {
        const data = result(
            ["w.location", "w.weather"],
            ["w.count"],
            [{ "w.location": "Seattle", "w.weather": "rain", "w.count": 5 }],
            { "w.count": "Count" },
        );
        const config = { xAxisField: "w.location" };

        const option = transformData("line", data, config);

        expect(drillDownValuesFor("line", option, { seriesName: "rain" })).toEqual({
            "w.weather": ["rain"],
        });
    });

    it("leaves out the time dimension and the measure a series name carries", () => {
        const data = {
            query: {
                measures: ["w.count", "w.total"],
                dimensions: ["w.created", "w.weather"],
                timeDimensions: [
                    {
                        dimension: "w.created",
                        granularity: "day",
                        dateRange: ["2026-01-01", "2026-01-02"],
                    },
                ],
            },
            annotation: {
                measures: { "w.count": { title: "Count" }, "w.total": { title: "Total" } },
            },
            data: [
                {
                    "w.created.day": "2026-01-01T00:00:00.000",
                    "w.weather": "rain",
                    "w.count": 5,
                    "w.total": 9,
                },
            ],
        };

        const option = transformData("time_bar", data, {});
        const seriesName = option.series[0].name;

        expect(seriesName).toBe("rain, Count");
        expect(drillDownValuesFor("time_bar", option, { seriesName })).toEqual({
            "w.weather": ["rain"],
        });
    });

    it("uses only the first dimension for a pie", () => {
        const data = result(
            ["w.weather", "w.location"],
            ["w.count"],
            [{ "w.weather": "rain", "w.location": "Seattle", "w.count": 5 }],
            { "w.count": "Count" },
        );

        const option = transformData("pie", data, {});

        expect(drillDownValuesFor("pie", option, { name: "rain" })).toEqual({
            "w.weather": ["rain"],
        });
    });

    it("keeps NULL and the empty string apart behind their shared label", () => {
        const data = result(
            ["w.weather"],
            ["w.count"],
            [
                { "w.weather": null, "w.count": 5 },
                { "w.weather": "", "w.count": 2 },
            ],
            { "w.count": "Count" },
        );

        const option = transformData("bar", data, {});

        expect(option.xAxis.data).toHaveLength(1);
        expect(drillDownValuesFor("bar", option, { name: option.xAxis.data[0] })).toEqual({
            "w.weather": [null, ""],
        });
    });

    it("still finds the values once the chart is configured and sorted", () => {
        const data = result(
            ["w.weather"],
            ["w.count"],
            [
                { "w.weather": "rain", "w.count": 5 },
                { "w.weather": "sun", "w.count": 2 },
            ],
            { "w.count": "Count" },
        );
        const config = { sortBars: "asc" };

        const option = applyConfiguration("bar", transformData("bar", data, config), config);

        expect(option.xAxis.data).toEqual(["sun", "rain"]);
        expect(drillDownValuesFor("bar", option, { name: "rain" })).toEqual({
            "w.weather": ["rain"],
        });
    });

    it("returns nothing for a label the chart does not have", () => {
        const data = result(["w.weather"], ["w.count"], [{ "w.weather": "rain", "w.count": 5 }], {
            "w.count": "Count",
        });

        const option = transformData("bar", data, {});

        expect(drillDownValuesFor("bar", option, { name: "constructor" })).toEqual({});
        expect(drillDownValuesFor("bar", null, { name: "rain" })).toEqual({});
    });
});

describe("filterMembers and filterValues", () => {
    const nested = {
        or: [
            { member: "w.weather", operator: "notSet" },
            { member: "w.weather", operator: "equals", values: [""] },
        ],
    };

    it("reads a plain filter", () => {
        const plain = { member: "w.weather", operator: "equals", values: ["rain"] };

        expect(filterMembers(plain)).toEqual(["w.weather"]);
        expect(filterValues(plain)).toEqual(["rain"]);
    });

    it("reads through a nested filter", () => {
        expect(filterMembers(nested)).toEqual(["w.weather", "w.weather"]);
        expect(filterValues(nested)).toEqual([null, ""]);
    });
});

describe("drill down breadcrumbs", () => {
    const filter = (member, value) => ({
        member,
        operator: "equals",
        values: [value],
    });

    function breadcrumbsFor(trail, chartName = "Chart") {
        const crumbs = drillDownBreadcrumbs(trail, chartName);

        return crumbs.map((crumb) => ({ name: crumb.name, value: crumb.value }));
    }

    it("shows no value on the first level", () => {
        const crumbs = breadcrumbsFor([{ dimensions: ["w.weather"], filters: [] }]);

        expect(crumbs[0]).toEqual({ name: "weather", value: "" });
    });

    it("shows the value that led to each level", () => {
        const crumbs = breadcrumbsFor([
            { dimensions: ["w.weather"], filters: [] },
            {
                dimensions: ["w.borough"],
                filters: [filter("w.weather", "rain")],
            },
            {
                dimensions: ["w.payment"],
                filters: [filter("w.weather", "rain"), filter("w.borough", "Manhattan")],
            },
        ]);

        expect(crumbs).toEqual([
            { name: "weather", value: "" },
            { name: "borough", value: "rain" },
            { name: "payment", value: "Manhattan" },
        ]);
    });

    it("keeps filters the chart already had off the trail", () => {
        const existing = filter("w.year", "2024");

        const crumbs = breadcrumbsFor([
            { dimensions: ["w.weather"], filters: [existing] },
            {
                dimensions: ["w.borough"],
                filters: [existing, filter("w.weather", "rain")],
            },
        ]);

        expect(crumbs[1].value).toBe("rain");
    });

    it("joins several values added in one step", () => {
        const crumbs = breadcrumbsFor([
            { dimensions: ["w.a", "w.b"], filters: [] },
            {
                dimensions: ["w.c"],
                filters: [filter("w.a", "one"), filter("w.b", "two")],
            },
        ]);

        expect(crumbs[1].value).toBe("one, two");
    });

    it("names a level split by several dimensions", () => {
        const crumbs = breadcrumbsFor([{ dimensions: ["w.location", "w.weather"], filters: [] }]);

        expect(crumbs[0].name).toBe("location, weather");
    });

    it("names a level without dimensions after the chart", () => {
        const crumbs = breadcrumbsFor([{ dimensions: [], filters: [] }], "Sales by month");

        expect(crumbs[0].name).toBe("Sales by month");
    });

    it("shows (empty) for a value that was NULL", () => {
        const crumbs = breadcrumbsFor([
            { dimensions: ["w.product"], filters: [] },
            {
                dimensions: ["w.region"],
                filters: [{ member: "w.product", operator: "notSet" }],
            },
        ]);

        expect(crumbs[1].value).toBe("(empty)");
    });

    it("shows (empty) once when NULL and an empty string were both clicked", () => {
        const both = {
            or: [
                { member: "w.product", operator: "notSet" },
                { member: "w.product", operator: "equals", values: [""] },
            ],
        };

        const crumbs = breadcrumbsFor([
            { dimensions: ["w.product"], filters: [] },
            { dimensions: ["w.region"], filters: [both] },
        ]);

        expect(crumbs[1].value).toBe("(empty)");
    });
});
