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
            <span>{{ t("collimato.charts.new_chart.tabs.customize") }}</span>
            <ChevronUpIcon
                :class="open ? 'rotate-180 transform' : ''"
                class="h-5 w-5 text-purple-500"
            />
        </DisclosureButton>

        <DisclosurePanel class="px-4 pb-2 pt-4 text-sm text-gray-500">
            <div v-if="showOrientation" class="mb-5">
                <label class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.orientation")
                }}</label>
                <span class="isolate mt-2 inline-flex rounded-md shadow-sm">
                    <button
                        @click="((config.orientation = 'vertical'), update())"
                        type="button"
                        class="relative inline-flex items-center rounded-l-md px-3 py-2 text-sm font-semibold ring-1 ring-gray-300 focus:z-10"
                        :class="
                            config.orientation === 'vertical'
                                ? 'bg-indigo-600 text-white hover:bg-indigo-500'
                                : 'bg-white text-gray-900'
                        "
                    >
                        {{ t("collimato.charts.new_chart.customize.vertical") }}
                    </button>

                    <button
                        @click="((config.orientation = 'horizontal'), update())"
                        type="button"
                        class="relative -ml-px inline-flex items-center rounded-r-md px-3 py-2 text-sm font-semibold ring-1 ring-gray-300 focus:z-10"
                        :class="
                            config.orientation === 'horizontal'
                                ? 'bg-indigo-600 text-white hover:bg-indigo-500'
                                : 'bg-white text-gray-900'
                        "
                    >
                        {{ t("collimato.charts.new_chart.customize.horizontal") }}
                    </button>
                </span>
            </div>
            <BaseCheckbox
                class="mt-2 mb-5"
                :model-value="config.legend"
                :label="t('collimato.charts.new_chart.customize.legend')"
                :description="t('collimato.charts.new_chart.customize.legend_description')"
                @update:model-value="
                    config.legend = $event;
                    update();
                "
            />

            <SettingSelect
                v-if="dimensions.length"
                class="mb-5"
                v-model="config.xAxisField"
                :label="t('collimato.charts.new_chart.customize.x_axis_field')"
                :description="t('collimato.charts.new_chart.customize.x_axis_field_description')"
                :options="xAxisFields"
                @change="update()"
            />

            <SettingSelect
                class="mb-5"
                v-model="config.sortBars"
                :label="t('collimato.charts.new_chart.customize.sort_bars')"
                :options="sortBarsOptions"
                @change="update()"
            />

            <SettingSelect
                v-if="showMissingValues"
                class="mb-5"
                v-model="config.missingValues"
                :label="t('collimato.charts.new_chart.customize.missing_values')"
                :options="missingValuesOptions"
                @change="update()"
            />

            <SettingSelect
                v-if="config.legend"
                class="mb-5"
                v-model="config.legendPosition"
                :label="t('collimato.charts.new_chart.customize.legend_position')"
                :options="legendPositions"
                @change="update()"
            />

            <BaseCheckbox
                class="mt-2 mb-5"
                :model-value="config.stack"
                :label="t('collimato.charts.new_chart.customize.stack')"
                :description="t('collimato.charts.new_chart.customize.stack_description')"
                @update:model-value="
                    config.stack = $event;
                    update();
                "
            />

            <BaseCheckbox
                v-if="showBarValues"
                class="mt-2 mb-5"
                :model-value="config.showValues"
                :label="t('collimato.charts.new_chart.customize.bar_values')"
                :description="t('collimato.charts.new_chart.customize.bar_values_description')"
                @update:model-value="
                    config.showValues = $event;
                    update();
                "
            />

            <BaseCheckbox
                class="mt-2 mb-5"
                :model-value="config.dataZoom"
                :label="t('collimato.charts.new_chart.customize.data_zoom')"
                :description="t('collimato.charts.new_chart.customize.data_zoom_description')"
                @update:model-value="
                    config.dataZoom = $event;
                    update();
                "
            />

            <div class="flex flex-row justify-between items-center mb-2">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.y_axis") }}
                </h4>
            </div>

            <div>
                <label for="y axis label" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.y_axis_label")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.yAxisLabel"
                        type="text"
                        name="yAxisLabel"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        :placeholder="t('collimato.charts.new_chart.customize.y_axis_label')"
                        @input="update()"
                    />
                </div>
            </div>

            <div>
                <label for="bottom" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.y_axis_name_gap")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.yAxisNameGap"
                        type="number"
                        name="yAxisNameGap"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="50"
                        @input="removeNonDigitsYAxisNameGap($event)"
                        @change="update()"
                    />
                </div>
            </div>

            <div class="flex flex-row justify-between items-center mb-2 mt-5">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.x_axis") }}
                </h4>
            </div>

            <div>
                <label for="x axis label" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.x_axis_label")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.xAxisLabel"
                        type="text"
                        name="xAxisLabel"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        :placeholder="t('collimato.charts.new_chart.customize.x_axis_label')"
                        @input="update()"
                    />
                </div>
            </div>

            <div>
                <label for="bottom" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.x_axis_name_gap")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.xAxisNameGap"
                        type="number"
                        name="yAxisNameGap"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="50"
                        @input="removeNonDigitsXAxisNameGap($event)"
                        @change="update()"
                    />
                </div>
            </div>

            <SettingSelect
                v-model="config.angle"
                :label="t('collimato.charts.new_chart.customize.x_tick_layout')"
                :options="angleOptions"
                @change="update()"
            />

            <div class="flex flex-row justify-between items-center mb-2 mt-5">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.grid") }}
                </h4>
            </div>

            <div>
                <label for="top" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.top_margin")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.topMargin"
                        type="number"
                        name="top"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="75"
                        @input="removeNonDigitsTop($event)"
                        @change="update()"
                    />
                </div>
            </div>
            <div>
                <label for="left" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.left_margin")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.leftMargin"
                        type="number"
                        name="left"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="75"
                        @input="removeNonDigitsLeft($event)"
                        @change="update()"
                    />
                </div>
            </div>
            <div>
                <label for="right" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.right_margin")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.rightMargin"
                        type="number"
                        name="right"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="75"
                        @input="removeNonDigitsRight($event)"
                        @change="update()"
                    />
                </div>
            </div>
            <div>
                <label for="bottom" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.bottom_margin")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.bottomMargin"
                        type="number"
                        name="bottom"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="75"
                        @input="removeNonDigitsBottom($event)"
                        @change="update()"
                    />
                </div>
            </div>

            <BaseCheckbox
                class="mt-2"
                :model-value="config.containLabel"
                :label="t('collimato.charts.new_chart.customize.contain_label')"
                :description="t('collimato.charts.new_chart.customize.contain_label_description')"
                @update:model-value="
                    config.containLabel = $event;
                    update();
                "
            />
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { reactive, computed, onMounted } from "vue";
import {
    LEGEND_POSITIONS,
    DEFAULT_LEGEND_POSITION,
    X_AXIS_DEFAULT,
    X_AXIS_NONE,
    SORT_BARS_NONE,
    SORT_BARS_OPTIONS,
    MISSING_VALUES_CONNECT,
    MISSING_VALUES_OPTIONS,
} from "@/utils/collimato/charts/chartSettings.js";
import { t } from "@/i18n/index.js";
import BaseCheckbox from "@/components/BaseCheckbox.vue";
import SettingSelect from "@/components/Collimato/Charts/Configurations/SettingSelect.vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ChevronUpIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    configuration: {
        type: Object,
        default: () => ({}),
    },
    dimensions: {
        type: Array,
        default: () => [],
    },
    showOrientation: {
        type: Boolean,
        default: false,
    },
    showBarValues: {
        type: Boolean,
        default: false,
    },
    showMissingValues: {
        type: Boolean,
        default: false,
    },
});
const emit = defineEmits(["update:configuration"]);

