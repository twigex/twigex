// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { CHART_PALETTE, CHART_LABEL_COLOR } from "@/utils/collimato/chartPalette.js";

export const MAX_QUERY_LIMIT = 50000;

export const LEGEND_POSITIONS = {
    Top: { orient: "horizontal", top: "top", left: "center" },
    Bottom: { orient: "horizontal", top: "bottom", left: "center" },
    Left: { orient: "vertical", top: "middle", left: "left" },
    Right: { orient: "vertical", top: "middle", left: "right" },
};

export const DEFAULT_LEGEND_POSITION = "Top";

export const X_AXIS_DEFAULT = "";
export const X_AXIS_NONE = "none";

export const SORT_BARS_NONE = "";
export const SORT_BARS_OPTIONS = [SORT_BARS_NONE, "asc", "desc"];

export function applyLegendPosition(legend, position) {
    Object.assign(legend, LEGEND_POSITIONS[position] ?? LEGEND_POSITIONS.Top);

    return legend;
}

export function applyPieChartConfiguration(data, configuration) {
    if (Object.keys(configuration).length === 0) {
        return data;
    }

    data.legend.show = configuration.legend;
    data.series[0].label.formatter = function (params) {
        // Hide label if percentage is less than threshold (e.g., 5%)
        if (params.percent < configuration.threshold) {
            return "";
        }
        // Otherwise, display the label with name and percentage

        const labelTypes = {
            percentage: `${params.percent}%`,
            value: `${params.value}`,
            category: `${params.name}`,
            "category, value": `${params.name} - ${params.value}`,
            "category, percentage": `${params.name} - ${params.percent}%`,
            "category, value, percentage": `${params.name} - ${params.value} - ${params.percent}%`,
        };

        if (labelTypes[configuration.labelType]) {
            return labelTypes[configuration.labelType];
        }

        return `${params.name}`;
    };

    // Precompute the total value of the series.
    const seriesData = data.series[0].data;
    let total = 0;

    seriesData.forEach((item) => {
        total += Number(item.value);
    });

    // Update each data item to include a labelLine configuration.
    // The labelLine is only shown if the computed percentage meets the threshold.
    data.series[0].data = seriesData.map((item) => {
        const percent = (item.value / total) * 100;

        return {
            ...item,
            labelLine: {
                show: percent >= configuration.threshold,
            },
            emphasis: {
                labelLine: {
                    show: percent >= configuration.threshold,
                },
            },
        };
    });

    let position = configuration.labelPosition ? "outside" : "inside";

    data.series[0].label.position = position;
    if (configuration.showLabels != undefined) {
        data.series[0].label.show = configuration.showLabels;
    }

    const outerRadius = configuration.outerRadius ?? 55;
    const innerRadius = configuration.innerRadius ?? 0;

    data.series[0].radius[1] = String(outerRadius) + "%";
    data.series[0].radius[0] = String(innerRadius) + "%";

    applyLegendPosition(data.legend, configuration.legendPosition ?? configuration.orientation);

    return data;
}

export function applyBarChartConfiguration(data, configuration) {
    if (Object.keys(configuration).length === 0) {
        return data;
    }

    data.grid.bottom = configuration.bottomMargin;
    data.grid.top = configuration.topMargin;
    data.grid.left = configuration.leftMargin;
    data.grid.right = configuration.rightMargin;
    data.grid.containLabel = configuration.containLabel;
    data.legend.show = configuration.legend;
    data.dataZoom[0].show = configuration.dataZoom;
    data.dataZoom[1].show = configuration.dataZoom;
    if (configuration.stack) {
        data.series.forEach((series) => {
            series.stack = "stack";
        });
    } else {
        data.series.forEach((series) => {
            series.stack = "";
        });
    }

    //Change bar chart orientation if vertical or horizontal
    if (configuration.orientation == "horizontal") {
        let x = data.xAxis;

        data.xAxis = data.yAxis;
        data.yAxis = x;

        data.yAxis.axisLabel.rotate = configuration.angle;
        data.yAxis.name = configuration.yAxisLabel;
        data.yAxis.nameGap = configuration.yAxisNameGap;

        data.xAxis.name = configuration.xAxisLabel;
        data.xAxis.nameGap = configuration.xAxisNameGap;
    } else {
        data.xAxis.axisLabel.rotate = configuration.angle;
        data.xAxis.name = configuration.xAxisLabel;
        data.xAxis.nameGap = configuration.xAxisNameGap;
        data.yAxis.name = configuration.yAxisLabel;
        data.yAxis.nameGap = configuration.yAxisNameGap;
    }

    data.series.forEach((series) => {
        series.label.show = configuration.showValues;
    });

    applyLegendPosition(data.legend, configuration.legendPosition);

    return data;
}

