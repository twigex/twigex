// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";

function generateGrid(participants) {
    const participantCount = participants.length;

    if (participantCount <= 0) return { rows: 0, columns: 0 };
    if (participantCount === 1) return { rows: 1, columns: 1 };

    let rows, columns;

    if (participantCount <= 2) {
        rows = 1;
        columns = 2;
    } else if (participantCount <= 4) {
        rows = 2;
        columns = 2;
    } else if (participantCount <= 6) {
        rows = 2;
        columns = 3;
    } else {
        rows = Math.floor(Math.sqrt(participantCount));
        columns = Math.ceil(participantCount / rows);

        while (rows * columns - participantCount >= columns) {
            rows--;
            columns = Math.ceil(participantCount / rows);
        }
    }

    return { rows, columns };
}

export function useCallGrid(participants) {
    const grid = computed(() => generateGrid(participants.value));

    const gridRows = computed(() => {
        const rows = [];
        let participantIndex = 0;
        const participantsArray = participants.value;

        for (let i = 0; i < grid.value.rows; i++) {
            const row = [];
            const colsInRow = Math.min(
                grid.value.columns,
                participantsArray.length - participantIndex,
            );

            for (let j = 0; j < colsInRow; j++) {
                row.push(participantsArray[participantIndex++]);
            }

            rows.push(row);
        }

        return rows;
    });

    return {
        grid,
        gridRows,
    };
}
