// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function toTableData(data) {
    let tableData = {
        headers: [],
        data: [],
    };

    const reg = /\./g;

    // Add time dimensions to headers if they exist
    if (data.query.timeDimensions?.length > 0) {
        data.query.timeDimensions.forEach((timeDimension) => {
            if (data.annotation.timeDimensions[timeDimension.dimension]) {
                tableData.headers.push({
                    label: data.annotation.timeDimensions[timeDimension.dimension].title,
                    key: timeDimension.dimension.replace(reg, "_"),
                });
            }
        });
    }

    // Add measures to headers
    data.query.measures.forEach((measure) => {
        tableData.headers.push({
            label: data.annotation.measures[measure].title,
            key: measure.replace(reg, "_"),
        });
    });

    // Add dimensions to headers
    data.query.dimensions.forEach((dimension) => {
        tableData.headers.push({
            label: data.annotation.dimensions[dimension].title,
            key: dimension.replace(reg, "_"),
        });
    });

    // Populate data rows
    data.data.forEach((item) => {
        let row = {};

        data.query.measures.forEach((measure) => {
            row[measure.replace(reg, "_")] = item[measure];
        });

        data.query.dimensions.forEach((dimension) => {
            row[dimension.replace(reg, "_")] = item[dimension];
        });

        if (data.query.timeDimensions?.length > 0) {
            data.query.timeDimensions.forEach((timeDimension) => {
                row[timeDimension.dimension.replace(reg, "_")] = item[timeDimension.dimension];
            });
        }

        tableData.data.push(row);
    });

    return tableData;
}
