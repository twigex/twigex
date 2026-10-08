// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi } from "vitest";
import { useChartBuilder } from "@/composables/collimato/useChartBuilder.js";

function setup(chartType = "bar") {
    return { builder: useChartBuilder(chartType) };
}

describe("map configuration", () => {
    const saved = {
        map_style: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json",
        longitude: -119.4179,
        latitude: 36.7783,
        zoom: 6,
        layers: [
            {
                id: 0,
                type: "point",
                show: true,
                name: "Quakes",
                model: "earthquakes",
                filters: [],
                config: {
                    fill_color: "#ebdd24",
                    fill_color_field: null,
                    radius: 10,
                    radius_field: "earthquakes.magnitude",
                    min_radius: 1,
                    max_radius: 100,
                    coordinates: {
                        lat: "earthquakes.latitude",
                        long: "earthquakes.longitude",
                    },
                    popover: { fields: [] },
                    data: null,
                },
            },
            {
                id: 1,
                type: "heatmap",
                show: false,
                name: "Density",
                model: "earthquakes",
                filters: [],
                config: {
                    colorScheme: "interpolatePlasma",
                    weight: null,
                    radiusPixels: 30,
                    intensity: 1,
                    threshold: 0.03,
                    coordinates: {
                        lat: "earthquakes.latitude",
                        long: "earthquakes.longitude",
                    },
                    data: null,
                },
            },
        ],
    };

    it("exposes the persisted shape by default", () => {
        const { builder } = setup("map");

        expect(Object.keys(builder.mapConfiguration.value).sort()).toEqual([
            "latitude",
            "layers",
            "longitude",
            "map_style",
            "zoom",
        ]);
    });

    it("round-trips a saved configuration without loss", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration(structuredClone(saved));

        expect(builder.mapConfiguration.value).toEqual(saved);
    });

    it("survives a re-save of what it produced", () => {
        const { builder } = setup("map");

        const persist = () => JSON.parse(JSON.stringify(builder.mapConfiguration.value));

        builder.setMapConfiguration(structuredClone(saved));
        const first = persist();

        builder.setMapConfiguration(first);
        const second = persist();

        expect(first).toEqual(saved);
        expect(second).toEqual(first);
    });

    it("fills in a zoom for older configurations that carry none", () => {
        const { builder } = setup("map");
        const { zoom, ...withoutZoom } = structuredClone(saved);

        builder.setMapConfiguration(withoutZoom);

        expect(builder.mapConfiguration.value.zoom).toBe(1);
        expect(builder.mapConfiguration.value.layers).toEqual(saved.layers);
    });

    it("gives layers saved before filters existed an empty list", () => {
        const older = structuredClone(saved);

        older.layers.forEach((layer) => delete layer.filters);

        const { builder } = setup("map");

        builder.setMapConfiguration(older);

        expect(builder.mapLayers.value.map((layer) => layer.filters)).toEqual([[], []]);
    });

    it("preserves keys it does not model", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            ...structuredClone(saved),
            pitch: 45,
            bearing: 12,
        });

        expect(builder.mapConfiguration.value.pitch).toBe(45);
        expect(builder.mapConfiguration.value.bearing).toBe(12);
    });

    it("exposes the very array the layer editor mutates", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration(structuredClone(saved));

        expect(builder.mapConfiguration.value.layers).toBe(builder.mapLayers.value);

        builder.mapLayers.value.push({ id: 2 });

        expect(builder.mapConfiguration.value.layers).toHaveLength(3);
    });
});

