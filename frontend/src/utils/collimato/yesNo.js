// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index";

const YES_TEXTS = ["1", "true", "t"];
const NO_TEXTS = ["0", "false", "f"];

// yesNoLabel names a yes/no value in the reader's language. Rows hold it as
// true or false; filters hold it as text, and some saved filters as "1" or
// "0". Anything else is given back as it is.
export function yesNoLabel(value) {
    const text = String(value).toLowerCase();

    if (YES_TEXTS.includes(text)) {
        return t.value("collimato.charts.yes");
    }

    if (NO_TEXTS.includes(text)) {
        return t.value("collimato.charts.no");
    }

    return value;
}

// yesNoFieldsOf gives the keys of the yes/no fields in the data models.
export function yesNoFieldsOf(dataModels) {
    const keys = new Set();

    (dataModels ?? []).forEach((dataModel) => {
        dataModel.dimensions.forEach((dimension) => {
            if (dimension.type === "boolean") {
                keys.add(dimension.name);
            }
        });
    });

    return keys;
}
