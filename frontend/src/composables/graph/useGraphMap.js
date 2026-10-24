// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import Graph from "graphology";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";
import { columnOrder } from "@/utils/graph/typeOrder.js";
import { createColumnPlacer, freePosition, COLUMN_GAP } from "@/utils/graph/layout.js";
import { pathTo } from "@/utils/graph/path.js";
import { typeStyles, GROUP_STYLE } from "@/utils/graph/palette.js";

const CURVATURE = 0.15;
const SKIP_CURVATURE = 0.35;

// Entities picked from one box fan out this many to a ring, a ring further
// out each time the last one fills.
const BOX_RING_SIZE = 6;

export function nodeKey(type, id) {
    return `${type}:${id}`;
}

function groupKey(parentKey, group) {
    return `${parentKey}:${group.relation_type}:${group.direction}`;
}

function capitalize(text) {
    if (!text) {
        return "";
    }

    return text.charAt(0).toUpperCase() + text.slice(1);
}

// useGraphMap holds the map's dots and lines and everything that changes
// them. The drawing (GraphCanvas) only reads from it.
export function useGraphMap() {
    const graph = new Graph();
    const placer = createColumnPlacer();
    const loadedRelations = new Map();
    const loadingRelations = new Map();

    const layout = ref("columns");
    const startKey = ref("");
    const selectedKey = ref("");
    const errorMessage = ref("");

    // The type chips: the keys of the types to show. Empty shows them all.
    const shownTypes = ref([]);

    // Bumped whenever dots or lines change, so the drawing and anything
    // computed from the map know to look again.
    const version = ref(0);

    // Bumped by start(), so an answer that arrives after the map was started
    // again is dropped instead of landing on the new map.
    let generation = 0;

    let typesByKey = new Map();
    let relationTypesByKey = new Map();
    let columns = new Map();
    let styles = {};

    function setTypes(types) {
        typesByKey = new Map(types.entity_types.map((entityType) => [entityType.key, entityType]));
        relationTypesByKey = new Map(
            types.relation_types.map((relationType) => [relationType.key, relationType]),
        );
        styles = typeStyles(types.entity_types);

        const ordered = columnOrder(types.entity_types, types.relation_types);

        const keysByID = new Map(
            types.entity_types.map((entityType) => [entityType.id, entityType.key]),
        );

        columns = new Map();
        ordered.order.forEach((typeID, index) => {
            columns.set(keysByID.get(typeID), index);
        });

        if (ordered.hasLoop) {
            layout.value = "free";
        }
    }

    function typeName(typeKey) {
        const entityType = typesByKey.get(typeKey);

        if (!entityType) {
            return typeKey;
        }

        return entityType.name;
    }

    function styleFor(typeKey) {
        return styles[typeKey] ?? GROUP_STYLE;
    }

    function columnOf(typeKey) {
        return columns.get(typeKey) ?? columns.size;
    }

    function relationName(group) {
        const relationType = relationTypesByKey.get(group.relation_type);

        if (!relationType) {
            return group.relation_type;
        }

        if (group.direction === "out") {
            return capitalize(relationType.name);
        }

        return capitalize(relationType.reverse_name);
    }

    function place(attributes, typeKey, parentKey, free) {
        let parentRow;

        if (parentKey) {
            parentRow = graph.getNodeAttribute(parentKey, "row");
        }

        const column = placer.place(columnOf(typeKey), parentRow);

        const placed = {
            ...attributes,
            row: column.row,
            columnX: column.x,
            columnY: column.y,
            freeX: free.x,
            freeY: free.y,
            originX: free.originX,
            originY: free.originY,
            x: column.x,
            y: column.y,
            parentKey: parentKey,
        };

        if (layout.value === "free") {
            placed.x = free.x;
            placed.y = free.y;
        }

        return placed;
    }

    function addEntity(entity, parentKey, free) {
        const key = nodeKey(entity.type, entity.id);

        if (graph.hasNode(key)) {
            return key;
        }

        const attributes = {
            size: 8,
            label: entity.title,
            color: styleFor(entity.type).color,
            columnType: entity.type,
            entityType: entity.type,
            entityId: entity.id,
            expanded: false,
        };

        graph.addNode(key, place(attributes, entity.type, parentKey, free));

        return key;
    }

    function addGroup(parentKey, group, free) {
        const key = groupKey(parentKey, group);

        if (graph.hasNode(key)) {
            return key;
        }

        let countText = `${group.count}`;

        if (group.count_capped) {
            countText = `${group.count}+`;
        }

        const attributes = {
            size: 18,
            label: `${typeName(group.entity_type)} (${countText})`,
            countText: countText,
            color: GROUP_STYLE.color,
            columnType: group.entity_type,
            isGroup: true,
            group: { ...group, parentKey: parentKey },
        };

        graph.addNode(key, place(attributes, group.entity_type, parentKey, free));

        return key;
    }

    // A line always runs from the relation's "from" entity to its "to" entity;
    // in columns that is left to right. A line that skips a column bends
    // more, so it goes around the dots in between.
    function relationEdge(parentKey, otherKey, group) {
        let source = parentKey;
        let target = otherKey;

        if (group.direction === "in") {
            source = otherKey;
            target = parentKey;
        }

        const relationType = relationTypesByKey.get(group.relation_type);

        let label = group.relation_type;

        if (relationType) {
            label = relationType.name;
        }

        const sourceColumn = columnOf(graph.getNodeAttribute(source, "columnType"));
        const targetColumn = columnOf(graph.getNodeAttribute(target, "columnType"));

        let curvature = CURVATURE;

        if (Math.abs(targetColumn - sourceColumn) > 1) {
            curvature = SKIP_CURVATURE;
        }

        graph.mergeEdge(source, target, { label: label, size: 1.5, curvature: curvature });
    }

    // A dot clicked again while its relations load waits for the same request.
    async function loadRelations(key) {
        if (loadedRelations.has(key)) {
            return loadedRelations.get(key);
        }

        if (!loadingRelations.has(key)) {
            loadingRelations.set(key, fetchRelations(key));
        }

        return loadingRelations.get(key);
    }

    async function fetchRelations(key) {
        const node = graph.getNodeAttributes(key);
        const startedIn = generation;

        errorMessage.value = "";

        try {
            const response = await graphService.getRelations(node.entityType, node.entityId);

            if (startedIn !== generation) {
                return null;
            }

            const relations = response.data;

            loadingRelations.delete(key);
            loadedRelations.set(key, relations);
            graph.setNodeAttribute(key, "label", relations.entity.title);
            version.value = version.value + 1;

            return relations;
        } catch (e) {
            if (startedIn !== generation) {
                return null;
            }

            loadingRelations.delete(key);
            errorMessage.value = extractErrorMessage(e);

            return null;
        }
    }

    function listNeighbours(key, groups) {
        const neighbours = [];

        for (const group of groups) {
            if (group.items.length === 0) {
                neighbours.push({
                    group: group,
                    item: null,
                    key: groupKey(key, group),
                    type: group.entity_type,
                    title: "",
                });

                continue;
            }

            for (const item of group.items) {
                const entity = item.entity;

                neighbours.push({
                    group: group,
                    item: entity,
                    key: nodeKey(entity.type, entity.id),
                    type: entity.type,
                    title: entity.title,
                });
            }
        }

        neighbours.sort((first, second) => {
            const byColumn = columnOf(first.type) - columnOf(second.type);

            if (byColumn !== 0) {
                return byColumn;
            }

            return first.title.localeCompare(second.title);
        });

        return neighbours;
    }

    async function expand(key) {
        const node = graph.getNodeAttributes(key);

        if (node.isGroup || node.expanded) {
            return;
        }

        const startedIn = generation;
        const relations = await loadRelations(key);

        if (!relations || startedIn !== generation) {
            return;
        }

        const neighbours = listNeighbours(key, relations.groups);
        const newNeighbours = neighbours.filter((neighbour) => !graph.hasNode(neighbour.key));

        newNeighbours.forEach((neighbour, index) => {
            const free = freePosition(graph.getNodeAttributes(key), index, newNeighbours.length);

            if (neighbour.item) {
                addEntity(neighbour.item, key, free);
            } else {
                addGroup(key, neighbour.group, free);
            }
        });

        for (const neighbour of neighbours) {
            relationEdge(key, neighbour.key, neighbour.group);
        }

        graph.setNodeAttribute(key, "expanded", true);
        version.value = version.value + 1;
    }

    // An entity picked from a box's list goes on the map next to the box,
    // related to the entity the box belongs to.
    function addFromGroup(boxKey, item) {
        const box = graph.getNodeAttributes(boxKey);
        const parentKey = box.group.parentKey;

        const picked = box.picked ?? 0;
        const ring = 1 + Math.floor(picked / BOX_RING_SIZE);

        const free = freePosition(box, picked % BOX_RING_SIZE, BOX_RING_SIZE);

        free.x = box.freeX + (free.x - box.freeX) * ring;
        free.y = box.freeY + (free.y - box.freeY) * ring;

        graph.setNodeAttribute(boxKey, "picked", picked + 1);

        const key = addEntity(item, parentKey, free);

        relationEdge(parentKey, key, box.group);
        version.value = version.value + 1;

        return key;
    }

    async function start(type, id) {
        generation = generation + 1;

        graph.clear();
        placer.clear();
        loadedRelations.clear();
        loadingRelations.clear();

        startKey.value = "";
        selectedKey.value = "";
        errorMessage.value = "";

        if (!type || !id) {
            version.value = version.value + 1;

            return;
        }

        const key = addEntity({ type: type, id: id, title: "…" }, "", freePosition(null, 0, 1));

        startKey.value = key;
        selectedKey.value = key;
        version.value = version.value + 1;

        await loadRelations(key);
    }

    function select(key) {
        selectedKey.value = key;
    }

    function isTypeShown(typeKey) {
        if (shownTypes.value.length === 0) {
            return true;
        }

        return shownTypes.value.includes(typeKey);
    }

    function toggleType(typeKey) {
        if (shownTypes.value.includes(typeKey)) {
            shownTypes.value = shownTypes.value.filter((key) => key !== typeKey);

            return;
        }

        shownTypes.value = [...shownTypes.value, typeKey];
    }

    function showAllTypes() {
        shownTypes.value = [];
    }

    // What the card header counts: entities (not group boxes) and lines.
    function counts() {
        if (version.value < 0) {
            return { entities: 0, relations: 0 };
        }

        let entities = 0;

        graph.forEachNode((key, attributes) => {
            if (!attributes.isGroup) {
                entities = entities + 1;
            }
        });

        return { entities: entities, relations: graph.size };
    }

    function setLayout(value) {
        layout.value = value;

        graph.updateEachNodeAttributes((key, attributes) => {
            if (value === "free") {
                return { ...attributes, x: attributes.freeX, y: attributes.freeY };
            }

            return { ...attributes, x: attributes.columnX, y: attributes.columnY };
        });

        version.value = version.value + 1;
    }

    function parentOf(key) {
        if (!graph.hasNode(key)) {
            return "";
        }

        return graph.getNodeAttribute(key, "parentKey");
    }

    // The path from the starting dot to the selected one: its dots, and the
    // lines between each dot and the one it was opened from.
    function currentPath() {
        const keys = [];
        const edges = new Set();

        if (version.value < 0 || !selectedKey.value || !graph.hasNode(selectedKey.value)) {
            return { keys: keys, edges: edges };
        }

        const path = pathTo(selectedKey.value, parentOf);

        for (let index = 0; index < path.length; index++) {
            keys.push(path[index]);

            if (index === 0) {
                continue;
            }

            const from = path[index - 1];
            const to = path[index];

            if (graph.hasEdge(from, to)) {
                edges.add(graph.edge(from, to));
            }

            if (graph.hasEdge(to, from)) {
                edges.add(graph.edge(to, from));
            }
        }

        return { keys: keys, edges: edges };
    }

    // Reading version makes anything computed from these functions update
    // when the map changes.
    function nodeAttributes(key) {
        if (version.value < 0 || !graph.hasNode(key)) {
            return null;
        }

        return graph.getNodeAttributes(key);
    }

    // The selected dot's relations, in words from its own side ("Placed by",
    // "Supplies"), once they have been loaded.
    function connections(key) {
        if (version.value < 0) {
            return [];
        }

        const relations = loadedRelations.get(key);

        if (!relations) {
            return [];
        }

        return relations.groups.map((group) => {
            const items = group.items.map((item) => {
                const entity = item.entity;

                return {
                    key: nodeKey(entity.type, entity.id),
                    title: entity.title,
                    type: entity.type,
                };
            });

            return {
                id: `${group.relation_type}:${group.direction}`,
                label: relationName(group),
                entityType: group.entity_type,
                count: group.count,
                capped: Boolean(group.count_capped),
                items: items,
                boxKey: groupKey(key, group),
            };
        });
    }

    function columnHeaders() {
        const headers = [];
        const used = new Set();

        graph.forEachNode((key, attributes) => {
            used.add(attributes.columnType);
        });

        for (const [typeKey, column] of columns) {
            if (!used.has(typeKey)) {
                continue;
            }

            headers.push({
                type: typeKey,
                title: typeName(typeKey),
                color: styleFor(typeKey).color,
                x: column * COLUMN_GAP,
            });
        }

        return headers;
    }

    return {
        graph,
        layout,
        shownTypes,
        startKey,
        selectedKey,
        errorMessage,
        version,
        setTypes,
        typeName,
        styleFor,
        start,
        expand,
        addFromGroup,
        select,
        isTypeShown,
        toggleType,
        showAllTypes,
        counts,
        setLayout,
        currentPath,
        nodeAttributes,
        connections,
        columnHeaders,
        loadRelations,
    };
}
