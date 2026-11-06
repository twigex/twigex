<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="absolute inset-0">
        <div
            v-if="map.layout.value === 'columns' && headers.length > 0"
            class="absolute inset-x-0 top-0 h-9 border-b border-gray-100 bg-white"
        >
            <div
                v-for="header in headers"
                :key="header.type"
                class="absolute top-1/2 flex -translate-x-1/2 -translate-y-1/2 items-center gap-x-1.5 whitespace-nowrap text-xs font-semibold uppercase tracking-wide text-gray-500"
                :style="{ left: header.left + 'px' }"
            >
                <span class="size-2.5 rounded-full" :style="{ background: header.color }"></span>
                {{ header.title }}
            </div>
        </div>

        <div
            ref="container"
            :class="[
                map.layout.value === 'columns' ? 'top-9' : 'top-0',
                'absolute inset-x-0 bottom-0',
            ]"
        ></div>

        <div class="absolute bottom-4 right-4 flex flex-col gap-y-1.5">
            <button
                v-for="control in zoomControls"
                :key="control.label"
                type="button"
                class="flex size-9 items-center justify-center rounded-md bg-white text-gray-700 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                :title="control.label"
                @click="control.run"
            >
                <span class="sr-only">{{ control.label }}</span>
                <component :is="control.icon" class="size-4" aria-hidden="true" />
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, watch } from "vue";
import Sigma from "sigma";
import { EdgeCurvedArrowProgram } from "@sigma/edge-curve";
import { ArrowsPointingOutIcon, MinusIcon, PlusIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";

const props = defineProps({
    map: {
        type: Object,
        required: true,
    },
});

const emit = defineEmits(["inspect", "expand"]);

const EDGE_COLOR = "#d1d5db";
const EDGE_FOCUS_COLOR = "#6366f1";
const EDGE_PATH_COLOR = "#a5b4fc";
const EDGE_FADED_COLOR = "#f3f4f6";
const NODE_FADED_COLOR = "#e5e7eb";
const FREE_CURVATURE = 0.15;
const LABEL_FONT = "Inter, ui-sans-serif, system-ui, sans-serif";
const LABEL_COLOR = "#1f2937";
const ZOOM_DURATION = 200;

const container = ref(null);
const headers = ref([]);

let renderer = null;
let resizeObserver = null;
let hoveredKey = "";
let path = { keys: [], edges: new Set() };

const zoomControls = computed(() => [
    {
        label: t.value("graph.zoom.in"),
        icon: PlusIcon,
        run: () => camera().animatedZoom({ duration: ZOOM_DURATION }),
    },
    {
        label: t.value("graph.zoom.out"),
        icon: MinusIcon,
        run: () => camera().animatedUnzoom({ duration: ZOOM_DURATION }),
    },
    {
        label: t.value("graph.zoom.fit"),
        icon: ArrowsPointingOutIcon,
        run: () => camera().animatedReset({ duration: ZOOM_DURATION }),
    },
]);

function camera() {
    return renderer.getCamera();
}

function focusKey() {
    if (hoveredKey) {
        return hoveredKey;
    }

    return props.map.selectedKey.value;
}

function isNearFocus(key, focus) {
    if (key === focus) {
        return true;
    }

    return props.map.graph.areNeighbors(key, focus);
}

// With a dot in focus (pointed at, or else selected), its own lines and
// neighbours stay strong and the rest fades, so each line can be followed.
// The starting dot and the path from it to the selected dot never fade, and
// the type chips never hide them.
function drawNode(key, data) {
    const drawn = { ...data };
    const focus = focusKey();

    const isStart = key === props.map.startKey.value;
    const isSelected = key === props.map.selectedKey.value;
    const onPath = isStart || path.keys.includes(key);

    if (isStart) {
        drawn.forceLabel = true;
    }

    if (isStart || isSelected) {
        drawn.highlighted = true;
        drawn.size = data.size + 6;
    }

    if (onPath) {
        return drawn;
    }

    if (!props.map.isTypeShown(data.columnType)) {
        drawn.hidden = true;

        return drawn;
    }

    if (focus && !isNearFocus(key, focus)) {
        drawn.color = NODE_FADED_COLOR;
        drawn.label = "";
        drawn.faded = true;
    }

    return drawn;
}

// Only the lines in focus carry a label, so labels don't pile up.
function drawEdge(edge, data) {
    const drawn = { ...data, color: EDGE_COLOR, label: "" };
    const focus = focusKey();

    if (props.map.layout.value === "free") {
        drawn.curvature = FREE_CURVATURE;
    }

    if (!focus) {
        return drawn;
    }

    const graph = props.map.graph;

    if (graph.hasExtremity(edge, focus)) {
        drawn.color = EDGE_FOCUS_COLOR;
        drawn.size = 3;
        drawn.label = data.label.toUpperCase();
        drawn.zIndex = 1;

        return drawn;
    }

    if (path.edges.has(edge)) {
        drawn.color = EDGE_PATH_COLOR;
        drawn.size = 2.5;
        drawn.zIndex = 1;

        return drawn;
    }

    drawn.color = EDGE_FADED_COLOR;

    return drawn;
}

// Labels go under the dots, and a group box shows its count inside.
function drawLabel(context, data, settings) {
    if (data.isGroup && !data.faded) {
        context.font = `700 ${settings.labelSize - 1}px ${LABEL_FONT}`;
        context.fillStyle = "#ffffff";
        context.textAlign = "center";
        context.textBaseline = "middle";
        context.fillText(data.countText, data.x, data.y);
    }

    if (!data.label) {
        return;
    }

    let weight = 500;

    if (data.highlighted) {
        weight = 600;
    }

    context.font = `${weight} ${settings.labelSize}px ${LABEL_FONT}`;
    context.fillStyle = LABEL_COLOR;
    context.textAlign = "center";
    context.textBaseline = "top";
    context.fillText(data.label, data.x, data.y + data.size + 4);
}

// A dot in focus gets a soft ring and its label on a white background, so it
// stays readable over lines.
function drawHover(context, data, settings) {
    context.beginPath();
    context.arc(data.x, data.y, data.size + 6, 0, Math.PI * 2);
    context.fillStyle = "rgba(79, 70, 229, 0.12)";
    context.fill();

    if (data.label) {
        context.font = `600 ${settings.labelSize}px ${LABEL_FONT}`;

        const width = context.measureText(data.label).width;
        const top = data.y + data.size + 3;

        context.fillStyle = "#ffffff";
        context.fillRect(data.x - width / 2 - 4, top, width + 8, settings.labelSize + 4);
    }

    drawLabel(context, data, settings);
}

// The column titles sit above the map, so they follow it when it is zoomed
// or moved.
function updateHeaders() {
    if (props.map.layout.value !== "columns") {
        headers.value = [];

        return;
    }

    headers.value = props.map.columnHeaders().map((header) => {
        const point = renderer.graphToViewport({ x: header.x, y: 0 });

        return { ...header, left: point.x };
    });
}

function redraw() {
    if (!renderer) {
        return;
    }

    path = props.map.currentPath();
    renderer.refresh();
}

function onEnterNode(event) {
    hoveredKey = event.node;
    container.value.style.cursor = "pointer";

    renderer.refresh({ skipIndexation: true });
}

function onLeaveNode() {
    hoveredKey = "";
    container.value.style.cursor = "default";

    renderer.refresh({ skipIndexation: true });
}

function onClickNode(event) {
    emit("inspect", event.node);
}

// A double click opens the dot instead of zooming in, as Sigma would.
function onDoubleClickNode(event) {
    event.preventSigmaDefault();

    emit("expand", event.node);
}

onMounted(() => {
    renderer = new Sigma(props.map.graph, container.value, {
        renderEdgeLabels: true,
        defaultEdgeType: "curved",
        edgeProgramClasses: {
            curved: EdgeCurvedArrowProgram,
        },
        defaultEdgeColor: EDGE_COLOR,
        labelRenderedSizeThreshold: 0,
        labelFont: LABEL_FONT,
        labelSize: 12,
        labelColor: { color: LABEL_COLOR },
        edgeLabelFont: LABEL_FONT,
        edgeLabelSize: 10,
        edgeLabelColor: { color: "#4b5563" },
        zIndex: true,
        nodeReducer: drawNode,
        edgeReducer: drawEdge,
        defaultDrawNodeLabel: drawLabel,
        defaultDrawNodeHover: drawHover,
    });

    renderer.on("clickNode", onClickNode);
    renderer.on("doubleClickNode", onDoubleClickNode);
    renderer.on("enterNode", onEnterNode);
    renderer.on("leaveNode", onLeaveNode);
    renderer.on("afterRender", updateHeaders);

    // Sigma only follows window resizes, and the map also changes size when
    // the sidebar or the details panel opens.
    resizeObserver = new ResizeObserver(() => {
        renderer.resize();
        renderer.refresh();
    });
    resizeObserver.observe(container.value);
});

onBeforeUnmount(() => {
    if (resizeObserver) {
        resizeObserver.disconnect();
    }

    if (renderer) {
        renderer.kill();
    }
});

watch(
    () => [
        props.map.version.value,
        props.map.selectedKey.value,
        props.map.layout.value,
        props.map.shownTypes.value,
    ],
    () => {
        hoveredKey = "";
        redraw();
    },
);
</script>
