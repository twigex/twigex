// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ScatterplotLayer, ArcLayer, GeoJsonLayer } from "@deck.gl/layers";
import { HeatmapLayer } from "@deck.gl/aggregation-layers";
import { scaleLinear } from "d3";
import * as d3 from "d3";
import { hexToRGB } from "@/utils/utils";

export function extent(values) {
    let min = Infinity;
    let max = -Infinity;

    values.forEach((value) => {
        if (!Number.isFinite(value)) {
            return;
        }

        min = Math.min(min, value);
        max = Math.max(max, value);
    });

    return Number.isFinite(min) ? [min, max] : [0, 0];
}

export function createGetRadius(data, field, range = [4, 100], radius) {
    if (field == null || data == null) {
        return () => radius;
    }

    const [minValue, maxValue] = extent(data.map((d) => Number(d[field])));
    const radiusScale = scaleLinear()
        .domain([0, Math.sqrt(maxValue - minValue)])
        .range(range);

    return (d) => {
        const value = Number(d[field]);

        if (!Number.isFinite(value)) {
            return range[0];
        }

        return radiusScale(Math.sqrt(value - minValue));
    };
}

export function createColorAccessor(
    data,
    field,
    domain = null,
    colorScheme = "interpolateViridis",
    fill_color = "#fff",
) {
    if (field == null || data == null) {
        return () => hexToRGB(fill_color);
    }

    let dataDomain;

    if (!domain) {
        dataDomain = extent(data.map((d) => Number(d[field])));
    } else {
        dataDomain = domain;
    }

    const valueToDomainScale = d3.scaleLinear().domain(dataDomain).range([0, 20]);

    const colorScale = d3.scaleSequential(d3[colorScheme]).domain([0, 20]);

    return (d) => {
        const value = Number(d[field]);
        const scaledValue = valueToDomainScale(value);
        const colorStr = colorScale(scaledValue);
        const c = d3.color(colorStr);

        return [c.r, c.g, c.b];
    };
}

function baseDeckProps(layer, layerType) {
    const { lat, long } = layer.config.coordinates;

    return {
        id: `${layer.name}-${layerType}`,
        data: layer.config.data,
        getPosition: (d) => [Number(d[long]), Number(d[lat])],
    };
}

export function buildPointLayer(layer, layerType) {
    return new ScatterplotLayer({
        ...baseDeckProps(layer, layerType),
        getRadius: createGetRadius(
            layer.config.data,
            layer.config.radius_field,
            [layer.config.min_radius, layer.config.max_radius],
            layer.config.radius,
        ),
        radiusUnits: "pixels",
        radiusMinPixels: 1,
        getFillColor: createColorAccessor(
            layer.config.data,
            layer.config.fill_color_field,
            null,
            layer.config.color_scheme,
            layer.config.fill_color,
        ),
        autoHighlight: true,
        highlightColor: [255, 255, 0],
        opacity: 0.8,
        pickable: true,
        tooltipFields: [...(layer.config.popover?.fields ?? [])],
        updateTriggers: {
            getRadius: [
                layer.config.radius_field,
                layer.config.min_radius,
                layer.config.max_radius,
                layer.config.radius,
            ],
            getFillColor: [
                layer.config.color_scheme,
                layer.config.fill_color_field,
                layer.config.fill_color,
            ],
        },
    });
}

export function buildHeatmapLayer(layer, layerType) {
    const steps = 10;
    const colorRange = Array.from({ length: steps }, (_, i) => {
        const c = d3.rgb(d3[layer.config.colorScheme](i / (steps - 1)));

        return [c.r, c.g, c.b, Math.round(c.opacity * 255)];
    });

    return new HeatmapLayer({
        ...baseDeckProps(layer, layerType),
        getWeight: (d) => (layer.config.weight != null ? Number(d[layer.config.weight]) : 1),
        radiusPixels: layer.config.radiusPixels,
        colorRange: colorRange,
        intensity: layer.config.intensity,
        threshold: layer.config.threshold,
        updateTriggers: {
            getWeight: [layer.config.weight],
            colorRange: [layer.config.colorScheme],
            radiusPixels: [layer.config.radiusPixels],
        },
    });
}

