// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { lineChart, barChart, timeBarChart } from "@/utils/collimato/charts/axisChart.js";
import { pieChart } from "@/utils/collimato/charts/pieChart.js";
import { number } from "@/utils/collimato/charts/bigNumber.js";
import {
    applyBarChartConfiguration,
    applyLineChartConfiguration,
    applyPieChartConfiguration,
} from "@/utils/collimato/charts/chartSettings.js";

export const CHART_DRAWING = {
    big_number: {
        transform: number,
        applyConfig: null,
    },
    line: {
        transform: lineChart,
        applyConfig: applyLineChartConfiguration,
    },
    bar: {
        transform: barChart,
        applyConfig: applyBarChartConfiguration,
    },
    time_bar: {
        transform: timeBarChart,
        applyConfig: applyBarChartConfiguration,
    },
    pie: {
        transform: pieChart,
        applyConfig: applyPieChartConfiguration,
    },
    map: {
        transform: null,
        applyConfig: null,
    },
};

export function transformData(chartType, data, configuration = {}, display = {}) {
    const transform = CHART_DRAWING[chartType]?.transform;

    return transform ? transform(data, configuration, display) : {};
}

// A JSON copy would drop functions, such as the y-axis number formatter.
function copyOption(value) {
    if (Array.isArray(value)) {
        return value.map((item) => copyOption(item));
    }

    if (value === null || typeof value !== "object") {
        return value;
    }

    const copy = {};

    Object.keys(value).forEach((key) => {
        copy[key] = copyOption(value[key]);
    });

    return copy;
}

export function applyConfiguration(chartType, data, configuration) {
    if (!data || Object.keys(data).length === 0) {
        return data;
    }

    const clone = copyOption(data);
    const apply = CHART_DRAWING[chartType]?.applyConfig;

    return apply ? apply(clone, configuration) : clone;
}
