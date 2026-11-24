// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

function fieldsOfKind(dataset, kind) {
    return dataset.fields
        .filter((field) => field.kind === kind)
        .map((field) => ({
            name: field.key,
            title: field.title,
            fullTitle: field.fullTitle,
            type: field.type,
        }));
}

// toDataModels lists the server's field list the way the chart, dashboard and
// role screens use it: each dataset or view by name (also as table), its join
// group, and its dimensions and measures, each field named by its key.
export function toDataModels(fieldList) {
    return fieldList.datasets.map((dataset) => ({
        name: dataset.name,
        table: dataset.name,
        title: dataset.title,
        isView: dataset.isView,
        join: dataset.joinGroup,
        dimensions: fieldsOfKind(dataset, "dimension"),
        measures: fieldsOfKind(dataset, "measure"),
    }));
}
