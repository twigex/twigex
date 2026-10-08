// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { MAP_DEFAULT_FILL } from "@/utils/collimato/chartPalette.js";

function hasCoordinates(coordinates) {
    return Boolean(coordinates?.lat && coordinates?.long);
}

export const MAP_LAYER_TYPES = {
    point: {
        labelKey: "collimato.charts.new_chart.map.types.point",
        ready: (config) => hasCoordinates(config?.coordinates),
        positions: (config) => [config.coordinates],
        queryFields: (config) => [
            config.coordinates?.lat,
            config.coordinates?.long,
            config.radius_field,
            config.fill_color_field,
            ...(config.popover?.fields ?? []),
        ],
        defaults: {
            limit: null,
            fill_color: MAP_DEFAULT_FILL,
            fill_color_field: null,
            color_scheme: "interpolatePlasma",
            color_palette_steps: 20,
            line_color: "#ffffff",
            line_width: 5,
            radius: 10,
            radius_scale: 6,
            radius_field: null,
            min_radius: 1,
            max_radius: 10,
            coordinates: { lat: null, long: null },
            popover: { fields: [] },
            data: [],
        },
    },
    arc: {
        labelKey: "collimato.charts.new_chart.map.types.arc",
        ready: (config) => hasCoordinates(config?.source) && hasCoordinates(config?.target),
        positions: (config) => [config.source, config.target],
        queryFields: (config) => [
            config.source?.lat,
            config.source?.long,
            config.target?.lat,
            config.target?.long,
            config.width_field,
            config.source_color_field,
            config.target_color_field,
            ...(config.popover?.fields ?? []),
        ],
        defaults: {
            limit: null,
            source: { lat: null, long: null },
            target: { lat: null, long: null },
            source_color: "#4f46e5",
            target_color: "#f43f5e",
            source_color_field: null,
            target_color_field: null,
            source_color_scheme: "interpolatePlasma",
            target_color_scheme: "interpolatePlasma",
            width: 1,
            width_field: null,
            width_units: "pixels",
            width_scale: 1,
            min_width: 0,
            max_width: 10,
            height: 1,
            tilt: 0,
            great_circle: false,
            num_segments: 50,
            opacity: 0.8,
            popover: { fields: [] },
            data: [],
        },
    },
    geojson: {
        labelKey: "collimato.charts.new_chart.map.types.geojson",
        ready: (config) => Boolean(config?.geojson_field),
        positions: () => [],
        queryFields: (config) => [
            config.geojson_field,
            config.fill_color_field,
            ...(config.popover?.fields ?? []),
        ],
        defaults: {
            limit: null,
            geojson_field: null,
            filled: true,
            stroked: true,
            extruded: false,
            wireframe: false,
            fill_color: MAP_DEFAULT_FILL,
            fill_color_field: null,
            color_scheme: "interpolatePlasma",
            line_color: "#ffffff",
            line_width: 1,
            line_width_units: "meters",
            line_width_scale: 1,
            min_line_width: 0,
            max_line_width: 10,
            point_type: "circle",
            point_radius: 1,
            point_radius_units: "meters",
            point_radius_scale: 1,
            min_point_radius: 1,
            max_point_radius: 100,
            elevation: 1000,
            elevation_scale: 1,
            opacity: 0.8,
            popover: { fields: [] },
            data: [],
        },
    },
    heatmap: {
        labelKey: "collimato.charts.new_chart.map.types.heatmap",
        ready: (config) => hasCoordinates(config?.coordinates),
        positions: (config) => [config.coordinates],
        queryFields: (config) => [config.coordinates?.lat, config.coordinates?.long, config.weight],
        defaults: {
            limit: null,
            colorScheme: "interpolatePlasma",
            weight: null,
            radiusPixels: 30,
            intensity: 1,
            threshold: 0.03,
            coordinates: { lat: null, long: null },
            data: [],
        },
    },
};

export const MAP_LAYER_TYPE_IDS = Object.keys(MAP_LAYER_TYPES);

export function layerTypeValue(layer) {
    return typeof layer.type === "string" ? layer.type : layer.type?.value;
}

export function layerTypeIsReady(typeId, config) {
    const type = MAP_LAYER_TYPES[typeId];

    return type ? Boolean(type.ready(config)) : false;
}

export function withCountMeasure(cubes, modelName, title) {
    return cubes.map((cube) =>
        cube.table === modelName
            ? {
                  ...cube,
                  dimensions: [
                      {
                          name: `${modelName}.count`,
                          title: title,
                          type: "number",
                      },
                      ...cube.dimensions,
                  ],
              }
            : cube,
    );
}

export function layerQueryFields(typeId, config) {
    const type = MAP_LAYER_TYPES[typeId];

    if (!type?.queryFields) {
        return [];
    }

    return [...new Set(type.queryFields(config).filter(Boolean))];
}

export function layerPositions(typeId, config) {
    const type = MAP_LAYER_TYPES[typeId];

    return type?.positions ? type.positions(config).filter(Boolean) : [];
}

export function defaultLayerConfig(typeId) {
    const type = MAP_LAYER_TYPES[typeId] ?? MAP_LAYER_TYPES.point;

    return JSON.parse(JSON.stringify(type.defaults));
}
