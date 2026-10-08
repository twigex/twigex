// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import PointLayer from "@/components/Collimato/Charts/Map/LayerTypes/PointLayer.vue";
import HeatLayer from "@/components/Collimato/Charts/Map/LayerTypes/HeatLayer.vue";
import ArcLayer from "@/components/Collimato/Charts/Map/LayerTypes/ArcLayer.vue";
import GeoJsonLayer from "@/components/Collimato/Charts/Map/LayerTypes/GeoJsonLayer.vue";

export const MAP_LAYER_EDITORS = {
    point: PointLayer,
    heatmap: HeatLayer,
    arc: ArcLayer,
    geojson: GeoJsonLayer,
};

export function editorFor(typeId) {
    return MAP_LAYER_EDITORS[typeId] ?? null;
}
