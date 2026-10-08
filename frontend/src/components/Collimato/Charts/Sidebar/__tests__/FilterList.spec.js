// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import FilterList from "@/components/Collimato/Charts/Sidebar/FilterList.vue";
import ChartFilters from "@/components/Collimato/Charts/Sidebar/ChartFilters.vue";

const cubes = [
    {
        table: "earthquakes",
        dimensions: [{ name: "earthquakes.magnitude", title: "Magnitude" }],
        measures: [],
    },
];

const filters = [{ member: "earthquakes.magnitude", operator: "gt", value: ["5"] }];

describe("FilterList", () => {
    it("shows its filters without needing to be expanded first", () => {
        const w = mount(FilterList, { props: { filters, cubes } });

        expect(w.text()).toContain("magnitude");
        expect(w.text()).toContain("gt");
    });

    it("offers a way to add one when empty", () => {
        const w = mount(FilterList, { props: { filters: [], cubes } });

        expect(w.find("button").exists()).toBe(true);
    });

    it("emits the filter it is asked to remove", async () => {
        const w = mount(FilterList, { props: { filters, cubes } });

        await w.findAll("button").at(-1).trigger("click");

        expect(w.emitted("remove-filter")).toEqual([[filters[0]]]);
    });

    it("leaves the heading to its caller when given no title", () => {
        const w = mount(FilterList, { props: { filters, cubes } });

        expect(w.find("h4").exists()).toBe(false);
    });
});

describe("ChartFilters", () => {
    it("keeps the sidebar collapsed until opened", () => {
        const w = mount(ChartFilters, { props: { filters, cubes } });

        expect(w.findComponent(FilterList).exists()).toBe(false);
    });

    it("counts the filters on its collapsed header", () => {
        const w = mount(ChartFilters, { props: { filters, cubes } });

        expect(w.text()).toContain("1");
    });

    it("shows the list once opened", async () => {
        const w = mount(ChartFilters, { props: { filters, cubes } });

        await w.find("button").trigger("click");

        expect(w.findComponent(FilterList).exists()).toBe(true);
    });
});
