// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi } from "vitest";
import { watch } from "vue";
import { useChartBuilder } from "@/composables/collimato/useChartBuilder.js";
import { defaultLayerConfig } from "@/utils/collimato/mapLayerTypes.js";

function setup(chartType = "bar") {
    return { builder: useChartBuilder(chartType) };
}

describe("loadLayer", () => {
    const model = {
        table: "earthquakes",
        join: 0,
        measures: [],
        dimensions: [{ name: "earthquakes.latitude" }, { name: "earthquakes.longitude" }],
    };

    const rows = [{ "earthquakes.latitude": 35.1 }];

    function mapSetup({ respond } = {}) {
        const statuses = [];
        const calls = [];
        const loadData = vi.fn((query, signal) => {
            calls.push({ query, signal });

            return respond
                ? respond(calls.length, signal)
                : Promise.resolve({ data: { data: rows } });
        });

        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [model];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "point",
                    model: "earthquakes",
                    config: {
                        coordinates: {
                            lat: "earthquakes.latitude",
                            long: "earthquakes.longitude",
                        },
                        data: null,
                    },
                },
            ],
        });

        watch(
            () => builder.layerStates.value.map((l) => [l.id, l.status]),
            (now, before) => {
                now.forEach(([id, status], i) => {
                    if (before?.[i]?.[1] !== status) {
                        statuses.push([id, status]);
                    }
                });
            },
            { flush: "sync" },
        );

        return { builder, statuses, calls, loadData };
    }

    it("writes the rows onto the layer and clears the status", async () => {
        const { builder, statuses } = mapSetup();

        await builder.loadLayer(0);

        expect(builder.mapLayers.value[0].config.data).toEqual(rows);
        expect(statuses).toEqual([
            [0, "loading"],
            [0, "loaded"],
        ]);
    });

    it("asks for every dimension of the layer's model", async () => {
        const { builder, calls } = mapSetup();

        await builder.loadLayer(0);

        expect(calls[0].query).toMatchObject({
            measures: ["earthquakes.count"],
            dimensions: ["earthquakes.latitude", "earthquakes.longitude"],
        });
    });

    it("does nothing until both coordinates are chosen", async () => {
        const { builder, loadData, statuses } = mapSetup();

        builder.mapLayers.value[0].config.coordinates.long = null;

        expect(await builder.loadLayer(0)).toBeNull();
        expect(loadData).not.toHaveBeenCalled();
        expect(statuses).toEqual([]);
    });

    it("does nothing when the layer has no model", async () => {
        const { builder, loadData } = mapSetup();

        builder.mapLayers.value[0].model = null;

        expect(await builder.loadLayer(0)).toBeNull();
        expect(loadData).not.toHaveBeenCalled();
    });

    it("clears the status even when the request fails", async () => {
        const { builder, statuses } = mapSetup({
            respond: () => Promise.reject(new Error("boom")),
        });

        await expect(builder.loadLayer(0)).rejects.toThrow("boom");
        expect(statuses).toEqual([
            [0, "loading"],
            [0, "loaded"],
        ]);
    });

    it("clears the status when the engine reports an error", async () => {
        const { builder, statuses } = mapSetup({
            respond: () => Promise.resolve({ data: { error: "Table not found" } }),
        });

        await expect(builder.loadLayer(0)).rejects.toThrow("Table not found");
        expect(statuses).toEqual([
            [0, "loading"],
            [0, "loaded"],
        ]);
    });

    it("aborts an in-flight load when the same layer reloads", async () => {
        let firstSignal = null;
        const { builder } = mapSetup({
            respond: (call, signal) => {
                if (call === 1) {
                    firstSignal = signal;

                    return new Promise(() => {});
                }

                return Promise.resolve({ data: { data: rows } });
            },
        });

        const first = builder.loadLayer(0);
        const second = builder.loadLayer(0);

        await expect(second).resolves.toEqual(rows);
        expect(firstSignal.aborted).toBe(true);
        await expect(first).resolves.toBeNull();
    });

    it("a superseded load does not clear the replacement's status", async () => {
        const { builder, statuses } = mapSetup({
            respond: (call) =>
                call === 1 ? new Promise(() => {}) : Promise.resolve({ data: { data: rows } }),
        });

        const first = builder.loadLayer(0);
        const second = builder.loadLayer(0);

        expect(builder.layerStates.value[0].status).toBe("loading");

        await second;
        await first;

        expect(statuses).toEqual([
            [0, "loading"],
            [0, "loaded"],
        ]);
        expect(builder.layerStates.value[0].status).toBe("loaded");
    });

    it("cancelling clears the status so the spinner cannot stick", async () => {
        const { builder, statuses } = mapSetup({
            respond: () => new Promise(() => {}),
        });

        const pending = builder.loadLayer(0);

        expect(builder.layerStates.value[0].status).toBe("loading");

        builder.cancelAllLayerLoads();

        await expect(pending).resolves.toBeNull();

        expect(builder.layerStates.value[0].status).toBe("");
        expect(statuses).toEqual([
            [0, "loading"],
            [0, ""],
        ]);
    });

    it("cancelling one layer leaves the others alone", async () => {
        const { builder } = mapSetup({
            respond: () => new Promise(() => {}),
        });

        const second = builder.addLayer({ name: "B", type: "point" });

        builder.setLayerModel(second.id, "earthquakes");
        builder.mapLayers.value[1].config.coordinates = {
            lat: "earthquakes.latitude",
            long: "earthquakes.longitude",
        };

        const first = builder.loadLayer(0);
        const other = builder.loadLayer(second.id);

        builder.cancelLayerLoad(0);
        await expect(first).resolves.toBeNull();

        expect(builder.layerStates.value[0].status).toBe("");
        expect(builder.layerStates.value[1].status).toBe("loading");

        builder.cancelAllLayerLoads();
        await other;
    });

    it("a load that takes over keeps its own spinner after the old one is cancelled", async () => {
        const { builder } = mapSetup({
            respond: (call) => (call === 1 ? new Promise(() => {}) : new Promise(() => {})),
        });

        const first = builder.loadLayer(0);
        const second = builder.loadLayer(0);

        await expect(first).resolves.toBeNull();

        expect(builder.layerStates.value[0].status).toBe("loading");

        builder.cancelAllLayerLoads();
        await second;
    });

    it("loadAllLayers reports per-layer failures without stopping", async () => {
        const { builder } = mapSetup({
            respond: (call) =>
                call === 1
                    ? Promise.reject(new Error("boom"))
                    : Promise.resolve({ data: { data: rows } }),
        });

        builder.mapLayers.value.push({
            id: 1,
            type: "point",
            model: "earthquakes",
            config: {
                coordinates: {
                    lat: "earthquakes.latitude",
                    long: "earthquakes.longitude",
                },
                data: null,
            },
        });

        const results = await builder.loadAllLayers();

        expect(results.map((r) => r.status)).toEqual(["rejected", "fulfilled"]);
        expect(builder.mapLayers.value[1].config.data).toEqual(rows);
    });
});

