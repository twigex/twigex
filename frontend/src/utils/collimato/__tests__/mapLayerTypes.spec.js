// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";
import {
    MAP_LAYER_TYPES,
    MAP_LAYER_TYPE_IDS,
    layerTypeValue,
    defaultLayerConfig,
    layerTypeIsReady,
    layerPositions,
    layerQueryFields,
    withCountMeasure,
} from "@/utils/collimato/mapLayerTypes.js";
import { MAP_LAYER_EDITORS, editorFor } from "@/components/Collimato/Charts/Map/mapLayerEditors.js";
import { createDeckLayer, toFeatureCollection } from "@/utils/collimato/mapDeckLayers.js";

const LOCALES = { en, lv, kk, pl };

describe("every map layer type", () => {
    it.each(MAP_LAYER_TYPE_IDS)("%s declares defaults", (id) => {
        expect(MAP_LAYER_TYPES[id].defaults).toBeTruthy();
    });

    it.each(["point", "heatmap"])("%s starts with empty coordinates", (id) => {
        expect(MAP_LAYER_TYPES[id].defaults.coordinates).toEqual({
            lat: null,
            long: null,
        });
    });

    it.each(MAP_LAYER_TYPE_IDS)("%s has an editor component", (id) => {
        expect(editorFor(id)).toBeTruthy();
    });

    it("has no editor without a layer type behind it", () => {
        expect(Object.keys(MAP_LAYER_EDITORS).sort()).toEqual([...MAP_LAYER_TYPE_IDS].sort());
    });

    it.each(MAP_LAYER_TYPE_IDS)("%s builds a deck layer", (id) => {
        const layer = {
            name: "L",
            type: id,
            config: { ...defaultLayerConfig(id), data: [] },
        };

        expect(createDeckLayer(layer, id)).not.toBeNull();
    });

    it.each(MAP_LAYER_TYPE_IDS)("%s labels every default it declares", (id) => {
        expect(Object.keys(MAP_LAYER_TYPES[id].defaults).length).toBeGreaterThan(0);
    });

    it.each(MAP_LAYER_TYPE_IDS)("%s label resolves in every locale", (id) => {
        const { labelKey } = MAP_LAYER_TYPES[id];

        Object.entries(LOCALES).forEach(([locale, messages]) => {
            expect(messages[labelKey], `${locale} ${labelKey}`).toBeTruthy();
        });
    });
});

describe("defaultLayerConfig", () => {
    it("gives each caller its own copy", () => {
        const a = defaultLayerConfig("point");
        const b = defaultLayerConfig("point");

        a.radius = 99;
        a.coordinates.lat = "x";

        expect(b.radius).toBe(10);
        expect(b.coordinates.lat).toBeNull();
    });

    it("falls back to point for an unknown type", () => {
        expect(defaultLayerConfig("nope")).toEqual(defaultLayerConfig("point"));
    });
});

describe("layerTypeValue", () => {
    it("reads a plain string", () => {
        expect(layerTypeValue({ type: "point" })).toBe("point");
    });

    it("reads the legacy object shape saved by older maps", () => {
        expect(layerTypeValue({ type: { name: "Punkts", value: "point" } })).toBe("point");
    });

    it("returns nothing for a layer with no type", () => {
        expect(layerTypeValue({})).toBeUndefined();
    });
});

describe("createDeckLayer", () => {
    it("returns null for a type this build does not know", () => {
        expect(createDeckLayer({ name: "L", config: {} }, "hexagon")).toBeNull();
    });
});

describe("layer readiness", () => {
    it.each(MAP_LAYER_TYPE_IDS)("%s declares when it is ready", (id) => {
        expect(typeof MAP_LAYER_TYPES[id].ready).toBe("function");
    });

    it("is not ready with an empty config", () => {
        MAP_LAYER_TYPE_IDS.forEach((id) => {
            expect(layerTypeIsReady(id, defaultLayerConfig(id))).toBe(false);
        });
    });

    it("is ready once a point layer has both coordinates", () => {
        const config = defaultLayerConfig("point");

        config.coordinates = { lat: "t.lat", long: "t.long" };

        expect(layerTypeIsReady("point", config)).toBe(true);
    });

    it("is not ready with only one coordinate", () => {
        const config = defaultLayerConfig("point");

        config.coordinates = { lat: "t.lat", long: null };

        expect(layerTypeIsReady("point", config)).toBe(false);
    });

    it("is not ready for a type this build does not know", () => {
        expect(layerTypeIsReady("hexagon", {})).toBe(false);
    });
});

