// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { buildLayerQuery, layerQuery } from "@/utils/collimato/mapQuery.js";
import { defaultLayerConfig } from "@/utils/collimato/mapLayerTypes.js";
import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";

describe("buildLayerQuery", () => {
    it("asks for the model's count grouped by the given dimensions", () => {
        const query = buildLayerQuery("earthquakes", [
            "earthquakes.latitude",
            "earthquakes.longitude",
            "earthquakes.magnitude",
        ]);

        expect(query).toEqual({
            measures: ["earthquakes.count"],
            dimensions: ["earthquakes.latitude", "earthquakes.longitude", "earthquakes.magnitude"],
            order: [["earthquakes.count", "desc"]],
            filters: [],
            timeDimensions: [],
            limit: MAX_QUERY_LIMIT,
        });
    });

    it("keeps the dimension order it was given", () => {
        const names = ["m.c", "m.a", "m.b"];

        expect(buildLayerQuery("m", names).dimensions).toEqual(names);
    });

    it("copies the dimension list so callers cannot be mutated", () => {
        const names = ["m.a"];
        const query = buildLayerQuery("m", names);

        query.dimensions.push("m.b");

        expect(names).toEqual(["m.a"]);
    });

    it("returns a fresh object each call", () => {
        const a = buildLayerQuery("m", ["m.a"]);
        const b = buildLayerQuery("m", ["m.a"]);

        expect(a).not.toBe(b);
        expect(a.filters).not.toBe(b.filters);
    });

    it("still asks for count when the model exposes no dimensions", () => {
        const query = buildLayerQuery("m");

        expect(query.measures).toEqual(["m.count"]);
        expect(query.dimensions).toEqual([]);
    });
});

describe("layer filters", () => {
    it("sends no filters when the layer has none", () => {
        expect(buildLayerQuery("earthquakes", []).filters).toEqual([]);
    });

    it("passes a filter through to the query", () => {
        const query = buildLayerQuery(
            "earthquakes",
            ["earthquakes.latitude"],
            [
                {
                    member: "earthquakes.magnitude",
                    operator: "gt",
                    value: ["5"],
                },
            ],
        );

        expect(query.filters).toEqual([
            {
                member: "earthquakes.magnitude",
                operator: "gt",
                values: ["5"],
            },
        ]);
    });

    it("merges filters on the same member and operator", () => {
        const query = buildLayerQuery(
            "earthquakes",
            [],
            [
                {
                    member: "earthquakes.region",
                    operator: "equals",
                    value: "a",
                },
                {
                    member: "earthquakes.region",
                    operator: "equals",
                    value: "b",
                },
            ],
        );

        expect(query.filters).toEqual([
            {
                member: "earthquakes.region",
                operator: "equals",
                values: ["a", "b"],
            },
        ]);
    });

    it("treats a layer saved before filters existed as unfiltered", () => {
        expect(buildLayerQuery("earthquakes", [], undefined).filters).toEqual([]);
    });

    it("keeps a different operator on the same member apart", () => {
        const query = buildLayerQuery(
            "earthquakes",
            [],
            [
                { member: "earthquakes.depth", operator: "gt", value: "1" },
                { member: "earthquakes.depth", operator: "lt", value: "9" },
            ],
        );

        expect(query.filters).toHaveLength(2);
    });
});

describe("row limit", () => {
    it("uses the maximum when none is set", () => {
        expect(buildLayerQuery("t", []).limit).toBe(MAX_QUERY_LIMIT);
    });

    it("uses a smaller limit when one is given", () => {
        expect(buildLayerQuery("t", [], [], 500).limit).toBe(500);
    });

    it("will not exceed the maximum the backend allows", () => {
        expect(buildLayerQuery("t", [], [], MAX_QUERY_LIMIT * 10).limit).toBe(MAX_QUERY_LIMIT);
    });

    it("treats zero and nonsense as unset", () => {
        [0, -5, null, undefined, "abc"].forEach((value) => {
            expect(buildLayerQuery("t", [], [], value).limit).toBe(MAX_QUERY_LIMIT);
        });
    });
});

describe("the count measure", () => {
    it("is never sent as a dimension, since it is already a measure", () => {
        const query = buildLayerQuery("taxi_trips", ["pickup_zone.latitude", "taxi_trips.count"]);

        expect(query.measures).toEqual(["taxi_trips.count"]);
        expect(query.dimensions).toEqual(["pickup_zone.latitude"]);
    });

    it("leaves another model's count alone", () => {
        const query = buildLayerQuery("taxi_trips", ["pickup_zone.count"]);

        expect(query.dimensions).toEqual(["pickup_zone.count"]);
    });
});

describe("layerQuery", () => {
    function arc() {
        return {
            model: "taxi_trips",
            type: "arc",
            filters: [],
            config: {
                ...defaultLayerConfig("arc"),
                source: {
                    lat: "pickup_zone.latitude",
                    long: "pickup_zone.longitude",
                },
                target: {
                    lat: "dropoff_zone.latitude",
                    long: "dropoff_zone.longitude",
                },
            },
        };
    }

    it("asks for the columns the layer draws with, wherever they live", () => {
        expect(layerQuery(arc()).dimensions).toEqual([
            "pickup_zone.latitude",
            "pickup_zone.longitude",
            "dropoff_zone.latitude",
            "dropoff_zone.longitude",
        ]);
    });

    it("counts the layer's own model", () => {
        expect(layerQuery(arc()).measures).toEqual(["taxi_trips.count"]);
    });

    it("carries the layer's filters", () => {
        const layer = arc();

        layer.filters = [{ member: "pickup_zone.zone", operator: "equals", value: ["JFK"] }];

        expect(layerQuery(layer).filters).toEqual([
            {
                member: "pickup_zone.zone",
                operator: "equals",
                values: ["JFK"],
            },
        ]);
    });

    it("carries the layer's row limit", () => {
        const layer = arc();

        layer.config.limit = 250;

        expect(layerQuery(layer).limit).toBe(250);
    });

    it("reads the legacy object shape of a layer type", () => {
        const layer = arc();

        layer.type = { name: "Loks", value: "arc" };

        expect(layerQuery(layer).dimensions).toHaveLength(4);
    });
});
