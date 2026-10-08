// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { moveCardToColumn, setCardField } from "../board";

const board = () => [
    {
        id: "todo",
        tasks: [{ id: "a", status: { id: "todo" } }, { id: "b" }],
        order: [{ id: "a" }, { id: "b" }],
    },
    { id: "done", tasks: [{ id: "c" }], order: [{ id: "c" }] },
    { id: null, isUnassigned: true, tasks: [], order: [] },
];

const ids = (column) => column.tasks.map((t) => t.id);

describe("setCardField", () => {
    it("sets the field on the card and its place in the order", () => {
        const columns = setCardField(board(), "a", "priority", "high");

        expect(columns[0].tasks[0].priority).toBe("high");
        expect(columns[0].order[0].priority).toBe("high");
    });

    it("leaves columns without the card as they were", () => {
        const before = board();
        const after = setCardField(before, "a", "priority", "high");

        expect(after[1]).toBe(before[1]);
    });

    it("clears a field to nothing, not to an empty object", () => {
        const columns = setCardField(board(), "a", "status", null);

        expect(columns[0].tasks[0].status).toBeNull();
    });
});

describe("moveCardToColumn", () => {
    it("moves the card to the column of its new value", () => {
        const columns = moveCardToColumn(board(), "a", "status", { id: "done" });

        expect(ids(columns[0])).toEqual(["b"]);
        expect(columns[0].order).toEqual([{ id: "b" }]);
        expect(ids(columns[1])).toEqual(["c", "a"]);
        expect(columns[1].order).toEqual([{ id: "c" }, { id: "a" }]);
        expect(columns[1].tasks[1].status).toEqual({ id: "done" });
    });

    it("moves the card to the column without a value when its value has none", () => {
        const columns = moveCardToColumn(board(), "a", "status", null);

        expect(ids(columns[2])).toEqual(["a"]);
        expect(columns[2].tasks[0].status).toBeNull();
    });

    it("keeps the card once, where it was, when its value names its own column", () => {
        const columns = moveCardToColumn(board(), "a", "status", { id: "todo" });

        expect(ids(columns[0])).toEqual(["a", "b"]);
    });

    it("leaves the board alone when the card is not on it", () => {
        const before = board();

        expect(moveCardToColumn(before, "zz", "status", { id: "done" })).toBe(before);
    });
});
