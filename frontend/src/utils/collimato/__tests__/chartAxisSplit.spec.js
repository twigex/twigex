// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData } from "@/utils/collimato/chartTypes.js";

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