const defaultConfig = {
    containLabel: true,
    bottomMargin: 5,
    topMargin: 30,
    leftMargin: 5,
    rightMargin: 5,
    yAxisLabel: "",
    yAxisNameGap: 50,
    xAxisLabel: "",
    xAxisNameGap: 50,
    angle: 0,
    legend: true,
    legendPosition: DEFAULT_LEGEND_POSITION,
    xAxisField: X_AXIS_DEFAULT,
    sortBars: SORT_BARS_NONE,
    stack: false,
    orientation: "vertical",
    dataZoom: false,
    showValues: false,
    missingValues: MISSING_VALUES_CONNECT,
};

const config = reactive({
    ...defaultConfig,
    ...props.configuration,
});

const angles = [0, 45, 90];

const angleOptions = angles.map((angle) => ({
    value: angle,
    label: String(angle),
}));

const CUSTOMIZE = "collimato.charts.new_chart.customize.";

const xAxisFields = computed(() => [
    { value: X_AXIS_DEFAULT, label: t.value(`${CUSTOMIZE}x_axis_field_auto`) },
    ...props.dimensions.map((dimension) => ({
        value: dimension.name,
        label: dimension.title || dimension.name,
    })),
    { value: X_AXIS_NONE, label: t.value(`${CUSTOMIZE}x_axis_field_none`) },
]);

