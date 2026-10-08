// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import { ListboxButton } from "@headlessui/vue";
import PointLayer from "@/components/Collimato/Charts/Map/LayerTypes/PointLayer.vue";
import HeatLayer from "@/components/Collimato/Charts/Map/LayerTypes/HeatLayer.vue";
import ArcLayer from "@/components/Collimato/Charts/Map/LayerTypes/ArcLayer.vue";
import { createDeckLayer } from "@/utils/collimato/mapDeckLayers.js";
import { defaultLayerConfig } from "@/utils/collimato/mapLayerTypes.js";

const cubes = [
    {
        table: "taxi_trips",
        join: 1,
        dimensions: [{ name: "taxi_trips.fare_amount", title: "Fare" }],
        measures: [],
    },
];

const COUNT = "taxi_trips.count";
const COUNT_LABEL = "Number of records";
const TOOLTIP_PLACEHOLDER = "Select tooltip fields";

function mountWith(component, config) {
    return mount(component, {
        props: { cubes, selectedCube: "taxi_trips", config, layer: {} },
    });
}

function pickerLabels(w) {
    return w.findAllComponents(ListboxButton).map((b) => b.text());
}

function openPicker(w, label) {
    const button = w.findAllComponents(ListboxButton).find((b) => b.text() === label);

    return button.trigger("click");
}

describe("value pickers resolve the count they offer", () => {
    it("point radius shows the count once chosen", () => {
        const w = mountWith(PointLayer, {
            ...defaultLayerConfig("point"),
            radius_field: COUNT,
        });

        expect(pickerLabels(w)).toContain(COUNT_LABEL);
    });

    it("point radius shows nothing chosen until it is", () => {
        const w = mountWith(PointLayer, defaultLayerConfig("point"));

        expect(pickerLabels(w)).not.toContain(COUNT_LABEL);
    });

    it("heat weight shows the count once chosen", () => {
        const w = mountWith(HeatLayer, {
            ...defaultLayerConfig("heatmap"),
            weight: COUNT,
        });

        expect(pickerLabels(w)).toContain(COUNT_LABEL);
    });

    it("arc width shows the count once chosen", () => {
        const w = mountWith(ArcLayer, {
            ...defaultLayerConfig("arc"),
            width_field: COUNT,
        });

        expect(pickerLabels(w)).toContain(COUNT_LABEL);
    });

    it("point radius still resolves an ordinary dimension", () => {
        const w = mountWith(PointLayer, {
            ...defaultLayerConfig("point"),
            radius_field: "taxi_trips.fare_amount",
        });

        expect(pickerLabels(w)).toContain("Fare");
    });

    it("offers the count in the same list it resolves against", async () => {
        const w = mountWith(PointLayer, defaultLayerConfig("point"));

        await openPicker(w, TOOLTIP_PLACEHOLDER);

        expect(w.findAll("li").map((li) => li.text())).toContain(COUNT_LABEL);
    });
});

describe("tooltip fields", () => {
    it("can show the record count", () => {
        const w = mountWith(PointLayer, {
            ...defaultLayerConfig("point"),
            popover: { fields: [COUNT] },
        });

        expect(pickerLabels(w)).toContain(COUNT);
    });

    it("still resolves ordinary dimensions", () => {
        const w = mountWith(PointLayer, {
            ...defaultLayerConfig("point"),
            popover: { fields: ["taxi_trips.fare_amount"] },
        });

        expect(pickerLabels(w)).toContain("taxi_trips.fare_amount");
    });

    it("drops a field that no longer exists", () => {
        const w = mountWith(PointLayer, {
            ...defaultLayerConfig("point"),
            popover: { fields: ["taxi_trips.gone"] },
        });

        expect(pickerLabels(w)).toContain(TOOLTIP_PLACEHOLDER);
    });
});

describe("arc colours are independent", () => {
    function arc(overrides = {}) {
        return mount(ArcLayer, {
            props: {
                cubes,
                selectedCube: "taxi_trips",
                config: { ...defaultLayerConfig("arc"), ...overrides },
                layer: {},
            },
        });
    }

    it("changes only the start colour", async () => {
        const w = arc();

        await w.findAll('input[type="color"]')[0].setValue("#00ff00");

        const payload = w.emitted("update:config").at(-1)[0];

        expect(payload.source_color).toBe("#00ff00");
        expect(payload.target_color).toBe(defaultLayerConfig("arc").target_color);
    });

    it("gives each end its own colour input", () => {
        const ids = arc()
            .findAll('input[type="color"]')
            .map((i) => i.attributes("id"));

        expect(ids).toHaveLength(2);
        expect(new Set(ids).size).toBe(2);
    });

    it("keeps the two colour schemes apart", () => {
        const config = defaultLayerConfig("arc");

        expect(config.source_color_scheme).toBeTruthy();
        expect(config.target_color_scheme).toBeTruthy();
        expect(config).not.toHaveProperty("color_scheme");
    });

    it("renders each end with its own scheme", () => {
        const layer = {
            name: "trips",
            type: "arc",
            config: {
                ...defaultLayerConfig("arc"),
                source: { lat: "a.lat", long: "a.long" },
                target: { lat: "b.lat", long: "b.long" },
                source_color_scheme: "interpolateViridis",
                target_color_scheme: "interpolateInferno",
                source_color_field: "taxi_trips.count",
                target_color_field: "taxi_trips.count",
                data: [{ "taxi_trips.count": 1 }, { "taxi_trips.count": 9 }],
            },
        };

        const deckLayer = createDeckLayer(layer, "arc");
        const row = layer.config.data[1];

        expect(deckLayer.props.getSourceColor(row)).not.toEqual(
            deckLayer.props.getTargetColor(row),
        );
    });

    it("falls back to the old shared scheme for arcs saved before this", () => {
        const layer = {
            name: "trips",
            type: "arc",
            config: {
                ...defaultLayerConfig("arc"),
                source: { lat: "a.lat", long: "a.long" },
                target: { lat: "b.lat", long: "b.long" },
                source_color_scheme: undefined,
                target_color_scheme: undefined,
                color_scheme: "interpolatePlasma",
                source_color_field: "taxi_trips.count",
                data: [{ "taxi_trips.count": 1 }],
            },
        };

        expect(() => createDeckLayer(layer, "arc")).not.toThrow();
    });
});