export function applyLineChartConfiguration(data, configuration) {
    if (Object.keys(configuration).length === 0) {
        return data;
    }

    data.grid.bottom = configuration.bottomMargin;
    data.grid.top = configuration.topMargin;
    data.grid.left = configuration.leftMargin;
    data.grid.right = configuration.rightMargin;
    data.grid.containLabel = configuration.containLabel;
    data.legend.show = configuration.legend;
    data.dataZoom[0].show = configuration.dataZoom;
    data.dataZoom[1].show = configuration.dataZoom;
    if (configuration.stack) {
        data.series.forEach((series) => {
            series.stack = "stack";
        });
    } else {
        data.series.forEach((series) => {
            series.stack = "";
        });
    }

    data.xAxis.axisLabel.rotate = configuration.angle;
    data.xAxis.name = configuration.xAxisLabel;
    data.xAxis.nameGap = configuration.xAxisNameGap;
    data.yAxis.name = configuration.yAxisLabel;
    data.yAxis.nameGap = configuration.yAxisNameGap;

    applyLegendPosition(data.legend, configuration.legendPosition);

    return data;
}

function numberFormatter(value) {
    if (value >= 1e9) {
        return (value / 1e9).toFixed(1) + "B"; // Billions
    } else if (value >= 1e6) {
        return (value / 1e6).toFixed(1) + "M"; // Millions
    } else if (value >= 1e3) {
        return (value / 1e3).toFixed(1) + "k"; // Thousands
    }

    return value.toString(); // Small numbers
}

function incrementDate(date, granularity) {
    switch (granularity) {
        case "second":
            date.setUTCSeconds(date.getUTCSeconds() + 1);
            break;
        case "minute":
            date.setUTCMinutes(date.getUTCMinutes() + 1);
            break;
        case "hour":
            date.setUTCHours(date.getUTCHours() + 1);
            break;
        case "day":
            date.setUTCDate(date.getUTCDate() + 1);
            date.setUTCHours(0, 0, 0, 0); // Reset hours, minutes, seconds, and milliseconds
            break;
        case "week": {
            const dayOfWeek = date.getUTCDay(); // 0 (Sunday) to 6 (Saturday)
            const daysToMonday = (dayOfWeek === 0 ? -6 : 1) - dayOfWeek; // Adjust to Monday

            if (daysToMonday !== 0) {
                date.setUTCDate(date.getUTCDate() + daysToMonday); // Align to Monday
            }

            date.setUTCDate(date.getUTCDate() + 7); // Increment by 7 days
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        }

        case "month":
            date.setUTCMonth(date.getUTCMonth() + 1);
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        case "quarter":
            date.setUTCMonth(date.getUTCMonth() + 3);
            date.setUTCHours(0, 0, 0, 0); // Reset time components
            break;
        case "year":
            date.setUTCFullYear(date.getUTCFullYear() + 1);
            date.setUTCMonth(0, 1); // Reset to January 1
            date.setUTCHours(0, 0, 0, 0);
            break;
        default:
            throw new Error("Unsupported granularity");
    }
}

function generateTimeArray(startTime, endTime, granularity, withTime = true) {
    const startDate = new Date(startTime + "Z"); // Ensure UTC by adding 'Z'
    const endDate = new Date(endTime + "Z"); // Ensure UTC by adding 'Z'
    const times = [];

    for (
        let currentDate = new Date(startDate.getTime());
        currentDate <= endDate;
        incrementDate(currentDate, granularity)
    ) {
        if (withTime) {
            // Include full timestamp without the trailing 'Z'
            times.push(currentDate.toISOString().slice(0, -1));
        } else {
            // Include only the date part
            times.push(currentDate.toISOString().split("T")[0]);
        }
    }

    return times;
}

