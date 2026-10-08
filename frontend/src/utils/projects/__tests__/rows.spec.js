// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import {
    clearReferences,
    countSubtasks,
    linkedDialogTitle,
    newFieldColumn,
    normalizeTaskRow,
    optionLookup,
    refreshEmbedded,
    withOptions,
    withoutTask,
} from "../rows";

describe("linkedDialogTitle", () => {
    const tables = [
        {
            id: "t2",
            name: "t0123456789abcdef0123456789abcdef",
            display_name: { String: "Table 2", Valid: true },
        },
        { id: "junction", parent_table_id: "t2" },
        { id: "old", name: "Clients" },
    ];

    it("names a link after the table it shows, through its junction for a one-way link", () => {
        expect(linkedDialogTitle({ name: "c1", parent_table_id: "t2" }, tables)).toBe("Table 2");
        expect(linkedDialogTitle({ name: "c1", linked_id: "junction" }, tables)).toBe("Table 2");
        expect(linkedDialogTitle({ name: "c1", parent_table_id: "old" }, tables)).toBe("Clients");
    });

    it("names a single select, or a table it cannot find, after the field", () => {
        expect(
            linkedDialogTitle({ name: "c1", display_name: "Status", single_select: true }, tables),
        ).toBe("Status");
        expect(
            linkedDialogTitle(
                { name: "c1", display_name: "Owner", parent_table_id: "gone" },
                tables,
            ),
        ).toBe("Owner");
    });
});

describe("newFieldColumn", () => {
    it("types a new field by its column, as a reload does", () => {
        expect(newFieldColumn("single select").headerType).toBe("VARCHAR");
        expect(newFieldColumn("assignee").headerType).toBe("VARCHAR");
        expect(newFieldColumn("number").headerType).toBe("DECIMAL");
        expect(newFieldColumn("bool").headerType).toBe("TINYINT");
        expect(newFieldColumn("link").headerType).toBe("VARCHAR");
        expect(newFieldColumn("text").headerType).toBe("TEXT");
    });

    it("gives every row its own empty list", () => {
        const { empty } = newFieldColumn("link");

        expect(empty()).toEqual([]);
        expect(empty()).not.toBe(empty());
    });
});

describe("normalizeTaskRow", () => {
    const headers = [
        { name: "name", header_type: "VARCHAR" },
        { name: "done", header_type: "TINYINT" },
        { name: "status", header_type: "VARCHAR" },
    ];

    it("shapes a row the way the grid holds it", () => {
        const row = normalizeTaskRow(
            { id: " t1 ", name: "Task", done: "1", parent_task_id: "p1" },
            headers,
        );

        expect(row).toEqual({
            id: "t1",
            name: "Task",
            done: true,
            status: "",
            parent_task_id: "p1",
            _original_parent_task_id: "p1",
            _is_subtask: true,
        });
    });

    it("treats a task that is its own parent, or has none, as a root", () => {
        expect(normalizeTaskRow({ id: "t1", parent_task_id: "t1" }, headers)._is_subtask).toBe(
            false,
        );
        expect(
            normalizeTaskRow({ id: "t1", parent_task_id: "" }, headers).parent_task_id,
        ).toBeNull();
    });
});

describe("clearReferences", () => {
    const rows = [
        {
            id: "r1",
            links: [{ id: "gone" }, { id: "kept" }],
            status: { id: "gone" },
            label: "gone",
        },
        { id: "r2", links: [{ id: "kept" }], status: { id: "kept" }, label: "kept" },
    ];

    it("drops a deleted task from link lists only", () => {
        const next = clearReferences(rows, "gone");

        expect(next[0].links).toEqual([{ id: "kept" }]);
        expect(next[0].status).toEqual({ id: "gone" });
        expect(next[1]).toBe(rows[1]);
    });

    it("also clears a deleted option held as a whole cell", () => {
        const next = clearReferences(rows, "gone", { isOption: true });

        expect(next[0]).toMatchObject({ links: [{ id: "kept" }], status: null, label: null });
        expect(next[1]).toBe(rows[1]);
    });

    it("returns the same array when nothing points at the id", () => {
        expect(clearReferences(rows, "unknown")).toBe(rows);
    });
});