describe("layer config and coordinates", () => {
    function layerSetup() {
        const loadData = vi.fn(() => Promise.resolve({ data: { data: [{ row: 1 }] } }));
        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [
            {
                table: "earthquakes",
                join: 0,
                measures: [],
                dimensions: [{ name: "earthquakes.latitude" }, { name: "earthquakes.longitude" }],
            },
        ];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "point",
                    model: "earthquakes",
                    config: {
                        radius: 10,
                        coordinates: { lat: null, long: null },
                        data: null,
                    },
                },
            ],
        });

        return { builder, loadData };
    }

    const layer = (builder) => builder.mapLayers.value[0];

    it("merges a config patch and leaves the rest alone", () => {
        const { builder } = layerSetup();

        builder.setLayerConfig(0, { radius: 42 });

        expect(layer(builder).config.radius).toBe(42);
        expect(layer(builder).config.coordinates).toEqual({
            lat: null,
            long: null,
        });
    });

    it("setting one coordinate cannot null the other", async () => {
        const { builder } = layerSetup();

        await builder.setLayerCoordinates(0, {
            lat: "earthquakes.latitude",
        });
        await builder.setLayerCoordinates(0, {
            long: "earthquakes.longitude",
        });

        expect(layer(builder).config.coordinates).toEqual({
            lat: "earthquakes.latitude",
            long: "earthquakes.longitude",
        });
    });

    it("does not load until both coordinates are present", async () => {
        const { builder, loadData } = layerSetup();

        await builder.setLayerCoordinates(0, {
            lat: "earthquakes.latitude",
        });
        expect(loadData).not.toHaveBeenCalled();

        await builder.setLayerCoordinates(0, {
            long: "earthquakes.longitude",
        });
        expect(loadData).toHaveBeenCalledTimes(1);
        expect(layer(builder).config.data).toEqual([{ row: 1 }]);
    });

    it("reloads when a coordinate is changed after both are set", async () => {
        const { builder, loadData } = layerSetup();

        await builder.setLayerCoordinates(0, {
            lat: "earthquakes.latitude",
            long: "earthquakes.longitude",
        });
        await builder.setLayerCoordinates(0, {
            lat: "earthquakes.longitude",
        });

        expect(loadData).toHaveBeenCalledTimes(2);
    });

    it("clearing a coordinate stops the layer loading again", async () => {
        const { builder, loadData } = layerSetup();

        await builder.setLayerCoordinates(0, {
            lat: "earthquakes.latitude",
            long: "earthquakes.longitude",
        });
        loadData.mockClear();

        await builder.setLayerCoordinates(0, { long: null });

        expect(loadData).not.toHaveBeenCalled();
        expect(layer(builder).config.coordinates.lat).toBe("earthquakes.latitude");
    });

    it("ignores operations for a layer that is gone", async () => {
        const { builder } = layerSetup();

        expect(() => builder.setLayerConfig(99, { radius: 1 })).not.toThrow();
        await expect(builder.setLayerCoordinates(99, { lat: "x" })).resolves.toBeNull();
    });
});

