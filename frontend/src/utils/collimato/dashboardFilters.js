// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function filterApplies(filter, chartId) {
    const scope = filter?.apply_to ?? [];

    return scope.length === 0 || scope.some((id) => id === chartId);
}

export function filtersForChart(filters, chartId) {
    return (filters ?? []).filter((filter) => filterApplies(filter, chartId));
}

export function queryFilterFor(filter, values) {
    return {
        member: filter.column,
        operator: filter.operator || "equals",
        values: values,
    };
}

export function withFilterApplied(queryFilters, filter, values) {
    const rest = (queryFilters ?? []).filter((existing) => existing.member !== filter.column);

    return values?.length ? [...rest, queryFilterFor(filter, values)] : rest;
}

export function withFilterRemoved(queryFilters, filter) {
    return (queryFilters ?? []).filter((existing) => existing.member !== filter.column);
}
