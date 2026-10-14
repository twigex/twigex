// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { drillDownKeyFor } from "@/utils/collimato/charts/chartTypes.js";

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
