// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function formatTooltip(fields, row) {
    if (!Array.isArray(fields) || fields.length === 0 || !row) {
        return null;
    }

    return fields.map((field) => `${field}: ${row[field]}`).join("\n");
}
