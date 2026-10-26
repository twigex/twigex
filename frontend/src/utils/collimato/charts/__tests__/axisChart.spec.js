// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData } from "@/utils/collimato/charts/chartDrawing.js";
import { cubeResult } from "@/utils/collimato/charts/__tests__/cubeResult.js";
import { plainSpaces } from "@/utils/collimato/charts/__tests__/plainSpaces.js";

const weather = {
    query: {
        measures: ["weather.count"],
        dimensions: ["weather.weather"],
        timeDimensions: [],
    },
    annotation: { measures: { "weather.count": { title: "Count" } } },
    data: [
        { "weather.weather": "rain", "weather.count": 5 },
        { "weather.weather": "sun", "weather.count": 3 },
        { "weather.weather": "fog", "weather.count": 1 },
    ],
};

describe("barChart", () => {
    it("treats a time dimension without granularity as a plain dimension", () => {
        const option = transformData(
            "bar",
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-03T00:00:00.000"],
                    },
                ],
                rows: [
                    { "orders.status": "new", "orders.count": 5 },
                    { "orders.status": "done", "orders.count": 3 },
                ],
            }),
        );

        expect(option.xAxis.data).toEqual(["new", "done"]);
        expect(option.series[0].data).toEqual([5, 3]);
    });

    it("buckets rows onto the time axis when a granularity is set", () => {
        const option = transformData(
            "bar",
            cubeResult({
                measures: ["orders.count"],
                dimensions: [],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        granularity: "day",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-03T00:00:00.000"],
                    },
                ],
                rows: [
                    {
                        "orders.created.day": "2026-01-02T00:00:00.000",
                        "orders.count": 7,
                    },
                ],
            }),
        );

        expect(option.xAxis.data).toEqual(["1 Jan 2026", "2 Jan 2026", "3 Jan 2026"]);
        expect(option.series[0].data).toEqual([null, 7, null]);
    });
});

describe("axis charts with several measures", () => {
    const twoMeasures = () =>
        cubeResult({
            measures: ["orders.count", "orders.revenue"],
            dimensions: ["orders.status"],
            rows: [
                {
                    "orders.status": "new",
                    "orders.count": 5,
                    "orders.revenue": 90,
                },
                {
                    "orders.status": "done",
                    "orders.count": 3,
                    "orders.revenue": 40,
                },
            ],
            titles: { "orders.count": "Count", "orders.revenue": "Revenue" },
        });

    it.each(["line", "time_bar", "bar"])("%s renders every measure", (type) => {
        const option = transformData(type, twoMeasures());

        expect(option.series).toHaveLength(2);
        expect(option.series.map((s) => s.data)).toEqual([
            [5, 3],
            [90, 40],
        ]);
    });

    it("puts the dimension values on the axis, one series per measure", () => {
        const option = transformData("line", twoMeasures());

        expect(option.xAxis.data).toEqual(["new", "done"]);
        expect(option.legend.data).toEqual(["Count", "Revenue"]);
    });

    it("names series by dimension alone when there is one measure", () => {
        const option = transformData("line", {
            ...twoMeasures(),
            query: {
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        granularity: "day",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-01T00:00:00.000"],
                    },
                ],
            },
            data: [
                {
                    "orders.created.day": "2026-01-01T00:00:00.000",
                    "orders.status": "new",
                    "orders.count": 5,
                },
            ],
        });

        expect(option.legend.data).toEqual(["new"]);
    });

    it("appends the measure name when there is more than one", () => {
        const option = transformData("line", {
            ...twoMeasures(),
            query: {
                measures: ["orders.count", "orders.revenue"],
                dimensions: ["orders.status"],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        granularity: "day",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-01T00:00:00.000"],
                    },
                ],
            },
            data: [
                {
                    "orders.created.day": "2026-01-01T00:00:00.000",
                    "orders.status": "new",
                    "orders.count": 5,
                    "orders.revenue": 90,
                },
            ],
        });

        expect(option.legend.data).toEqual(["new, Count", "new, Revenue"]);
    });
});

describe("axis charts without a time dimension", () => {
    it("makes one series per measure, not one per row", () => {
        const option = transformData(
            "line",
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                rows: [
                    { "orders.status": "a", "orders.count": 1 },
                    { "orders.status": "b", "orders.count": 2 },
                    { "orders.status": "c", "orders.count": 3 },
                ],
            }),
        );

        expect(option.series).toHaveLength(1);
        expect(option.series[0].data).toEqual([1, 2, 3]);
        expect(option.xAxis.data).toEqual(["a", "b", "c"]);
    });
});

