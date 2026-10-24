// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Types an admin creates have no colour of their own, so each takes the next
// one from this list, in the order of the type list.
const PALETTE = [
    { color: "#0891b2", badge: "bg-cyan-50 text-cyan-700 ring-cyan-700/10" },
    { color: "#7c3aed", badge: "bg-violet-50 text-violet-700 ring-violet-700/10" },
    { color: "#ea580c", badge: "bg-orange-50 text-orange-700 ring-orange-600/20" },
    { color: "#16a34a", badge: "bg-green-50 text-green-700 ring-green-600/20" },
    { color: "#e11d48", badge: "bg-rose-50 text-rose-700 ring-rose-600/10" },
    { color: "#2563eb", badge: "bg-blue-50 text-blue-700 ring-blue-700/10" },
    { color: "#ca8a04", badge: "bg-yellow-50 text-yellow-800 ring-yellow-600/20" },
    { color: "#0d9488", badge: "bg-teal-50 text-teal-700 ring-teal-600/20" },
];

const BUILT_IN = {
    chart: { color: "#4f46e5", badge: "bg-indigo-50 text-indigo-700 ring-indigo-700/10" },
    dashboard: { color: "#059669", badge: "bg-emerald-50 text-emerald-700 ring-emerald-600/20" },
    dataset: { color: "#d97706", badge: "bg-amber-50 text-amber-800 ring-amber-600/20" },
    dataset_field: { color: "#db2777", badge: "bg-pink-50 text-pink-700 ring-pink-700/10" },
};

export const GROUP_STYLE = { color: "#9ca3af", badge: "bg-gray-50 text-gray-600 ring-gray-500/10" };

export function typeStyles(entityTypes) {
    const styles = {};

    let next = 0;

    for (const entityType of entityTypes) {
        if (BUILT_IN[entityType.key]) {
            styles[entityType.key] = BUILT_IN[entityType.key];

            continue;
        }

        styles[entityType.key] = PALETTE[next % PALETTE.length];
        next = next + 1;
    }

    return styles;
}
