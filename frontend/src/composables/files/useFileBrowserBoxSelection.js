// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, onBeforeUnmount } from "vue";

export function useFileBrowserBoxSelection({ container, fileItemsRefs, fileList, selectedFiles }) {
    const isDragging = ref(false);
    const isSelecting = ref(false);
    const startX = ref(0);
    const startY = ref(0);
    const currentX = ref(0);
    const currentY = ref(0);

    const boxLeft = ref(0);
    const boxTop = ref(0);
    const boxWidth = ref(0);
    const boxHeight = ref(0);

    const onMouseDown = (e) => {
        if (e.button !== 0) return;
        if (e.target.closest("[data-drag-handle]")) return;

        isDragging.value = false;
        isSelecting.value = true;

        const rect = container.value.getBoundingClientRect();

        startX.value = Math.max(0, e.clientX - rect.left);
        startY.value = Math.max(0, e.clientY - rect.top);

        currentX.value = startX.value;
        currentY.value = startY.value;

        updateSelectionBox();

        document.addEventListener("mousemove", onMouseMove);
        document.addEventListener("mouseup", onMouseUp);
    };

    const onMouseMove = (e) => {
        if (!isSelecting.value) return;

        isDragging.value = true;

        const rect = container.value.getBoundingClientRect();

        currentX.value = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
        currentY.value = Math.max(0, Math.min(e.clientY - rect.top, rect.height));

        updateSelectionBox();
        checkSelection();
    };

    const onMouseUp = () => {
        if (!isSelecting.value) return;

        isSelecting.value = false;

        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
    };

    const updateSelectionBox = () => {
        boxLeft.value = Math.min(startX.value, currentX.value);
        boxTop.value = Math.min(startY.value, currentY.value);
        boxWidth.value = Math.abs(currentX.value - startX.value);
        boxHeight.value = Math.abs(currentY.value - startY.value);
    };

    const checkSelection = () => {
        selectedFiles.value = [];

        const filesById = new Map(fileList.value.map((file) => [file.id, file]));

        fileItemsRefs.value.forEach((fileItem) => {
            if (!fileItem) return; // Safety check in case a ref is null

            const fileRect = fileItem.$el.getBoundingClientRect(); // Use $el for component root
            const containerRect = container.value.getBoundingClientRect();

            const relativeFileRect = {
                left: fileRect.left - containerRect.left,
                top: fileRect.top - containerRect.top,
                right: fileRect.right - containerRect.left,
                bottom: fileRect.bottom - containerRect.top,
            };

            if (
                relativeFileRect.right > boxLeft.value &&
                relativeFileRect.left < boxLeft.value + boxWidth.value &&
                relativeFileRect.bottom > boxTop.value &&
                relativeFileRect.top < boxTop.value + boxHeight.value
            ) {
                const file = filesById.get(fileItem.$el.dataset.id);

                if (file) {
                    selectedFiles.value.push(file);
                }
            }
        });
    };

    const onContainerClick = () => {
        if (!isDragging.value) {
            selectedFiles.value = [];
        }
    };

    onBeforeUnmount(() => {
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
    });

    return {
        isSelecting,
        boxLeft,
        boxTop,
        boxWidth,
        boxHeight,
        onMouseDown,
        onContainerClick,
    };
}
