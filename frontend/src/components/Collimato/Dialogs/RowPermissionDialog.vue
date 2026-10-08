<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="closeDialog" :initialFocus="searchInput">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-md"
                        >
                            <!-- Header -->
                            <div
                                class="flex items-center justify-between px-6 py-4 border-b border-gray-200"
                            >
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{ t("collimato.roles.row_permission_dialog.title") }}
                                </DialogTitle>
                                <button
                                    @click="closeDialog"
                                    type="button"
                                    class="rounded-md text-gray-400 hover:text-gray-500"
                                >
                                    <XMarkIcon class="h-5 w-5" />
                                </button>
                            </div>

                            <!-- Search -->
                            <div class="px-4 pt-3 pb-2">
                                <div class="relative">
                                    <MagnifyingGlassIcon
                                        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                    />
                                    <input
                                        ref="searchInput"
                                        v-model="tableSearch"
                                        type="text"
                                        class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        :placeholder="
                                            t(
                                                'collimato.roles.row_permission_dialog.search_placeholder',
                                            )
                                        "
                                    />
                                </div>
                            </div>

                            <!-- Table list -->
                            <div class="max-h-52 overflow-y-auto px-2 pb-2">
                                <p
                                    class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                >
                                    {{ t("collimato.roles.row_permission_dialog.data_model") }}
                                </p>
                                <button
                                    v-for="table in filteredTables"
                                    :key="table.name"
                                    type="button"
                                    @click="selectTable(table)"
                                    class="flex w-full items-center gap-x-2 rounded-md px-3 py-2 text-sm text-left transition-colors"
                                    :class="
                                        selectedFilter.table?.name === table.name
                                            ? 'bg-indigo-50 text-indigo-700'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    "
                                >
                                    <CircleStackIcon
                                        class="h-4 w-4 shrink-0 opacity-60"
                                        aria-hidden="true"
                                    />
                                    <span class="flex-1 truncate">{{ table.name }}</span>
                                    <CheckIcon
                                        v-if="selectedFilter.table?.name === table.name"
                                        class="h-4 w-4 shrink-0 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </button>
                                <p
                                    v-if="filteredTables.length === 0"
                                    class="py-6 text-center text-sm text-gray-500"
                                >
                                    {{ t("common.label.no_results") }}
                                </p>
                            </div>

                            <!-- Member list (once table selected) -->
                            <div
                                v-if="selectedFilter.table"
                                class="border-t border-gray-200 max-h-40 overflow-y-auto px-2 pb-2"
                            >
                                <p
                                    class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                >
                                    {{ t("collimato.roles.row_permission_dialog.member") }}
                                </p>
                                <button
                                    v-for="col in modelMembers"
                                    :key="col.name"
                                    type="button"
                                    @click="selectMember(col)"
                                    class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                    :class="
                                        selectedFilter.member?.name === col.name
                                            ? 'bg-indigo-50 text-indigo-700'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    "
                                >
                                    <span>{{ col.name }}</span>
                                    <CheckIcon
                                        v-if="selectedFilter.member?.name === col.name"
                                        class="h-4 w-4 shrink-0 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </button>
                            </div>

                            <!-- Operator list (once member selected) -->
                            <div
                                v-if="selectedFilter.member"
                                class="border-t border-gray-200 max-h-40 overflow-y-auto px-2 pb-2"
                            >
                                <p
                                    class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                >
                                    {{ t("collimato.roles.row_permission_dialog.operator") }}
                                </p>
                                <button
                                    v-for="op in operators"
                                    :key="op.value"
                                    type="button"
                                    @click="selectOperator(op)"
                                    class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                    :class="
                                        selectedFilter.operator?.value === op.value
                                            ? 'bg-indigo-50 text-indigo-700'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    "
                                >
                                    <span>{{ op.name }}</span>
                                    <CheckIcon
                                        v-if="selectedFilter.operator?.value === op.value"
                                        class="h-4 w-4 shrink-0 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </button>
                            </div>

                            <!-- Values (once operator selected and operator needs values) -->
                            <div v-if="showValues" class="border-t border-gray-200 px-4 pt-3 pb-2">
                                <div class="flex items-center justify-between mb-2">
                                    <p
                                        class="text-xs font-semibold uppercase tracking-wider text-gray-400"
                                    >
                                        {{ t("collimato.roles.row_permission_dialog.values") }}
                                    </p>
                                    <button
                                        type="button"
                                        @click="addValue"
                                        class="text-xs font-medium text-indigo-600 hover:text-indigo-500"
                                    >
                                        + {{ t("collimato.roles.row_permission_dialog.add_value") }}
                                    </button>
                                </div>
                                <div class="max-h-32 overflow-y-auto space-y-1.5">
                                    <div
                                        v-for="(value, index) in selectedFilter.values"
                                        :key="index"
                                        class="flex rounded-md shadow-sm"
                                    >
                                        <input
                                            v-model="selectedFilter.values[index]"
                                            type="text"
                                            class="block w-full rounded-none rounded-l-md border-0 py-1.5 px-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                            placeholder="Enter value"
                                        />
                                        <button
                                            type="button"
                                            @click="selectedFilter.values.splice(index, 1)"
                                            class="relative -ml-px inline-flex items-center rounded-r-md px-2.5 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 hover:text-red-500"
                                        >
                                            <XMarkIcon class="h-4 w-4" aria-hidden="true" />
                                        </button>
                                    </div>
                                    <p
                                        v-if="selectedFilter.values.length === 0"
                                        class="py-2 text-center text-xs text-gray-400"
                                    >
                                        No values added yet.
                                    </p>
                                </div>
                            </div>

                            <!-- Footer -->
                            <div class="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
                                <button
                                    type="button"
                                    @click="closeDialog"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    type="button"
                                    @click="addFilter"
                                    :disabled="!canSave"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white transition-colors"
                                    :class="
                                        !canSave
                                            ? 'bg-gray-300 cursor-not-allowed'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                >
                                    {{ t("common.button.save") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { reactive, computed, ref } from "vue";
import { t } from "@/i18n/index.js";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, XMarkIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { CircleStackIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    dataModels: { type: Array, required: true },
});

const emit = defineEmits(["close", "add-filter"]);

const operators = [
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.set"),
        value: "set",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.not_set"),
        value: "notSet",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.equals"),
        value: "equals",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.not_equals"),
        value: "notEquals",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.contains"),
        value: "contains",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.not_contains"),
        value: "notContains",
    },
    { name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.gt"), value: "gt" },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.gte"),
        value: "gte",
    },
    { name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.lt"), value: "lt" },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.lte"),
        value: "lte",
    },
];

