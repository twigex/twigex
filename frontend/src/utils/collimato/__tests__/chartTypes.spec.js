// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";
import {
    CHART_TYPES,
    CHART_TYPE_IDS,
    RENDERER_ECHARTS,
    RENDERER_NUMBER,
    RENDERER_MAP,
    hasTimeAxis,
    isEchartsType,
    drillDownKeyFor,
    transformData,
    applyConfiguration,
    maxMeasuresFor,
} from "@/utils/collimato/chartTypes.js";
import { CHART_TYPE_PRESENTATION } from "@/components/Collimato/Charts/chartTypePresentation.js";
import { buildQuery } from "@/utils/collimato/chartQuery.js";

const RENDERERS = [RENDERER_ECHARTS, RENDERER_NUMBER, RENDERER_MAP];
const LOCALES = { en, lv, kk, pl };

describe("every registered chart type", () => {
    it.each(CHART_TYPE_IDS)("%s declares a known renderer", (id) => {
        expect(RENDERERS).toContain(CHART_TYPES[id].renderer);
    });

    it.each(CHART_TYPE_IDS)("%s has a transform unless it renders itself", (id) => {
        const type = CHART_TYPES[id];

        if (type.renderer === RENDERER_MAP) {
            expect(type.transform).toBeNull();
        } else {
            expect(typeof type.transform).toBe("function");
        }
    });

    it.each(CHART_TYPE_IDS)("%s only applies config when ECharts renders it", (id) => {
        const type = CHART_TYPES[id];

        if (type.applyConfig) {
            expect(type.renderer).toBe(RENDERER_ECHARTS);
        }
    });

    it.each(CHART_TYPE_IDS)("%s only drills down when ECharts renders it", (id) => {
        const type = CHART_TYPES[id];

        if (type.drillDownKey) {
            expect(type.renderer).toBe(RENDERER_ECHARTS);
            expect(["name", "seriesName"]).toContain(type.drillDownKey);
        }
    });

    it.each(CHART_TYPE_IDS)("%s has presentation to match", (id) => {
        expect(CHART_TYPE_PRESENTATION[id]).toBeDefined();
    });

    it("has no presentation entry without a type behind it", () => {
        expect(Object.keys(CHART_TYPE_PRESENTATION).sort()).toEqual([...CHART_TYPE_IDS].sort());
    });
});

describe("chart type labels", () => {
    it.each(CHART_TYPE_IDS)("%s resolves in every locale", (id) => {
        const { labelKey, shortLabelKey } = CHART_TYPE_PRESENTATION[id];

        Object.entries(LOCALES).forEach(([locale, messages]) => {
            expect(messages[labelKey], `${locale} ${labelKey}`).toBeTruthy();
            expect(messages[shortLabelKey], `${locale} ${shortLabelKey}`).toBeTruthy();
        });
    });
});

describe("time axis is declared once", () => {
    const time = {
        dimension: { name: "orders.created_at" },
        dateRange: "this month",
        granularity: "day",
    };

    it.each(CHART_TYPE_IDS)("buildQuery agrees with the registry for %s", (id) => {
        const query = buildQuery({ time, chartType: id });

        expect(query.timeDimensions.length > 0).toBe(hasTimeAxis(id));
    });
});

describe("dispatch", () => {
    it("returns an empty option for an unknown type", () => {
        expect(transformData("nope", { query: {}, data: [] })).toEqual({});
        expect(isEchartsType("nope")).toBe(false);
        expect(hasTimeAxis("nope")).toBe(false);
        expect(drillDownKeyFor("nope")).toBeNull();
    });

    it("leaves the option untouched for a type with no config step", () => {
        const option = { series: [] };

        expect(applyConfiguration("big_number", option, { a: 1 })).toEqual(option);
    });

    it("does not mutate the option it is given", () => {
        const option = { legend: { show: false }, series: [] };
        const before = JSON.stringify(option);

        applyConfiguration("pie", option, {});

        expect(JSON.stringify(option)).toBe(before);
    });
});

describe("maxMeasuresFor", () => {
    it.each(CHART_TYPE_IDS)("%s declares a limit or none at all", (id) => {
        const limit = maxMeasuresFor(id);

        expect(limit === null || limit > 0).toBe(true);
    });

    it("holds a pie to one measure", () => {
        expect(maxMeasuresFor("pie")).toBe(1);
    });

    it("has no limit for a type this build does not know", () => {
        expect(maxMeasuresFor("candlestick")).toBeNull();
    });
});
