// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { unref } from "vue";

// Lets a table's columns be resized by dragging their header's edge. While
// dragging, the cells are sized directly, a frame at a time, without
// rendering the table again; the widths are set once the drag ends. tableId
// is the data-table-id of the table's root element.
export function useColumnResize({ headers, widths, tableId, onResized }) {
    const startResize = (index, event) => {
        const header = headers.value[index];

        if (!header) {
            console.error("No header found at index:", index);

            return;
        }

        const columnName = header.name;
        const initialX = event.clientX;
        const initialWidth = widths.value[index];

        const nonReactiveWidths = [...widths.value];

        let rafId = null;
        let lastWidth = initialWidth;

        const handleMouseMove = (event) => {
            const dx = event.clientX - initialX;
            const newWidth = Math.max(70, initialWidth + dx);

            if (newWidth !== lastWidth) {
                lastWidth = newWidth;
                nonReactiveWidths[index] = newWidth;

                if (rafId) cancelAnimationFrame(rafId);

                rafId = requestAnimationFrame(() => {
                    const styleWidth = `${newWidth}px`;

                    let elements = document.querySelectorAll(
                        `[data-table-id="${unref(tableId)}"] [data-column-name="${columnName}"]`,
                    );

                    if (elements.length === 0) {
                        elements = document.querySelectorAll(
                            `[data-table-id="${unref(tableId)}"] [data-column-index="${index}"]`,
                        );
                    }

                    elements.forEach((el) => {
                        el.style.width = styleWidth;
                        el.style.minWidth = styleWidth;
                    });
                    rafId = null;
                });
            }
        };

        const handleMouseUp = () => {
            if (rafId) cancelAnimationFrame(rafId);

            window.removeEventListener("mousemove", handleMouseMove);
            window.removeEventListener("mouseup", handleMouseUp);

            // Always update reactive widths (dialog: updates localDialogWidths; non-dialog: updates store)
            widths.value = nonReactiveWidths;

            // A width the server refuses goes back to what it was.
            Promise.resolve(onResized?.(columnName, nonReactiveWidths[index])).then((saved) => {
                if (saved !== false) return;
                const restored = [...widths.value];

                restored[index] = initialWidth;
                widths.value = restored;
                document
                    .querySelectorAll(
                        `[data-table-id="${unref(tableId)}"] [data-column-name="${columnName}"]`,
                    )
                    .forEach((el) => {
                        el.style.width = `${initialWidth}px`;
                        el.style.minWidth = `${initialWidth}px`;
                    });
            });
        };

        window.addEventListener("mousemove", handleMouseMove);
        window.addEventListener("mouseup", handleMouseUp);
    };

    return { startResize };
}
