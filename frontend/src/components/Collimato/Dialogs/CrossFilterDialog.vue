<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close()">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
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
                            class="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-xl sm:p-6"
                        >
                            <div class="absolute right-0 top-0 hidden pr-4 pt-4 sm:block"></div>
                            <div class="sm:flex sm:items-start">
                                <div class="mt-3 w-full text-center sm:ml-4 sm:mt-0 sm:text-left">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                        >{{ t("collimato.dashboard.filter.title") }}</DialogTitle
                                    >
                                    <div class="mt-2 px-2 w-full">
                                        <div>
                                            <label
                                                for="name"
                                                class="block text-sm/6 font-medium text-gray-900"
                                                >{{ t("collimato.dashboard.filter.name") }}</label
                                            >
                                            <div class="mt-2">
                                                <input
                                                    v-model="name"
                                                    type="text"
                                                    name="name"
                                                    id="name"
                                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm/6"
                                                    placeholder="filter name"
                                                    :class="
                                                        v$.name.$error
                                                            ? 'ring-red-600 focus:ring-red-600'
                                                            : 'ring-gray-300 focus:ring-indigo-600'
                                                    "
                                                />
                                                <p
                                                    v-if="v$.name.$error"
                                                    class="mt-2 text-sm text-red-600"
                                                >
                                                    This field is required
                                                </p>
                                            </div>
                                        </div>

                                        <FilterFieldSelect
                                            v-model="selected"
                                            :options="dataModels"
                                            :label="t('collimato.dashboard.filter.data_model')"
                                            :error="v$.selected.$error"
                                        />

                                        <FilterFieldSelect
                                            :model-value="selectedColumn"
                                            @update:model-value="selectColumn"
                                            :options="selected?.dimensions ?? []"
                                            :label="t('collimato.dashboard.filter.column')"
                                            :error="v$.selectedColumn.$error"
                                        />

                                        <FilterFieldSelect
                                            :model-value="selectedOperator"
                                            @update:model-value="selectOperator"
                                            :options="computedOperators"
                                            :label="t('collimato.dashboard.filter.operator')"
                                            :error="v$.selectedOperator.$error"
                                        />

                                        <FilterValuePicker
                                            v-if="
                                                selectedOperator &&
                                                [
                                                    'equals',
                                                    'notEquals',
                                                    'contains',
                                                    'notContains',
                                                    'gt',
                                                    'gte',
                                                    'lt',
                                                    'lte',
                                                ].includes(selectedOperator.value)
                                            "
                                            v-model="selectedValues"
                                            :values="values"
                                            :error="v$.selectedValues.$error"
                                        />

                                        <!-- inDateRange and notInDateRange -->
                                        <div
                                            v-if="
                                                selectedOperator &&
                                                ['inDateRange', 'notInDateRange'].includes(
                                                    selectedOperator.value,
                                                )
                                            "
                                        >
                                            <div class="flex flex-col w-full gap-y-2 mt-2">
                                                <DateField
                                                    v-model="start_time"
                                                    :label="
                                                        t('collimato.dashboard.filter.start_date')
                                                    "
                                                    :title="
                                                        t(
                                                            'collimato.dashboard.filter.start_date_title',
                                                        )
                                                    "
                                                    :error="v$.start_time.$error"
                                                />

                                                <DateField
                                                    v-model="end_time"
                                                    :label="
                                                        t('collimato.dashboard.filter.end_date')
                                                    "
                                                    :title="
                                                        t(
                                                            'collimato.dashboard.filter.end_date_title',
                                                        )
                                                    "
                                                    :error="v$.end_time.$error"
                                                />
                                            </div>
                                        </div>

                                        <!-- Before and After date -->
                                        <div
                                            v-if="
                                                selectedOperator &&
                                                ['beforeDate', 'afterDate'].includes(
                                                    selectedOperator.value,
                                                )
                                            "
                                            class="mt-2"
                                        >
                                            <DateField
                                                v-model="date_time"
                                                :label="t('collimato.dashboard.filter.date')"
                                                :title="t('collimato.dashboard.filter.date_title')"
                                                :error="v$.date_time.$error"
                                            />
                                        </div>

                                        <FilterScope
                                            v-model:scope="selectedOption"
                                            v-model:charts="selectedCharts"
                                            :options="options"
                                            :available-charts="filteredCharts"
                                        />
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                    @click="add"
                                >
                                    {{
                                        editFilter
                                            ? t("common.button.save")
                                            : t("common.button.add")
                                    }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="close()"
                                >
                                    {{ t("common.button.cancel") }}
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
import { ref, watch, computed } from "vue";
import { t } from "@/i18n/index.js";
import { useRoute } from "vue-router";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import collimatoService from "@/services/collimatoService";
import FilterFieldSelect from "@/components/Collimato/Dialogs/CrossFilter/FilterFieldSelect.vue";
import DateField from "@/components/Collimato/Dialogs/CrossFilter/DateField.vue";
import FilterScope from "@/components/Collimato/Dialogs/CrossFilter/FilterScope.vue";
import FilterValuePicker from "@/components/Collimato/Dialogs/CrossFilter/FilterValuePicker.vue";
import { pollQuery } from "@/utils/collimato/queryPolling.js";

const filteredCharts = computed(() => {
    //if chart type is not map
    return props.charts.filter((chart) => chart.chart_type != "map");
});

const computedOperators = computed(() => {
    if (selectedColumn.value) {
        if (selectedColumn.value.type == "time") {
            return operators.filter(
                (operator) =>
                    operator.value == "inDateRange" ||
                    operator.value == "notInDateRange" ||
                    operator.value == "beforeDate" ||
                    operator.value == "afterDate",
            );
        }

        return operators.filter(
            (operator) =>
                operator.value != "inDateRange" &&
                operator.value != "notInDateRange" &&
                operator.value != "beforeDate" &&
                operator.value != "afterDate",
        );
    }

    return [];
});

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    dataModels: {
        type: Array,
        default: () => [],
    },
    charts: {
        type: Array,
        default: () => [],
    },
    editFilter: {
        type: Object,
        required: false,
        default: () => null,
    },
});

