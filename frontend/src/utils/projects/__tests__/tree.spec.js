// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import {
    canMoveInto,
    findTreeNode,
    findTreeParent,
    flattenTaskTree,
    moveTreeNode,
    removeTreeNode,
    workspaceTables,
} from "../tree";

function tree() {
    return [
        {
            id: "ws",
            children: [
                {
                    id: "a",
                    isFolder: true,
                    children: [
                        { id: "b", isFolder: true, children: [{ id: "in-b", type: "project" }] },
                        { id: "in-a", type: "project" },
                    ],
                },
                { id: "d", isFolder: true, children: [] },
                { id: "at-root", type: "project" },
            ],
        },
    ];
}

const ids = (node) => node.children.map((c) => c.id);

describe("canMoveInto", () => {
    it("allows a table or a folder into another folder or the root", () => {
        const roots = tree();
        const node = (id) => findTreeNode(roots, id);

        expect(canMoveInto(roots, node("at-root"), node("b"))).toBe(true);
        expect(canMoveInto(roots, node("in-b"), node("ws"))).toBe(true);
        expect(canMoveInto(roots, node("b"), node("d"))).toBe(true);
        expect(canMoveInto(roots, node("b"), node("ws"))).toBe(true);
    });

    it("refuses a table as the target, the current folder, and a folder's own subtree", () => {
        const roots = tree();
        const node = (id) => findTreeNode(roots, id);

        expect(canMoveInto(roots, node("at-root"), node("in-a"))).toBe(false);
        expect(canMoveInto(roots, node("d"), node("at-root"))).toBe(false);
        expect(canMoveInto(roots, node("in-a"), node("a"))).toBe(false);
        expect(canMoveInto(roots, node("a"), node("a"))).toBe(false);
        expect(canMoveInto(roots, node("a"), node("b"))).toBe(false);
        expect(canMoveInto(roots, node("d"), { id: "elsewhere", isFolder: true })).toBe(false);
    });
});

describe("moveTreeNode", () => {
    it("moves a table and keeps folders ahead of tables", () => {
        const roots = tree();

        expect(moveTreeNode(roots, "in-b", "ws")).toBe(true);
        expect(ids(roots[0])).toEqual(["a", "d", "at-root", "in-b"]);
        expect(findTreeNode(roots, "b").children).toEqual([]);

        expect(moveTreeNode(roots, "d", "a")).toBe(true);
        expect(ids(findTreeNode(roots, "a"))).toEqual(["b", "d", "in-a"]);
        expect(findTreeParent(roots, "d").id).toBe("a");
    });

    it("leaves the tree alone for a move it cannot make", () => {
        const roots = tree();
        const before = JSON.stringify(roots);

        expect(moveTreeNode(roots, "a", "b")).toBe(false);
        expect(moveTreeNode(roots, "at-root", "in-a")).toBe(false);
        expect(moveTreeNode(roots, "in-a", "a")).toBe(false);
        expect(moveTreeNode(roots, "missing", "a")).toBe(false);
        expect(JSON.stringify(roots)).toBe(before);
    });
});

describe("removeTreeNode", () => {
    it("removes a node from wherever it sits", () => {
        const roots = tree();

        expect(removeTreeNode(roots, "in-b")).toBe(true);
        expect(findTreeNode(roots, "b").children).toEqual([]);

        expect(removeTreeNode(roots, "a")).toBe(true);
        expect(ids(roots[0])).toEqual(["d", "at-root"]);
    });

    it("removes only a node that matches", () => {
        const roots = tree();
        const before = JSON.stringify(roots);

        expect(removeTreeNode(roots, "d", (node) => !node.isFolder)).toBe(false);
        expect(removeTreeNode(roots, "missing")).toBe(false);
        expect(removeTreeNode(undefined, "a")).toBe(false);
        expect(JSON.stringify(roots)).toBe(before);
    });
});

describe("flattenTaskTree", () => {
    const rows = () => [
        { id: 1, parent_task_id: null },
        { id: 2, parent_task_id: 1 },
        { id: 3, parent_task_id: 2 },
        { id: 4, parent_task_id: null },
    ];
    const ids = (list) => list.map((r) => r.id);

    it("lists only the top level when nothing is expanded", () => {
        expect(ids(flattenTaskTree(rows(), new Set()))).toEqual([1, 4]);
    });

    it("puts an expanded task's subtasks under it, at their depth", () => {
        const list = flattenTaskTree(rows(), new Set([1, 2]));

        expect(ids(list)).toEqual([1, 2, 3, 4]);
        expect(list.map((r) => r._indent)).toEqual([0, 1, 2, 0]);
    });

    it("matches expanded ids whether they are strings or numbers", () => {
        expect(ids(flattenTaskTree(rows(), new Set(["1"])))).toEqual([1, 2, 4]);
    });

    it("lists a task whose parent is not among the rows at the top", () => {
        expect(ids(flattenTaskTree([{ id: 5, parent_task_id: 99 }], new Set()))).toEqual([5]);
    });

    it("lists a task that is its own parent once, even when expanded", () => {
        expect(ids(flattenTaskTree([{ id: 6, parent_task_id: 6 }], new Set([6])))).toEqual([6]);
    });

    it("does not add the depth to the row's own keys", () => {
        const [first] = flattenTaskTree(rows(), new Set());

        expect(Object.keys(first)).toEqual(["id", "parent_task_id"]);
    });
});

describe("workspaceTables", () => {
    it("lists each table once, from the top and from folders within folders", () => {
        const workspace = {
            tables: [{ id: "t1" }],
            folders: [
                { tables: [{ id: "t1" }, { id: "t2" }], children: [{ tables: [{ id: "t3" }] }] },
                {},
            ],
        };

        expect(workspaceTables(workspace).map((t) => t.id)).toEqual(["t1", "t2", "t3"]);
        expect(workspaceTables(null)).toEqual([]);
    });
});
