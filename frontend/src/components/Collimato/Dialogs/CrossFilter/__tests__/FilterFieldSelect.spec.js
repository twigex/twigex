// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import FilterFieldSelect from "@/components/Collimato/Dialogs/CrossFilter/FilterFieldSelect.vue";
import en from "@/i18n/en.json";

const options = [
    { id: 1, name: "orders" },
    { id: 2, name: "customers" },
];

describe("FilterFieldSelect", () => {
    it("shows the label when nothing is chosen", () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options, label: "Select data model" },
        });

        expect(w.find("button").text()).toContain("Select data model");
    });

    it("shows the chosen option's name", () => {
        const w = mount(FilterFieldSelect, {
            props: {
                modelValue: options[1],
                options,
                label: "Select data model",
            },
        });

        expect(w.find("button").text()).toContain("customers");
    });

    it("shows an error message when asked to", () => {
        const w = mount(FilterFieldSelect, {
            props: {
                modelValue: null,
                options,
                label: "Select data model",
                error: true,
            },
        });

        expect(w.text()).toContain(en["collimato.dashboard.filter.required"]);
    });

    it("says nothing about errors otherwise", () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options, label: "x" },
        });

        expect(w.find("p").exists()).toBe(false);
    });

    it("emits the whole option, not a key", async () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options, label: "x" },
        });

        await w.find("button").trigger("click");
        await w.findAll("li")[1].trigger("click");

        expect(w.emitted("update:modelValue").at(-1)[0]).toEqual(options[1]);
    });

    it("labels every option by name", async () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options, label: "x" },
        });

        await w.find("button").trigger("click");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["orders", "customers"]);
    });
});

describe("FilterFieldSelect with no options yet", () => {
    it("renders before anything can populate it", () => {
        expect(() =>
            mount(FilterFieldSelect, {
                props: { modelValue: null, options: [], label: "Select column" },
            }),
        ).not.toThrow();
    });

    it("still shows its label", () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options: [], label: "Select column" },
        });

        expect(w.find("button").text()).toContain("Select column");
    });

    it("offers nothing", async () => {
        const w = mount(FilterFieldSelect, {
            props: { modelValue: null, options: [], label: "Select column" },
        });

        await w.find("button").trigger("click");

        expect(w.findAll("li")).toHaveLength(0);
    });
});