const emits = defineEmits(["update:modelValue", "add-filter", "update-filter", "close"]);

const operators = [
    { name: "is set", value: "set" },
    { name: "is not set", value: "notSet" },
    { name: "equals", value: "equals" },
    { name: "not equals", value: "notEquals" },
    { name: "contains", value: "contains" },
    { name: "not contains", value: "notContains" },
    { name: ">", value: "gt" },
    { name: ">=", value: "gte" },
    { name: "<", value: "lt" },
    { name: "<=", value: "lte" },
    { name: "date range", value: "inDateRange" },
    { name: "not in date range", value: "notInDateRange" },
    { name: "before date", value: "beforeDate" },
    { name: "after date", value: "afterDate" },
];

const route = useRoute();
const selected = ref(null);
const selectedColumn = ref(null);
const selectedValues = ref([]);
const selectedOperator = ref(null);
const values = ref([]);
const name = ref("");

const date_time = ref(null);
const start_time = ref(null);
const end_time = ref(null);

const options = computed(() => [
    { id: "all", title: t.value("collimato.dashboard.filter.scope_all") },
    {
        id: "specific",
        title: t.value("collimato.dashboard.filter.scope_specific"),
    },
]);
const selectedOption = ref(options.value[0]);
const selectedCharts = ref([]);

const rules = {
    name: { required },
    selected: { required },
    selectedColumn: { required },
    selectedOperator: { required },
    selectedValues: { required },
    start_time: { required },
    end_time: { required },
    date_time: { required },
};
const v$ = useVuelidate(rules, {
    name,
    selected,
    selectedColumn,
    selectedOperator,
    selectedValues,
    start_time,
    end_time,
    date_time,
});

