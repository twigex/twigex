<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full flex flex-col justify-between" @contextmenu.prevent="rightClick">
        <ul
            role="list"
            class="h-full overflow-y-auto"
            :class="
                grid
                    ? 'grid auto-rows-min grid-cols-[repeat(auto-fill,minmax(8.5rem,1fr))] gap-3 p-3 sm:grid-cols-[repeat(auto-fill,minmax(12rem,1fr))]'
                    : 'px-2'
            "
        >
            <slot name="default"></slot>
        </ul>
        <slot name="context"></slot>

        <slot name="pagination"></slot>
    </div>
</template>

<script setup>
defineProps({
    grid: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["right-click"]);

function rightClick(event) {
    //stop propagation so that the parent context menu doesn't open
    event.stopPropagation();

    emit("right-click", { x: event.x, y: event.y, file: null });
}
</script>
