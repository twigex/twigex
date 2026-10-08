// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";
import { LEGEND_POSITIONS, DEFAULT_LEGEND_POSITION } from "@/utils/collimato/chartUtils.js";
import {
    CHART_TYPE_IDS,
    isEchartsType,
    transformData,
    applyConfiguration,
} from "@/utils/collimato/chartTypes.js";

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
