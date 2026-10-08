// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { appliedFlatFilters, filtersForSave, withoutFilterOptions } from "../filterCombine";

describe("withoutFilterOptions", () => {
    it("drops the option lists from flat and grouped filters and keeps the rest", () => {
        const options = [{ id: ["s1", "s2"], name: "Done" }];
        const payload = {
            sort: [{ field: "name", direction: "asc" }],
            flatFilters: [{ field: "status", operator: "is", value: "s1", linkedOptions: options }],
            groups: [
                {
                    relationToNext: "OR",
                    filters: [
                        {
                            field: "status",
                            operator: "is_not",
                            value: "s2",
                            linkedOptions: options,
                        },
                    ],
                },
            ],
        };

        expect(withoutFilterOptions(payload)).toEqual({
            sort: payload.sort,
            flatFilters: [{ field: "status", operator: "is", value: "s1" }],
            groups: [
                {
                    relationToNext: "OR",
                    filters: [{ field: "status", operator: "is_not", value: "s2" }],
                },
            ],
        });
        expect(payload.flatFilters[0].linkedOptions).toBe(options);
    });

    it("gives empty lists for a payload without filters", () => {
        expect(withoutFilterOptions({})).toEqual({ flatFilters: [], groups: [] });
    });
});

describe("appliedFlatFilters", () => {
    it("keeps a row with a field, an operator and a value", () => {
        const row = { field: "status", operator: "is", value: "a" };

        expect(appliedFlatFilters([row])).toEqual([row]);
    });

    it("keeps a row whose operator takes no value", () => {
        const rows = [
            { field: "due_date", operator: "is_set", value: "" },
            { field: "due_date", operator: "is_not_set" },
        ];

        expect(appliedFlatFilters(rows)).toEqual(rows);
    });

    it("leaves out a row still being filled in", () => {
        expect(
            appliedFlatFilters([
                { field: "status", operator: "is", value: "" },
                { field: "", operator: "is", value: "a" },
                { field: "status", operator: "", value: "a" },
            ]),
        ).toEqual([]);
    });

    it("takes a missing list as empty", () => {
        expect(appliedFlatFilters(undefined)).toEqual([]);
    });
});

describe("filtersForSave", () => {
    const zero = { field: "c1", operator: "equals", value: 0 };
    const unfinished = { field: "c1", operator: "equals", value: "" };
    const many = { field: "status", operator: "is_any_of", value: ["a", "b"] };

    it("keeps the rows that are applied, a zero among them", () => {
        expect(filtersForSave([zero, unfinished], []).flatFilters).toEqual([zero]);
    });

    it("sends a row's list of values as values on the report only", () => {
        expect(filtersForSave([many], [], { report: true }).flatFilters).toEqual([
            { field: "status", operator: "is_any_of", values: ["a", "b"] },
        ]);
        expect(filtersForSave([many], []).flatFilters).toEqual([many]);
    });

    it("drops empty groups and the last group's relation", () => {
        const groups = [
            { filters: [zero], nextRelation: "OR" },
            { filters: [], nextRelation: "AND" },
            { filters: [many], nextRelation: "AND" },
        ];

        expect(filtersForSave([], groups).groups).toEqual([
            { filters: [zero], relationToNext: "OR" },
            { filters: [many], relationToNext: null },
        ]);
    });
});
