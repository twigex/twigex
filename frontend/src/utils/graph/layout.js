// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export const COLUMN_GAP = 12;
export const ROW_GAP = 1.5;
export const RING_RADIUS = 3;

// Keeps the next free row of each column. A new dot goes level with the dot
// it was opened from, or the first free row below that, so its lines stay
// short; dots already on the map never move.
export function createColumnPlacer() {
    const nextRows = new Map();

    function place(column, parentRow) {
        let row = nextRows.get(column) ?? 0;

        if (parentRow !== undefined && parentRow > row) {
            row = parentRow;
        }

        nextRows.set(column, row + 1);

        return {
            x: column * COLUMN_GAP,
            y: -row * ROW_GAP,
            row: row,
        };
    }

    function clear() {
        nextRows.clear();
    }

    return { place, clear };
}

// In the free layout the starting dot gets a full ring. A dot opened later
// fans its new neighbours out on the side facing away from where it came
// from, so the new lines don't cross back over the old ones.
function freeAngle(parent, index, total) {
    if (parent.originX === undefined) {
        return (2 * Math.PI * index) / total;
    }

    const outward = Math.atan2(parent.freeY - parent.originY, parent.freeX - parent.originX);
    const spread = Math.min(Math.PI, (Math.PI / 6) * total);
    const step = spread / total;

    return outward - spread / 2 + step * (index + 0.5);
}

export function freePosition(parent, index, total) {
    if (!parent) {
        return { x: 0, y: 0 };
    }

    const angle = freeAngle(parent, index, total);

    return {
        x: parent.freeX + RING_RADIUS * Math.cos(angle),
        y: parent.freeY + RING_RADIUS * Math.sin(angle),
        originX: parent.freeX,
        originY: parent.freeY,
    };
}