describe("sub-day granularity", () => {
    const hourly = (type) =>
        transformData(type, {
            query: {
                measures: ["orders.count"],
                dimensions: [],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        granularity: "hour",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-01T02:00:00.000"],
                    },
                ],
            },
            annotation: { measures: { "orders.count": { title: "Count" } } },
            data: [
                {
                    "orders.created.hour": "2026-01-01T00:00:00.000",
                    "orders.count": 1,
                },
                {
                    "orders.created.hour": "2026-01-01T01:00:00.000",
                    "orders.count": 2,
                },
                {
                    "orders.created.hour": "2026-01-01T02:00:00.000",
                    "orders.count": 4,
                },
            ],
        });

    it.each(["line", "time_bar"])("%s keeps each hour separate", (type) => {
        const option = hourly(type);

        expect(option.xAxis.data).toHaveLength(3);
        expect(option.series[0].data).toEqual([1, 2, 4]);
    });
});

describe("series shape", () => {
    it.each(["line", "bar", "time_bar"])("%s gives every series a label object", (type) => {
        const option = transformData(
            type,
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                timeDimensions: [
                    {
                        dimension: "orders.created",
                        granularity: "day",
                        dateRange: ["2026-01-01T00:00:00.000", "2026-01-02T00:00:00.000"],
                    },
                ],
                rows: [
                    {
                        "orders.created.day": "2026-01-01T00:00:00.000",
                        "orders.status": "new",
                        "orders.count": 5,
                    },
                ],
            }),
        );

        option.series.forEach((s) => expect(s.label).toBeDefined());
    });
});

describe("x-axis field unset", () => {
    it("keeps every dimension on the axis, as before", () => {
        const option = transformData("bar", weather);

        expect(option.xAxis.data).toEqual(["rain", "sun", "fog"]);
        expect(option.legend.data).toEqual(["Count"]);
        expect(option.series).toHaveLength(1);
        expect(option.series[0].data).toEqual([5, 3, 1]);
    });

    it("ignores a field that is not one of the dimensions", () => {
        const option = transformData("bar", weather, {
            xAxisField: "weather.nonsense",
        });

        expect(option.xAxis.data).toEqual(["rain", "sun", "fog"]);
        expect(option.legend.data).toEqual(["Count"]);
    });
});

describe("dimension as the series breakdown", () => {
    const twoDimensions = {
        query: {
            measures: ["weather.count"],
            dimensions: ["weather.location", "weather.weather"],
            timeDimensions: [],
        },
        annotation: { measures: { "weather.count": { title: "Count" } } },
        data: [
            {
                "weather.location": "Seattle",
                "weather.weather": "rain",
                "weather.count": 5,
            },
            {
                "weather.location": "Seattle",
                "weather.weather": "sun",
                "weather.count": 2,
            },
            {
                "weather.location": "Oslo",
                "weather.weather": "rain",
                "weather.count": 4,
            },
            {
                "weather.location": "Oslo",
                "weather.weather": "sun",
                "weather.count": 1,
            },
        ],
    };

    it("puts the chosen field on the axis and the rest in the legend", () => {
        const option = transformData("bar", twoDimensions, {
            xAxisField: "weather.location",
        });

        expect(option.xAxis.data).toEqual(["Seattle", "Oslo"]);
        expect(option.legend.data).toEqual(["rain", "sun"]);
    });

    it("lines each series up with the right axis position", () => {
        const option = transformData("bar", twoDimensions, {
            xAxisField: "weather.location",
        });

        const byName = Object.fromEntries(option.series.map((s) => [s.name, s.data]));

        expect(byName.rain).toEqual([5, 4]);
        expect(byName.sun).toEqual([2, 1]);
    });

    it("swaps axis and legend when the other field is chosen", () => {
        const option = transformData("bar", twoDimensions, {
            xAxisField: "weather.weather",
        });

        expect(option.xAxis.data).toEqual(["rain", "sun"]);
        expect(option.legend.data).toEqual(["Seattle", "Oslo"]);
    });

    it("appends the measure once there is more than one", () => {
        const option = transformData(
            "bar",
            {
                ...twoDimensions,
                query: {
                    ...twoDimensions.query,
                    measures: ["weather.count", "weather.total"],
                },
                annotation: {
                    measures: {
                        "weather.count": { title: "Count" },
                        "weather.total": { title: "Total" },
                    },
                },
            },
            { xAxisField: "weather.location" },
        );

        expect(option.legend.data).toEqual([
            "rain, Count",
            "sun, Count",
            "rain, Total",
            "sun, Total",
        ]);
    });

    it("gives every series a slot for every axis value", () => {
        const option = transformData("bar", twoDimensions, {
            xAxisField: "weather.location",
        });

        option.series.forEach((s) => expect(s.data).toHaveLength(option.xAxis.data.length));
    });

    it("reads string values from the engine as numbers", () => {
        const option = transformData(
            "bar",
            {
                ...weather,
                data: [
                    { "weather.weather": "rain", "weather.count": "5" },
                    { "weather.weather": "sun", "weather.count": "3" },
                ],
            },
            {},
        );

        expect(option.series[0].data).toEqual([5, 3]);
    });
});

