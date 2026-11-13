// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// The colours a type may have, by the names the server accepts. Each keeps
// its classes written out in full, so Tailwind includes them.
export const COLORS = {
    indigo: {
        color: "#4f46e5",
        badge: "bg-indigo-50 text-indigo-700 ring-indigo-700/10",
        tile: "bg-indigo-50 text-indigo-600",
    },
    emerald: {
        color: "#059669",
        badge: "bg-emerald-50 text-emerald-700 ring-emerald-600/20",
        tile: "bg-emerald-50 text-emerald-600",
    },
    amber: {
        color: "#d97706",
        badge: "bg-amber-50 text-amber-800 ring-amber-600/20",
        tile: "bg-amber-50 text-amber-600",
    },
    pink: {
        color: "#db2777",
        badge: "bg-pink-50 text-pink-700 ring-pink-700/10",
        tile: "bg-pink-50 text-pink-600",
    },
    cyan: {
        color: "#0891b2",
        badge: "bg-cyan-50 text-cyan-700 ring-cyan-700/10",
        tile: "bg-cyan-50 text-cyan-600",
    },
    violet: {
        color: "#7c3aed",
        badge: "bg-violet-50 text-violet-700 ring-violet-700/10",
        tile: "bg-violet-50 text-violet-600",
    },
    orange: {
        color: "#ea580c",
        badge: "bg-orange-50 text-orange-700 ring-orange-600/20",
        tile: "bg-orange-50 text-orange-600",
    },
    green: {
        color: "#16a34a",
        badge: "bg-green-50 text-green-700 ring-green-600/20",
        tile: "bg-green-50 text-green-600",
    },
    rose: {
        color: "#e11d48",
        badge: "bg-rose-50 text-rose-700 ring-rose-600/10",
        tile: "bg-rose-50 text-rose-600",
    },
    blue: {
        color: "#2563eb",
        badge: "bg-blue-50 text-blue-700 ring-blue-700/10",
        tile: "bg-blue-50 text-blue-600",
    },
    yellow: {
        color: "#ca8a04",
        badge: "bg-yellow-50 text-yellow-800 ring-yellow-600/20",
        tile: "bg-yellow-50 text-yellow-700",
    },
    teal: {
        color: "#0d9488",
        badge: "bg-teal-50 text-teal-700 ring-teal-600/20",
        tile: "bg-teal-50 text-teal-600",
    },
    slate: {
        color: "#475569",
        badge: "bg-slate-50 text-slate-700 ring-slate-600/20",
        tile: "bg-slate-100 text-slate-600",
    },
};

// The order in which types without a chosen colour are given one.
const AUTO_ORDER = ["cyan", "violet", "orange", "green", "rose", "blue", "yellow", "teal", "slate"];

export const GROUP_STYLE = {
    color: "#9ca3af",
    badge: "bg-gray-50 text-gray-600 ring-gray-500/10",
    tile: "bg-gray-100 text-gray-500",
};

// Each type gets the colour its admin chose. A type without one takes the
// next colour no other type uses yet, in the order of the type list.
export function typeStyles(entityTypes) {
    const styles = {};
    const used = new Set();

    for (const entityType of entityTypes) {
        if (COLORS[entityType.color]) {
            used.add(entityType.color);
        }
    }

    let next = 0;

    for (const entityType of entityTypes) {
        if (COLORS[entityType.color]) {
            styles[entityType.key] = COLORS[entityType.color];

            continue;
        }

        let name = AUTO_ORDER[next % AUTO_ORDER.length];

        for (let tries = 0; tries < AUTO_ORDER.length && used.has(name); tries++) {
            next = next + 1;
            name = AUTO_ORDER[next % AUTO_ORDER.length];
        }

        used.add(name);
        next = next + 1;

        styles[entityType.key] = COLORS[name];
    }

    return styles;
}
