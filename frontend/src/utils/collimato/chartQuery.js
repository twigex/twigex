// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";
import { hasTimeAxis } from "@/utils/collimato/chartTypes.js";

export const ORDER_OPTIONS = ["asc", "desc", "none"];
export const ORDER_NONE = "none";

export function mergeFilters(filters) {
    const merged = [];

    filters.forEach((filter) => {
        const values = Array.isArray(filter.value) ? [...filter.value] : [filter.value];

        const existing = merged.find(
            (f) => f.member == filter.member && f.operator == filter.operator,
        );

        if (existing) {
            existing.values.push(...values);

            return;
        }

        merged.push({
            member: filter.member,
            operator: filter.operator,
            values: values,
        });
    });

    return merged;
}

export function buildQuery({
    measures = [],
    dimensions = [],
    orders = [],
    filters = [],
    time = {},
    chartType = "",
}) {
    const query = {
        measures: measures.map((field) => field.name),
        dimensions: dimensions.map((field) => field.name),
        filters: mergeFilters(filters),
        timeDimensions: [],
        order: orders
            .filter((entry) => entry.direction != ORDER_NONE)
            .map((entry) => [entry.name, entry.direction]),
        limit: MAX_QUERY_LIMIT,
    };

    if (time.dimension && hasTimeAxis(chartType)) {
        query.timeDimensions.push({
            dimension: time.dimension.name,
            dateRange: time.dateRange != "custom" ? time.dateRange : [time.startTime, time.endTime],
            granularity: time.granularity ? time.granularity : null,
        });
    }

    return query;
}

export function joinableDataModels(dataModels, selectedModel) {
    if (!selectedModel) {
        return [];
    }

    if (!(selectedModel.join > 0)) {
        return [selectedModel];
    }

    return [
        selectedModel,
        ...dataModels.filter(
            (model) => model.table != selectedModel.table && model.join == selectedModel.join,
        ),
    ];
}
