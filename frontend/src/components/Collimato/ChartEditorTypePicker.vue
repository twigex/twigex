<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="px-4 pt-4 pb-2">
        <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">
            {{ t("collimato.charts.new_chart.chart_type") }}
        </p>
    </div>
    <div class="px-4 pb-4 grid grid-cols-3 gap-2">
        <button
            v-for="c in charts"
            :key="c.value"
            type="button"
            @click="emit('select', c.value)"
            class="flex flex-col items-center gap-1.5 rounded-md border p-2.5 text-xs font-medium transition-colors focus:outline-none"
            :class="
                chartType === c.value
                    ? 'border-indigo-600 bg-indigo-50 text-indigo-700'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-gray-300 hover:bg-gray-50'
            "
        >
            <component :is="c.icon" class="h-5 w-5" aria-hidden="true" />
            <span>{{ c.shortName }}</span>
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

defineProps({
    charts: {
        type: Array,
        default: () => [],
    },
    chartType: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["select"]);
</script>
