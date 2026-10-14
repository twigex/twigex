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

export function drillDownValuesFor(chartType, option, params) {
    const key = drillDownKeyFor(chartType);

    if (!option || !option.drillDownValues) {
        return {};
    }

    const known = option.drillDownValues[key];

    if (!known) {
        return {};
    }

    const label = params[key];

    if (!Object.hasOwn(known, label)) {
        return {};
    }

    return known[label];
}

function drillDownFilter(member, values) {
    const hasNull = values.includes(null);
    const withoutNull = values.filter((value) => value !== null);
    const named = withoutNull.map((value) => String(value));

    const equals = { member: member, operator: "equals", values: named };
    const notSet = { member: member, operator: "notSet" };

    if (!hasNull) {
        return equals;
    }

    if (named.length === 0) {
        return notSet;
    }

    return { or: [notSet, equals] };
}

export function drillDownFilters(clicked) {
    const filters = [];

    if (!clicked) {
        return filters;
    }

    Object.keys(clicked).forEach((member) => {
        const values = clicked[member];

        filters.push(drillDownFilter(member, values));
    });

    return filters;
}

export function filterMembers(filter) {
    if (filter.or) {
        return filter.or.flatMap((inner) => filterMembers(inner));
    }

    if (!filter.member) {
        return [];
    }

    return [filter.member];
}

export function filterValues(filter) {
    if (filter.operator === "notSet") {
        return [null];
    }

    if (filter.or) {
        return filter.or.flatMap((inner) => filterValues(inner));
    }

    if (!filter.values) {
        return [];
    }

    return filter.values;
}
