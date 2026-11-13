<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div ref="container" class="absolute inset-0"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from "vue";
import Graph from "graphology";
import Sigma from "sigma";
import { columnOrder } from "@/utils/graph/typeOrder.js";
import { typeStyles } from "@/utils/graph/palette.js";
import { baseSigmaSettings } from "@/utils/graph/drawing.js";

const props = defineProps({
    types: {
        type: Object,
        required: true,
    },
});

const COLUMN_GAP = 10;
const ROW_GAP = 4;
const CIRCLE_RADIUS = 8;
const PARALLEL_SPREAD = 0.5;

const container = ref(null);

let renderer = null;
let resizeObserver = null;

// Each type goes one column to the right of every type that relates to it, so
// the arrows read left to right; types in the same column are stacked. When
// the types form a loop there is no such order, and they go in a circle.
function placeTypes(entityTypes, relationTypes) {
    const ordered = columnOrder(entityTypes, relationTypes);
    const positions = new Map();

    if (ordered.hasLoop) {
        ordered.order.forEach((typeID, index) => {
            const angle = (2 * Math.PI * index) / ordered.order.length;

            positions.set(typeID, {
                x: CIRCLE_RADIUS * Math.cos(angle),
                y: CIRCLE_RADIUS * Math.sin(angle),
            });
        });

        return positions;
    }

    const columns = new Map();
    const rowsUsed = new Map();

    for (const typeID of ordered.order) {
        let column = 0;

        for (const relationType of relationTypes) {
            const relationsIn =
                relationType.to_type_id === typeID && relationType.from_type_id !== typeID;

            if (relationsIn && columns.has(relationType.from_type_id)) {
                column = Math.max(column, columns.get(relationType.from_type_id) + 1);
            }
        }

        const row = rowsUsed.get(column) ?? 0;

        columns.set(typeID, column);
        rowsUsed.set(column, row + 1);
        positions.set(typeID, { x: column * COLUMN_GAP, y: -row * ROW_GAP });
    }

    return positions;
}

// Two relation types between the same two types get curves bent apart, so
// both stay visible. A relation type from a type to itself can't be drawn as
// a line and is only listed in the table.
function buildGraph() {
    const entityTypes = props.types.entity_types;
    const relationTypes = props.types.relation_types;

    const styles = typeStyles(entityTypes);
    const positions = placeTypes(entityTypes, relationTypes);
    const graph = new Graph({ multi: true, type: "directed" });

    for (const entityType of entityTypes) {
        const position = positions.get(entityType.id);

        graph.addNode(entityType.id, {
            ...position,
            size: 12,
            label: entityType.name,
            color: styles[entityType.key].color,
        });
    }

    const pairs = new Map();

    for (const relationType of relationTypes) {
        const from = relationType.from_type_id;
        const to = relationType.to_type_id;

        if (from === to || !graph.hasNode(from) || !graph.hasNode(to)) {
            continue;
        }

        const pair = `${from}|${to}`;
        const list = pairs.get(pair) ?? [];

        list.push(relationType);
        pairs.set(pair, list);
    }

    for (const list of pairs.values()) {
        list.forEach((relationType, index) => {
            const curvature = (index - (list.length - 1) / 2) * PARALLEL_SPREAD;

            graph.addDirectedEdge(relationType.from_type_id, relationType.to_type_id, {
                label: relationType.name.toUpperCase(),
                size: 2,
                color: "#9ca3af",
                curvature: curvature,
            });
        });
    }

    return graph;
}

function draw() {
    if (renderer) {
        renderer.kill();
    }

    renderer = new Sigma(buildGraph(), container.value, {
        ...baseSigmaSettings(),
        labelSize: 13,
    });
}

onMounted(() => {
    draw();

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

watch(() => props.types, draw);
</script>
