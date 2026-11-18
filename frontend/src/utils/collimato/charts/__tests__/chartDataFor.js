// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function chartDataFor({ measures, dimensions, timeDimensions, rows, titles }) {
    return {
        query: {
            measures,
            dimensions,
            timeDimensions: timeDimensions ?? [],
        },
        columns: [],
        titles: Object.fromEntries(measures.map((m) => [m, titles?.[m] ?? m])),
        rows,
    };
}
