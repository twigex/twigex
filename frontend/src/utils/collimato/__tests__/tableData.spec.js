// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { toTableData } from "@/utils/collimato/tableData.js";

describe("toTableData", () => {
    it("lays out date fields, then totals, then fields, with their titles", () => {
        const data = {
            query: {
                measures: ["orders.revenue"],
                dimensions: ["orders.country"],
                timeDimensions: [{ dimension: "orders.ordered_at", granularity: "month" }],
            },
            columns: [
                { key: "orders.ordered_at", type: "time" },
                { key: "orders.revenue", type: "number" },
                { key: "orders.country", type: "string" },
            ],
            titles: {
                "orders.ordered_at": "Orders Order date",
                "orders.revenue": "Orders Revenue",
                "orders.country": "Orders Country",
            },
            rows: [
                {
                    "orders.ordered_at": "2025-11-01T00:00:00Z",
                    "orders.revenue": 43412.27,
                    "orders.country": "Latvia",
                },
            ],
        };

        expect(toTableData(data)).toEqual({
            headers: [
                { label: "Orders Order date", key: "orders_ordered_at" },
                { label: "Orders Revenue", key: "orders_revenue" },
                { label: "Orders Country", key: "orders_country" },
            ],
            data: [
                {
                    orders_ordered_at: "2025-11-01T00:00:00.000",
                    orders_revenue: 43412.27,
                    orders_country: "Latvia",
                },
            ],
        });
    });

    it("leaves out a date field that only filters", () => {
        const data = {
            query: {
                measures: ["orders.count"],
                dimensions: [],
                timeDimensions: [{ dimension: "orders.ordered_at", dateRange: "this year" }],
            },
            columns: [{ key: "orders.count", type: "number" }],
            titles: { "orders.count": "Orders Count" },
            rows: [{ "orders.count": 5 }],
        };

        expect(toTableData(data).headers).toEqual([{ label: "Orders Count", key: "orders_count" }]);
    });

    it("shows yes/no values as Yes and No", () => {
        const data = {
            query: {
                measures: ["orders.count"],
                dimensions: ["orders.is_large"],
                timeDimensions: [],
            },
            columns: [
                { key: "orders.count", type: "number" },
                { key: "orders.is_large", type: "boolean" },
            ],
            titles: { "orders.count": "Orders Count", "orders.is_large": "Orders Large order" },
            rows: [
                { "orders.count": 12, "orders.is_large": true },
                { "orders.count": 30, "orders.is_large": false },
            ],
        };

        expect(toTableData(data).data).toEqual([
            { orders_count: 12, orders_is_large: "Yes" },
            { orders_count: 30, orders_is_large: "No" },
        ]);
    });
});
