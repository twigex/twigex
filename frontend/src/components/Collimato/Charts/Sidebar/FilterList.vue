<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div class="flex flex-row justify-between items-center">
            <h4 v-if="title" class="text-small font-semibold leading-4 text-gray-400">
                {{ title }}
            </h4>
            <div v-else></div>

            <div class="flex border-gray-100 pt-2">
                <button
                    @click="filterDialog = true"
                    type="button"
                    class="text-sm font-semibold leading-6 text-indigo-600 hover:text-indigo-500"
                >
                    <span aria-hidden="true">+</span>
                    {{ t("collimato.charts.new_chart.filters.add_filter") }}
                </button>
            </div>
        </div>

        <ul
            role="list"
            class="mt-2 divide-y divide-gray-100 border-t border-gray-200 text-sm leading-6"
        >
            <li
                v-for="filter in filters"
                :key="filter.member"
                class="flex justify-between gap-x-6 py-2 px-2 hover:bg-gray-50 rounded cursor-pointer"
            >
                <div class="flex flex-col truncate">
                    <div class="font-medium text-gray-900">
                        <BaseTooltip position="top">
                            {{ filter.member.split(".")[1] }}
                        </BaseTooltip>
                    </div>
                    <div class="text-xs text-gray-600 truncate">
                        <BaseTooltip position="top">
                            {{ filter.member.split(".")[0] }}
                        </BaseTooltip>
                    </div>
                    <div>{{ filter.operator }} {{ filter.value }}</div>
                </div>

                <div class="flex flex-row items-center justify-center gap-x-1">
                    <button
                        @click="
                            editFilter = filter;
                            filterDialog = true;
                        "
                        type="button"
                        class="font-semibold text-gray-600 hover:text-gray-500"
                    >
                        <PencilIcon class="h-4 w-4" />
                    </button>
                    <button
                        @click="$emit('remove-filter', filter)"
                        type="button"
                        class="font-semibold text-red-600 hover:text-indigo-500"
                    >
                        <XCircleIcon class="h-4 w-4" />
                    </button>
                </div>
            </li>
        </ul>
        <p v-if="filters.length < 1" class="mt-1 text-sm leading-6 text-gray-500">
            {{ t("collimato.charts.new_chart.filters.click_add_filter") }}
        </p>

        <FilterDialog
            :filterDialog="filterDialog"
            :options="cubes"
            :edit-filter="editFilter"
            @close="((filterDialog = false), (editFilter = null))"
            @add-filter="$emit('add-filter', $event)"
            @update-filter="$emit('update-filter', editFilter.member, $event)"
        />
    </div>
</template>

<script setup>
import { ref } from "vue";
import { t } from "@/i18n/index.js";
import FilterDialog from "@/components/Collimato/Dialogs/FilterDialog.vue";
import BaseTooltip from "@/components/BaseTooltip.vue";
import { PencilIcon, XCircleIcon } from "@heroicons/vue/24/outline";

defineProps({
    filters: {
        type: Array,
        default: () => [],
    },
    cubes: {
        type: Array,
        default: () => [],
    },
    title: {
        type: String,
        default: "",
    },
});

defineEmits(["add-filter", "update-filter", "remove-filter"]);

const editFilter = ref(null);
const filterDialog = ref(false);
</script>
