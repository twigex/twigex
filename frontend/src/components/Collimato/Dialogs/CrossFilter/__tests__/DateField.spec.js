// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import DateField from "@/components/Collimato/Dialogs/CrossFilter/DateField.vue";
import en from "@/i18n/en.json";

function field(props = {}) {
    return mount(DateField, {
        props: {
            modelValue: null,
            label: "Select start date",
            title: "Choose start date",
            ...props,
        },
        attachTo: document.body,
    });
}

describe("DateField", () => {
    it("shows the label as the placeholder when empty", () => {
        const w = field();

        expect(w.find("input").attributes("placeholder")).toBe("Select start date");
    });

    it("shows the chosen date", () => {
        const w = field({ modelValue: "2026-01-01" });

        expect(w.find("input").element.value).toBe("2026-01-01");
    });

    it("is read only, so a date can only come from the picker", () => {
        expect(field().find("input").attributes("readonly")).toBeDefined();
    });

    it("shows the error message when asked to", () => {
        const w = field({ error: true });

        expect(w.text()).toContain(en["collimato.dashboard.filter.required"]);
    });
});
