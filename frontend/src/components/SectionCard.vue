<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="overflow-hidden rounded-md border border-gray-200">
        <div
            class="flex items-center justify-between gap-x-4 border-b border-gray-200 bg-gray-50 px-4 py-3"
        >
            <div>
                <h2 class="text-sm font-semibold text-gray-900">
                    {{ title }}
                </h2>
                <p v-if="description" class="mt-0.5 text-xs text-gray-500">
                    {{ description }}
                </p>
            </div>

            <div class="shrink-0">
                <slot name="action" />
            </div>
        </div>

        <slot />
    </div>
</template>

<script setup>
defineProps({
    title: {
        type: String,
        required: true,
    },
    description: {
        type: String,
        default: "",
    },
});
</script>
