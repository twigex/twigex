// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { measureTitleOf, shortNumber } from "@/utils/collimato/charts/chartLabels.js";

export function number(data) {
    const values = data.query.measures.map((measure) => {
        const total = data.data.reduce((sum, item) => {
            if (item[measure] == null || item[measure] === "") {
                return sum;
            }

            return sum + Number(item[measure]);
        }, null);

        return {
            label: measureTitleOf(data, measure),
            formatted: total == null ? "-" : shortNumber(total),
            value: total,
        };
    });

    return { values: values };
}
