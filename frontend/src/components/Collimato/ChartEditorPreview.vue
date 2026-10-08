<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex flex-col overflow-hidden border-b border-gray-200"
        :class="chartOpen ? 'flex-[3] min-h-0' : 'shrink-0'"
    >
        <div
            class="shrink-0 flex items-center gap-2 bg-gray-50 hover:bg-gray-100 cursor-pointer select-none px-4 py-2"
            @click="chartOpen = !chartOpen"
        >
            <ChevronRightIcon
                class="h-4 w-4 text-gray-400 transition-transform duration-200"
                :class="chartOpen ? 'rotate-90' : ''"
                aria-hidden="true"
            />
            <component :is="chartIcon" class="h-4 w-4 text-gray-400" aria-hidden="true" />
            <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">{{
                t("collimato.charts.new_chart.chart_label")
            }}</span>
        </div>
        <div
            v-show="chartOpen"
            class="relative flex flex-1 min-h-0 items-center justify-center bg-white"
        >
            <div
                v-if="loading"
                class="absolute inset-0 z-10 flex items-center justify-center bg-white/75"
            >
                <div
                    class="h-10 w-10 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent"
                ></div>
            </div>

            <BigNumber v-if="options != null && chartType == 'big_number'" :options="options" />

            <ChartContainer
                v-else-if="options != null && isEchartsType(chartType)"
                :option="options"
                :chartType="chartType"
                :chartOptions="chartOptions"
            />

            <div v-else class="flex flex-col text-center items-center gap-2">
                <component :is="chartIcon" class="h-12 w-12 text-gray-300" aria-hidden="true" />
                <h3 class="text-sm font-semibold text-gray-900">
                    {{ t("collimato.charts.new_chart.no_data") }}
                </h3>
                <p class="text-sm text-gray-500">
                    {{ t("collimato.charts.new_chart.no_data_description") }}
                </p>
            </div>
        </div>
    </div>

    <div
        class="flex flex-col overflow-hidden"
        :class="resultsOpen ? (chartOpen ? 'flex-[2] min-h-0' : 'flex-1 min-h-0') : 'shrink-0'"
    >
        <div
            class="shrink-0 flex items-center gap-2 border-b border-gray-200 bg-gray-50 hover:bg-gray-100 cursor-pointer select-none px-4 py-2"
            @click="resultsOpen = !resultsOpen"
        >
            <ChevronRightIcon
                class="h-4 w-4 text-gray-400 transition-transform duration-200"
                :class="resultsOpen ? 'rotate-90' : ''"
                aria-hidden="true"
            />
            <TableCellsIcon class="h-4 w-4 text-gray-400" aria-hidden="true" />
            <span class="text-xs font-semibold uppercase tracking-wider text-gray-400">{{
                t("collimato.charts.new_chart.results_label")
            }}</span>
            <span
                v-if="tableData.data.length > 0"
                class="rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-medium text-indigo-700"
            >
                {{ tableData.data.length }}
            </span>
        </div>
        <div v-show="resultsOpen" class="flex-1 min-h-0 overflow-y-auto overflow-x-auto bg-white">
            <Table
                v-if="tableData.headers.length > 0"
                class="pb-1"
                :headers="tableData.headers"
                :items="tableData.data"
                :dense="true"
            />
            <div v-else class="flex flex-col items-center justify-center h-full gap-2">
                <TableCellsIcon class="h-10 w-10 text-gray-300" aria-hidden="true" />
                <p class="text-sm font-medium text-gray-500">
                    {{ t("collimato.charts.new_chart.no_results") }}
                </p>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import ChartContainer from "@/components/Collimato/Charts/ChartContainer.vue";
import BigNumber from "@/components/Collimato/Charts/BigNumber.vue";
import Table from "@/components/Collimato/ResultTable.vue";
import { isEchartsType } from "@/utils/collimato/chartTypes.js";
import { ChevronRightIcon } from "@heroicons/vue/20/solid";
import { TableCellsIcon } from "@heroicons/vue/24/outline";

defineProps({
    loading: {
        type: Boolean,
        default: false,
    },
    options: {
        type: Object,
        default: null,
    },
    chartType: {
        type: String,
        default: "",
    },
    chartOptions: {
        type: Object,
        default: () => ({}),
    },
    chartIcon: {
        type: [Object, Function],
        default: null,
    },
    tableData: {
        type: Object,
        default: () => ({
            headers: [],
            data: [],
        }),
    },
});

const chartOpen = defineModel("chartOpen", {
    type: Boolean,
    default: true,
});

const resultsOpen = defineModel("resultsOpen", {
    type: Boolean,
    default: true,
});
</script>
