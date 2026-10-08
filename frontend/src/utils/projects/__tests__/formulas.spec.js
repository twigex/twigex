// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import {
    calculateRowValueForRow,
    formatCellValue,
    columnTotal,
    formulaArgsFor,
    isComputedCalculations,
} from "../formulas";

const calc = (formula, extra = {}) => ({
    name: "total",
    header_usage: "calculations",
    formula,
    ...extra,
});
const fields = (...names) => names.map((value) => ({ type: "field", value }));

describe("calculateRowValueForRow", () => {
    it("adds the fields a SUM names", () => {
        const header = calc({ name: "SUM", args: fields("hours", "extra") });

        expect(calculateRowValueForRow(header, { hours: "2.5", extra: 4 })).toBe(6.5);
    });

    it("skips a field that holds no number", () => {
        const header = calc({ name: "SUM", args: fields("hours", "extra") });

        expect(calculateRowValueForRow(header, { hours: "3", extra: "" })).toBe(3);
    });

    it("averages the fields an AVERAGE names", () => {
        const header = calc({ name: "AVERAGE", args: fields("a", "b") });

        expect(calculateRowValueForRow(header, { a: 2, b: 4 })).toBe(3);
    });

    it("reads its own stored value for an aggregate over no fields", () => {
        const header = calc({ name: "SUM", args: [] });

        expect(calculateRowValueForRow(header, { total: "12" })).toBe(12);
    });

    it("works out a function of constants", () => {
        const header = calc({
            name: "ROUND",
            args: [
                { type: "const", value: "2.345" },
                { type: "const", value: "2" },
            ],
        });

        expect(calculateRowValueForRow(header, {})).toBe(2.35);
    });

    it("is empty for a field that is not a calculation", () => {
        expect(calculateRowValueForRow({ name: "x", header_usage: "text" }, { x: 1 })).toBe("");
    });
});

describe("isComputedCalculations", () => {
    it("is true for an aggregate over fields", () => {
        expect(isComputedCalculations(calc({ name: "SUM", args: fields("a") }))).toBe(true);
    });

    it("is false for an aggregate over no fields, which is typed in", () => {
        expect(isComputedCalculations(calc({ name: "SUM", args: [] }))).toBe(false);
    });

    it("is true for any other function", () => {
        expect(isComputedCalculations(calc({ name: "TODAY", args: [] }))).toBe(true);
    });

    it("is false for a field that is not a calculation", () => {
        expect(isComputedCalculations({ header_usage: "text" })).toBe(false);
    });
});

describe("formatCellValue", () => {
    it("shows up to two decimals", () => {
        expect(formatCellValue(2.5)).toBe("2.5");
        expect(formatCellValue(1 / 3)).toBe("0.33");
    });

    it("shows the decimals a field names", () => {
        expect(formatCellValue(2.5, { decimals: 2 })).toBe("2.50");
    });

    it("shows nothing for an empty value", () => {
        expect(formatCellValue("").trim()).toBe("");
        expect(formatCellValue(null).trim()).toBe("");
    });
});

describe("a field slot left empty", () => {
    const header = calc({ name: "SUM", args: fields("") });

    it("counts as no field, so the value is typed in", () => {
        expect(isComputedCalculations(header)).toBe(false);
    });

    it("reads the stored value", () => {
        expect(calculateRowValueForRow(header, { total: "7" })).toBe(7);
    });
});

describe("columnTotal", () => {
    const rows = [
        { a: 1, b: 2, total: "4" },
        { a: 3, b: 4, total: "6" },
        { a: "", b: "x", total: "" },
    ];

    it("adds a SUM over fields across the rows", () => {
        expect(columnTotal(calc({ name: "SUM", args: fields("a", "b") }), rows)).toBe(10);
    });

    it("averages each row's AVERAGE, counting a row with no numbers as 0", () => {
        expect(columnTotal(calc({ name: "AVERAGE", args: fields("a", "b") }), rows)).toBeCloseTo(
            5 / 3,
        );
    });

    it("takes the smallest of a MIN over fields, skipping a row with no numbers", () => {
        expect(columnTotal(calc({ name: "MIN", args: fields("a", "b") }), rows)).toBe(1);
    });

    it("combines the typed values of an aggregate over no fields", () => {
        expect(columnTotal(calc({ name: "MAX", args: [] }), rows)).toBe(6);
    });

    it("is empty for a column that is not a calculation, or without rows", () => {
        expect(columnTotal({ name: "a", header_usage: "number" }, rows)).toBe("");
        expect(columnTotal(calc({ name: "SUM", args: [] }), null)).toBe("");
    });
});

describe("formulaArgsFor", () => {
    const headers = [{ name: "c1" }, { header_name: "c2", name: "x" }];

    it("gives a variadic formula's field arguments their headers", () => {
        const spec = {
            args: [
                { type: "field", value: "c1" },
                { type: "constant", value: 3 },
                { type: "field", value: "gone" },
            ],
        };

        expect(formulaArgsFor({ arity: "variadic" }, spec, headers)).toEqual([headers[0], 3, null]);
    });

    it("reads a fixed formula's field arguments from its catalog entry", () => {
        const formula = { args: [{ kind: "field" }, { kind: "number" }] };
        const spec = { args: [{ value: "c2" }, { value: undefined }] };

        expect(formulaArgsFor(formula, spec, headers)).toEqual([headers[1], ""]);
    });
});
