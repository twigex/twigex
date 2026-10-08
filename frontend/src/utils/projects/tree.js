// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// workspaceTables lists each of a workspace's tables once. The workspace
// details send every table at the top and again in its folder, and folders
// hold folders.
export function workspaceTables(workspace) {
    const tables = new Map();
    const add = (table) => tables.has(table.id) || tables.set(table.id, table);
    const walk = (folders) =>
        (folders || []).forEach((folder) => {
            (folder.tables || []).forEach(add);
            walk(folder.children);
        });

    (workspace?.tables || []).forEach(add);
    walk(workspace?.folders);

    return [...tables.values()];
}

export function isTableNode(node) {
    return node?.type === "project";
}

export function findTreeNode(nodes, id) {
    for (const node of nodes || []) {
        if (node.id === id) return node;
        const found = findTreeNode(node.children, id);

        if (found) return found;
    }

    return null;
}

export function findTreeParent(nodes, id, parent = null) {
    for (const node of nodes || []) {
        if (node.id === id) return parent;
        const found = findTreeParent(node.children, id, node);

        if (found) return found;
    }

    return null;
}

export function removeTreeNode(nodes, id, matches = () => true) {
    for (const [i, node] of (nodes || []).entries()) {
        if (node.id === id && matches(node)) {
            nodes.splice(i, 1);

            return true;
        }

        if (removeTreeNode(node.children, id, matches)) return true;
    }

    return false;
}

function contains(node, id) {
    if (node.id === id) return true;

    return (node.children || []).some((child) => contains(child, id));
}

export function canMoveInto(roots, item, target) {
    if (!item || !target || isTableNode(target) || !findTreeNode(roots, target.id)) {
        return false;
    }

    if (findTreeParent(roots, item.id)?.id === target.id) return false;

    return isTableNode(item) || !contains(item, target.id);
}

// Folders are kept ahead of tables, the order the server's tree comes in.
export function moveTreeNode(roots, itemId, targetId) {
    const item = findTreeNode(roots, itemId);
    const target = findTreeNode(roots, targetId);

    if (
        !item ||
        !target ||
        isTableNode(target) ||
        (!isTableNode(item) && contains(item, targetId))
    ) {
        return false;
    }

    const parent = findTreeParent(roots, itemId);

    if (parent === target) return false;
    if (parent) {
        parent.children.splice(parent.children.indexOf(item), 1);
    }

    if (!target.children) target.children = [];
    if (isTableNode(item)) {
        target.children.push(item);
    } else {
        const firstTable = target.children.findIndex(isTableNode);

        target.children.splice(firstTable === -1 ? target.children.length : firstTable, 0, item);
    }

    return true;
}

const taskIndents = new WeakMap();

// Lists tasks with the subtasks of each expanded task under it, each row's
// depth readable as its _indent. A task whose parent is not among the rows,
// or is the task itself, is listed at the top. The depth is kept beside the
// row rather than written onto it, so the row is not changed.
export function flattenTaskTree(rows, expanded) {
    const children = new Map();
    const ids = new Set();

    for (const row of rows) {
        ids.add(String(row.id));
        const parent = row.parent_task_id == null ? null : String(row.parent_task_id);

        if (parent === null || parent === String(row.id)) continue;
        let list = children.get(parent);

        if (!list) children.set(parent, (list = []));
        list.push(row);
    }

    const open = new Set([...expanded].map(String));
    const roots = rows.filter((row) => {
        const parent = row.parent_task_id == null ? null : String(row.parent_task_id);

        return parent === null || parent === String(row.id) || !ids.has(parent);
    });

    const flat = [];
    const stack = [];

    for (let i = roots.length - 1; i >= 0; i--) stack.push([roots[i], 0]);

    while (stack.length) {
        const [node, depth] = stack.pop();

        setIndent(node, depth);
        flat.push(node);

        const kids = open.has(String(node.id)) ? children.get(String(node.id)) : null;

        if (kids) for (let i = kids.length - 1; i >= 0; i--) stack.push([kids[i], depth + 1]);
    }

    return flat;
}

function setIndent(node, depth) {
    taskIndents.set(node, depth);
    const own = Object.getOwnPropertyDescriptor(node, "_indent");

    if (!own) {
        Object.defineProperty(node, "_indent", {
            get() {
                return taskIndents.get(node) || 0;
            },
            set(val) {
                taskIndents.set(node, val);
            },
            configurable: true,
            enumerable: false,
        });
    } else if (!own.get) {
        node._indent = depth;
    }
}
