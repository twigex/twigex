<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div ref="mapContainer" :style="{ width: '100%', height: '100%' }"></div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from "vue";
import { Map as MaplibreMap, setWorkerUrl } from "maplibre-gl";
import maplibreWorkerUrl from "maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url";
import "maplibre-gl/dist/maplibre-gl.css";
import { MapboxOverlay } from "@deck.gl/mapbox";
import { formatTooltip } from "@/utils/collimato/mapTooltip.js";
import { createDeckLayer } from "@/utils/collimato/mapDeckLayers.js";
import { layerTypeValue, layerPositions } from "@/utils/collimato/mapLayerTypes.js";

setWorkerUrl(maplibreWorkerUrl);

const DEFAULT_MAP_STYLE = "https://basemaps.cartocdn.com/gl/dark-matter-gl-style/style.json";

const props = defineProps({
    data: {
        type: Object,
        required: true,
    },
});

const mapContainer = ref(null);
let map;
let deckOverlay;

const mapLayers = ref([]);

function renderLayers() {
    if (!deckOverlay) {
        return;
    }

    const deckLayers = [];

    mapLayers.value.forEach((layer) => {
        if (!layer.show) {
            return;
        }

        const deckLayer = createDeckLayer(layer, layerTypeValue(layer));

        if (deckLayer) {
            deckLayers.push(deckLayer);
        }
    });

    deckOverlay.setProps({ layers: deckLayers });
}

function updateLayers(layers) {
    mapLayers.value = layers ?? [];
    renderLayers();
}

watch(
    () => props.data.map_style,
    (style) => {
        if (map) {
            map.setStyle(style || DEFAULT_MAP_STYLE);
        }
    },
);

watch(
    () => [props.data.longitude, props.data.latitude, props.data.zoom].join("|"),
    () => {
        if (!map) {
            return;
        }

        const { longitude, latitude, zoom } = props.data;
        const center = [Number(longitude), Number(latitude)];
        const hasCenter = center.every(Number.isFinite) && longitude !== "" && latitude !== "";

        if (hasCenter) {
            map.jumpTo({ center: center, zoom: Number(zoom) || map.getZoom() });

            return;
        }

        if (Number.isFinite(Number(zoom))) {
            map.zoomTo(Number(zoom), { duration: 0 });
        }
    },
);

onMounted(() => {
    map = new MaplibreMap({
        container: mapContainer.value,
        style: props.data.map_style || DEFAULT_MAP_STYLE,
        center:
            props.data.longitude && props.data.latitude
                ? [props.data.longitude, props.data.latitude]
                : [24.104716, 56.949659],
        zoom: props.data.zoom ? props.data.zoom : 1,
    });

    map.on("load", () => {
        deckOverlay = new MapboxOverlay({
            interleaved: false,
            layers: [],
            getCursor: ({ isHovering }) => (isHovering ? "pointer" : "grab"),
            getTooltip: ({ object, layer }) => {
                const text = formatTooltip(layer?.props?.tooltipFields, object);

                if (!text) {
                    return null;
                }

                return {
                    text: text,
                    style: { padding: "8px", fontSize: "12px" },
                };
            },
        });

        map.addControl(deckOverlay);
        renderLayers();
    });
});

function layerBounds() {
    let west = Infinity;
    let south = Infinity;
    let east = -Infinity;
    let north = -Infinity;
    let found = false;

    mapLayers.value.forEach((layer) => {
        const rows = layer.config?.data;
        const positions = layerPositions(layerTypeValue(layer), layer.config);

        if (!positions.length || !Array.isArray(rows)) {
            return;
        }

        rows.forEach((row) => {
            positions.forEach((position) => {
                const lng = Number(row[position.long]);
                const lat = Number(row[position.lat]);

                if (!Number.isFinite(lng) || !Number.isFinite(lat)) {
                    return;
                }

                west = Math.min(west, lng);
                east = Math.max(east, lng);
                south = Math.min(south, lat);
                north = Math.max(north, lat);
                found = true;
            });
        });
    });

    return found ? [west, south, east, north] : null;
}

function fitToLayers() {
    const bounds = layerBounds();

    if (!map || !bounds) {
        return false;
    }

    const [west, south, east, north] = bounds;

    map.fitBounds(
        [
            [west, south],
            [east, north],
        ],
        { padding: 48, maxZoom: 12, duration: 0 },
    );

    return true;
}

onUnmounted(() => {
    if (map) map.remove();
});

defineExpose({
    updateLayers,
    fitToLayers,
});
</script>