describe("no axis field", () => {
    it("turns every dimension value into its own series", () => {
        const option = transformData("bar", weather, { xAxisField: "none" });

        expect(option.legend.data).toEqual(["rain", "sun", "fog"]);
        expect(option.xAxis.data).toEqual([""]);
        expect(option.series).toHaveLength(3);
    });

    it("gives each series its single value", () => {
        const option = transformData("bar", weather, { xAxisField: "none" });

        const byName = Object.fromEntries(option.series.map((s) => [s.name, s.data]));

        expect(byName.rain).toEqual([5]);
        expect(byName.sun).toEqual([3]);
        expect(byName.fog).toEqual([1]);
    });

    it("cannot be confused with a dimension, since names are qualified", () => {
        expect(weather.query.dimensions.every((d) => d.includes("."))).toBe(true);
    });
});

describe("sort bars by value", () => {
    it("leaves the order alone when unset", () => {
        const option = transformData("bar", weather, {});

        expect(option.xAxis.data).toEqual(["rain", "sun", "fog"]);
    });

    it("sorts axis categories descending", () => {
        const option = transformData("bar", weather, { sortBars: "desc" });

        expect(option.xAxis.data).toEqual(["rain", "sun", "fog"]);
        expect(option.series[0].data).toEqual([5, 3, 1]);
    });

    it("sorts axis categories ascending", () => {
        const option = transformData("bar", weather, { sortBars: "asc" });

        expect(option.xAxis.data).toEqual(["fog", "sun", "rain"]);
        expect(option.series[0].data).toEqual([1, 3, 5]);
    });

    it("keeps every series lined up with the reordered axis", () => {
        const option = transformData(
            "bar",
            {
                query: {
                    measures: ["w.count", "w.total"],
                    dimensions: ["w.kind"],
                    timeDimensions: [],
                },
                annotation: {
                    measures: {
                        "w.count": { title: "Count" },
                        "w.total": { title: "Total" },
                    },
                },
                data: [
                    { "w.kind": "a", "w.count": 1, "w.total": 10 },
                    { "w.kind": "b", "w.count": 9, "w.total": 20 },
                ],
            },
            { sortBars: "desc" },
        );

        expect(option.xAxis.data).toEqual(["b", "a"]);
        option.series.forEach((s) => expect(s.data).toHaveLength(2));
        const byName = Object.fromEntries(option.series.map((s) => [s.name, s.data]));

        expect(byName.Count).toEqual([9, 1]);
        expect(byName.Total).toEqual([20, 10]);
    });

    it("sorts the series instead when the categories are in the legend", () => {
        const option = transformData("bar", weather, {
            xAxisField: "none",
            sortBars: "desc",
        });

        expect(option.legend.data).toEqual(["rain", "sun", "fog"]);
        expect(option.series.map((s) => s.data[0])).toEqual([5, 3, 1]);
    });

    it("sorts the series the other way round", () => {
        const option = transformData("bar", weather, {
            xAxisField: "none",
            sortBars: "asc",
        });

        expect(option.legend.data).toEqual(["fog", "sun", "rain"]);
    });

    it("leaves a time axis in chronological order", () => {
        const option = transformData(
            "time_bar",
            {
                query: {
                    measures: ["w.count"],
                    dimensions: [],
                    timeDimensions: [
                        {
                            dimension: "w.created",
                            granularity: "day",
                            dateRange: ["2026-01-01T00:00:00.000", "2026-01-03T00:00:00.000"],
                        },
                    ],
                },
                annotation: { measures: { "w.count": { title: "Count" } } },
                data: [
                    {
                        "w.created.day": "2026-01-01T00:00:00.000",
                        "w.count": 9,
                    },
                    {
                        "w.created.day": "2026-01-02T00:00:00.000",
                        "w.count": 1,
                    },
                    {
                        "w.created.day": "2026-01-03T00:00:00.000",
                        "w.count": 5,
                    },
                ],
            },
            { sortBars: "desc" },
        );

        expect(option.xAxis.data).toHaveLength(3);
        expect(option.series[0].data).toEqual([9, 1, 5]);
    });
});

