// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { modelNameProblems, withDatabaseNames } from "@/utils/collimato/modelNames.js";

const tables = [
    {
        name: "orders",
        columns: [{ name: "id" }, { name: "orderdate" }, { name: "country" }, { name: "amount" }],
    },
];

const typedByHand = {
    name: "orders",
    sql_table: "Orders",
    connection_id: "c1",
    dimensions: [
        { name: "id", sql: "id", type: "number", primary_key: true },
        { name: "OrderDate", sql: "OrderDate", type: "time" },
        { name: "region", sql: "region", type: "string" },
        { name: "label", sql: "CASE WHEN country = 'LV' THEN 1 END", type: "number", custom: true },
    ],
    measures: [
        { name: "count", type: "count" },
        { name: "revenue", type: "sum", sql: "Amount" },
    ],
};

describe("modelNameProblems", () => {
    it("finds names Postgres has only in another case, and names it does not have", () => {
        expect(modelNameProblems(typedByHand, tables, "postgres")).toEqual([
            { kind: "table", saved: "Orders", found: "orders" },
            { kind: "column", field: "OrderDate", saved: "OrderDate", found: "orderdate" },
            { kind: "column", field: "region", saved: "region", found: null },
            { kind: "total", field: "revenue", saved: "Amount", found: "amount" },
        ]);
    });

    it("on MySQL, which ignores case in column names, flags only missing columns", () => {
        expect(modelNameProblems(typedByHand, tables, "mysql")).toEqual([
            { kind: "table", saved: "Orders", found: "orders" },
            { kind: "column", field: "region", saved: "region", found: null },
        ]);
    });

    it("finds nothing for a model with the database's names", () => {
        const model = {
            ...typedByHand,
            sql_table: "orders",
            dimensions: [{ name: "id", sql: "id", type: "number" }],
            measures: [{ name: "revenue", type: "sum", sql: "amount" }],
        };

        expect(modelNameProblems(model, tables, "postgres")).toEqual([]);
    });

    it("does not look at columns of a table the database does not have", () => {
        const model = { ...typedByHand, sql_table: "missing" };

        expect(modelNameProblems(model, tables, "postgres")).toEqual([
            { kind: "table", saved: "missing", found: null },
        ]);
    });
});

describe("withDatabaseNames", () => {
    it("uses the database's names and keeps each field's own name", () => {
        const fixed = withDatabaseNames(
            typedByHand,
            modelNameProblems(typedByHand, tables, "postgres"),
        );

        expect(fixed.sql_table).toBe("orders");
        expect(fixed.dimensions[1]).toEqual({
            name: "OrderDate",
            sql: "orderdate",
            type: "time",
            custom: true,
        });
        expect(fixed.dimensions[2].sql).toBe("region");
        expect(fixed.measures[1]).toEqual({ name: "revenue", type: "sum", sql: "amount" });
        expect(typedByHand.sql_table).toBe("Orders");
    });
});