const SUB_DAY_GRANULARITIES = ["second", "minute", "hour"];

const AXIS_SERIES_PRESETS = {
    line: { type: "line", symbolSize: 8, smooth: true },
    bar: { type: "bar", stack: "" },
};

function measureTitleOf(data, measure) {
    return data.annotation.measures[measure]?.title || measure;
}

function valuesOf(item, dimensions) {
    return dimensions.map((dimension) =>
        item[dimension] != null ? String(item[dimension]) : "null",
    );
}

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
        .map((dimension) => (item[dimension] != null ? String(item[dimension]) : "null"));
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
        data: Array.from({ length: length }, () => 0),
        label: { show: false, position: "top", color: CHART_LABEL_COLOR },
        ...(AXIS_SERIES_PRESETS[seriesType] ?? AXIS_SERIES_PRESETS.line),
    };
}

function sortBarsByValue(option, direction) {
    if (direction !== "asc" && direction !== "desc") {
        return option;
    }

    const sign = direction === "asc" ? 1 : -1;

    if (option.xAxis.data.length > 1) {
        const totals = option.xAxis.data.map((_, index) =>
            option.series.reduce((sum, entry) => sum + (entry.data[index] ?? 0), 0),
        );
        const order = totals
            .map((_, index) => index)
            .sort((a, b) => sign * (totals[a] - totals[b]));

        option.xAxis.data = order.map((index) => option.xAxis.data[index]);
        option.series.forEach((entry) => {
            entry.data = order.map((index) => entry.data[index]);
        });

        return option;
    }

    option.series = [...option.series].sort((a, b) => sign * ((a.data[0] ?? 0) - (b.data[0] ?? 0)));
    option.legend.data = option.series.map((entry) => entry.name);

    return option;
}

export function axisSeriesChart(data, seriesType, configuration = {}) {
    const { measures, dimensions, timeDimensions } = data.query;
    const timeDimension = timeDimensions?.[0];
    const granularity = timeDimension?.granularity;
    const withTime = SUB_DAY_GRANULARITIES.includes(granularity);
    const showMeasure = measures.length > 1;

    let xAxisData = [];
    const series = [];
    const legend = [];

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
        xAxisData = generateTimeArray(
            timeDimension.dateRange[0],
            timeDimension.dateRange[1],
            granularity,
            withTime,
        );

        const timeKey = `${timeDimension.dimension}.${granularity}`;

        measures.forEach((measure) => {
            const title = measureTitleOf(data, measure);

            data.data.forEach((item) => {
                const name = seriesNameOf(
                    title,
                    dimensionValuesOf(item, dimensions, timeDimension),
                    showMeasure,
                );
                const raw = item[timeKey];
                const bucket = withTime ? raw : raw?.split("T")[0];
                const index = xAxisData.indexOf(bucket);

                if (index !== -1) {
                    seriesFor(name).data[index] += item[measure] ?? 0;
                }
            });
        });
    } else {
        const axisDimensions = axisDimensionsOf(configuration, dimensions);
        const seriesDimensions = dimensions.filter(
            (dimension) => !axisDimensions.includes(dimension),
        );

        data.data.forEach((item) => {
            const label = valuesOf(item, axisDimensions).join(", ");

            if (!xAxisData.includes(label)) {
                xAxisData.push(label);
            }
        });

        measures.forEach((measure) => {
            const title = measureTitleOf(data, measure);

            data.data.forEach((item) => {
                const name = seriesNameOf(title, valuesOf(item, seriesDimensions), showMeasure);
                const index = xAxisData.indexOf(valuesOf(item, axisDimensions).join(", "));

                if (index !== -1) {
                    seriesFor(name).data[index] += Number(item[measure]) || 0;
                }
            });
        });
    }

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
            axisLabel: { formatter: numberFormatter },
        },
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

export function lineChart(data, configuration) {
    return axisSeriesChart(data, "line", configuration);
}

export function barChart(data, configuration) {
    return axisSeriesChart(data, "bar", configuration);
}

