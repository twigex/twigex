// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index";

export function measureTitleOf(data, measure) {
    return data.annotation.measures[measure]?.title || measure;
}

export function labelOf(value) {
    if (value == null || value === "") {
        return t.value("collimato.charts.empty_value");
    }

    return String(value);
}

export function valuesOf(item, dimensions) {
    return dimensions.map((dimension) => labelOf(item[dimension]));
}

const SHORT_NUMBER_STEPS = [
    { size: 1e21, suffix: "S" }, // Sextillion
    { size: 1e18, suffix: "Qi" }, // Quintillion
    { size: 1e15, suffix: "Q" }, // Quadrillion
    { size: 1e12, suffix: "T" }, // Trillion
    { size: 1e9, suffix: "B" }, // Billion
    { size: 1e6, suffix: "M" }, // Million
    { size: 1e3, suffix: "K" }, // Thousand
];

export function shortNumber(value) {
    const size = Math.abs(value);
    const step = SHORT_NUMBER_STEPS.find((candidate) => size >= candidate.size);

    if (!step) {
        return Number(value).toString();
    }

    let shortened = (value / step.size).toFixed(1);

    if (shortened.endsWith(".0")) {
        shortened = shortened.slice(0, -2);
    }

    return shortened + step.suffix;
}