describe("time axis keeps every value when the range starts mid-period", () => {
    function countBy(granularity, dateRange, rows) {
        return transformData(
            "line",
            cubeResult({
                measures: ["posts.count"],
                dimensions: [],
                timeDimensions: [
                    {
                        dimension: "posts.created",
                        granularity: granularity,
                        dateRange: dateRange,
                    },
                ],
                rows: rows,
            }),
        );
    }

    it("by month, starting mid-month", () => {
        const option = countBy(
            "month",
            ["2026-09-15T00:00:00.000", "2026-10-14T23:59:59.999"],
            [
                { "posts.created.month": "2026-09-01T00:00:00.000", "posts.count": "152" },
                { "posts.created.month": "2026-10-01T00:00:00.000", "posts.count": "67" },
            ],
        );

        expect(plainSpaces(option.xAxis.data)).toEqual(["15 – 30 Sept 2026", "1 – 14 Oct 2026"]);
        expect(option.series[0].data).toEqual([152, 67]);
    });

    it("by week, starting on a Thursday", () => {
        const option = countBy(
            "week",
            ["2026-09-17T00:00:00.000", "2026-09-27T23:59:59.999"],
            [
                { "posts.created.week": "2026-09-14T00:00:00.000", "posts.count": "40" },
                { "posts.created.week": "2026-09-21T00:00:00.000", "posts.count": "10" },
            ],
        );

        expect(plainSpaces(option.xAxis.data)).toEqual(["17 – 20 Sept 2026", "21 – 27 Sept 2026"]);
        expect(option.series[0].data).toEqual([40, 10]);
    });

    it("by quarter, starting mid-quarter", () => {
        const option = countBy(
            "quarter",
            ["2026-02-10T00:00:00.000", "2026-08-01T23:59:59.999"],
            [
                { "posts.created.quarter": "2026-01-01T00:00:00.000", "posts.count": "5" },
                { "posts.created.quarter": "2026-07-01T00:00:00.000", "posts.count": "3" },
            ],
        );

        expect(plainSpaces(option.xAxis.data)).toEqual([
            "10 Feb – 31 Mar 2026",
            "Q2 2026",
            "1 Jul – 1 Aug 2026",
        ]);
        expect(option.series[0].data).toEqual([5, null, 3]);
    });

    it("by year, starting mid-year", () => {
        const option = countBy(
            "year",
            ["2025-06-01T00:00:00.000", "2026-03-01T23:59:59.999"],
            [
                { "posts.created.year": "2025-01-01T00:00:00.000", "posts.count": "4" },
                { "posts.created.year": "2026-01-01T00:00:00.000", "posts.count": "6" },
            ],
        );

        expect(plainSpaces(option.xAxis.data)).toEqual([
            "1 Jun – 31 Dec 2025",
            "1 Jan – 1 Mar 2026",
        ]);
        expect(option.series[0].data).toEqual([4, 6]);
    });

    it("by hour, starting mid-hour", () => {
        const option = countBy(
            "hour",
            ["2026-09-15T10:30:00.000", "2026-09-15T11:59:59.999"],
            [
                { "posts.created.hour": "2026-09-15T10:00:00.000", "posts.count": "2" },
                { "posts.created.hour": "2026-09-15T11:00:00.000", "posts.count": "1" },
            ],
        );

        expect(option.xAxis.data).toEqual(["15 Sept 2026, 10:00", "15 Sept 2026, 11:00"]);
        expect(option.series[0].data).toEqual([2, 1]);
    });
});