describe("refreshEmbedded", () => {
    const headers = [
        { name: "status", linked_id: "statuses" },
        { name: "links", linked_id: "projects" },
    ];
    const rows = [
        {
            id: "r1",
            status: { id: "s1", name: "Open", color: "red" },
            links: [{ id: "p1", name: "Old" }],
        },
        { id: "r2", status: { id: "s2", name: "Done" }, links: [] },
    ];

    it("renames and recolours an option inside the rows that hold it", () => {
        const next = refreshEmbedded(rows, headers, "statuses", {
            id: "s1",
            name: "Started",
            color: "blue",
        });

        expect(next[0].status).toEqual({ id: "s1", name: "Started", color: "blue" });
        expect(next[1]).toBe(rows[1]);
    });

    it("renames a linked task inside link lists", () => {
        const next = refreshEmbedded(rows, headers, "projects", { id: "p1", name: "New" });

        expect(next[0].links).toEqual([{ id: "p1", name: "New" }]);
    });

    it("renames a linked task in a link column, which names its junction table", () => {
        const linkHeaders = [
            { name: "two_way", linked_id: "junction-1", parent_table_id: "clients" },
            { name: "one_way", linked_id: "junction-2", parent_table_id: "" },
        ];
        const linkRows = [
            {
                id: "r1",
                two_way: [{ id: "c1", name: "Acme" }],
                one_way: [{ id: "c1", name: "Acme" }],
            },
        ];
        const tables = [{ id: "junction-2", parent_table_id: "clients" }];

        const next = refreshEmbedded(
            linkRows,
            linkHeaders,
            "clients",
            { id: "c1", name: "Acme Ltd" },
            tables,
        );

        expect(next[0].two_way).toEqual([{ id: "c1", name: "Acme Ltd" }]);
        expect(next[0].one_way).toEqual([{ id: "c1", name: "Acme Ltd" }]);
    });

    it("leaves rows alone when no column points at the changed table", () => {
        expect(refreshEmbedded(rows, headers, "elsewhere", { id: "s1", name: "x" })).toBe(rows);
    });
});

describe("countSubtasks", () => {
    it("counts children per parent and skips rows without one", () => {
        const counts = countSubtasks([
            { id: "1", parent_task_id: null },
            { id: "2", parent_task_id: "1" },
            { id: "3", parent_task_id: 1 },
            { id: "4", parent_task_id: "2" },
            { id: "5" },
        ]);

        expect(counts.get("1")).toBe(2);
        expect(counts.get("2")).toBe(1);
        expect(counts.has("3")).toBe(false);
        expect(counts.has("null")).toBe(false);
    });

    it("handles missing rows", () => {
        expect(countSubtasks(undefined).size).toBe(0);
    });
});

describe("optionLookup and withOptions", () => {
    const headers = [
        { name: "status", single_select: true, linked_id: "statuses" },
        { name: "priority", single_select: true, linked_id: "missing" },
        { name: "notes" },
    ];
    const tables = [
        { id: "statuses", data_base: [{ id: 1, name: "Done", color: "#0f0" }] },
        { id: "statuses", data_base: [{ id: 1, name: "Later copy" }] },
    ];

    it("replaces an option id with the option from its field's table", () => {
        const lookup = optionLookup(headers, tables);
        const task = { id: "a", status: "1", priority: "7", notes: "x" };

        expect(withOptions(task, lookup)).toEqual({
            id: "a",
            status: { id: 1, name: "Done", color: "#0f0" },
            priority: "7",
            notes: "x",
        });
        expect(task.status).toBe("1");
    });

    it("leaves an empty or unknown option as it is", () => {
        const lookup = optionLookup(headers, tables);

        expect(withOptions({ status: "" }, lookup)).toEqual({ status: "" });
        expect(withOptions({ status: "9" }, lookup)).toEqual({ status: "9" });
        expect(optionLookup(undefined, undefined).size).toBe(0);
    });
});

describe("withoutTask", () => {
    it("removes the task and its subtasks at every depth", () => {
        const rows = [
            { id: "a" },
            { id: "b", parent_task_id: "a" },
            { id: "c", parent_task_id: "b" },
            { id: "d", parent_task_id: null },
            { id: "e", parent_task_id: "d" },
        ];

        expect(withoutTask(rows, "a").map((r) => r.id)).toEqual(["d", "e"]);
        expect(withoutTask(rows, "b").map((r) => r.id)).toEqual(["a", "d", "e"]);
    });

    it("keeps a task that is its own parent from taking the others", () => {
        const rows = [{ id: "a", parent_task_id: "a" }, { id: "b" }];

        expect(withoutTask(rows, "a").map((r) => r.id)).toEqual(["b"]);
        expect(withoutTask(undefined, "a")).toEqual([]);
    });
});
