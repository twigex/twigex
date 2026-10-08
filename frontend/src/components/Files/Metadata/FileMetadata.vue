<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div
            v-for="m in file.metadata"
            :key="m.id"
            class="min-w-0 flex flex-row justify-between text-sm leading-6 py-2"
        >
            <label for="comments" class="font-medium text-gray-900">{{ m.title }}:</label>
            <div v-if="m.type == 'color'" class="text-gray-500">
                <span
                    class="inline-flex items-center gap-x-1.5 rounded-md px-2 py-1 text-xs font-medium text-white"
                    :style="{
                        backgroundColor: m.value.color,
                    }"
                >
                    {{ m.value.name }}
                </span>
            </div>

            <div v-else class="text-gray-500">{{ m.value }}</div>
        </div>
    </div>
</template>

<script setup>
defineProps({
    file: Object,
});
</script>
