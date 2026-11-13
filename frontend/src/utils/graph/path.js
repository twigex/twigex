// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Follows each dot back to the dot it was opened from, up to the starting
// dot, and returns the keys in order from the start. parentOf gives a dot's
// parent key, or an empty value for the starting dot.
export function pathTo(key, parentOf) {
    const path = [];
    const seen = new Set();

    let current = key;

    while (current && !seen.has(current)) {
        path.unshift(current);
        seen.add(current);

        current = parentOf(current);
    }

    return path;
}
