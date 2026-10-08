// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { ListboxButton } from "@headlessui/vue";
import AxisChartCustomization from "@/components/Collimato/Charts/Configurations/AxisChartCustomization.vue";
import PieChartCustomization from "@/components/Collimato/Charts/Configurations/PieChartCustomization.vue";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";

const LOCALES = { en, lv, kk, pl };

const CONDITIONAL = ["orientation", "showValues"];

const K = "collimato.charts.new_chart.customize.";
const AUTO = en[K + "x_axis_field_auto"];
const NONE = en[K + "x_axis_field_none"];
const SORT_AS_QUERIED = en[K + "sort_bars_none"];
const SORT_ASC = en[K + "sort_bars_asc"];
const SORT_DESC = en[K + "sort_bars_desc"];

function openPicker(w, label) {
    const button = w.findAllComponents(ListboxButton).find((b) => b.text() === label);

    return button.trigger("click");
}

function optionLabels(w) {
    return w.findAll("li").map((li) => li.text());
}

function emittedPayload(w) {
    const events = w.emitted("update:configuration");

    expect(events).toBeTruthy();

    return events.at(-1)[0];
}

describe("AxisChartCustomization", () => {
    it("emits every setting it owns, not just some of them", () => {
        const w = mount(AxisChartCustomization, {
            props: { configuration: {}, dimensions: [] },
        });

        const payload = emittedPayload(w);
        const missing = Object.keys(w.vm.config).filter(
            (key) => !CONDITIONAL.includes(key) && !(key in payload),
        );

        expect(missing).toEqual([]);
    });

    it("carries the legend position through", () => {
        const w = mount(AxisChartCustomization, {
            props: {
                configuration: { legendPosition: "Bottom" },
                dimensions: [],
            },
        });

        expect(emittedPayload(w).legendPosition).toBe("Bottom");
    });

    it("carries the x-axis field through", () => {
        const w = mount(AxisChartCustomization, {
            props: {
                configuration: { xAxisField: "weather.weather" },
                dimensions: [{ name: "weather.weather" }],
            },
        });

        expect(emittedPayload(w).xAxisField).toBe("weather.weather");
    });

    it("offers the dimensions plus auto and none", async () => {
        const w = mount(AxisChartCustomization, {
            props: {
                configuration: {},
                dimensions: [
                    { name: "weather.weather", title: "Weather" },
                    { name: "weather.location" },
                ],
            },
        });

        await openPicker(w, AUTO);

        expect(optionLabels(w)).toEqual([AUTO, "Weather", "weather.location", NONE]);
    });

    it("labels the control in every locale", () => {
        [
            "collimato.charts.new_chart.customize.x_axis_field",
            "collimato.charts.new_chart.customize.x_axis_field_auto",
            "collimato.charts.new_chart.customize.x_axis_field_none",
        ].forEach((key) =>
            Object.entries(LOCALES).forEach(([locale, messages]) =>
                expect(messages[key], `${locale} ${key}`).toBeTruthy(),
            ),
        );
    });
});

describe("PieChartCustomization", () => {
    it("emits every setting it owns", () => {
        const w = mount(PieChartCustomization, {
            props: { configuration: {}, dimensions: [] },
        });

        const payload = emittedPayload(w);
        const missing = Object.keys(w.vm.config).filter((key) => !(key in payload));

        expect(missing).toEqual([]);
    });
});

describe("sort bars control", () => {
    it("offers as-queried, ascending and descending", async () => {
        const w = mount(AxisChartCustomization, {
            props: { configuration: {}, dimensions: [] },
        });

        await openPicker(w, SORT_AS_QUERIED);

        expect(optionLabels(w)).toEqual([SORT_AS_QUERIED, SORT_ASC, SORT_DESC]);
    });

    it("carries the choice through", () => {
        const w = mount(AxisChartCustomization, {
            props: { configuration: { sortBars: "desc" }, dimensions: [] },
        });

        expect(emittedPayload(w).sortBars).toBe("desc");
    });

    it("labels every option in every locale", () => {
        ["sort_bars", "sort_bars_none", "sort_bars_asc", "sort_bars_desc"].forEach((suffix) => {
            const key = `collimato.charts.new_chart.customize.${suffix}`;

            Object.entries(LOCALES).forEach(([locale, messages]) =>
                expect(messages[key], `${locale} ${key}`).toBeTruthy(),
            );
        });
    });
});