describe("layer filters", () => {
    const model = {
        table: "earthquakes",
        join: 0,
        measures: [],
        dimensions: [{ name: "earthquakes.latitude" }, { name: "earthquakes.longitude" }],
    };

    function filterSetup() {
        const calls = [];
        const loadData = vi.fn((query) => {
            calls.push(query);

            return Promise.resolve({ data: { data: [] } });
        });

        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [model];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "point",
                    model: "earthquakes",
                    config: {
                        coordinates: {
                            lat: "earthquakes.latitude",
                            long: "earthquakes.longitude",
                        },
                        data: null,
                    },
                },
            ],
        });

        return { builder, calls };
    }

    const magnitude = {
        member: "earthquakes.magnitude",
        operator: "gt",
        value: ["5"],
    };

    it("starts a new layer with no filters", () => {
        const { builder } = setup("map");

        expect(builder.addLayer({ name: "A", type: "point" }).filters).toEqual([]);
    });

    it("reloads the layer with the filter applied", async () => {
        const { builder, calls } = filterSetup();

        await builder.addLayerFilter(0, magnitude);

        expect(calls.at(-1).filters).toEqual([
            {
                member: "earthquakes.magnitude",
                operator: "gt",
                values: ["5"],
            },
        ]);
    });

    it("edits a filter in place rather than appending", async () => {
        const { builder, calls } = filterSetup();

        await builder.addLayerFilter(0, magnitude);
        await builder.updateLayerFilter(0, "earthquakes.magnitude", {
            ...magnitude,
            value: ["7"],
        });

        expect(builder.mapLayers.value[0].filters).toHaveLength(1);
        expect(calls.at(-1).filters[0].values).toEqual(["7"]);
    });

    it("reloads without the filter once it is removed", async () => {
        const { builder, calls } = filterSetup();

        await builder.addLayerFilter(0, magnitude);
        await builder.removeLayerFilter(0, builder.mapLayers.value[0].filters[0]);

        expect(builder.mapLayers.value[0].filters).toEqual([]);
        expect(calls.at(-1).filters).toEqual([]);
    });

    it("drops filters when the layer changes model", () => {
        const { builder } = filterSetup();

        builder.addLayerFilter(0, magnitude);
        builder.setLayerModel(0, "volcanoes");

        expect(builder.mapLayers.value[0].filters).toEqual([]);
    });

    it("keeps each layer's filters to itself", async () => {
        const { builder } = filterSetup();
        const second = builder.addLayer({ name: "B", type: "point" });

        await builder.addLayerFilter(0, magnitude);

        expect(second.filters).toEqual([]);
    });

    it("ignores a filter aimed at a layer that is gone", async () => {
        const { builder, calls } = filterSetup();

        await expect(builder.addLayerFilter(99, magnitude)).resolves.toBeNull();
        expect(calls).toHaveLength(0);
    });

    it("saves the filters with the map", async () => {
        const { builder } = filterSetup();

        await builder.addLayerFilter(0, magnitude);

        expect(builder.mapConfiguration.value.layers[0].filters).toEqual([magnitude]);
    });
});