describe("layer positions", () => {
    it("reports the one coordinate pair a point contributes", () => {
        const config = defaultLayerConfig("point");

        config.coordinates = { lat: "t.lat", long: "t.long" };

        expect(layerPositions("point", config)).toEqual([{ lat: "t.lat", long: "t.long" }]);
    });

    it("reports nothing for a type this build does not know", () => {
        expect(layerPositions("geojson", {})).toEqual([]);
    });

    it("supports a type that contributes two pairs per row", () => {
        const twoEnded = {
            ready: (config) => Boolean(config.from.lat && config.to.lat),
            positions: (config) => [config.from, config.to],
            defaults: {},
        };

        MAP_LAYER_TYPES.__probe = twoEnded;

        const config = {
            from: { lat: "a.lat", long: "a.long" },
            to: { lat: "b.lat", long: "b.long" },
        };

        expect(layerPositions("__probe", config)).toHaveLength(2);
        expect(layerTypeIsReady("__probe", config)).toBe(true);

        delete MAP_LAYER_TYPES.__probe;
    });
});

describe("arc layer", () => {
    const ready = () => {
        const config = defaultLayerConfig("arc");

        config.source = { lat: "t.pickup_lat", long: "t.pickup_long" };
        config.target = { lat: "t.drop_lat", long: "t.drop_long" };

        return config;
    };

    it("needs both ends before it will load", () => {
        const config = defaultLayerConfig("arc");

        expect(layerTypeIsReady("arc", config)).toBe(false);

        config.source = { lat: "t.pickup_lat", long: "t.pickup_long" };
        expect(layerTypeIsReady("arc", config)).toBe(false);

        config.target = { lat: "t.drop_lat", long: "t.drop_long" };
        expect(layerTypeIsReady("arc", config)).toBe(true);
    });

    it("contributes both ends to the map extent", () => {
        expect(layerPositions("arc", ready())).toEqual([
            { lat: "t.pickup_lat", long: "t.pickup_long" },
            { lat: "t.drop_lat", long: "t.drop_long" },
        ]);
    });

    it("reads both ends off a row", () => {
        const layer = {
            name: "trips",
            type: "arc",
            config: {
                ...ready(),
                data: [
                    {
                        "t.pickup_lat": "40.75",
                        "t.pickup_long": "-73.99",
                        "t.drop_lat": "40.64",
                        "t.drop_long": "-73.78",
                    },
                ],
            },
        };

        const deckLayer = createDeckLayer(layer, "arc");
        const row = layer.config.data[0];

        expect(deckLayer.props.getSourcePosition(row)).toEqual([-73.99, 40.75]);
        expect(deckLayer.props.getTargetPosition(row)).toEqual([-73.78, 40.64]);
    });

    it("carries deck.gl's own shape options through", () => {
        const layer = { name: "trips", type: "arc", config: ready() };
        const deckLayer = createDeckLayer(layer, "arc");

        expect(deckLayer.props.greatCircle).toBe(false);
        expect(deckLayer.props.numSegments).toBe(50);
        expect(deckLayer.props.widthUnits).toBe("pixels");
        expect(deckLayer.props.getHeight).toBe(1);
        expect(deckLayer.props.getTilt).toBe(0);
    });
});

describe("geojson layer", () => {
    const geometry = '{"type":"Point","coordinates":[-73.99,40.75]}';

    it("needs a geometry column before it will load", () => {
        const config = defaultLayerConfig("geojson");

        expect(layerTypeIsReady("geojson", config)).toBe(false);

        config.geojson_field = "t.geom";
        expect(layerTypeIsReady("geojson", config)).toBe(true);
    });

    it("contributes nothing to the map extent", () => {
        expect(layerPositions("geojson", { geojson_field: "t.geom" })).toEqual([]);
    });

    it("wraps a bare geometry into a feature, keeping the row as properties", () => {
        const collection = toFeatureCollection([{ "t.geom": geometry, "t.fare": 12 }], "t.geom");

        expect(collection.type).toBe("FeatureCollection");
        expect(collection.features).toHaveLength(1);
        expect(collection.features[0].type).toBe("Feature");
        expect(collection.features[0].geometry.type).toBe("Point");
        expect(collection.features[0].properties["t.fare"]).toBe(12);
    });

    it("accepts a row that already holds a full feature", () => {
        const collection = toFeatureCollection(
            [
                {
                    "t.geom": JSON.stringify({
                        type: "Feature",
                        geometry: JSON.parse(geometry),
                        properties: { name: "pickup" },
                    }),
                    "t.fare": 12,
                },
            ],
            "t.geom",
        );

        expect(collection.features[0].properties.name).toBe("pickup");
        expect(collection.features[0].properties["t.fare"]).toBe(12);
    });

    it("accepts geometry that is already an object", () => {
        const collection = toFeatureCollection([{ "t.geom": JSON.parse(geometry) }], "t.geom");

        expect(collection.features).toHaveLength(1);
    });

    it("skips rows that cannot be parsed rather than throwing", () => {
        const collection = toFeatureCollection(
            [
                { "t.geom": "not json" },
                { "t.geom": null },
                { "t.geom": '{"nope":1}' },
                { "t.geom": geometry },
            ],
            "t.geom",
        );

        expect(collection.features).toHaveLength(1);
    });

    it("survives no rows at all", () => {
        expect(toFeatureCollection(null, "t.geom").features).toEqual([]);
    });

    it("reads colour fields out of the feature properties", () => {
        const layer = {
            name: "zones",
            type: "geojson",
            config: {
                ...defaultLayerConfig("geojson"),
                geojson_field: "t.geom",
                fill_color_field: "t.fare",
                data: [
                    { "t.geom": geometry, "t.fare": 10 },
                    { "t.geom": geometry, "t.fare": 20 },
                ],
            },
        };

        const deckLayer = createDeckLayer(layer, "geojson");
        const feature = deckLayer.props.data.features[1];

        expect(deckLayer.props.getFillColor(feature)).toHaveLength(3);
    });

    it("carries deck.gl's own render options through", () => {
        const layer = {
            name: "zones",
            type: "geojson",
            config: {
                ...defaultLayerConfig("geojson"),
                geojson_field: "t.geom",
            },
        };
        const deckLayer = createDeckLayer(layer, "geojson");

        expect(deckLayer.props.filled).toBe(true);
        expect(deckLayer.props.stroked).toBe(true);
        expect(deckLayer.props.extruded).toBe(false);
        expect(deckLayer.props.pointType).toBe("circle");
        expect(deckLayer.props.lineWidthUnits).toBe("meters");
        expect(deckLayer.props.getElevation).toBe(1000);
    });
});

