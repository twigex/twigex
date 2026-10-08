// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { fieldLabel } from "./rows";

const DEFAULT_WIDTH = 150;

// pickView finds the view a grid opens: the one named, or the table's first.
export function pickView(table, viewId) {
    const options = Array.isArray(table?.options) ? table.options : [];
    const views = Array.isArray(table?.views) ? table.views : [];

    return (
        options.find((o) => o.id === viewId) ||
        views.find((v) => v.id === viewId) ||
        options[0] ||
        null
    );
}

// viewHeaders gives the table's fields as the view shows them, named,
// shown or hidden, in its saved order, with each column's width. Widths are
// matched by field name, as the saved order may hold columns the table no
// longer has.
export function viewHeaders(table, view) {
    if (!Array.isArray(table?.headers)) {
        return { headers: null, widths: Array(table?.headers?.length || 0).fill(DEFAULT_WIDTH) };
    }

    const order = view?.order ? JSON.parse(view.order) : null;
    const saved = Array.isArray(order) ? new Map(order.map((o) => [o.name, o])) : null;
    const headerOrder = Array.isArray(order)
        ? order.find((o) => o?.name === "headerOrder")?.value
        : null;

    let headers = table.headers.map((h) => {
        const o = saved?.get(h.name);

        return {
            ...h,
            display_name: fieldLabel(h, o?.display_name),
            visible: h.name === "name" ? true : (o?.visible ?? true),
        };
    });

    if (Array.isArray(headerOrder)) {
        const byName = Object.fromEntries(headers.map((h) => [h.name, h]));

        headers = headerOrder.map((name) => byName[name]).filter(Boolean);
    }

    const widths = headers.map((h) => {
        const o = saved?.get(h.name);

        return o ? Number(o.width) || DEFAULT_WIDTH : DEFAULT_WIDTH;
    });

    return { headers, widths };
}

// tableDisplayNames maps each table's id to the name it is shown by.
export function tableDisplayNames(tables) {
    const names = {};

    for (const t of tables) {
        const display = t.display_name?.String;

        names[t.id] = t.display_name?.Valid && display?.trim() ? display : t.name;
    }

    return names;
}

const withLinkedOptions = (f) => ({ ...f, linkedOptions: f.linkedOptions || [] });

// viewFilterState is the filter a view opens with: the saved filter that is
// active and matches the view's own, or else the view's own, made ready for
// the filter builder and returned as built too. It is null for a view
// without a filter.
export function viewFilterState(savedFilters, groups, flatFilters) {
    if (groups.length === 0 && flatFilters.length === 0) return null;

    const active = savedFilters.find(
        (f) =>
            f.is_active &&
            JSON.stringify(f.filters?.flatFilters) === JSON.stringify(flatFilters) &&
            JSON.stringify(f.filters?.groups) === JSON.stringify(groups),
    );

    if (active) return { filter: active, built: null };

    const built = {
        groups: groups.map((group) => ({
            ...group,
            filters: (group.filters || []).map(withLinkedOptions),
        })),
        flatFilters: flatFilters.map(withLinkedOptions),
    };

    return { filter: { filters: built }, built };
}

// parseViewSort reads a view's saved sort, which may be stored as text.
export function parseViewSort(raw) {
    try {
        const sort = typeof raw === "string" && raw ? JSON.parse(raw) : raw;

        return Array.isArray(sort) ? sort : [];
    } catch {
        return [];
    }
}
