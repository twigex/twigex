// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import FilterScope from "@/components/Collimato/Dialogs/CrossFilter/FilterScope.vue";

const options = [
    { id: "all", title: "Apply to all panels" },
    { id: "specific", title: "Apply to specific panels" },
];

const availableCharts = [
    { id: "a", name: "Revenue" },
    { id: "b", name: "Orders" },
];

function scope(props = {}) {
    return mount(FilterScope, {
        props: { scope: options[0], charts: [], options, availableCharts, ...props },
    });
}

describe("FilterScope", () => {
    it("offers every scope option", () => {
        expect(scope().findAll('input[type="radio"]')).toHaveLength(2);
    });

    it("marks the current scope", () => {
        const radios = scope({ scope: options[1] }).findAll('input[type="radio"]');

        expect(radios[0].element.checked).toBe(false);
        expect(radios[1].element.checked).toBe(true);
    });

    it("emits the whole option when the scope changes", async () => {
        const w = scope();

        await w.findAll('input[type="radio"]')[1].trigger("change");

        expect(w.emitted("update:scope").at(-1)[0]).toEqual(options[1]);
    });

    it("hides the chart picker while the scope is all panels", () => {
        expect(scope().text()).not.toContain("Revenue");
    });

    it("shows the chart picker once the scope is specific", async () => {
        const w = scope({ scope: options[1] });

        await w.find("button").trigger("click");

        expect(w.text()).toContain("Revenue");
    });

    it("does not write into the charts it was given", async () => {
        const charts = [];
        const w = scope({ scope: options[1], charts });

        await w.find("button").trigger("click");
        await w.findAll("li")[0].trigger("click");

        expect(charts).toEqual([]);
        expect(w.emitted("update:charts")).toBeTruthy();
    });
});