describe("layer query fields", () => {
    it.each(MAP_LAYER_TYPE_IDS)("%s declares what it needs", (id) => {
        expect(typeof MAP_LAYER_TYPES[id].queryFields).toBe("function");
    });

    it.each(MAP_LAYER_TYPE_IDS)("%s asks for nothing when empty", (id) => {
        expect(layerQueryFields(id, defaultLayerConfig(id))).toEqual([]);
    });

    it("asks for a point's coordinates and the fields it colours by", () => {
        const config = defaultLayerConfig("point");

        config.coordinates = { lat: "t.lat", long: "t.long" };
        config.fill_color_field = "t.fare";
        config.popover = { fields: ["t.zone"] };

        expect(layerQueryFields("point", config)).toEqual(["t.lat", "t.long", "t.fare", "t.zone"]);
    });

    it("asks for both ends of an arc", () => {
        const config = defaultLayerConfig("arc");

        config.source = { lat: "a.lat", long: "a.long" };
        config.target = { lat: "b.lat", long: "b.long" };

        expect(layerQueryFields("arc", config)).toEqual(["a.lat", "a.long", "b.lat", "b.long"]);
    });

    it("does not ask for the same column twice", () => {
        const config = defaultLayerConfig("point");

        config.coordinates = { lat: "t.lat", long: "t.long" };
        config.fill_color_field = "t.lat";

        expect(layerQueryFields("point", config)).toEqual(["t.lat", "t.long"]);
    });

    it("asks for the geometry column of a geojson layer", () => {
        const config = defaultLayerConfig("geojson");

        config.geojson_field = "t.geom";

        expect(layerQueryFields("geojson", config)).toEqual(["t.geom"]);
    });

    it("asks for nothing for a type this build does not know", () => {
        expect(layerQueryFields("hexagon", {})).toEqual([]);
    });
});

describe("withCountMeasure", () => {
    const cubes = [
        { table: "taxi_trips", dimensions: [{ name: "taxi_trips.fare" }] },
        { table: "pickup_zone", dimensions: [{ name: "pickup_zone.borough" }] },
    ];

    it("offers the model's count alongside its dimensions", () => {
        const result = withCountMeasure(cubes, "taxi_trips", "Records");

        expect(result[0].dimensions[0]).toEqual({
            name: "taxi_trips.count",
            title: "Records",
            type: "number",
        });
        expect(result[0].dimensions[1].name).toBe("taxi_trips.fare");
    });

    it("adds it only to the layer's own model", () => {
        const result = withCountMeasure(cubes, "taxi_trips", "Records");

        expect(result[1].dimensions).toHaveLength(1);
        expect(result[1].dimensions[0].name).toBe("pickup_zone.borough");
    });

    it("leaves the original list untouched", () => {
        withCountMeasure(cubes, "taxi_trips", "Records");

        expect(cubes[0].dimensions).toHaveLength(1);
    });

    it("adds nothing when the model is not in the list", () => {
        const result = withCountMeasure(cubes, "weather", "Records");

        expect(result.flatMap((c) => c.dimensions)).toHaveLength(2);
    });

    it("appears in the layer's field list", () => {
        const config = defaultLayerConfig("arc");

        config.source = { lat: "a.lat", long: "a.long" };
        config.target = { lat: "b.lat", long: "b.long" };
        config.width_field = "taxi_trips.count";

        expect(layerQueryFields("arc", config)).toContain("taxi_trips.count");
    });
});
