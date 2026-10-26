// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { drillDownKeyFor } from "@/utils/collimato/charts/chartTypes.js";
import { labelOf } from "@/utils/collimato/charts/chartLabels.js";

//Returns values for drillDownValues for options in axisSeriesChart without empty value translation
function valuesOfRaw(item, dimensions) {
    return dimensions.map((dimension) => item[dimension] ?? null);
}

export function recordDrillDownValues(known, label, item, dimensions) {
    if (!Object.hasOwn(known, label)) {
        known[label] = {};
    }

    const entry = known[label];
    const values = valuesOfRaw(item, dimensions);

    dimensions.forEach((dimension, index) => {
        const value = values[index];

        if (!entry[dimension]) {
            entry[dimension] = [];
        }

        if (!entry[dimension].includes(value)) {
            entry[dimension].push(value);
        }
    });
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

function breadcrumbName(dimensions, chartName) {
    if (dimensions.length === 0) {
        return chartName;
    }

    return dimensions.map((dimension) => dimension.split(".").at(-1)).join(", ");
}

function filtersAddedSince(previous, query) {
    const before = (previous?.filters ?? []).map((filter) => JSON.stringify(filter));

    return (query.filters ?? []).filter((filter) => !before.includes(JSON.stringify(filter)));
}

function breadcrumbValue(filters) {
    const values = filters.flatMap((filter) => filterValues(filter));
    const labels = values.map((value) => labelOf(value));
    const unique = [];

    labels.forEach((label) => {
        if (!unique.includes(label)) {
            unique.push(label);
        }
    });

    return unique.join(", ");
}

export function drillDownBreadcrumbs(trail, chartName) {
    return trail.map((query, index) => {
        const previous = trail[index - 1];
        const added = filtersAddedSince(previous, query);

        return {
            query: query,
            name: breadcrumbName(query.dimensions ?? [], chartName),
            value: breadcrumbValue(added),
        };
    });
}
