// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { CHART_PALETTE, CHART_LABEL_COLOR } from "@/utils/collimato/charts/chartPalette.js";
import { recordDrillDownValues } from "@/utils/collimato/charts/drillDown.js";
import {
    LEGEND_POSITIONS,
    DEFAULT_LEGEND_POSITION,
    X_AXIS_NONE,
    applyMissingValues,
    sortBarsByValue,
} from "@/utils/collimato/charts/chartSettings.js";
import { generateTimeArray, timeSlotLabels } from "@/utils/collimato/charts/timeSlots.js";
import {
    labelOf,
    valuesOf,
    measureTitleOf,
    shortNumber,
} from "@/utils/collimato/charts/chartLabels.js";

const SUB_DAY_GRANULARITIES = ["second", "minute", "hour"];

const AXIS_SERIES_PRESETS = {
    line: { type: "line", symbolSize: 8, smooth: true },
    bar: { type: "bar", stack: "" },
};

export function axisDimensionsOf(configuration, dimensions) {
    const chosen = configuration?.xAxisField;

    if (chosen === X_AXIS_NONE) {
        return [];
    }

    return chosen && dimensions.includes(chosen) ? [chosen] : dimensions;
}

function dimensionValuesOf(item, dimensions, timeDimension) {
    return dimensions
        .filter((dimension) => dimension !== timeDimension?.dimension)
        .map((dimension) => labelOf(item[dimension]));
}

function seriesNameOf(measureTitle, dimensionValues, showMeasure) {
    const parts = [...dimensionValues];

    if (showMeasure || parts.length === 0) {
        parts.push(measureTitle);
    }

    return parts.join(", ");
}

function emptySeries(name, seriesType, length) {
    return {
        name: name,
        data: Array.from({ length: length }, () => null),
        label: {
            show: false,
            position: "top",
            color: CHART_LABEL_COLOR,
            formatter: (params) => shortNumber(params.value),
        },
        ...(AXIS_SERIES_PRESETS[seriesType] ?? AXIS_SERIES_PRESETS.line),
    };
}

export function axisSeriesChart(data, seriesType, configuration = {}, display = {}) {
    const { measures, dimensions, timeDimensions } = data.query;
    const timeDimension = timeDimensions?.[0];
    const granularity = timeDimension?.granularity;
    const withTime = SUB_DAY_GRANULARITIES.includes(granularity);
    const showMeasure = measures.length > 1;

    let xAxisData = [];
    const series = [];
    const legend = [];
    const drillDownValues = { name: {}, seriesName: {} };

    const seriesFor = (name) => {
        let entry = series.find((s) => s.name === name);

        if (!entry) {
            entry = emptySeries(name, seriesType, xAxisData.length);
            series.push(entry);
            legend.push(name);
        }

        return entry;
    };

    if (granularity) {
        // Cube names each row after the start of its period (the Monday, the
        // 1st), even when the range starts later, but counts only the days
        // inside the range. A range from Sunday 20 Dec gives a row named
        // 14 Dec with only the 20th counted. The slots use Cube's names so
        // each row finds its place; the labels show the days counted.
        const slots = generateTimeArray(
            timeDimension.dateRange[0],
            timeDimension.dateRange[1],
            granularity,
            withTime,
        );

        xAxisData = slots;

        const timeKey = `${timeDimension.dimension}.${granularity}`;
        const seriesDimensions = dimensions.filter(
            (dimension) => dimension !== timeDimension.dimension,
        );

        measures.forEach((measure) => {
            const title = measureTitleOf(data, measure);

            data.data.forEach((item) => {
                const name = seriesNameOf(
                    title,
                    dimensionValuesOf(item, dimensions, timeDimension),
                    showMeasure,
                );

                recordDrillDownValues(drillDownValues.seriesName, name, item, seriesDimensions);

                const raw = item[timeKey];
                const bucket = withTime ? raw : raw?.split("T")[0];
                const index = slots.indexOf(bucket);

                const value = item[measure];

                if (index !== -1) {
                    const entry = seriesFor(name);

                    if (value != null && value !== "") {
                        entry.data[index] += Number(value);
                    }
                }
            });
        });

        const labels = timeSlotLabels(slots, granularity, timeDimension.dateRange, display);

        xAxisData = labels;
    } else {
        const axisDimensions = axisDimensionsOf(configuration, dimensions);
        const seriesDimensions = dimensions.filter(
            (dimension) => !axisDimensions.includes(dimension),
        );

        data.data.forEach((item) => {
            const label = valuesOf(item, axisDimensions).join(", ");

            recordDrillDownValues(drillDownValues.name, label, item, axisDimensions);

            if (!xAxisData.includes(label)) {
                xAxisData.push(label);
            }
        });

        measures.forEach((measure) => {
            const title = measureTitleOf(data, measure);

            data.data.forEach((item) => {
                const name = seriesNameOf(title, valuesOf(item, seriesDimensions), showMeasure);
                const index = xAxisData.indexOf(valuesOf(item, axisDimensions).join(", "));

                recordDrillDownValues(drillDownValues.seriesName, name, item, seriesDimensions);

                const value = item[measure];

                if (index !== -1) {
                    const entry = seriesFor(name);

                    if (value != null && value !== "") {
                        entry.data[index] += Number(value);
                    }
                }
            });
        });
    }

    // Done here rather than in applyLineChartConfiguration, which
    // returns early for charts saved without any configuration.
    applyMissingValues(series, configuration);

    //Accessible in ChartWidget.vue as currentData.
    const option = {
        color: CHART_PALETTE,
        legend: {
            type: "scroll",
            data: legend,
            show: true,
            ...LEGEND_POSITIONS[DEFAULT_LEGEND_POSITION],
        },
        tooltip: {
            trigger: "axis",
            confine: false,
            appendToBody: true,
            axisPointer: { type: "shadow" },
        },
        grid: {
            left: "5",
            right: "5",
            bottom: "5",
            top: "30",
            containLabel: true,
        },
        xAxis: {
            name: "",
            nameLocation: "middle",
            nameGap: 80,
            type: "category",
            data: xAxisData,
            axisLabel: { rotate: 0 },
        },
        yAxis: {
            name: "",
            nameLocation: "middle",
            nameGap: 80,
            type: "value",
            axisLabel: { formatter: (value) => shortNumber(value) },
        },
        drillDownValues: drillDownValues,
        dataZoom: [
            {
                type: "slider",
                show: false,
                xAxisIndex: [0],
                start: 0,
                end: 100,
            },
            {
                type: "inside",
                show: false,
                xAxisIndex: [0],
                start: 0,
                end: 100,
            },
        ],
        series,
    };

    return granularity ? option : sortBarsByValue(option, configuration?.sortBars);
}

export function lineChart(data, configuration, display) {
    return axisSeriesChart(data, "line", configuration, display);
}

export function barChart(data, configuration, display) {
    return axisSeriesChart(data, "bar", configuration, display);
}

export function timeBarChart(data, configuration, display) {
    return axisSeriesChart(data, "bar", configuration, display);
}
