// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { h } from "vue";
import BaseSelect from "@/components/BaseSelect.vue";

const options = [
    { value: "require", label: "Encrypted" },
    { value: "disable", label: "Not encrypted" },
];

function select(props = {}, slots = {}) {
    return mount(BaseSelect, {
        props: { modelValue: "", options, ...props },
        slots,
        global: { stubs: { Portal: { template: "<div><slot /></div>" } } },
    });
}

describe("Select", () => {
    it("shows the label of the chosen value", () => {
        expect(select({ modelValue: "disable" }).find("button").text()).toBe("Not encrypted");
    });

    it("shows the placeholder when nothing is chosen", () => {
        const w = select({ placeholder: "Choose one" });

        expect(w.find("button").text()).toBe("Choose one");
    });

    it("emits the value, not the option", async () => {
        const w = select();

        await w.find("button").trigger("click");
        await w.findAll("li")[1].trigger("click");

        expect(w.emitted("update:modelValue").at(-1)[0]).toBe("disable");
    });

    it("accepts plain strings as options", async () => {
        const w = select({ options: ["mysql", "postgres"] });

        await w.find("button").trigger("click");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["mysql", "postgres"]);
    });

    it("shows a label above itself when given one", () => {
        expect(select({ label: "SSL/TLS" }).text()).toContain("SSL/TLS");
    });

    it("renders no label element when not given one", () => {
        expect(select().find("label").exists()).toBe(false);
    });

    it("lets a caller render its own options", async () => {
        const w = select(
            {},
            {
                option: ({ option }) => h("b", `[${option.label}]`),
            },
        );

        await w.find("button").trigger("click");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["[Encrypted]", "[Not encrypted]"]);
    });

    it("lets a caller render the chosen value", () => {
        const w = select(
            { modelValue: "require" },
            { selected: ({ option }) => h("b", `>${option.label}<`) },
        );

        expect(w.find("button").text()).toBe(">Encrypted<");
    });

    it("cannot be opened when disabled", async () => {
        const w = select({ disabled: true });

        await w.find("button").trigger("click");

        expect(w.findAll("li")).toHaveLength(0);
    });
});
