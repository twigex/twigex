// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h } from "vue";
import BaseCheckbox from "@/components/BaseCheckbox.vue";

describe("Checkbox", () => {
    it("associates its label with its own checkbox", () => {
        const w = mount(BaseCheckbox, {
            props: { modelValue: false, label: "Legend" },
        });

        const id = w.find("input").attributes("id");

        expect(id).toBeTruthy();
        expect(w.find("label").attributes("for")).toBe(id);
    });

    it("gives every instance a different id", () => {
        const Host = defineComponent({
            render() {
                return h("div", [
                    h(BaseCheckbox, { modelValue: false, label: "A" }),
                    h(BaseCheckbox, { modelValue: false, label: "B" }),
                    h(BaseCheckbox, { modelValue: false, label: "C" }),
                ]);
            },
        });

        const w = mount(Host);
        const ids = w.findAll("input").map((i) => i.attributes("id"));

        expect(ids).toHaveLength(3);
        expect(new Set(ids).size).toBe(3);
    });

    it("points aria-describedby at its own description", () => {
        const w = mount(BaseCheckbox, {
            props: {
                modelValue: false,
                label: "Legend",
                description: "Whether to show the chart legend.",
            },
        });

        const describedBy = w.find("input").attributes("aria-describedby");

        expect(describedBy).toBe(w.find("p").attributes("id"));
    });

    it("omits aria-describedby when there is no description", () => {
        const w = mount(BaseCheckbox, {
            props: { modelValue: false, label: "Legend" },
        });

        expect(w.find("input").attributes("aria-describedby")).toBeUndefined();
        expect(w.find("p").exists()).toBe(false);
    });

    it("reflects the value it is given and emits on change", async () => {
        const w = mount(BaseCheckbox, {
            props: { modelValue: true, label: "Legend" },
        });

        expect(w.find("input").element.checked).toBe(true);

        await w.find("input").setValue(false);

        expect(w.emitted("update:modelValue")).toEqual([[false]]);
    });

    it("follows the prop rather than keeping its own state", async () => {
        const w = mount(BaseCheckbox, {
            props: { modelValue: false, label: "Legend" },
        });

        await w.setProps({ modelValue: true });

        expect(w.find("input").element.checked).toBe(true);
    });
});
