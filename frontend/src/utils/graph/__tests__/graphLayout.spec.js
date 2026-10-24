// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { columnOrder } from "@/utils/graph/typeOrder.js";
import { createColumnPlacer, freePosition, COLUMN_GAP, ROW_GAP } from "@/utils/graph/layout.js";
import { pathTo } from "@/utils/graph/path.js";
import { typeStyles } from "@/utils/graph/palette.js";
import { relationsOf } from "@/utils/graph/relations.js";
import { suggestKey, fallbackKey } from "@/utils/graph/keys.js";

const company = { id: "t-company", key: "company" };
const order = { id: "t-order", key: "order" };
const invoice = { id: "t-invoice", key: "invoice" };

function relation(from, to) {
    return { from_type_id: from.id, to_type_id: to.id };
}

describe("columnOrder", () => {
    it("puts every relation's from type left of its to type", () => {
        const result = columnOrder(
            [invoice, order, company],
            [relation(company, order), relation(order, invoice)],
        );

        expect(result.order).toEqual(["t-company", "t-order", "t-invoice"]);
        expect(result.hasLoop).toBe(false);
    });

    it("keeps a relation from a type to itself inside its column", () => {
        const result = columnOrder(
            [company, order],
            [relation(company, order), relation(order, order)],
        );

        expect(result.order).toEqual(["t-company", "t-order"]);
        expect(result.hasLoop).toBe(false);
    });

    it("reports a loop between two types and still gives both a column", () => {
        const result = columnOrder(
            [company, order],
            [relation(company, order), relation(order, company)],
        );

        expect(result.hasLoop).toBe(true);
        expect(result.order).toHaveLength(2);
    });
});

describe("createColumnPlacer", () => {
    it("puts a new dot level with its parent, and never on a taken row", () => {
        const placer = createColumnPlacer();

        const first = placer.place(1, 4);
        const second = placer.place(1, 4);
        const above = placer.place(1, 0);

        expect(first).toEqual({ x: COLUMN_GAP, y: -4 * ROW_GAP, row: 4 });
        expect(second.row).toBe(5);
        expect(above.row).toBe(6);
    });
});

describe("freePosition", () => {
    it("fans new dots out away from where the parent came from", () => {
        const parent = { freeX: 3, freeY: 0, originX: 0, originY: 0 };

        const position = freePosition(parent, 0, 1);

        expect(position.x).toBeGreaterThan(parent.freeX);
        expect(position.originX).toBe(3);
    });

    it("puts the starting dot in the middle", () => {
        expect(freePosition(null, 0, 1)).toEqual({ x: 0, y: 0 });
    });
});

describe("pathTo", () => {
    it("follows the parents back to the start, in order", () => {
        const parents = { c: "b", b: "a", a: "" };

        expect(pathTo("c", (key) => parents[key])).toEqual(["a", "b", "c"]);
    });

    it("stops at a loop instead of running forever", () => {
        const parents = { a: "b", b: "a" };

        expect(pathTo("a", (key) => parents[key])).toEqual(["b", "a"]);
    });
});

describe("typeStyles", () => {
    it("uses the chosen colour, and gives a type without one a colour nobody uses", () => {
        const styles = typeStyles([
            { key: "chart", color: "indigo" },
            { key: "company", color: "cyan" },
            { key: "order" },
        ]);

        expect(styles.chart.color).toBe("#4f46e5");
        expect(styles.company.color).toBe("#0891b2");
        expect(styles.order.color).not.toBe(styles.company.color);
        expect(styles.order.color).not.toBe(styles.chart.color);
    });
});

describe("relationsOf", () => {
    it("lists a relation type from a type to itself from both ends", () => {
        const fulfilled = {
            key: "order_fulfilled_by_order",
            name: "is fulfilled by",
            reverse_name: "fulfils",
            from_type_id: order.id,
            to_type_id: order.id,
            cardinality: "one_to_many",
        };

        const relations = relationsOf(order, [fulfilled], new Map([[order.id, order]]));

        expect(relations.map((relation) => relation.name)).toEqual(["is fulfilled by", "fulfils"]);
        expect(relations.map((relation) => relation.cardinality)).toEqual([
            "one_to_many",
            "many_to_one",
        ]);
    });
});

describe("suggestKey", () => {
    it("turns names into keys the server accepts", () => {
        expect(suggestKey("Purchase order")).toBe("purchase_order");
        expect(suggestKey("Ēka 2")).toBe("eka_2");
        expect(suggestKey("2nd hand")).toBe("nd_hand");
        expect(suggestKey("company", "places", "order")).toBe("company_places_order");
        expect(suggestKey("a".repeat(80))).toHaveLength(64);
        expect(suggestKey("Spółka")).toBe("spolka");
        expect(suggestKey("Компания")).toBe("");
    });
});

describe("fallbackKey", () => {
    it("makes a key the server accepts", () => {
        expect(fallbackKey("type")).toMatch(/^type_[a-z0-9]+$/);
    });
});