export function timeBarChart(data, configuration) {
    return axisSeriesChart(data, "bar", configuration);
}

export function pieChart(data) {
    const measures = data.query.measures;
    const dimensions = data.query.dimensions;
    const legend = new Set();
    const seriesData = [];

    if (dimensions.length === 0 || measures.length === 0) {
        console.warn("No dimensions or measures provided.");

        return {};
    }

    const primaryDimension = dimensions[0];
    const measure = measures[0];
    const measureTitle = measureTitleOf(data, measure);

    data.data.forEach((item) => {
        const name = `${item[primaryDimension] ?? "null"}`;

        legend.add(name);
        seriesData.push({ value: item[measure] ?? 0, name: name });
    });

    const option = {
        color: CHART_PALETTE,
        tooltip: {
            trigger: "item",
            formatter: "{a} <br/>{b}: {c} ({d}%)",
        },
        legend: {
            type: "scroll",
            show: true,
            data: Array.from(legend),
            ...LEGEND_POSITIONS[DEFAULT_LEGEND_POSITION],
        },
        series: [
            {
                name: measureTitle,
                type: "pie",
                radius: ["0%", "55%"],
                label: {
                    //red
                    show: true,
                    formatter: null,
                    position: "outside",
                },
                labelLine: {
                    show: true,
                    formatter: null,
                },
                data: seriesData,
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: "rgba(0, 0, 0, 0.5)",
                    },
                },
            },
        ],
    };

    return option;
}

export function toTableData(data) {
    let tableData = {
        headers: [],
        data: [],
    };

    const reg = /\./g;

    // Add time dimensions to headers if they exist
    if (data.query.timeDimensions?.length > 0) {
        data.query.timeDimensions.forEach((timeDimension) => {
            if (data.annotation.timeDimensions[timeDimension.dimension]) {
                tableData.headers.push({
                    label: data.annotation.timeDimensions[timeDimension.dimension].title,
                    key: timeDimension.dimension.replace(reg, "_"),
                });
            }
        });
    }

    // Add measures to headers
    data.query.measures.forEach((measure) => {
        tableData.headers.push({
            label: data.annotation.measures[measure].title,
            key: measure.replace(reg, "_"),
        });
    });

    // Add dimensions to headers
    data.query.dimensions.forEach((dimension) => {
        tableData.headers.push({
            label: data.annotation.dimensions[dimension].title,
            key: dimension.replace(reg, "_"),
        });
    });

    // Populate data rows
    data.data.forEach((item) => {
        let row = {};

        data.query.measures.forEach((measure) => {
            row[measure.replace(reg, "_")] = item[measure];
        });

        data.query.dimensions.forEach((dimension) => {
            row[dimension.replace(reg, "_")] = item[dimension];
        });

        if (data.query.timeDimensions?.length > 0) {
            data.query.timeDimensions.forEach((timeDimension) => {
                row[timeDimension.dimension.replace(reg, "_")] = item[timeDimension.dimension];
            });
        }

        tableData.data.push(row);
    });

    return tableData;
}

export function number(data) {
    const values = data.query.measures.map((measure) => {
        const total = data.data.reduce((sum, item) => sum + (Number(item[measure]) || 0), 0);

        return {
            label: measureTitleOf(data, measure),
            formatted: formatNumber(total),
            value: total,
        };
    });

    return { values: values };
}

function formatNumber(num) {
    if (num >= 1e21) return (num / 1e21).toFixed(1) + "S"; // Sextillion
    if (num >= 1e18) return (num / 1e18).toFixed(1) + "Qi"; // Quintillion
    if (num >= 1e15) return (num / 1e15).toFixed(1) + "Q"; // Quadrillion
    if (num >= 1e12) return (num / 1e12).toFixed(1) + "T"; // Trillion
    if (num >= 1e9) return (num / 1e9).toFixed(1) + "B"; // Billion
    if (num >= 1e6) return (num / 1e6).toFixed(1) + "M"; // Million
    if (num >= 1e3) return (num / 1e3).toFixed(1) + "K"; // Thousand

    // Return number below 1000 as a whole number without leading zeros
    return Number(num).toString();
}
