<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }" :defaultOpen="true">
        <DisclosureButton
            class="flex w-full justify-between px-4 py-2 text-left text-sm font-medium text-gray-700"
            :class="open ? 'bg-indigo-50 hover:bg-indigo-100' : 'bg-gray-50 hover:bg-gray-100'"
        >
            <span>{{ t("collimato.charts.new_chart.chart_query.query") }}</span>
            <ChevronUpIcon
                :class="open ? 'rotate-180 transform' : ''"
                class="h-5 w-5 text-indigo-400"
            />
        </DisclosureButton>
        <DisclosurePanel class="px-4 pb-2 pt-4 text-sm text-gray-500">
            <div class="space-y-7">
                <div>
                    <div class="flex flex-row justify-between items-center">
                        <h4 class="text-small font-semibold leading-4 text-gray-400">
                            {{ t("collimato.charts.new_chart.chart_query.measures") }}
                        </h4>
                        <div class="flex border-gray-100 pt-2">
                            <button
                                @click="measureDialog = true"
                                type="button"
                                class="text-sm font-semibold leading-3 text-indigo-600 hover:text-indigo-500"
                            >
                                <span aria-hidden="true">+</span>
                                {{ t("collimato.charts.new_chart.chart_query.add_measure") }}
                            </button>
                        </div>
                    </div>

                    <ul
                        role="list"
                        class="mt-2 divide-y divide-gray-100 border-t border-gray-200 text-sm leading-6"
                    >
                        <li
                            v-for="measure in measures"
                            :key="measure.name"
                            class="flex justify-between gap-x-6 py-2 cursor-pointer hover:bg-gray-50 px-2 rounded"
                        >
                            <div class="flex flex-col truncate">
                                <div class="font-medium text-gray-900">
                                    {{ measure.name.split(".")[1] }}
                                </div>
                                <div class="text-xs text-gray-600">
                                    {{ measure.name.split(".")[0] }}
                                </div>
                            </div>
                            <button
                                @click="$emit('remove-measure', measure)"
                                type="button"
                                class="font-semibold text-red-600 hover:text-indigo-500"
                            >
                                <XCircleIcon class="h-5 w-5" />
                            </button>
                        </li>
                    </ul>
                    <p v-if="measures.length < 1" class="mt-1 text-sm leading-6 text-gray-500">
                        {{ t("collimato.charts.new_chart.chart_query.click_add_measure") }}
                    </p>
                </div>

                <div>
                    <div class="flex flex-row justify-between items-center">
                        <h4 class="text-small font-semibold leading-4 text-gray-400">
                            {{ t("collimato.charts.new_chart.chart_query.dimensions") }}
                        </h4>

                        <div class="flex border-gray-100 pt-2">
                            <button
                                @click="dimensionDialog = true"
                                type="button"
                                class="text-sm font-semibold leading-3 text-indigo-600 hover:text-indigo-500"
                            >
                                <span aria-hidden="true">+</span>
                                {{ t("collimato.charts.new_chart.chart_query.add_dimension") }}
                            </button>
                        </div>
                    </div>

                    <ul
                        role="list"
                        class="mt-2 divide-y divide-gray-100 border-t border-gray-200 text-sm leading-6"
                    >
                        <li
                            v-for="dimension in dimensions"
                            :key="dimension.name"
                            class="flex justify-between gap-x-6 py-2 cursor-pointer hover:bg-gray-50 px-2 rounded"
                        >
                            <div class="flex flex-col truncate">
                                <div class="font-medium text-gray-900">
                                    {{ dimension.name.split(".")[1] }}
                                </div>
                                <div class="text-xs text-gray-600">
                                    {{ dimension.name.split(".")[0] }}
                                </div>
                            </div>
                            <button
                                @click="$emit('remove-dimension', dimension)"
                                type="button"
                                class="font-semibold text-red-600 hover:text-indigo-500"
                            >
                                <XCircleIcon class="h-5 w-5" />
                            </button>
                        </li>
                    </ul>

                    <p v-if="dimensions.length < 1" class="mt-1 text-sm leading-6 text-gray-500">
                        {{ t("collimato.charts.new_chart.chart_query.click_add_dimension") }}
                    </p>
                </div>

                <div>
                    <h4 class="text-small font-semibold leading-4 text-gray-400 mb-2">
                        {{ t("collimato.charts.new_chart.chart_query.orders") }}
                    </h4>

                    <draggable
                        :model-value="orders"
                        @update:model-value="$emit('reorder', $event)"
                        @start="drag = true"
                        @end="drag = false"
                        item-key="name"
                    >
                        <template #item="{ element }">
                            <li
                                class="flex items-center justify-between border-t border-gray-200 hover:bg-gray-50 gap-x-6 py-2 px-2 cursor-pointer rounded"
                            >
                                <div>
                                    <div class="flex flex-col truncate">
                                        <div class="font-medium text-gray-900">
                                            {{ element.name.split(".")[1] }}
                                        </div>
                                        <div class="text-xs text-gray-600">
                                            {{ element.name.split(".")[0] }}
                                        </div>
                                    </div>
                                    <span class="isolate inline-flex rounded-md shadow-sm mt-1">
                                        <button
                                            @click="
                                                $emit('set-order', element.name, orderOptions[0])
                                            "
                                            type="button"
                                            class="relative inline-flex items-center rounded-l-md px-2 py-1 text-sm font-semibold text-gray-600 ring-1 ring-inset ring-gray-300 focus:z-10"
                                            :class="{
                                                'bg-indigo-600 hover:bg-indigo-500 text-white':
                                                    element.direction == orderOptions[0],
                                            }"
                                        >
                                            {{
                                                t(
                                                    "collimato.charts.new_chart.chart_query.orders.asc",
                                                )
                                            }}
                                        </button>
                                        <button
                                            @click="
                                                $emit('set-order', element.name, orderOptions[1])
                                            "
                                            type="button"
                                            class="relative -ml-px inline-flex items-center px-2 py-1 text-sm font-semibold text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-10"
                                            :class="{
                                                'bg-indigo-600 hover:bg-indigo-500 text-white':
                                                    element.direction == orderOptions[1],
                                            }"
                                        >
                                            {{
                                                t(
                                                    "collimato.charts.new_chart.chart_query.orders.desc",
                                                )
                                            }}
                                        </button>
                                        <button
                                            @click="
                                                $emit('set-order', element.name, orderOptions[2])
                                            "
                                            type="button"
                                            class="relative -ml-px inline-flex items-center rounded-r-md px-2 py-1 text-sm font-semibold text-gray-600 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-10"
                                            :class="{
                                                'bg-indigo-600 hover:bg-indigo-500 text-white':
                                                    element.direction == orderOptions[2],
                                            }"
                                        >
                                            {{
                                                t(
                                                    "collimato.charts.new_chart.chart_query.orders.none",
                                                )
                                            }}
                                        </button>
                                    </span>
                                </div>
                            </li>
                        </template>
                    </draggable>

                    <div class="flex border-t border-gray-100 pt-6"></div>
                </div>
            </div>

            <MeasureDialog
                :open="measureDialog"
                :data="cubes"
                @add="$emit('add-measure', $event)"
                @close="measureDialog = false"
            />

            <DimensionDialog
                :open="dimensionDialog"
                :data="cubes"
                @add="$emit('add-dimension', $event)"
                @close="dimensionDialog = false"
            />
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref } from "vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ChevronUpIcon } from "@heroicons/vue/20/solid";
import { XCircleIcon } from "@heroicons/vue/24/outline";
import draggable from "vuedraggable";
import { ORDER_OPTIONS } from "@/utils/collimato/chartQuery.js";
import MeasureDialog from "@/components/Collimato/Dialogs/MeasureDialog.vue";
import DimensionDialog from "@/components/Collimato/Dialogs/DimensionDialog.vue";

defineProps({
    cubes: { type: Array, default: () => [] },
    measures: { type: Array, default: () => [] },
    dimensions: { type: Array, default: () => [] },
    orders: { type: Array, default: () => [] },
});

defineEmits([
    "add-measure",
    "remove-measure",
    "add-dimension",
    "remove-dimension",
    "set-order",
    "reorder",
]);

const orderOptions = ORDER_OPTIONS;
const drag = ref(false);
const measureDialog = ref(false);
const dimensionDialog = ref(false);
</script>