describe("layer lifecycle", () => {
    const POINT = "point";
    const HEATMAP = "heatmap";

    function lifecycleSetup() {
        const loadData = vi.fn(() => Promise.resolve({ data: { data: [{ row: 1 }] } }));
        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [
            {
                table: "earthquakes",
                join: 0,
                measures: [],
                dimensions: [
                    { name: "earthquakes.latitude" },
                    { name: "earthquakes.longitude" },
                    { name: "earthquakes.magnitude" },
                ],
            },
            { table: "orders", join: 0, measures: [], dimensions: [] },
        ];

        return { builder, loadData };
    }

    it("adds a layer with the defaults for its type", () => {
        const { builder } = lifecycleSetup();

        const layer = builder.addLayer({ name: "Quakes", type: POINT });

        expect(layer).toMatchObject({
            id: 0,
            name: "Quakes",
            show: true,
            model: null,
        });
        expect(layer.config.coordinates).toEqual({ lat: null, long: null });
        expect(layer.config.radius).toBe(10);
        expect(builder.mapLayers.value).toHaveLength(1);
    });

    it("gives each layer its own config, not a shared default", () => {
        const { builder } = lifecycleSetup();

        const a = builder.addLayer({ name: "A", type: POINT });
        const b = builder.addLayer({ name: "B", type: POINT });

        a.config.radius = 99;

        expect(b.config.radius).toBe(10);
        expect(b.id).toBe(1);
    });

    it("reuses no id already taken", () => {
        const { builder } = lifecycleSetup();

        builder.setMapConfiguration({ layers: [{ id: 7, config: {} }] });

        expect(builder.addLayer({ name: "X", type: POINT }).id).toBe(8);
    });

    it("removes a layer by id", () => {
        const { builder } = lifecycleSetup();
        const layer = builder.addLayer({ name: "A", type: POINT });

        builder.addLayer({ name: "B", type: POINT });

        builder.removeLayer(layer.id);

        expect(builder.mapLayers.value.map((l) => l.name)).toEqual(["B"]);
    });

    it("toggles visibility and renames", () => {
        const { builder } = lifecycleSetup();
        const layer = builder.addLayer({ name: "A", type: POINT });

        builder.setLayerVisible(layer.id, false);
        builder.setLayerName(layer.id, "Renamed");

        expect(layer.show).toBe(false);
        expect(layer.name).toBe("Renamed");
    });

    describe("changing the data model", () => {
        async function withLoadedLayer() {
            const { builder, loadData } = lifecycleSetup();
            const layer = builder.addLayer({ name: "A", type: POINT });

            builder.setLayerModel(layer.id, "earthquakes");
            builder.setLayerConfig(layer.id, {
                radius_field: "earthquakes.magnitude",
                fill_color_field: "earthquakes.magnitude",
                popover: { fields: ["earthquakes.magnitude"] },
            });
            await builder.setLayerCoordinates(layer.id, {
                lat: "earthquakes.latitude",
                long: "earthquakes.longitude",
            });

            return { builder, loadData, layer };
        }

        it("clears the rows so stale points cannot render", async () => {
            const { builder, layer } = await withLoadedLayer();

            expect(layer.config.data).toEqual([{ row: 1 }]);

            builder.setLayerModel(layer.id, "orders");

            expect(builder.mapLayers.value[0].config.data).toBeNull();
        });

        it("clears every field that named the old model", async () => {
            const { builder, layer } = await withLoadedLayer();

            builder.setLayerModel(layer.id, "orders");
            const config = builder.mapLayers.value[0].config;

            expect(config.coordinates).toEqual({ lat: null, long: null });
            expect(config.radius_field).toBeNull();
            expect(config.fill_color_field).toBeNull();
            expect(config.popover.fields).toEqual([]);
        });

        it("keeps appearance settings that are not model-scoped", async () => {
            const { builder, layer } = await withLoadedLayer();

            builder.setLayerConfig(layer.id, { radius: 42 });

            builder.setLayerModel(layer.id, "orders");
            const config = builder.mapLayers.value[0].config;

            expect(config.radius).toBe(42);
        });

        it("does not reload until coordinates are picked again", async () => {
            const { builder, loadData, layer } = await withLoadedLayer();

            loadData.mockClear();

            builder.setLayerModel(layer.id, "orders");
            await builder.loadLayer(layer.id);

            expect(loadData).not.toHaveBeenCalled();
        });
    });

    describe("changing the layer type", () => {
        it("swaps in the defaults for the new type", () => {
            const { builder } = lifecycleSetup();
            const layer = builder.addLayer({ name: "A", type: POINT });

            builder.setLayerModel(layer.id, "earthquakes");

            builder.setLayerType(layer.id, HEATMAP);
            const updated = builder.mapLayers.value[0];

            expect(updated.type).toBe(HEATMAP);
            expect(updated.config.radiusPixels).toBe(30);
            expect(updated.config.radius).toBeUndefined();
            expect(updated.config.coordinates).toEqual({
                lat: null,
                long: null,
            });
        });

        it("leaves the chosen model alone", () => {
            const { builder } = lifecycleSetup();
            const layer = builder.addLayer({ name: "A", type: POINT });

            builder.setLayerModel(layer.id, "earthquakes");

            builder.setLayerType(layer.id, HEATMAP);

            expect(builder.mapLayers.value[0].model).toBe("earthquakes");
        });
    });
});

describe("map centre", () => {
    it("reports no centre for a fresh map", () => {
        const { builder } = setup("map");

        expect(builder.hasMapCenter.value).toBe(false);
    });

    it("reports no centre when the saved one is blank", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            latitude: "",
            longitude: "",
            layers: [],
        });

        expect(builder.hasMapCenter.value).toBe(false);
    });

    it("reports a centre once real coordinates are saved", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            latitude: 36.7783,
            longitude: -119.4179,
            layers: [],
        });

        expect(builder.hasMapCenter.value).toBe(true);
    });

    it("treats a zero centre as a real one", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            latitude: 0,
            longitude: 0,
            layers: [],
        });

        expect(builder.hasMapCenter.value).toBe(true);
    });
});

describe("legacy layer type", () => {
    it("migrates the persisted translated label to a plain value", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: { name: "Punkts", value: "point" },
                    config: {},
                },
                {
                    id: 1,
                    type: { name: "Siltuma", value: "heatmap" },
                    config: {},
                },
            ],
        });

        expect(builder.mapLayers.value.map((l) => l.type)).toEqual(["point", "heatmap"]);
    });

    it("persists the value only, so no translation is written back", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: { name: "Punkts", value: "point" },
                    config: {},
                },
            ],
        });

        const persisted = JSON.parse(JSON.stringify(builder.mapConfiguration.value));

        expect(persisted.layers[0].type).toBe("point");
    });

    it("leaves an already-migrated configuration alone", () => {
        const { builder } = setup("map");

        builder.setMapConfiguration({
            layers: [{ id: 0, type: "point", config: {} }],
        });

        expect(builder.mapLayers.value[0].type).toBe("point");
    });
});