async function queryData() {
    if (selectedColumn.value.type == "time") {
        return;
    }

    const query = {
        measures: [selected.value.name + ".count"],
        dimensions: [selectedColumn.value.name],
        filters: [],
        timeDimensions: [],
        order: [],
        limit: 1000,
    };

    try {
        const result = await pollQuery(
            (q) => collimatoService.loadData(route.params.workspaceId, q),
            query,
        );

        values.value = result.data.data.map((row) => row[selectedColumn.value.name]);
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

async function add() {
    // Get the selected operator value
    // Touch common required fields

    v$.value.name.$touch();
    v$.value.selected.$touch();
    v$.value.selectedColumn.$touch();
    v$.value.selectedOperator.$touch();

    if (
        v$.value.name.$error ||
        v$.value.selected.$error ||
        v$.value.selectedColumn.$error ||
        v$.value.selectedOperator.$error
    ) {
        return;
    }

    if (["inDateRange", "notInDateRange"].includes(selectedOperator.value.value)) {
        v$.value.start_time.$touch();
        v$.value.end_time.$touch();

        if (v$.value.start_time.$error || v$.value.end_time.$error) {
            return;
        }
    }

    if (["beforeDate", "afterDate"].includes(selectedOperator.value.value)) {
        v$.value.date_time.$touch();

        if (v$.value.date_time.$error) {
            return;
        }
    }

    if (
        !["set", "notSet", "beforeDate", "afterDate", "inDateRange", "notInDateRange"].includes(
            selectedOperator.value.value,
        )
    ) {
        v$.value.selectedValues.$touch();

        if (v$.value.selectedValues.$error) {
            return;
        }
    }

    let stringValues = [];

    selectedValues.value.forEach((value) => {
        if (typeof value === "number") {
            stringValues.push(String(value));
        } else {
            stringValues.push(value);
        }
    });

    let chartsToApply = selectedCharts.value;

    if (selectedOption.value === "all") {
        chartsToApply = props.charts;
    }

    if (
        selectedOperator.value.value === "inDateRange" ||
        selectedOperator.value.value === "notInDateRange"
    ) {
        stringValues = [start_time.value, end_time.value];
    } else if (
        selectedOperator.value.value === "beforeDate" ||
        selectedOperator.value.value === "afterDate"
    ) {
        stringValues = [date_time.value];
    }

    if (["set", "notSet"].includes(selectedOperator.value.value)) {
        stringValues = [];
    }

    let filter = {
        name: name.value,
        table: selected.value.name,
        column: selectedColumn.value.name,
        operator: selectedOperator.value.value,
        values: stringValues,
        apply_to: chartsToApply.map((chart) => chart.id),
    };

    if (props.editFilter) {
        emits("update-filter", { id: props.editFilter.id, filter: filter });

        return;
    }

    emits("add-filter", filter);
}

function selectColumn(column) {
    selectedColumn.value = column;
    onColumnSelect(column);
    queryData();
}

function selectOperator(operator) {
    selectedOperator.value = operator;
    clearValues(operator);
}

function clearValues(o) {
    if (
        o.value === "inDateRange" ||
        o.value === "notInDateRange" ||
        o.value === "beforeDate" ||
        o.value === "afterDate"
    ) {
        selectedValues.value = [];
    }
}

function onColumnSelect(column) {
    if (column.type == "time") {
        selectedValues.value = [];
    }
}

function reset() {
    name.value = "";
    selected.value = null;
    selectedColumn.value = null;
    selectedOperator.value = null;
    selectedValues.value = [];
    selectedCharts.value = [];
    selectedOption.value = options.value[0];
    values.value = [];
    start_time.value = null;
    end_time.value = null;
    date_time.value = null;

    v$.value.$reset();
}

function close() {
    emits("close");
}

watch(
    () => props.modelValue,
    (open) => {
        if (!open) {
            setTimeout(reset, 300);
        }
    },
);

//watch editFilter
watch(
    () => props.editFilter,
    (newValue) => {
        if (newValue) {
            selected.value = props.dataModels.find((model) => model.name === newValue.table);

            selectedOperator.value = operators.find(
                (operator) => operator.value === newValue.operator,
            );

            selectedColumn.value = selected.value.dimensions.find(
                (column) => column.name === newValue.column,
            );

            if (newValue.operator === "afterDate" || newValue.operator === "beforeDate") {
                date_time.value = newValue.values[0];
            } else if (
                newValue.operator === "inDateRange" ||
                newValue.operator === "notInDateRange"
            ) {
                start_time.value = newValue.values[0];
                end_time.value = newValue.values[1];
            } else {
                queryData().then(() => {
                    selectedValues.value = newValue.values;
                });
            }

            name.value = newValue.name;
            selectedCharts.value = props.charts.filter((chart) =>
                newValue.apply_to.includes(chart.id),
            );

            if (selectedCharts.value.length > 0) {
                selectedOption.value = options.value[1];
            }
        }
    },
);
</script>