describe("map layers over joined models", () => {
    const trips = {
        table: "taxi_trips",
        join: 1,
        measures: [],
        dimensions: [{ name: "taxi_trips.pu_location_id" }, { name: "taxi_trips.fare_amount" }],
    };

    const pickup = {
        table: "pickup_zone",
        join: 1,
        measures: [],
        dimensions: [{ name: "pickup_zone.latitude" }, { name: "pickup_zone.longitude" }],
    };

    const unrelated = {
        table: "weather",
        join: 2,
        measures: [],
        dimensions: [{ name: "weather.temp" }],
    };

    function setupTrips() {
        const calls = [];
        const loadData = vi.fn((query) => {
            calls.push(query);

            return Promise.resolve({ data: { data: [] } });
        });

        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [trips, pickup, unrelated];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "arc",
                    model: "taxi_trips",
                    config: {
                        source: {
                            lat: "pickup_zone.latitude",
                            long: "pickup_zone.longitude",
                        },
                        target: {
                            lat: "pickup_zone.latitude",
                            long: "pickup_zone.longitude",
                        },
                        data: null,
                    },
                },
            ],
        });

        return { builder, calls };
    }

    it("asks only for the columns the layer draws with", async () => {
        const { builder, calls } = setupTrips();

        await builder.loadLayer(0);

        expect(calls[0].dimensions).toEqual(["pickup_zone.latitude", "pickup_zone.longitude"]);
    });

    it("reaches columns that live on a joined model", async () => {
        const { builder, calls } = setupTrips();

        builder.mapLayers.value[0].config.target = {
            lat: "dropoff_zone.latitude",
            long: "dropoff_zone.longitude",
        };
        await builder.loadLayer(0);

        expect(calls.at(-1).dimensions).toEqual([
            "pickup_zone.latitude",
            "pickup_zone.longitude",
            "dropoff_zone.latitude",
            "dropoff_zone.longitude",
        ]);
    });

    it("leaves out columns the layer does not use", async () => {
        const { builder, calls } = setupTrips();

        await builder.loadLayer(0);

        expect(calls[0].dimensions).not.toContain("taxi_trips.fare_amount");
        expect(calls[0].dimensions).not.toContain("weather.temp");
    });

    it("includes tooltip fields, since they must survive the grouping", async () => {
        const { builder, calls } = setupTrips();

        builder.mapLayers.value[0].config.popover = {
            fields: ["taxi_trips.fare_amount"],
        };
        await builder.loadLayer(0);

        expect(calls.at(-1).dimensions).toContain("taxi_trips.fare_amount");
    });

    it("still asks for the model's count measure", async () => {
        const { builder, calls } = setupTrips();

        await builder.loadLayer(0);

        expect(calls[0].measures).toEqual(["taxi_trips.count"]);
    });
});

