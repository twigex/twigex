// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import FilterValuePicker from "@/components/Collimato/Dialogs/CrossFilter/FilterValuePicker.vue";

const values = ["rain", "sun", "fog", "snow"];

function picker(props = {}) {
    return mount(FilterValuePicker, {
        props: { modelValue: [], values, ...props },
    });
}

describe("FilterValuePicker", () => {
    it("offers every value when nothing is chosen", async () => {
        const w = picker();

        await w.find("button").trigger("click");

        expect(w.findAll("li").map((li) => li.text())).toEqual(values);
    });

    it("stops offering a value once it is chosen", async () => {
        const w = picker({ modelValue: ["rain"] });

        await w.find("button").trigger("click");

        expect(w.findAll("li").map((li) => li.text())).not.toContain("rain");
    });

    it("narrows the list as you type", async () => {
        const w = picker();

        await w.find("button").trigger("click");
        await w.find("input").setValue("sn");
        await w.find("input").trigger("change");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["snow"]);
    });

    it("matches regardless of case", async () => {
        const w = picker();

        await w.find("button").trigger("click");
        await w.find("input").setValue("RAIN");
        await w.find("input").trigger("change");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["rain"]);
    });

    it("copes with values that are not strings", async () => {
        const w = picker({ values: [1, 2, 30], modelValue: [] });

        await w.find("button").trigger("click");
        await w.find("input").setValue("3");
        await w.find("input").trigger("change");

        expect(w.findAll("li").map((li) => li.text())).toEqual(["30"]);
    });

    it("emits a new list rather than editing the one it was given", async () => {
        const chosen = ["rain", "sun"];
        const w = picker({ modelValue: chosen });

        await w.findAll("button")[0].trigger("click");

        expect(chosen).toEqual(["rain", "sun"]);
    });
});
