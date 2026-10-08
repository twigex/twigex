// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { transformData, applyConfiguration } from "@/utils/collimato/chartTypes.js";

function cubeResult({ measures, dimensions, timeDimensions, rows, titles }) {
    return {
        query: {
            measures,
            dimensions,
            timeDimensions: timeDimensions ?? [],
        },
        annotation: {
            measures: Object.fromEntries(measures.map((m) => [m, { title: titles?.[m] ?? m }])),
            dimensions: Object.fromEntries(dimensions.map((d) => [d, { title: d }])),
            timeDimensions: {},
        },
        data: rows,
    };
}

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

        expect(option.xAxis.data).toEqual(["2026-01-01", "2026-01-02", "2026-01-03"]);
        expect(option.series[0].data).toEqual([0, 7, 0]);
    });
});

describe("applyConfiguration", () => {
    const barConfig = {
        legend: true,
        bottomMargin: 5,
        topMargin: 30,
        leftMargin: 5,
        rightMargin: 5,
        containLabel: true,
        dataZoom: false,
        stack: false,
        orientation: "vertical",
        showValues: false,
        angle: 0,
    };

    it.each(["pie", "bar", "line", "time_bar"])(
        "returns an empty option untouched for %s",
        (chartType) => {
            expect(applyConfiguration(chartType, {}, barConfig)).toEqual({});
        },
    );

    it("returns a null option untouched", () => {
        expect(applyConfiguration("bar", null, barConfig)).toBeNull();
    });
});

describe("pie configuration", () => {
    let warn;

    beforeEach(() => {
        warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    });

    afterEach(() => {
        warn.mockRestore();
    });

    function pieOption() {
        return transformData(
            "pie",
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                rows: [
                    { "orders.status": "new", "orders.count": 5 },
                    { "orders.status": "done", "orders.count": 5 },
                ],
            }),
        );
    }

    it("falls back to default radii when the configuration omits them", () => {
        const option = applyConfiguration("pie", pieOption(), {
            legend: true,
            threshold: 5,
            labelType: "category",
            orientation: "Top",
        });

        expect(option.series[0].radius).toEqual(["0%", "55%"]);
    });

    it("honours configured radii", () => {
        const option = applyConfiguration("pie", pieOption(), {
            legend: true,
            threshold: 5,
            labelType: "category",
            orientation: "Top",
            innerRadius: 20,
            outerRadius: 70,
        });

        expect(option.series[0].radius).toEqual(["20%", "70%"]);
    });

    it("yields an empty option when there is nothing to plot", () => {
        const option = transformData("pie", cubeResult({ measures: [], dimensions: [], rows: [] }));

        expect(option).toEqual({});
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
