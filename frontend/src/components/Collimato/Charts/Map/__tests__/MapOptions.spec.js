// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import MapOptions from "@/components/Collimato/Charts/Map/MapOptions.vue";
import en from "@/i18n/en.json";

const cubes = [
    {
        table: "earthquakes",
        join: 0,
        dimensions: [
            {
                name: "earthquakes.magnitude",
                title: "Magnitude",
                type: "number",
            },
        ],
        measures: [],
    },
];

const layer = (overrides = {}) => ({
    id: 0,
    type: "point",
    show: true,
    name: "Quakes",
    model: "earthquakes",
    filters: [],
    config: { coordinates: { lat: null, long: null }, data: [] },
    ...overrides,
});

function mountOptions(layers) {
    return mount(MapOptions, {
        props: {
            chartOptions: {
                map_style: "s",
                latitude: "",
                longitude: "",
                zoom: 1,
                layers: layers,
            },
            cubes: cubes,
        },
    });
}

const reloadButton = (w) =>
    w
        .findAll("button")
        .find((b) => b.attributes("title") === en["collimato.charts.new_chart.map.reload_layer"]);

describe("MapOptions reload", () => {
    it("asks to reload the layer that was clicked", async () => {
        const w = mountOptions([layer({ id: 7 })]);

        await reloadButton(w).trigger("click");

        expect(w.emitted("reload-layer")).toEqual([[7]]);
    });

    it("offers no reload until a model is picked", () => {
        const w = mountOptions([layer({ model: null })]);

        expect(reloadButton(w)).toBeUndefined();
    });

    it("does not expand the layer when reloading it", async () => {
        const w = mountOptions([layer()]);

        await reloadButton(w).trigger("click");

        expect(w.text()).not.toContain(en["collimato.charts.new_chart.map.filters_description"]);
    });
});

describe("MapOptions layer status", () => {
    const spinner = (w) => w.find("span.animate-spin");

    it("spins on the layer that is loading", () => {
        const w = mount(MapOptions, {
            props: {
                chartOptions: {
                    map_style: "s",
                    latitude: "",
                    longitude: "",
                    zoom: 1,
                    layers: [layer()],
                },
                cubes: cubes,
                statuses: [{ id: 0, name: "Quakes", status: "loading" }],
            },
        });

        expect(spinner(w).exists()).toBe(true);
        expect(reloadButton(w)).toBeUndefined();
    });

    it("goes back to the reload button once loaded", () => {
        const w = mount(MapOptions, {
            props: {
                chartOptions: {
                    map_style: "s",
                    latitude: "",
                    longitude: "",
                    zoom: 1,
                    layers: [layer()],
                },
                cubes: cubes,
                statuses: [{ id: 0, name: "Quakes", status: "loaded" }],
            },
        });

        expect(spinner(w).exists()).toBe(false);
        expect(reloadButton(w)).toBeTruthy();
    });

    it("does not spin a layer because a different one is loading", () => {
        const w = mount(MapOptions, {
            props: {
                chartOptions: {
                    map_style: "s",
                    latitude: "",
                    longitude: "",
                    zoom: 1,
                    layers: [layer({ id: 0 })],
                },
                cubes: cubes,
                statuses: [{ id: 9, name: "Other", status: "loading" }],
            },
        });

        expect(spinner(w).exists()).toBe(false);
    });
});
