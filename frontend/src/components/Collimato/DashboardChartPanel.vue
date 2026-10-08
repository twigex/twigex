<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col w-64 shrink-0 border-l overflow-hidden">
        <div class="px-3 py-2 border-b">
            <span class="text-xs font-semibold uppercase tracking-wide text-gray-400">
                {{ t("collimato.dashboard.chart_panel.title") }}
            </span>
        </div>
        <div class="flex-1 overflow-y-auto px-3 py-3 space-y-2">
            <div
                v-for="chart in availableCharts"
                :key="chart.id"
                class="rounded-lg border bg-white p-3"
            >
                <div class="flex items-start justify-between gap-x-2">
                    <h3 class="flex-1 truncate text-sm font-medium text-gray-900">
                        {{ chart.name }}
                    </h3>
                    <button
                        type="button"
                        class="shrink-0 rounded px-2 py-1 text-xs font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2"
                        :class="
                            isAddedToGrid(chart.id)
                                ? 'bg-red-600 hover:bg-red-500'
                                : 'bg-indigo-600 hover:bg-indigo-500'
                        "
                        @click="
                            isAddedToGrid(chart.id) ? emit('remove', chart) : emit('add', chart)
                        "
                    >
                        {{
                            isAddedToGrid(chart.id)
                                ? t("collimato.dashboard.chart_panel.remove")
                                : t("collimato.dashboard.chart_panel.add")
                        }}
                    </button>
                </div>
                <div class="mt-2 flex items-center justify-between text-xs text-gray-500">
                    <span>{{ chartTypeLabel(chart.chart_type) }}</span>
                    <span>{{ getDateAndTime(chart.updated_at) }}</span>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import useDateOperations from "@/composables/useDateOperations.js";

const props = defineProps({
    availableCharts: {
        type: Array,
        default: () => [],
    },
    charts: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["add", "remove"]);

const { getDateAndTime } = useDateOperations();

const chartTypeLabels = {
    big_number: t.value("collimato.charts.new_chart.types.short.big_number"),
    line: t.value("collimato.charts.new_chart.types.short.line"),
    bar: t.value("collimato.charts.new_chart.types.short.bar"),
    time_bar: t.value("collimato.charts.new_chart.types.short.time_bar"),
    pie: t.value("collimato.charts.new_chart.types.short.pie"),
    map: t.value("collimato.charts.new_chart.types.short.map"),
};

function chartTypeLabel(type) {
    return chartTypeLabels[type] ?? type;
}

function isAddedToGrid(id) {
    return props.charts.some((c) => c.id === id);
}
</script>
