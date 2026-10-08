<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="filterDialog">
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
                            class="relative transform rounded-lg bg-white shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-md overflow-hidden"
                        >
                            <!-- Header -->
                            <div
                                class="flex items-center justify-between px-6 py-4 border-b border-gray-200"
                            >
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{
                                        t("collimato.charts.new_chart.filters.filter_dialog.title")
                                    }}
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
                                        v-model="query"
                                        type="text"
                                        class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        :placeholder="t('common.placeholder.search')"
                                    />
                                </div>
                            </div>

                            <!-- Member list -->
                            <div class="max-h-56 overflow-y-auto px-2 pb-2">
                                <template v-for="group in filteredOptions" :key="group.table">
                                    <p
                                        class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                    >
                                        {{ group.table }}
                                    </p>
                                    <button
                                        v-for="item in [...group.measures, ...group.dimensions]"
                                        :key="item.name"
                                        type="button"
                                        @click="
                                            selectedFilter.member = item;
                                            selectedFilter.operator = null;
                                            selectedFilter.value = [];
                                        "
                                        class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                        :class="
                                            selectedFilter.member?.name === item.name
                                                ? 'bg-indigo-50 text-indigo-700'
                                                : 'text-gray-700 hover:bg-gray-50'
                                        "
                                    >
                                        <span>{{ item.title }}</span>
                                        <CheckIcon
                                            v-if="selectedFilter.member?.name === item.name"
                                            class="h-4 w-4 shrink-0 text-indigo-600"
                                        />
                                    </button>
                                </template>
                                <p
                                    v-if="filteredOptions.length === 0"
                                    class="py-6 text-center text-sm text-gray-500"
                                >
                                    {{ t("common.label.no_results") }}
                                </p>
                            </div>

                            <!-- Operator + value (only when member selected) -->
                            <div
                                v-if="selectedFilter.member"
                                class="border-t border-gray-200 px-4 pt-3 pb-2 space-y-3"
                            >
                                <!-- Operator list -->
                                <div class="max-h-40 overflow-y-auto px-0">
                                    <p
                                        class="px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                    >
                                        {{
                                            t(
                                                "collimato.charts.new_chart.filters.filter_dialog.select_operator",
                                            )
                                        }}
                                    </p>
                                    <button
                                        v-for="o in computedOperators"
                                        :key="o.value"
                                        type="button"
                                        @click="
                                            selectedFilter.operator = o;
                                            selectedFilter.value = [];
                                        "
                                        class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                        :class="
                                            selectedFilter.operator?.value === o.value
                                                ? 'bg-indigo-50 text-indigo-700'
                                                : 'text-gray-700 hover:bg-gray-50'
                                        "
                                    >
                                        <span>{{ o.name }}</span>
                                        <CheckIcon
                                            v-if="selectedFilter.operator?.value === o.value"
                                            class="h-4 w-4 shrink-0 text-indigo-600"
                                        />
                                    </button>
                                </div>

                                <!-- Date range -->
                                <div v-if="isDateRange" class="grid grid-cols-2 gap-3">
                                    <div>
                                        <label
                                            class="block text-xs font-medium text-gray-500 mb-1"
                                            >{{
                                                t(
                                                    "collimato.charts.new_chart.filters.filter_dialog.select_start_date",
                                                )
                                            }}</label
                                        >
                                        <DatePicker v-model="start_time" />
                                    </div>
                                    <div>
                                        <label
                                            class="block text-xs font-medium text-gray-500 mb-1"
                                            >{{
                                                t(
                                                    "collimato.charts.new_chart.filters.filter_dialog.select_end_date",
                                                )
                                            }}</label
                                        >
                                        <DatePicker v-model="end_time" />
                                    </div>
                                </div>

                                <!-- Single date -->
                                <div v-else-if="isSingleDate">
                                    <label class="block text-xs font-medium text-gray-500 mb-1">{{
                                        t(
                                            "collimato.charts.new_chart.filters.filter_dialog.select_date",
                                        )
                                    }}</label>
                                    <DatePicker v-model="date_time" />
                                </div>

                                <!-- Multi-value input -->
                                <div v-else-if="showValueInput">
                                    <p
                                        class="px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                    >
                                        {{
                                            t(
                                                "collimato.charts.new_chart.filters.filter_dialog.select_values",
                                            )
                                        }}
                                    </p>
                                    <div class="max-h-32 overflow-y-auto">
                                        <div
                                            v-for="(val, idx) in selectedFilter.value"
                                            :key="idx"
                                            class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-gray-700 hover:bg-gray-50"
                                        >
                                            <span class="truncate">{{ val }}</span>
                                            <button
                                                type="button"
                                                @click="removeValue(val)"
                                                class="ml-2 shrink-0 text-gray-400 hover:text-red-500"
                                            >
                                                <XMarkIcon class="h-4 w-4" />
                                            </button>
                                        </div>
                                    </div>
                                    <div class="relative mt-1">
                                        <input
                                            ref="valueInput"
                                            v-model="comboQuery"
                                            type="text"
                                            @keydown.enter.prevent="addQueryToSelected"
                                            class="block w-full rounded-md border-0 py-1.5 pl-3 pr-10 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                            :placeholder="
                                                t(
                                                    'collimato.charts.new_chart.filters.filter_dialog.select_values',
                                                )
                                            "
                                        />
                                        <button
                                            type="button"
                                            @click="addQueryToSelected"
                                            class="absolute inset-y-0 right-0 flex items-center pr-2 text-gray-400 hover:text-indigo-600"
                                        >
                                            <PlusIcon class="h-4 w-4" />
                                        </button>
                                    </div>
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
                                    :disabled="disableSave()"
                                    @click="addFilter"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white transition-colors"
                                    :class="
                                        disableSave()
                                            ? 'bg-gray-300 cursor-not-allowed'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                >
                                    {{
                                        t(
                                            "collimato.charts.new_chart.filters.filter_dialog.save_filter",
                                        )
                                    }}
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
import { t } from "@/i18n/index.js";
import { ref, reactive, computed, watch } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, XMarkIcon, MagnifyingGlassIcon, PlusIcon } from "@heroicons/vue/20/solid";
import DatePicker from "@/components/DatePicker/DatePicker.vue";

