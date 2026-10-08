// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// What decides a filter's result, and nothing else: a saved filter comes back
// from the server without the options the builder shows, and with its keys in
// another order, so two filters match when their keys do.
export function filterConditionsKey(payload) {
    const pick = (f) => [
        f.field,
        f.operator,
        JSON.stringify(f.values ?? f.value ?? ""),
        f.operatorBetween || "AND",
        f.date || "",
    ];

    return JSON.stringify({
        flat: (payload?.flatFilters || []).map(pick),
        groups: (payload?.groups || []).map((g) => [
            (g.filters || []).map(pick),
            g.relationToNext || null,
        ]),
    });
}

// Drops the option lists the filter editor attaches to filters. The server
// finds a field's options itself, and every status of every table would
// otherwise travel with each request, and in the report's CSV link.
export function withoutFilterOptions(payload) {
    const strip = ({ linkedOptions, ...filter }) => filter;

    return {
        ...payload,
        flatFilters: (payload?.flatFilters || []).map(strip),
        groups: (payload?.groups || []).map((group) => ({
            ...group,
            filters: (group.filters || []).map(strip),
        })),
    };
}

const VALUELESS_OPERATORS = new Set(["is_set", "is_not_set"]);

// The rows of a flat filter that say something: a field, an operator, and a
// value unless the operator takes none. A row still being filled in is left
// out rather than sent half made.
export function appliedFlatFilters(flatFilters) {
    return (flatFilters || []).filter(
        (f) =>
            f.field &&
            f.operator &&
            (VALUELESS_OPERATORS.has(f.operator) ||
                (f.value !== "" && f.value !== null && f.value !== undefined)),
    );
}

// filtersForSave is the filter as it is saved: the rows that say something,
// by the same rule as when it is applied, and on the report a row's list of
// values sent as values. A group's relation to the next is kept unless it is
// the last group.
export function filtersForSave(flatFilters, groups, { report = false } = {}) {
    const row = (f) => {
        if (!report || !Array.isArray(f.value)) return { ...f };
        const { value, ...rest } = f;

        return { ...rest, values: value };
    };

    const kept = (groups || []).filter((g) => g.filters.length > 0);

    return {
        flatFilters: appliedFlatFilters(flatFilters).map(row),
        groups: kept.map((g, i) => ({
            filters: appliedFlatFilters(g.filters).map(row),
            relationToNext: i < kept.length - 1 ? g.nextRelation : null,
        })),
    };
}
