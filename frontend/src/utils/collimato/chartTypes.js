// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import {
    lineChart,
    barChart,
    timeBarChart,
    pieChart,
    number,
    applyBarChartConfiguration,
    applyLineChartConfiguration,
    applyPieChartConfiguration,
    axisDimensionsOf,
} from "@/utils/collimato/chartUtils.js";

export const RENDERER_ECHARTS = "echarts";
export const RENDERER_NUMBER = "number";
export const RENDERER_MAP = "map";

export const CHART_TYPES = {
    big_number: {
        renderer: RENDERER_NUMBER,
        transform: number,
        applyConfig: null,
        maxMeasures: null,
        hasTimeAxis: false,
        drillDownKey: null,
    },
    line: {
        renderer: RENDERER_ECHARTS,
        transform: lineChart,
        applyConfig: applyLineChartConfiguration,
        maxMeasures: null,
        hasTimeAxis: true,
        drillDownKey: "seriesName",
    },
    bar: {
        renderer: RENDERER_ECHARTS,
        transform: barChart,
        applyConfig: applyBarChartConfiguration,
        maxMeasures: null,
        hasTimeAxis: false,
        drillDownKey: "name",
    },
    time_bar: {
        renderer: RENDERER_ECHARTS,
        transform: timeBarChart,
        applyConfig: applyBarChartConfiguration,
        maxMeasures: null,
        hasTimeAxis: true,
        drillDownKey: "seriesName",
    },
    pie: {
        renderer: RENDERER_ECHARTS,
        transform: pieChart,
        applyConfig: applyPieChartConfiguration,
        maxMeasures: 1,
        hasTimeAxis: false,
        drillDownKey: "name",
    },
    map: {
        renderer: RENDERER_MAP,
        transform: null,
        applyConfig: null,
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

export function transformData(chartType, data, configuration = {}) {
    const transform = CHART_TYPES[chartType]?.transform;

    return transform ? transform(data, configuration) : {};
}

export function applyConfiguration(chartType, data, configuration) {
    if (!data || Object.keys(data).length === 0) {
        return data;
    }

    const clone = JSON.parse(JSON.stringify(data));
    const apply = CHART_TYPES[chartType]?.applyConfig;

    return apply ? apply(clone, configuration) : clone;
}

export function drillDownMembers(chartType, configuration, query) {
    const dimensions = query?.dimensions ?? [];

    if (chartType === "pie") {
        return dimensions.slice(0, 1);
    }

    const usesSeries = drillDownKeyFor(chartType) === "seriesName";
    const timeDimension = query?.timeDimensions?.[0]?.dimension;

    if (timeDimension) {
        return usesSeries ? dimensions.filter((dimension) => dimension !== timeDimension) : [];
    }

    const axis = axisDimensionsOf(configuration, dimensions);

    return usesSeries ? dimensions.filter((dimension) => !axis.includes(dimension)) : axis;
}

export function drillDownFilters(chartType, configuration, query, clicked) {
    const members = drillDownMembers(chartType, configuration, query);
    const values = String(clicked ?? "").split(", ");
    const carriesMeasure =
        drillDownKeyFor(chartType) === "seriesName" && (query?.measures?.length ?? 0) > 1;

    const named = carriesMeasure ? values.slice(0, -1) : values;

    return members
        .map((member, index) => ({
            member: member,
            operator: "equals",
            values: [named[index]],
        }))
        .filter((filter) => filter.values[0] !== undefined);
}
