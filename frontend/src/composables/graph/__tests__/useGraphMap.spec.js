// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach } from "vitest";
import graphService from "@/services/graphService";
import { useGraphMap, nodeKey } from "@/composables/graph/useGraphMap.js";

// The relations the API answers with, by entity.
const answers = {
    "order:o-1264": {
        entity: { type: "order", id: "o-1264", title: "Order #1264" },
        groups: [
            {
                relation_type: "company_places_order",
                direction: "in",
                entity_type: "company",
                count: 1,
                items: [
                    {
                        entity: { type: "company", id: "c-05", title: "Tallinn Tech" },
                        origin: "structured",
                    },
                ],
            },
            {
                relation_type: "order_fulfilled_by_order",
                direction: "out",
                entity_type: "order",
                count: 1,
                items: [
                    {
                        entity: { type: "order", id: "o-1268", title: "Order #1268" },
                        origin: "structured",
                    },
                ],
            },
        ],
    },
    "company:c-05": {
        entity: { type: "company", id: "c-05", title: "Tallinn Tech" },
        groups: [
            {
                relation_type: "company_places_order",
                direction: "out",
                entity_type: "order",
                count: 1000,
                count_capped: true,
                items: [],
            },
        ],
    },
};

vi.mock("@/services/graphService", () => ({
    default: {
        getRelations: vi.fn(async (type, id) => ({ data: answers[`${type}:${id}`] })),
    },
}));

const types = {
    entity_types: [
        { id: "t-company", key: "company", name: "Company", source: "dataset" },
        { id: "t-order", key: "order", name: "Order", source: "dataset" },
    ],
    relation_types: [
        {
            key: "company_places_order",
            name: "places",
            reverse_name: "placed by",
            from_type_id: "t-company",
            to_type_id: "t-order",
        },
        {
            key: "order_fulfilled_by_order",
            name: "is fulfilled by",
            reverse_name: "fulfils",
            from_type_id: "t-order",
            to_type_id: "t-order",
        },
    ],
};

let map;

beforeEach(() => {
    graphService.getRelations.mockClear();

    map = useGraphMap();
    map.setTypes(types);
});

describe("useGraphMap", () => {
    it("starts with only the starting dot, named from the API", async () => {
        await map.start("order", "o-1264");

        expect(map.graph.order).toBe(1);
        expect(map.startKey.value).toBe("order:o-1264");
        expect(map.graph.getNodeAttribute("order:o-1264", "label")).toBe("Order #1264");
    });

    it("opens a dot and draws each line from the relation's from end", async () => {
        await map.start("order", "o-1264");
        await map.expand("order:o-1264");

        expect(map.graph.hasNode("company:c-05")).toBe(true);
        expect(map.graph.hasEdge("company:c-05", "order:o-1264")).toBe(true);
        expect(map.graph.hasEdge("order:o-1264", "order:o-1268")).toBe(true);
    });

    it("words each relation from the selected dot's side", async () => {
        await map.start("order", "o-1264");

        const labels = map.connections("order:o-1264").map((connection) => connection.label);

        expect(labels).toEqual(["Placed by", "Is fulfilled by"]);
    });

    it("shows a big group as one box, and adds a picked entity next to it", async () => {
        await map.start("order", "o-1264");
        await map.expand("order:o-1264");
        await map.expand("company:c-05");

        const boxKey = "company:c-05:company_places_order:out";

        expect(map.graph.getNodeAttribute(boxKey, "countText")).toBe("1000+");

        const key = map.addFromGroup(boxKey, { type: "order", id: "o-1", title: "Order #1" });

        expect(key).toBe(nodeKey("order", "o-1"));
        expect(map.graph.hasEdge("company:c-05", key)).toBe(true);
    });

    it("fans out entities picked from the same box", async () => {
        await map.start("order", "o-1264");
        await map.expand("order:o-1264");
        await map.expand("company:c-05");

        const boxKey = "company:c-05:company_places_order:out";

        const first = map.addFromGroup(boxKey, { type: "order", id: "o-1", title: "Order #1" });
        const second = map.addFromGroup(boxKey, { type: "order", id: "o-2", title: "Order #2" });

        const firstNode = map.graph.getNodeAttributes(first);
        const secondNode = map.graph.getNodeAttributes(second);

        const samePlace =
            firstNode.freeX === secondNode.freeX && firstNode.freeY === secondNode.freeY;

        expect(samePlace).toBe(false);
    });

    it("asks once for a dot clicked again while it loads", async () => {
        await map.start("order", "o-1264");
        await map.expand("order:o-1264");

        graphService.getRelations.mockClear();

        await Promise.all([map.loadRelations("company:c-05"), map.loadRelations("company:c-05")]);

        expect(graphService.getRelations).toHaveBeenCalledTimes(1);
    });

    it("drops an answer that arrives after the map was started again", async () => {
        const first = map.start("order", "o-1264");
        const second = map.start("company", "c-05");

        await Promise.all([first, second]);

        expect(map.errorMessage.value).toBe("");
        expect(map.graph.hasNode("order:o-1264")).toBe(false);
        expect(map.graph.getNodeAttribute("company:c-05", "label")).toBe("Tallinn Tech");
    });

    it("keeps the path from the start to the selected dot", async () => {
        await map.start("order", "o-1264");
        await map.expand("order:o-1264");
        await map.expand("company:c-05");

        map.select("company:c-05:company_places_order:out");

        expect(map.currentPath().keys).toEqual([
            "order:o-1264",
            "company:c-05",
            "company:c-05:company_places_order:out",
        ]);
    });

    it("starts in the free layout when two types relate to each other both ways", () => {
        const looping = useGraphMap();

        looping.setTypes({
            entity_types: types.entity_types,
            relation_types: [
                ...types.relation_types,
                {
                    key: "order_billed_to_company",
                    name: "billed to",
                    reverse_name: "billed",
                    from_type_id: "t-order",
                    to_type_id: "t-company",
                },
            ],
        });

        expect(looping.layout.value).toBe("free");
        expect(map.layout.value).toBe("columns");
    });
});