describe("row limits", () => {
    const model = {
        table: "t",
        join: 0,
        measures: [],
        dimensions: [{ name: "t.lat" }, { name: "t.long" }],
    };

    function setupLimit(rowCount, limit) {
        const calls = [];
        const loadData = vi.fn((query) => {
            calls.push(query);

            return Promise.resolve({
                data: { data: Array.from({ length: rowCount }, () => ({})) },
            });
        });

        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [model];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "point",
                    model: "t",
                    config: {
                        coordinates: { lat: "t.lat", long: "t.long" },
                        limit: limit,
                        data: null,
                    },
                },
            ],
        });

        return { builder, calls };
    }

    it("sorts by the count so a cut result keeps the busiest rows", async () => {
        const { builder, calls } = setupLimit(1, null);

        await builder.loadLayer(0);

        expect(calls[0].order).toEqual([["t.count", "desc"]]);
    });

    it("asks for the limit the layer was given", async () => {
        const { builder, calls } = setupLimit(1, 100);

        await builder.loadLayer(0);

        expect(calls[0].limit).toBe(100);
    });

    it("flags a layer that came back at its limit", async () => {
        const { builder } = setupLimit(100, 100);

        await builder.loadLayer(0);

        expect(builder.layerStates.value[0].truncated).toBe(true);
    });

    it("does not flag a layer that came back under its limit", async () => {
        const { builder } = setupLimit(42, 100);

        await builder.loadLayer(0);

        expect(builder.layerStates.value[0].truncated).toBe(false);
    });

    it("forgets the flag when the layer is removed", async () => {
        const { builder } = setupLimit(100, 100);

        await builder.loadLayer(0);
        builder.removeLayer(0);

        expect(builder.layerStates.value).toEqual([]);
    });
});

describe("config changes that alter the query", () => {
    const model = {
        table: "t",
        join: 0,
        measures: [],
        dimensions: [{ name: "t.lat" }, { name: "t.long" }, { name: "t.zone" }, { name: "t.fare" }],
    };

    function setupConfig() {
        const calls = [];
        const loadData = vi.fn((query) => {
            calls.push(query);

            return Promise.resolve({ data: { data: [] } });
        });

        const builder = useChartBuilder("map", { loadData });

        builder.dataModels.value = [model];
        builder.setMapConfiguration({
            layers: [
                {
                    id: 0,
                    type: "point",
                    model: "t",
                    config: {
                        ...defaultLayerConfig("point"),
                        coordinates: { lat: "t.lat", long: "t.long" },
                        data: null,
                    },
                },
            ],
        });

        return { builder, calls };
    }

    it("refetches when a tooltip field is added, so the column exists", async () => {
        const { builder, calls } = setupConfig();

        await builder.setLayerConfig(0, { popover: { fields: ["t.zone"] } });

        expect(calls).toHaveLength(1);
        expect(calls[0].dimensions).toContain("t.zone");
    });

    it("refetches when the colour field changes", async () => {
        const { builder, calls } = setupConfig();

        await builder.setLayerConfig(0, { fill_color_field: "t.fare" });

        expect(calls.at(-1).dimensions).toContain("t.fare");
    });

    it("refetches when the row limit changes", async () => {
        const { builder, calls } = setupConfig();

        await builder.setLayerConfig(0, { limit: 100 });

        expect(calls.at(-1).limit).toBe(100);
    });

    it("does not refetch for a change the query cannot see", async () => {
        const { builder, calls } = setupConfig();

        await builder.setLayerConfig(0, { fill_color: "#ff0000" });
        await builder.setLayerConfig(0, { radius: 20 });

        expect(calls).toHaveLength(0);
    });

    it("still applies a change that does not refetch", async () => {
        const { builder } = setupConfig();

        await builder.setLayerConfig(0, { fill_color: "#ff0000" });

        expect(builder.mapLayers.value[0].config.fill_color).toBe("#ff0000");
    });
});
