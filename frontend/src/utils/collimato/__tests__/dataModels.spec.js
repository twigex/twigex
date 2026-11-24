// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { toDataModels } from "@/utils/collimato/dataModels.js";

describe("toDataModels", () => {
    it("lists datasets with their dimensions and totals, each field named by its key", () => {
        const fieldList = {
            datasets: [
                {
                    name: "orders",
                    title: "Orders",
                    isView: false,
                    joinGroup: 1,
                    fields: [
                        {
                            key: "orders.country",
                            title: "Country",
                            fullTitle: "Orders Country",
                            kind: "dimension",
                            type: "string",
                        },
                        {
                            key: "orders.count",
                            title: "Orders",
                            fullTitle: "Orders Orders",
                            kind: "measure",
                            type: "number",
                            measureType: "count",
                        },
                    ],
                },
            ],
            modelErrors: [],
        };

        expect(toDataModels(fieldList)).toEqual([
            {
                name: "orders",
                table: "orders",
                title: "Orders",
                isView: false,
                join: 1,
                dimensions: [
                    {
                        name: "orders.country",
                        title: "Country",
                        fullTitle: "Orders Country",
                        type: "string",
                    },
                ],
                measures: [
                    {
                        name: "orders.count",
                        title: "Orders",
                        fullTitle: "Orders Orders",
                        type: "number",
                    },
                ],
            },
        ]);
    });
});
