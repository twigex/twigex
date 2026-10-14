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
