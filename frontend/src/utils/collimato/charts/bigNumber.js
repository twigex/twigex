// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { measureTitleOf } from "@/utils/collimato/charts/chartLabels.js";

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
            formatted: total == null ? "-" : formatNumber(total),
            value: total,
        };
    });

    return { values: values };
}

function formatNumber(num) {
    if (num >= 1e21) return (num / 1e21).toFixed(1) + "S"; // Sextillion
    if (num >= 1e18) return (num / 1e18).toFixed(1) + "Qi"; // Quintillion
    if (num >= 1e15) return (num / 1e15).toFixed(1) + "Q"; // Quadrillion
    if (num >= 1e12) return (num / 1e12).toFixed(1) + "T"; // Trillion
    if (num >= 1e9) return (num / 1e9).toFixed(1) + "B"; // Billion
    if (num >= 1e6) return (num / 1e6).toFixed(1) + "M"; // Million
    if (num >= 1e3) return (num / 1e3).toFixed(1) + "K"; // Thousand

    // Return number below 1000 as a whole number without leading zeros
    return Number(num).toString();
}