export function buildArcLayer(layer, layerType) {
    const { source, target } = layer.config;

    return new ArcLayer({
        id: `${layer.name}-${layerType}`,
        data: layer.config.data,
        getSourcePosition: (d) => [Number(d[source.long]), Number(d[source.lat])],
        getTargetPosition: (d) => [Number(d[target.long]), Number(d[target.lat])],
        getSourceColor: createColorAccessor(
            layer.config.data,
            layer.config.source_color_field,
            null,
            layer.config.source_color_scheme ?? layer.config.color_scheme,
            layer.config.source_color,
        ),
        getTargetColor: createColorAccessor(
            layer.config.data,
            layer.config.target_color_field,
            null,
            layer.config.target_color_scheme ?? layer.config.color_scheme,
            layer.config.target_color,
        ),
        getWidth: createGetRadius(
            layer.config.data,
            layer.config.width_field,
            [layer.config.min_width, layer.config.max_width],
            layer.config.width,
        ),
        getHeight: layer.config.height,
        getTilt: layer.config.tilt,
        widthUnits: layer.config.width_units,
        widthScale: layer.config.width_scale,
        widthMinPixels: layer.config.min_width,
        widthMaxPixels: layer.config.max_width,
        greatCircle: layer.config.great_circle,
        numSegments: layer.config.num_segments,
        opacity: layer.config.opacity,
        autoHighlight: true,
        highlightColor: [255, 255, 0],
        pickable: true,
        tooltipFields: [...(layer.config.popover?.fields ?? [])],
        updateTriggers: {
            getSourcePosition: [source.lat, source.long],
            getTargetPosition: [target.lat, target.long],
            getSourceColor: [
                layer.config.source_color,
                layer.config.source_color_field,
                layer.config.source_color_scheme,
            ],
            getTargetColor: [
                layer.config.target_color,
                layer.config.target_color_field,
                layer.config.target_color_scheme,
            ],
            getWidth: [
                layer.config.width,
                layer.config.width_field,
                layer.config.min_width,
                layer.config.max_width,
            ],
        },
    });
}

export function toFeatureCollection(rows, field) {
    const features = [];

    (rows ?? []).forEach((row) => {
        const raw = row[field];

        if (!raw) {
            return;
        }

        let parsed;

        try {
            parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
        } catch {
            return;
        }

        if (!parsed?.type) {
            return;
        }

        features.push(
            parsed.type === "Feature"
                ? {
                      ...parsed,
                      properties: { ...row, ...(parsed.properties ?? {}) },
                  }
                : { type: "Feature", geometry: parsed, properties: row },
        );
    });

    return { type: "FeatureCollection", features: features };
}

function fromProperties(accessor) {
    return (feature) => accessor(feature.properties ?? {});
}

export function buildGeoJsonLayer(layer, layerType) {
    const config = layer.config;

    return new GeoJsonLayer({
        id: `${layer.name}-${layerType}`,
        data: toFeatureCollection(config.data, config.geojson_field),
        filled: config.filled,
        stroked: config.stroked,
        extruded: config.extruded,
        wireframe: config.wireframe,
        pointType: config.point_type,
        lineWidthUnits: config.line_width_units,
        lineWidthScale: config.line_width_scale,
        lineWidthMinPixels: config.min_line_width,
        lineWidthMaxPixels: config.max_line_width,
        pointRadiusUnits: config.point_radius_units,
        pointRadiusScale: config.point_radius_scale,
        pointRadiusMinPixels: config.min_point_radius,
        pointRadiusMaxPixels: config.max_point_radius,
        elevationScale: config.elevation_scale,
        getFillColor: fromProperties(
            createColorAccessor(
                config.data,
                config.fill_color_field,
                null,
                config.color_scheme,
                config.fill_color,
            ),
        ),
        getLineColor: () => hexToRGB(config.line_color),
        getLineWidth: config.line_width,
        getPointRadius: config.point_radius,
        getElevation: config.elevation,
        opacity: config.opacity,
        autoHighlight: true,
        highlightColor: [255, 255, 0],
        pickable: true,
        tooltipFields: [...(config.popover?.fields ?? [])],
        updateTriggers: {
            data: [config.geojson_field],
            getFillColor: [config.fill_color, config.fill_color_field, config.color_scheme],
            getLineColor: [config.line_color],
        },
    });
}

export function createDeckLayer(layer, layerType) {
    if (layerType === "point") {
        return buildPointLayer(layer, layerType);
    }

    if (layerType === "heatmap") {
        return buildHeatmapLayer(layer, layerType);
    }

    if (layerType === "arc") {
        return buildArcLayer(layer, layerType);
    }

    if (layerType === "geojson") {
        return buildGeoJsonLayer(layer, layerType);
    }

    return null;
}
