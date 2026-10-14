// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

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

export const MISSING_VALUES_CONNECT = "connect";

export const MISSING_VALUES_GAP = "gap";

export const MISSING_VALUES_ZERO = "zero";

export const MISSING_VALUES_OPTIONS = [
    MISSING_VALUES_CONNECT,
    MISSING_VALUES_GAP,
    MISSING_VALUES_ZERO,
];

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

function fillMissingWithZero(entry) {
    entry.data = entry.data.map((value) => {
        if (value === null) {
            return 0;
        }

        return value;
    });
}

function missingValuesOf(configuration) {
    // Charts saved before this setting existed have no value and
    // must keep connecting.
    if (!configuration || !configuration.missingValues) {
        return MISSING_VALUES_CONNECT;
    }

    return configuration.missingValues;
}

export function applyMissingValues(series, configuration) {
    const missingValues = missingValuesOf(configuration);

    series.forEach((entry) => {
        if (missingValues === MISSING_VALUES_ZERO) {
            fillMissingWithZero(entry);
        }

        entry.connectNulls = missingValues === MISSING_VALUES_CONNECT;
    });
}

export function sortBarsByValue(option, direction) {
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
