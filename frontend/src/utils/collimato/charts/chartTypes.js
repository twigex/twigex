// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export const RENDERER_ECHARTS = "echarts";
export const RENDERER_NUMBER = "number";
export const RENDERER_MAP = "map";

export const CHART_TYPES = {
    big_number: {
        renderer: RENDERER_NUMBER,
        maxMeasures: null,
        hasTimeAxis: false,
        drillDownKey: null,
    },
    line: {
        renderer: RENDERER_ECHARTS,
        maxMeasures: null,
        hasTimeAxis: true,
        drillDownKey: "seriesName",
    },
    bar: {
        renderer: RENDERER_ECHARTS,
        maxMeasures: null,
        hasTimeAxis: false,
        drillDownKey: "name",
    },
    time_bar: {
        renderer: RENDERER_ECHARTS,
        maxMeasures: null,
        hasTimeAxis: true,
        drillDownKey: "seriesName",
    },
    pie: {
        renderer: RENDERER_ECHARTS,
        maxMeasures: 1,
        hasTimeAxis: false,
        drillDownKey: "name",
    },
    map: {
        renderer: RENDERER_MAP,
        maxMeasures: null,
        hasTimeAxis: false,
        drillDownKey: null,
    },
};

export const CHART_TYPE_IDS = Object.keys(CHART_TYPES);

export function rendererFor(chartType) {
    return CHART_TYPES[chartType]?.renderer ?? null;
}

export function isEchartsType(chartType) {
    return rendererFor(chartType) === RENDERER_ECHARTS;
}

export function hasTimeAxis(chartType) {
    return CHART_TYPES[chartType]?.hasTimeAxis ?? false;
}

export function maxMeasuresFor(chartType) {
    return CHART_TYPES[chartType]?.maxMeasures ?? null;
}

export function drillDownKeyFor(chartType) {
    return CHART_TYPES[chartType]?.drillDownKey ?? null;
}
