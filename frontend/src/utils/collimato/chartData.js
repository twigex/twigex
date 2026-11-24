// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The time axis names each slot in UTC without the zone, with milliseconds:
// 2025-03-01T00:00:00.000 (see timeSlots.js). A time already written without
// a zone is in UTC too.
export function slotTime(value) {
    const hasZone = /(Z|[+-]\d\d:\d\d)$/.test(value);
    const utcValue = hasZone ? value : value + "Z";

    return new Date(utcValue).toISOString().slice(0, -1);
}

// filterValueOf writes a value from a row as a filter takes it: as text, so
// a number 1205 or a yes/no true filters as "1205" or "true". Empty stays
// empty.
export function filterValueOf(value) {
    if (value == null) {
        return value;
    }

    return String(value);
}

// toChartData puts the server's answer together with the query that asked for
// it, in the form the charts draw from: the query with each date field's
// range worked out (the days "this month" stands for), the rows, and each
// field's title.
export function toChartData(query, answer) {
    const titles = {};

    answer.columns.forEach((column) => {
        titles[column.key] = column.fullTitle;
    });

    const ranges = {};

    answer.dateRanges.forEach((range) => {
        ranges[range.field] = [slotTime(range.start), slotTime(range.end)];
    });

    const timeDimensions = (query.timeDimensions ?? []).map((timeDimension) => {
        const dateRange = ranges[timeDimension.dimension];

        return dateRange ? { ...timeDimension, dateRange } : { ...timeDimension };
    });

    return {
        query: {
            measures: query.measures ?? [],
            dimensions: query.dimensions ?? [],
            timeDimensions,
            filters: query.filters ?? [],
        },
        columns: answer.columns,
        rows: answer.rows,
        titles,
    };
}

function whenAborted(signal) {
    return new Promise((_, reject) => {
        if (signal.aborted) {
            reject(new DOMException("Aborted", "AbortError"));

            return;
        }

        signal.addEventListener("abort", () => reject(new DOMException("Aborted", "AbortError")));
    });
}

// withoutDateAxisInFields leaves out of the plain fields a date field that is
// also grouped by a step: the date axis shows it. Some saved charts ask for it
// both ways, and the server refuses a field asked for twice.
export function withoutDateAxisInFields(query) {
    const axisFields = (query.timeDimensions ?? [])
        .filter((timeDimension) => timeDimension.granularity)
        .map((timeDimension) => timeDimension.dimension);

    if (axisFields.length === 0) {
        return query;
    }

    const dimensions = (query.dimensions ?? []).filter((field) => !axisFields.includes(field));

    return { ...query, dimensions };
}

// fetchChartData asks the server for a query's data, through loadData(query,
// signal), and gives it as toChartData does. Aborting the signal stops the
// wait at once, even if the request does not end.
export async function fetchChartData(loadData, query, { signal } = {}) {
    const asked = withoutDateAxisInFields(query);
    const request = loadData(asked, signal);
    const response = signal ? await Promise.race([request, whenAborted(signal)]) : await request;

    return toChartData(asked, response.data);
}
