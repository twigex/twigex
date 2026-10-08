// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, onBeforeUnmount, ref, watch } from "vue";

// Shows a long list a batch at a time: the first rows at once, the rest a
// batch per frame. Changes to a list already shown do not restart it.
export function useProgressiveRender(rows, batch) {
    const renderedCount = ref(batch);
    const shown = computed(() => rows.value.slice(0, renderedCount.value));

    let renderFrame = 0;

    function start() {
        cancelAnimationFrame(renderFrame);
        renderedCount.value = batch;
        const addBatch = () => {
            renderedCount.value = Math.min(renderedCount.value + batch, rows.value.length);
            if (renderedCount.value < rows.value.length)
                renderFrame = requestAnimationFrame(addBatch);
        };

        renderFrame = requestAnimationFrame(addBatch);
    }

    onBeforeUnmount(() => cancelAnimationFrame(renderFrame));

    watch(
        () => rows.value.length,
        (newLen) => {
            if (newLen > renderedCount.value) renderedCount.value = newLen;
        },
    );

    return { renderedCount, shown, start };
}
