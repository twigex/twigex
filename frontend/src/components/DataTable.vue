<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="space-y-4">
        <div v-if="$slots.toolbar">
            <slot name="toolbar" />
        </div>

        <div class="overflow-hidden rounded-lg bg-white shadow-sm ring-1 ring-gray-900/5">
            <div class="overflow-x-auto">
                <table class="min-w-full divide-y divide-gray-200">
                    <thead class="bg-gray-50">
                        <tr>
                            <th
                                v-for="(column, index) in columns"
                                :key="column.key"
                                scope="col"
                                :aria-sort="ariaSort(column)"
                                :class="[
                                    cellPadding(index),
                                    alignClass(column),
                                    hiddenClass(column),
                                    'whitespace-nowrap py-3.5 text-sm font-semibold text-gray-900',
                                ]"
                            >
                                <span v-if="column.srOnly" class="sr-only">{{ column.label }}</span>
                                <button
                                    v-else-if="column.sortable"
                                    type="button"
                                    class="group inline-flex items-center"
                                    @click="toggleSort(column)"
                                >
                                    {{ column.label }}
                                    <span
                                        :class="[
                                            activeSort.key === column.key
                                                ? 'bg-gray-200 text-gray-900 group-hover:bg-gray-300'
                                                : 'invisible text-gray-400 group-hover:visible group-focus:visible',
                                            'ml-2 flex-none rounded-sm',
                                        ]"
                                    >
                                        <ChevronUpIcon
                                            v-if="activeSort.key === column.key && !activeSort.desc"
                                            class="size-5"
                                            aria-hidden="true"
                                        />
                                        <ChevronDownIcon v-else class="size-5" aria-hidden="true" />
                                    </span>
                                </button>
                                <template v-else>{{ column.label }}</template>
                            </th>
                        </tr>
                    </thead>

                    <tbody class="divide-y divide-gray-200">
                        <tr v-if="loading && rows.length === 0">
                            <td :colspan="columns.length" class="py-10">
                                <div class="flex justify-center">
                                    <BaseSpinner />
                                </div>
                            </td>
                        </tr>
                        <tr v-else-if="rows.length === 0">
                            <td
                                :colspan="columns.length"
                                class="px-6 py-10 text-center text-sm text-gray-500"
                            >
                                <slot name="empty">{{ t("common.label.no_results") }}</slot>
                            </td>
                        </tr>
                        <tr
                            v-for="item in rows"
                            :key="item[itemKey]"
                            :class="clickable ? 'cursor-pointer hover:bg-gray-50' : ''"
                            @click="onRowClick(item)"
                        >
                            <td
                                v-for="(column, index) in columns"
                                :key="column.key"
                                :class="[
                                    cellPadding(index),
                                    alignClass(column),
                                    hiddenClass(column),
                                    'whitespace-nowrap py-4 text-sm text-gray-500',
                                ]"
                            >
                                <slot :name="column.key" :item="item">
                                    {{ valueOf(item, column.key) }}
                                </slot>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <BasePagination
            v-if="totalCount > itemsPerPage"
            :currentPage="page"
            :totalItems="totalCount"
            :itemsPerPage="itemsPerPage"
            @update:currentPage="changePage"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, ref } from "vue";
import BasePagination from "@/components/BasePagination.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import { ChevronDownIcon, ChevronUpIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    columns: {
        type: Array,
        required: true,
    },
    items: {
        type: Array,
        required: true,
    },
    itemKey: {
        type: String,
        default: "id",
    },
    sort: {
        type: Object,
        default: null,
    },
    loading: {
        type: Boolean,
        default: false,
    },
    clickable: {
        type: Boolean,
        default: false,
    },
    itemsPerPage: {
        type: Number,
        default: 20,
    },
    // Set to page and sort on the server: items is then the current page only.
    totalItems: {
        type: Number,
        default: null,
    },
    currentPage: {
        type: Number,
        default: null,
    },
});

const emit = defineEmits(["row-click", "update:sort", "update:currentPage"]);

const HIDDEN_BELOW = {
    sm: "hidden sm:table-cell",
    md: "hidden md:table-cell",
    lg: "hidden lg:table-cell",
};

const internalSort = ref({ key: "", desc: false });
const internalPage = ref(1);

const serverSide = computed(() => props.totalItems !== null);
const activeSort = computed(() => props.sort ?? internalSort.value);
const totalCount = computed(() => (serverSide.value ? props.totalItems : props.items.length));

const page = computed(() => {
    const requested = props.currentPage ?? internalPage.value;

    if (serverSide.value) return requested;

    const lastPage = Math.max(1, Math.ceil(totalCount.value / props.itemsPerPage));

    return Math.min(requested, lastPage);
});

const sortedItems = computed(() => {
    const { key, desc } = activeSort.value;

    if (serverSide.value || !key) return props.items;

    const column = props.columns.find((c) => c.key === key);
    const path = column?.sortKey ?? key;
    const direction = desc ? -1 : 1;

    return [...props.items].sort((a, b) => direction * compare(valueOf(a, path), valueOf(b, path)));
});

const rows = computed(() => {
    if (serverSide.value) return props.items;

    const start = (page.value - 1) * props.itemsPerPage;

    return sortedItems.value.slice(start, start + props.itemsPerPage);
});

function compare(a, b) {
    if (typeof a === "number" && typeof b === "number") return a - b;

    return String(a ?? "").localeCompare(String(b ?? ""), undefined, {
        numeric: true,
        sensitivity: "base",
    });
}

function toggleSort(column) {
    const current = activeSort.value;
    const next =
        current.key === column.key
            ? { key: column.key, desc: !current.desc }
            : { key: column.key, desc: false };

    internalSort.value = next;
    internalPage.value = 1;

    emit("update:sort", next);

    if (!serverSide.value) emit("update:currentPage", 1);
}

function ariaSort(column) {
    if (!column.sortable || activeSort.value.key !== column.key) return undefined;

    return activeSort.value.desc ? "descending" : "ascending";
}

function changePage(newPage) {
    internalPage.value = newPage;
    emit("update:currentPage", newPage);
}

function onRowClick(item) {
    if (props.clickable) emit("row-click", item);
}

function cellPadding(index) {
    if (index === 0) return "pl-4 pr-3 sm:pl-6";
    if (index === props.columns.length - 1) return "pl-3 pr-4 sm:pr-6";

    return "px-3";
}

function alignClass(column) {
    return column.align === "right" ? "text-right" : "text-left";
}

function hiddenClass(column) {
    return HIDDEN_BELOW[column.hiddenBelow] ?? "";
}

function valueOf(item, path) {
    return path.split(".").reduce((o, p) => (o ? o[p] : undefined), item);
}
</script>
