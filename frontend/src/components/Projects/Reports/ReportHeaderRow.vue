<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="grid-row">
        <div
            v-for="(header, index) in headers"
            :key="index"
            :class="[
                'grid-column relative hover:!bg-gray-100',
                header.name === 'id' || header.name === 'link_to_table' ? 'hidden-column' : '',
            ]"
            :style="{ width: columnWidths[index] + 'px' }"
            :data-column-index="index"
        >
            <div v-if="header.name !== 'id'" class="header-cell group">
                <label
                    class="block text-xs font-medium text-gray-500 header-label"
                    :title="header.name"
                >
                    {{ header.display_name || header.name }}
                </label>
                <button
                    class="group/handle absolute inset-y-0 right-0 z-10 flex w-1.5 cursor-col-resize items-center justify-center border-0 bg-transparent p-0 outline-none"
                    @mousedown="emit('resize-start', index, $event)"
                    title="Drag to resize"
                >
                    <span
                        class="h-[calc(100%-10px)] w-0.5 rounded-full bg-indigo-500 opacity-0 transition duration-150 group-hover/handle:opacity-100 group-active/handle:bg-indigo-600 group-active/handle:opacity-100"
                    ></span>
                </button>
            </div>
        </div>
        <div class="plus-icon-container"></div>
    </div>
</template>

<script setup>
defineProps({
    headers: { type: Array, default: () => [] },
    columnWidths: { type: Array, default: () => [] },
});

const emit = defineEmits(["resize-start"]);
</script>

<style scoped src="../projectsTable.css"></style>

<style scoped>
.grid-row .plus-icon-container {
    border-bottom: none !important;
    margin-left: 1px;
}

.grid-column {
    height: 36px;
    padding: 0 8px;
    display: flex;
    align-items: center;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.grid-column:nth-child(1),
.grid-column:nth-child(2) {
    position: sticky;
    left: 0;
    z-index: 1;
    background-color: #f9fafb;
}
</style>
