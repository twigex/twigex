// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { columnOrder } from "@/utils/graph/typeOrder.js";
import { createColumnPlacer, freePosition, COLUMN_GAP, ROW_GAP } from "@/utils/graph/layout.js";
import { pathTo } from "@/utils/graph/path.js";
import { typeStyles } from "@/utils/graph/palette.js";

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
    it("keeps the built-in colours and gives new types the next one each", () => {
        const styles = typeStyles([{ key: "chart" }, company, order]);

        expect(styles.chart.color).toBe("#4f46e5");
        expect(styles.company.color).not.toBe(styles.order.color);
    });
});
