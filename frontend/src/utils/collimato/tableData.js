// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { slotTime } from "@/utils/collimato/chartData.js";
import { yesNoLabel } from "@/utils/collimato/yesNo.js";

// toTableData lays a chart's data out as a table: date fields, then totals,
// then the other fields, each headed by its title. Dates are shown as the
// time axis names them, in UTC, and yes/no values as Yes and No.
export function toTableData(data) {
    let tableData = {
        headers: [],
        data: [],
    };

    const reg = /\./g;

    const dateFields = data.query.timeDimensions
        .map((timeDimension) => timeDimension.dimension)
        .filter((dimension) => Object.hasOwn(data.titles, dimension));

    const fields = [...dateFields, ...data.query.measures, ...data.query.dimensions];

    fields.forEach((field) => {
        tableData.headers.push({
            label: data.titles[field] ?? field,
            key: field.replace(reg, "_"),
        });
    });

    const timeFields = new Set(
        data.columns.filter((column) => column.type === "time").map((column) => column.key),
    );

    data.rows.forEach((item) => {
        let row = {};

        fields.forEach((field) => {
            const value = item[field];
            const isTime = timeFields.has(field) && value != null;
            const isYesNo = typeof value === "boolean";

            let shown = value;

            if (isTime) {
                shown = slotTime(value);
            }

            if (isYesNo) {
                shown = yesNoLabel(value);
            }

            row[field.replace(reg, "_")] = shown;
        });

        tableData.data.push(row);
    });

    return tableData;
}
