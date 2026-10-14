// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function cubeResult({ measures, dimensions, timeDimensions, rows, titles }) {
    return {
        query: {
            measures,
            dimensions,
            timeDimensions: timeDimensions ?? [],
        },
        annotation: {
            measures: Object.fromEntries(measures.map((m) => [m, { title: titles?.[m] ?? m }])),
            dimensions: Object.fromEntries(dimensions.map((d) => [d, { title: d }])),
            timeDimensions: {},
        },
        data: rows,
    };
}
