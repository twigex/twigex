// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";

function breadcrumbName(dimensions) {
    return dimensions.map((d) => d.split(".").at(-1)).join(", ");
}

function filtersAddedSince(previous, query) {
    const before = (previous?.filters ?? []).map((f) => JSON.stringify(f));

    return (query.filters ?? []).filter((f) => !before.includes(JSON.stringify(f)));
}

function breadcrumbsFor(trail) {
    return trail.map((query, index) => ({
        name: breadcrumbName(query.dimensions ?? []),
        value: filtersAddedSince(trail[index - 1], query)
            .flatMap((f) => f.values ?? [])
            .join(", "),
    }));
}

const filter = (member, value) => ({
    member,
    operator: "equals",
    values: [value],
});

describe("drill down breadcrumbs", () => {
    it("shows no value on the first level", () => {
        const crumbs = breadcrumbsFor([{ dimensions: ["w.weather"], filters: [] }]);

        expect(crumbs[0]).toEqual({ name: "weather", value: "" });
    });

    it("shows the value that led to each level", () => {
        const crumbs = breadcrumbsFor([
            { dimensions: ["w.weather"], filters: [] },
            {
                dimensions: ["w.borough"],
                filters: [filter("w.weather", "rain")],
            },
            {
                dimensions: ["w.payment"],
                filters: [filter("w.weather", "rain"), filter("w.borough", "Manhattan")],
            },
        ]);

        expect(crumbs).toEqual([
            { name: "weather", value: "" },
            { name: "borough", value: "rain" },
            { name: "payment", value: "Manhattan" },
        ]);
    });

    it("keeps filters the chart already had off the trail", () => {
        const existing = filter("w.year", "2024");

        const crumbs = breadcrumbsFor([
            { dimensions: ["w.weather"], filters: [existing] },
            {
                dimensions: ["w.borough"],
                filters: [existing, filter("w.weather", "rain")],
            },
        ]);

        expect(crumbs[1].value).toBe("rain");
    });

    it("joins several values added in one step", () => {
        const crumbs = breadcrumbsFor([
            { dimensions: ["w.a", "w.b"], filters: [] },
            {
                dimensions: ["w.c"],
                filters: [filter("w.a", "one"), filter("w.b", "two")],
            },
        ]);

        expect(crumbs[1].value).toBe("one, two");
    });

    it("names a level split by several dimensions", () => {
        const crumbs = breadcrumbsFor([{ dimensions: ["w.location", "w.weather"], filters: [] }]);

        expect(crumbs[0].name).toBe("location, weather");
    });
});
