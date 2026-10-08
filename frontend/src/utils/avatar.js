// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Full class names: Tailwind scans source text, so a name built at runtime
// produces no CSS.
const AVATAR_COLORS = [
    "bg-indigo-500",
    "bg-sky-600",
    "bg-teal-600",
    "bg-emerald-600",
    "bg-amber-600",
    "bg-rose-500",
    "bg-fuchsia-600",
    "bg-violet-500",
    "bg-cyan-700",
];

export function initialsFor(name) {
    const first = [...String(name ?? "").trim()][0];

    return first ? first.toUpperCase() : "?";
}

export function colorClassFor(id) {
    const key = String(id ?? "");

    let hash = 0;

    for (let i = 0; i < key.length; i++) {
        hash = (hash * 31 + key.charCodeAt(i)) | 0;
    }

    return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}
