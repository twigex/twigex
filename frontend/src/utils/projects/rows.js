// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const norm = (v) => (v == null ? "" : String(v).trim());

export const isBlank = (value) => norm(value) === "";

// What a field is called in a view: the view's own name for it, or the
// field's, or its column only for a field saved without a name. Views made
// before fields had generated columns saved the column as their name for it,
// so that does not count as one.
export function fieldLabel(header, viewLabel) {
    const own =
        typeof viewLabel === "string" && viewLabel && viewLabel !== header?.name ? viewLabel : "";
    const field =
        typeof header?.display_name === "string"
            ? header.display_name
            : header?.display_name?.String;

    return own || field || header?.name || "";
}

const GENERATED_TABLE_NAME = /^t[0-9a-f]{32}$/;

// What a linked record dialog is called: a link after the table it shows, a
// single select after its field, since its options live in a table that has
// only a generated name. tables are the workspace's tables as they are cached.
export function linkedDialogTitle(header, tables = []) {
    if (!header || header.single_select) return fieldLabel(header);

    const junction = tables.find((t) => t.id === header.linked_id);
    const target = tables.find(
        (t) => t.id === (header.parent_table_id || junction?.parent_table_id),
    );
    const shown =
        typeof target?.display_name === "string"
            ? target.display_name
            : target?.display_name?.String;

    if (shown) return shown;

    return target?.name && !GENERATED_TABLE_NAME.test(target.name)
        ? target.name
        : fieldLabel(header);
}

// A new field comes back typed as it was created ("single select", "number"),
// but the grid shows a field by the type of its column, as a reload gives it.
// empty gives each row its own value, so rows never share one list.
export function newFieldColumn(headerType) {
    const type = String(headerType || "").toUpperCase();

    switch (type) {
        case "BOOL":
            return { headerType: "TINYINT", empty: () => false };
        case "TEXT":
        case "VARCHAR":
            return { headerType: type, empty: () => "" };
        case "DATE":
            return { headerType: type, empty: () => null };
        case "DECIMAL":
        case "INT":
            return { headerType: type, empty: () => 0 };
        case "NUMBER":
            return { headerType: "DECIMAL", empty: () => 0 };
        case "LINK":
            return { headerType: "VARCHAR", empty: () => [] };
        case "ASSIGNEE":
        case "SINGLE SELECT":
        case "CALCULATIONS":
        case "MASTER LINK":
        case "URL":
            return { headerType: "VARCHAR", empty: () => "" };
        case "FILE":
            return { headerType: "FILE", empty: () => [] };
        default:
            return { headerType: type, empty: () => null };
    }
}

// Shapes a task from the server the way the grid holds its rows.
export function normalizeTaskRow(item, headers) {
    const row = {};

    for (const h of headers) {
        row[h.name] =
            h.header_type === "TINYINT"
                ? item[h.name] === "1" || item[h.name] === 1 || item[h.name] === true
                : (item[h.name] ?? "");
    }

    row.id = norm(item.id);
    const rawParent = item.parent_task_id;
    const parent = rawParent == null ? null : norm(rawParent) || null;

    row.parent_task_id = parent;
    row._original_parent_task_id = parent;
    row._is_subtask = !!(parent && parent !== row.id);

    return row;
}

// The options of each single select field, by id, from the tables they live
// in. A board's cards show an option by its name and colour, not its id.
export function optionLookup(headers, tables) {
    const tableById = new Map();

    for (const table of tables || []) {
        if (!tableById.has(table.id)) tableById.set(table.id, table);
    }

    const lookup = new Map();

    for (const header of headers || []) {
        if (!header.single_select || !header.linked_id) continue;
        const options = tableById.get(header.linked_id)?.data_base;

        if (options)
            lookup.set(header.name, new Map(options.map((option) => [String(option.id), option])));
    }

    return lookup;
}

// A copy of the task with each option id it holds replaced by the option.
export function withOptions(task, lookup) {
    const result = { ...task };

    for (const [name, options] of lookup) {
        const option = task[name] ? options.get(String(task[name])) : undefined;

        if (option) result[name] = option;
    }

    return result;
}

// The rows left once a task is deleted: its subtasks, at every depth, are
// deleted with it.
export function withoutTask(rows, taskId) {
    const gone = new Set([String(taskId)]);
    let grew = true;

    while (grew) {
        grew = false;
        for (const row of rows || []) {
            const id = String(row.id);
            const parent = row.parent_task_id == null ? "" : String(row.parent_task_id);

            if (parent && parent !== id && gone.has(parent) && !gone.has(id)) {
                gone.add(id);
                grew = true;
            }
        }
    }

    return (rows || []).filter((row) => !gone.has(String(row.id)));
}

// Removes a deleted task or option from rows that point at it. A task only
// appears in link lists; an option can also be the whole value of a cell.
export function clearReferences(rows, deletedId, { isOption = false } = {}) {
    if (!Array.isArray(rows) || !deletedId) return rows;
    let changed = false;
    const next = rows.map((row) => {
        const cleared = {};

        for (const key of Object.keys(row)) {
            const val = row[key];

            if (Array.isArray(val)) {
                const filtered = val.filter((v) => v?.id !== deletedId && v !== deletedId);

                if (filtered.length !== val.length) cleared[key] = filtered;
            } else if (isOption && val && typeof val === "object" && val.id === deletedId) {
                cleared[key] = null;
            } else if (isOption && val === deletedId) {
                cleared[key] = null;
            }
        }

        if (Object.keys(cleared).length === 0) return row;
        changed = true;

        return { ...row, ...cleared };
    });

    return changed ? next : rows;
}

// Refreshes the copies of a changed option or linked task that other rows
// embed, for the columns of those rows that point at its table. A single
// select's linked_id is its option table; a link's is the junction table,
// and the table it shows is its parent, or for a one-way link the
// junction's, which tables gives.
export function refreshEmbedded(rows, headers, sourceTableId, changed, tables = []) {
    if (!Array.isArray(rows) || !changed?.id) return rows;
    const source = String(sourceTableId);
    const shows = (h) => {
        if (String(h.linked_id) === source) return true;
        if (h.single_select) return false;
        const target =
            h.parent_table_id ||
            tables.find((t) => String(t.id) === String(h.linked_id))?.parent_table_id;

        return target != null && String(target) === source;
    };
    const columns = (headers || []).filter(shows);

    if (columns.length === 0) return rows;
    const id = String(changed.id);
    let touched = false;
    const next = rows.map((row) => {
        const patch = {};

        for (const h of columns) {
            const cell = row[h.name];

            if (Array.isArray(cell)) {
                if (cell.some((v) => String(v?.id) === id)) {
                    patch[h.name] = cell.map((v) =>
                        String(v?.id) === id ? { ...v, name: changed.name } : v,
                    );
                }
            } else if (cell && typeof cell === "object" && String(cell.id) === id) {
                patch[h.name] = { ...cell, ...changed };
            }
        }

        if (Object.keys(patch).length === 0) return row;
        touched = true;

        return { ...row, ...patch };
    });

    return touched ? next : rows;
}

// Counts each task's subtasks in one pass, keyed by the parent's id as a string.
export function countSubtasks(rows) {
    const counts = new Map();

    for (const row of rows || []) {
        if (row.parent_task_id == null) continue;
        const parent = String(row.parent_task_id);

        counts.set(parent, (counts.get(parent) || 0) + 1);
    }

    return counts;
}
