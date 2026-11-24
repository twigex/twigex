// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    fetchChartData,
    filterValueOf,
    slotTime,
    toChartData,
    withoutDateAxisInFields,
} from "@/utils/collimato/chartData.js";

const answer = {
    columns: [
        {
            key: "orders.ordered_at",
            title: "Order date",
            fullTitle: "Orders Order date",
            kind: "dimension",
            type: "time",
            dateStep: "month",
        },
        {
            key: "orders.revenue",
            title: "Revenue",
            fullTitle: "Orders Revenue",
            kind: "measure",
            type: "number",
        },
    ],
    rows: [{ "orders.ordered_at": "2025-11-01T00:00:00Z", "orders.revenue": 43412.27 }],
    dateRanges: [
        {
            field: "orders.ordered_at",
            start: "2025-01-01T00:00:00Z",
            end: "2025-12-31T23:59:59.999Z",
        },
    ],
};

const query = {
    measures: ["orders.revenue"],
    dimensions: [],
    timeDimensions: [
        { dimension: "orders.ordered_at", granularity: "month", dateRange: "this year" },
    ],
    filters: [],
    limit: 100,
};

describe("toChartData", () => {
    it("puts the worked out date range into the query, as the time axis names times", () => {
        const data = toChartData(query, answer);

        expect(data.query.timeDimensions).toEqual([
            {
                dimension: "orders.ordered_at",
                granularity: "month",
                dateRange: ["2025-01-01T00:00:00.000", "2025-12-31T23:59:59.999"],
            },
        ]);
    });

    it("titles each field with its dataset's title in front", () => {
        const data = toChartData(query, answer);

        expect(data.titles).toEqual({
            "orders.ordered_at": "Orders Order date",
            "orders.revenue": "Orders Revenue",
        });
    });

    it("keeps the rows as the server sent them", () => {
        expect(toChartData(query, answer).rows).toBe(answer.rows);
    });

    it("leaves a date field without a range as it was asked for", () => {
        const withoutRange = { ...answer, dateRanges: [] };
        const data = toChartData(query, withoutRange);

        expect(data.query.timeDimensions[0].dateRange).toBe("this year");
    });
});

describe("slotTime", () => {
    it("writes a time in UTC without the zone", () => {
        expect(slotTime("2025-03-02T10:00:00Z")).toBe("2025-03-02T10:00:00.000");
    });

    it("reads a time without a zone as UTC", () => {
        expect(slotTime("2025-03-02T10:00:00.000")).toBe("2025-03-02T10:00:00.000");
    });
});

describe("filterValueOf", () => {
    it("writes numbers and yes/no as text, and keeps empty", () => {
        expect(filterValueOf(1205)).toBe("1205");
        expect(filterValueOf(true)).toBe("true");
        expect(filterValueOf("Latvia")).toBe("Latvia");
        expect(filterValueOf(null)).toBe(null);
    });
});

describe("fetchChartData", () => {
    it("gives the answer as chart data", async () => {
        const data = await fetchChartData(() => Promise.resolve({ data: answer }), query);

        expect(data.rows).toEqual(answer.rows);
    });

    it("stops waiting when cancelled, even if the request never ends", async () => {
        const controller = new AbortController();
        const loading = fetchChartData(() => new Promise(() => {}), query, {
            signal: controller.signal,
        });

        controller.abort();

        await expect(loading).rejects.toThrow("Aborted");
    });
});

describe("withoutDateAxisInFields", () => {
    it("leaves a date field on the axis out of the plain fields", () => {
        const both = {
            measures: ["orders.count"],
            dimensions: ["orders.ordered_at", "orders.country"],
            timeDimensions: [{ dimension: "orders.ordered_at", granularity: "month" }],
        };

        expect(withoutDateAxisInFields(both).dimensions).toEqual(["orders.country"]);
    });

    it("keeps a date field that only filters by its range", () => {
        const filterOnly = {
            dimensions: ["orders.ordered_at"],
            timeDimensions: [{ dimension: "orders.ordered_at", dateRange: "this year" }],
        };

        expect(withoutDateAxisInFields(filterOnly).dimensions).toEqual(["orders.ordered_at"]);
    });

    it("is what fetchChartData asks for", async () => {
        const asked = [];
        const both = {
            measures: ["orders.count"],
            dimensions: ["orders.ordered_at"],
            timeDimensions: [{ dimension: "orders.ordered_at", granularity: "month" }],
        };

        await fetchChartData((q) => {
            asked.push(q);

            return Promise.resolve({ data: { columns: [], rows: [], dateRanges: [] } });
        }, both);

        expect(asked[0].dimensions).toEqual([]);
    });
});