const sortBarsOptions = computed(() =>
    SORT_BARS_OPTIONS.map((value) => ({
        value: value,
        label: t.value(`${CUSTOMIZE}sort_bars_${value || "none"}`),
    })),
);

const missingValuesOptions = computed(() =>
    MISSING_VALUES_OPTIONS.map((value) => ({
        value: value,
        label: t.value(`${CUSTOMIZE}missing_values_${value}`),
    })),
);

const legendPositions = computed(() =>
    Object.keys(LEGEND_POSITIONS).map((value) => ({
        value: value,
        label: t.value(
            `collimato.charts.new_chart.customize.legend_position.${value.toLowerCase()}`,
        ),
    })),
);

function removeNonDigitsBottom(event) {
    let text = removeNonDigits(event);

    config.bottomMargin = Number(text);

    update();
}

function removeNonDigitsTop(event) {
    let text = removeNonDigits(event);

    config.topMargin = Number(text);

    update();
}

function removeNonDigitsLeft(event) {
    let text = removeNonDigits(event);

    config.leftMargin = Number(text);

    update();
}

function removeNonDigitsRight(event) {
    let text = removeNonDigits(event);

    config.rightMargin = Number(text);

    update();
}

function removeNonDigitsYAxisNameGap(event) {
    let text = removeNonDigits(event);

    config.yAxisNameGap = Number(text);

    update();
}

function removeNonDigitsXAxisNameGap(event) {
    let text = removeNonDigits(event);

    config.xAxisNameGap = Number(text);

    update();
}

const removeNonDigits = (event) => {
    event.target.value = event.target.value.replace(/\D/g, "");

    return event.target.value;
};

onMounted(() => update());

function update() {
    const payload = {
        angle: config.angle,
        yAxisLabel: config.yAxisLabel,
        yAxisNameGap: config.yAxisNameGap,
        xAxisLabel: config.xAxisLabel,
        xAxisNameGap: config.xAxisNameGap,
        bottomMargin: config.bottomMargin,
        topMargin: config.topMargin,
        rightMargin: config.rightMargin,
        leftMargin: config.leftMargin,
        containLabel: config.containLabel,
        legend: config.legend,
        legendPosition: config.legendPosition,
        xAxisField: config.xAxisField,
        sortBars: config.sortBars,
        dataZoom: config.dataZoom,
        stack: config.stack,
    };

    if (props.showOrientation) {
        payload.orientation = config.orientation;
    }

    if (props.showBarValues) {
        payload.showValues = config.showValues;
    }

    if (props.showMissingValues) {
        payload.missingValues = config.missingValues;
    }

    emit("update:configuration", payload);
}
</script>
