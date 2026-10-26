// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { CHART_PALETTE } from "@/utils/collimato/charts/chartPalette.js";
import { recordDrillDownValues } from "@/utils/collimato/charts/drillDown.js";
import {
    LEGEND_POSITIONS,
    DEFAULT_LEGEND_POSITION,
} from "@/utils/collimato/charts/chartSettings.js";
import { labelOf, measureTitleOf } from "@/utils/collimato/charts/chartLabels.js";

export function pieChart(data) {
    const measures = data.query.measures;
    const dimensions = data.query.dimensions;
    const legend = new Set();
    const seriesData = [];
    const drillDownValues = { name: {} };

    if (dimensions.length === 0 || measures.length === 0) {
        console.warn("No dimensions or measures provided.");

        return {};
    }

    const primaryDimension = dimensions[0];
    const measure = measures[0];
    const measureTitle = measureTitleOf(data, measure);

    data.data.forEach((item) => {
        const value = item[measure];

        if (value == null || value === "") {
            return;
        }

        const name = labelOf(item[primaryDimension]);
        const slice = seriesData.find((entry) => entry.name === name);

        recordDrillDownValues(drillDownValues.name, name, item, [primaryDimension]);

        if (slice) {
            slice.value += Number(value);

            return;
        }

        legend.add(name);
        seriesData.push({ value: Number(value), name: name });
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
        drillDownValues: drillDownValues,
    };

    return option;
}
