// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData, applyConfiguration } from "@/utils/collimato/charts/chartDrawing.js";
import { cubeResult } from "@/utils/collimato/charts/__tests__/cubeResult.js";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";
import {
    LEGEND_POSITIONS,
    DEFAULT_LEGEND_POSITION,
} from "@/utils/collimato/charts/chartSettings.js";
import { CHART_TYPE_IDS, isEchartsType } from "@/utils/collimato/charts/chartTypes.js";

const LOCALES = { en, lv, kk, pl };
const POSITIONS = Object.keys(LEGEND_POSITIONS);

const result = {
    query: {
        measures: ["orders.count"],
        dimensions: ["orders.status"],
        timeDimensions: [],
    },
    annotation: { measures: { "orders.count": { title: "Count" } } },
    data: [
        { "orders.status": "new", "orders.count": 5 },
        { "orders.status": "done", "orders.count": 3 },
    ],
};

const axisConfig = {
    legend: true,
    bottomMargin: 5,
    topMargin: 30,
    leftMargin: 5,
    rightMargin: 5,
    containLabel: true,
    dataZoom: false,
    stack: false,
    angle: 0,
    xAxisLabel: "",
    xAxisNameGap: 30,
    yAxisLabel: "",
    yAxisNameGap: 30,
    showValues: false,
    orientation: "vertical",
};

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

describe("missing values setting", () => {
    function postsPerDay() {
        return cubeResult({
            measures: ["posts.count"],
            dimensions: [],
            timeDimensions: [
                {
                    dimension: "posts.created",
                    granularity: "day",
                    dateRange: ["2026-01-01T00:00:00.000", "2026-01-03T00:00:00.000"],
                },
            ],
            rows: [{ "posts.created.day": "2026-01-02T00:00:00.000", "posts.count": "7" }],
        });
    }

    it("connects the line across missing slots by default", () => {
        const option = transformData("line", postsPerDay(), {});

        expect(option.series[0].data).toEqual([null, 7, null]);
        expect(option.series[0].connectNulls).toBe(true);
    });

    it("leaves a gap when asked to", () => {
        const option = transformData("line", postsPerDay(), { missingValues: "gap" });

        expect(option.series[0].data).toEqual([null, 7, null]);
        expect(option.series[0].connectNulls).toBe(false);
    });

    it("draws missing slots as zero when asked to", () => {
        const option = transformData("line", postsPerDay(), { missingValues: "zero" });

        expect(option.series[0].data).toEqual([0, 7, 0]);
    });

    it("keeps a real zero as zero", () => {
        const data = postsPerDay();

        data.data.push({ "posts.created.day": "2026-01-03T00:00:00.000", "posts.count": "0" });

        const option = transformData("line", data, { missingValues: "gap" });

        expect(option.series[0].data).toEqual([null, 7, 0]);
    });
});

describe("legend position", () => {
    it.each(CHART_TYPE_IDS.filter(isEchartsType))(
        "%s puts its legend at the top by default",
        (id) => {
            const legend = transformData(id, result).legend;

            expect(legend.top).toBe("top");
            expect(legend.left).toBe("center");
        },
    );

    it.each(POSITIONS)("moves an axis chart legend to %s", (position) => {
        const option = applyConfiguration("bar", transformData("bar", result), {
            ...axisConfig,
            legendPosition: position,
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS[position]);
    });

    it("falls back to the top for a saved chart with no position", () => {
        const option = applyConfiguration("bar", transformData("bar", result), {
            ...axisConfig,
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS[DEFAULT_LEGEND_POSITION]);
    });

    it.each(POSITIONS)("moves a pie legend to %s", (position) => {
        const option = applyConfiguration("pie", transformData("pie", result), {
            legend: true,
            showLabels: true,
            labelPosition: "outside",
            legendPosition: position,
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS[position]);
    });

    it("still honours a pie saved with the old orientation key", () => {
        const option = applyConfiguration("pie", transformData("pie", result), {
            legend: true,
            showLabels: true,
            labelPosition: "outside",
            orientation: "Bottom",
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS.Bottom);
    });

    it("prefers the current key when a chart carries both", () => {
        const option = applyConfiguration("pie", transformData("pie", result), {
            legend: true,
            showLabels: true,
            labelPosition: "outside",
            legendPosition: "Left",
            orientation: "Bottom",
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS.Left);
    });

    it("means chart orientation, not legend position, on an axis chart", () => {
        const option = applyConfiguration("bar", transformData("bar", result), {
            ...axisConfig,
            orientation: "horizontal",
            legendPosition: "Bottom",
        });

        expect(option.legend).toMatchObject(LEGEND_POSITIONS.Bottom);
        expect(option.yAxis.type).toBe("category");
    });

    it("stands a side legend up vertically", () => {
        expect(LEGEND_POSITIONS.Left.orient).toBe("vertical");
        expect(LEGEND_POSITIONS.Right.orient).toBe("vertical");
    });

    it.each(POSITIONS)("%s is named in every locale", (position) => {
        const key = `collimato.charts.new_chart.customize.legend_position.${position.toLowerCase()}`;

        Object.entries(LOCALES).forEach(([locale, messages]) => {
            expect(messages[key], `${locale} ${key}`).toBeTruthy();
        });
    });

    it("labels the control in every locale", () => {
        Object.entries(LOCALES).forEach(([locale, messages]) => {
            expect(
                messages["collimato.charts.new_chart.customize.legend_position"],
                locale,
            ).toBeTruthy();
        });
    });
});

describe("y-axis number labels", () => {
    function yAxisFormatter(chartType, configuration) {
        const option = transformData(
            chartType,
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                rows: [{ "orders.status": "new", "orders.count": "5" }],
            }),
        );

        const applied = applyConfiguration(chartType, option, configuration);

        return applied.yAxis.axisLabel.formatter;
    }

    it("keeps the formatter when the settings are applied", () => {
        const formatter = yAxisFormatter("line", { legend: true });

        expect(formatter(1500)).toBe("1.5K");
    });

    it("keeps the formatter on a horizontal bar chart, where the axes swap", () => {
        const option = transformData(
            "bar",
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                rows: [{ "orders.status": "new", "orders.count": "5" }],
            }),
        );

        const applied = applyConfiguration("bar", option, {
            legend: true,
            orientation: "horizontal",
        });

        expect(applied.xAxis.axisLabel.formatter(1500)).toBe("1.5K");
    });
});

describe("bar value labels", () => {
    it("shows bar values shortened, like the axis", () => {
        const option = transformData(
            "bar",
            cubeResult({
                measures: ["finance.profit"],
                dimensions: ["finance.region"],
                rows: [
                    { "finance.region": "East", "finance.profit": "-13319284.31" },
                    { "finance.region": "West", "finance.profit": "3903576.24" },
                ],
            }),
        );

        const applied = applyConfiguration("bar", option, { legend: true, showValues: true });
        const label = applied.series[0].label;

        expect(label.show).toBe(true);
        expect(label.formatter({ value: -13319284.31 })).toBe("-13.3M");
        expect(label.formatter({ value: 3903576.24 })).toBe("3.9M");
    });
});