const searchInput = ref(null);
const tableSearch = ref("");

const selectedFilter = reactive({
    table: null,
    member: null,
    operator: null,
    values: [],
});

const filteredTables = computed(() => {
    const q = tableSearch.value.trim().toLowerCase();

    if (!q) return props.dataModels;

    return props.dataModels.filter((m) => m.name.toLowerCase().includes(q));
});

const modelMembers = computed(() => {
    if (!selectedFilter.table) return [];

    return (selectedFilter.table.dimensions ?? []).concat(selectedFilter.table.measures ?? []);
});

const showValues = computed(
    () =>
        selectedFilter.operator !== null &&
        !["set", "notSet"].includes(selectedFilter.operator.value),
);

const canSave = computed(
    () =>
        selectedFilter.member !== null &&
        selectedFilter.operator !== null &&
        (["set", "notSet"].includes(selectedFilter.operator?.value) ||
            selectedFilter.values.length > 0),
);

function selectTable(table) {
    selectedFilter.table = table;
    selectedFilter.member = null;
    selectedFilter.operator = null;
    selectedFilter.values = [];
}

function selectMember(col) {
    selectedFilter.member = col;
    selectedFilter.operator = null;
    selectedFilter.values = [];
}

function selectOperator(op) {
    selectedFilter.operator = op;
    selectedFilter.values = [];
}

function addValue() {
    const isNumber = selectedFilter.member?.type === "number";

    selectedFilter.values.push(isNumber ? 0 : "");
}

function closeDialog() {
    selectedFilter.table = null;
    selectedFilter.member = null;
    selectedFilter.operator = null;
    selectedFilter.values = [];
    tableSearch.value = "";
    emit("close");
}

function addFilter() {
    if (!canSave.value) return;
    emit("add-filter", {
        table: selectedFilter.table.name,
        member: selectedFilter.member.name,
        operator: selectedFilter.operator.value,
        values: [...selectedFilter.values],
    });
    closeDialog();
}
</script>