const props = defineProps({
    filterDialog: { type: Boolean, required: true },
    options: { type: Array, required: true },
    editFilter: { type: Object, default: null },
});

const emit = defineEmits(["close", "add-filter", "update-filter"]);

const query = ref("");
const comboQuery = ref("");
const searchInput = ref(null);
const date_time = ref(null);
const start_time = ref(null);
const end_time = ref(null);

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
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.in_date_range"),
        value: "inDateRange",
    },
    {
        name: t.value(
            "collimato.charts.new_chart.filters.filter_dialog.operators.not_in_date_range",
        ),
        value: "notInDateRange",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.before_date"),
        value: "beforeDate",
    },
    {
        name: t.value("collimato.charts.new_chart.filters.filter_dialog.operators.after_date"),
        value: "afterDate",
    },
];

const selectedFilter = reactive({ member: null, operator: null, value: [] });

const filteredOptions = computed(() => {
    const q = query.value.toLowerCase().replace(/\s+/g, "");

    return props.options
        .map((g) => ({
            ...g,
            measures:
                q === ""
                    ? g.measures
                    : g.measures.filter((m) =>
                          m.title.toLowerCase().replace(/\s+/g, "").includes(q),
                      ),
            dimensions:
                q === ""
                    ? g.dimensions
                    : g.dimensions.filter((d) =>
                          d.title.toLowerCase().replace(/\s+/g, "").includes(q),
                      ),
        }))
        .filter((g) => g.measures.length > 0 || g.dimensions.length > 0);
});

const computedOperators = computed(() => {
    if (!selectedFilter.member) return operators;
    const dateOps = ["inDateRange", "notInDateRange", "beforeDate", "afterDate"];

    return selectedFilter.member.type === "time"
        ? operators.filter((o) => dateOps.includes(o.value))
        : operators.filter((o) => !dateOps.includes(o.value));
});

const isDateRange = computed(
    () =>
        selectedFilter.operator &&
        ["inDateRange", "notInDateRange"].includes(selectedFilter.operator.value),
);
const isSingleDate = computed(
    () =>
        selectedFilter.operator &&
        ["beforeDate", "afterDate"].includes(selectedFilter.operator.value),
);
const showValueInput = computed(
    () =>
        selectedFilter.operator &&
        !isDateRange.value &&
        !isSingleDate.value &&
        !["set", "notSet"].includes(selectedFilter.operator.value),
);

function addQueryToSelected() {
    if (comboQuery.value.trim() && !selectedFilter.value.includes(comboQuery.value.trim())) {
        selectedFilter.value.push(comboQuery.value.trim());
    }

    comboQuery.value = "";
}

function removeValue(val) {
    selectedFilter.value = selectedFilter.value.filter((v) => v !== val);
}

function disableSave() {
    if (!selectedFilter.member || !selectedFilter.operator) return true;
    if (isDateRange.value) return !start_time.value || !end_time.value;
    if (isSingleDate.value) return !date_time.value;

    return false;
}

function addFilter() {
    let value = selectedFilter.value;

    if (isDateRange.value) value = [start_time.value, end_time.value];
    else if (isSingleDate.value) value = [date_time.value];

    if (props.editFilter) {
        emit("update-filter", {
            member: selectedFilter.member.name,
            operator: selectedFilter.operator.value,
            value,
        });
    } else {
        emit("add-filter", {
            member: selectedFilter.member.name,
            operator: selectedFilter.operator.value,
            value,
        });
    }

    closeDialog();
}

function closeDialog() {
    emit("close");
    selectedFilter.member = null;
    selectedFilter.operator = null;
    selectedFilter.value = [];
    date_time.value = null;
    start_time.value = null;
    end_time.value = null;
    query.value = "";
    comboQuery.value = "";
}

function allItems() {
    return props.options.flatMap((g) => [...g.dimensions, ...g.measures]);
}

watch(
    () => props.editFilter,
    (value) => {
        if (value) {
            selectedFilter.member = allItems().find((o) => o.name === value.member) ?? null;
            selectedFilter.operator = operators.find((o) => o.value === value.operator) ?? null;
            selectedFilter.value = value.value ?? [];
        }
    },
);
</script>
