<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex justify-center items-center">
        <div
            class="relative group cursor-pointer"
            :style="{ width: size + 'px', height: size + 'px' }"
        >
            <!-- PROGRESS / STATUS -->
            <div class="transition-opacity duration-200 group-hover:opacity-20">
                <svg
                    v-if="status === 'pending' || status === 'uploading'"
                    :width="size"
                    :height="size"
                    class="transform -rotate-90"
                >
                    <circle
                        :cx="size / 2"
                        :cy="size / 2"
                        :r="radius"
                        fill="none"
                        class="stroke-current text-gray-300"
                        :stroke-width="strokeWidth"
                    />
                    <circle
                        :cx="size / 2"
                        :cy="size / 2"
                        :r="radius"
                        fill="none"
                        class="stroke-current text-indigo-600"
                        :stroke-width="strokeWidth"
                        :stroke-dasharray="circumference"
                        :stroke-dashoffset="circumference * (1 - progress / 100)"
                    />
                </svg>

                <svg
                    v-else-if="status === 'completed'"
                    :width="size"
                    :height="size"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                    class="text-green-400"
                    fill="none"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                    />
                </svg>

                <svg
                    v-else-if="status === 'error' || status === 'cancelled'"
                    :width="size"
                    :height="size"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                    :class="status === 'error' ? 'text-red-500' : 'text-amber-500'"
                    fill="none"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="m9.75 9.75 4.5 4.5m0-4.5-4.5 4.5M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                    />
                </svg>
            </div>

            <!-- CANCEL OVERLAY -->
            <button
                v-if="status === 'pending' || status === 'uploading'"
                @click.stop="$emit('cancel')"
                class="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200 text-gray-700 hover:text-red-500"
            >
                <XMarkIcon :style="{ width: size / 2 + 'px' }" />
            </button>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    size: { type: Number, default: 100 },
    progress: { type: Number, default: 0 },
    status: { type: String, default: "pending" },
});

defineEmits(["cancel"]);

const strokeWidth = computed(() => Math.max(4, props.size / 15));
const radius = computed(() => props.size / 2 - strokeWidth.value / 2);
const circumference = computed(() => 2 * Math.PI * radius.value);
</script>
