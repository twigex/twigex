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

            <Listbox as="div" v-if="dimensions.length" class="mb-5" v-model="config.xAxisField">
                <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.x_axis_field")
                }}</ListboxLabel>
                <p class="text-sm leading-6 text-gray-500">
                    {{ t("collimato.charts.new_chart.customize.x_axis_field_description") }}
                </p>
                <div class="relative mt-2">
                    <ListboxButton
                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    >
                        <span class="col-start-1 row-start-1 truncate pr-6">{{
                            selectedXAxisField
                        }}</span>
                        <ChevronUpDownIcon
                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                            aria-hidden="true"
                        />
                    </ListboxButton>

                    <transition
                        leave-active-class="transition ease-in duration-100"
                        leave-from-class="opacity-100"
                        leave-to-class="opacity-0"
                    >
                        <ListboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                as="template"
                                v-for="field in xAxisFields"
                                :key="field.value"
                                :value="field.value"
                                v-slot="{ active, selected }"
                                @click="update()"
                            >
                                <li
                                    :class="[
                                        active
                                            ? 'bg-indigo-600 text-white outline-none'
                                            : 'text-gray-900',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                        >{{ field.label }}</span
                                    >
                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                        ]"
                                    >
                                        <CheckIcon class="size-5" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </transition>
                </div>
            </Listbox>

            <Listbox as="div" class="mb-5" v-model="config.sortBars">
                <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.sort_bars")
                }}</ListboxLabel>
                <div class="relative mt-2">
                    <ListboxButton
                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    >
                        <span class="col-start-1 row-start-1 truncate pr-6">{{
                            selectedSortBars
                        }}</span>
                        <ChevronUpDownIcon
                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                            aria-hidden="true"
                        />
                    </ListboxButton>

                    <transition
                        leave-active-class="transition ease-in duration-100"
                        leave-from-class="opacity-100"
                        leave-to-class="opacity-0"
                    >
                        <ListboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                as="template"
                                v-for="sort in sortBarsOptions"
                                :key="sort.value"
                                :value="sort.value"
                                v-slot="{ active, selected }"
                                @click="update()"
                            >
                                <li
                                    :class="[
                                        active
                                            ? 'bg-indigo-600 text-white outline-none'
                                            : 'text-gray-900',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                        >{{ sort.label }}</span
                                    >
                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                        ]"
                                    >
                                        <CheckIcon class="size-5" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </transition>
                </div>
            </Listbox>

            <Listbox as="div" v-if="config.legend" class="mb-5" v-model="config.legendPosition">
                <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.legend_position")
                }}</ListboxLabel>
                <div class="relative mt-2">
                    <ListboxButton
                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    >
                        <span class="col-start-1 row-start-1 truncate pr-6">{{
                            selectedLegendPosition
                        }}</span>
                        <ChevronUpDownIcon
                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                            aria-hidden="true"
                        />
                    </ListboxButton>

                    <transition
                        leave-active-class="transition ease-in duration-100"
                        leave-from-class="opacity-100"
                        leave-to-class="opacity-0"
                    >
                        <ListboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                as="template"
                                v-for="position in legendPositions"
                                :key="position.value"
                                :value="position.value"
                                v-slot="{ active, selected }"
                                @click="update()"
                            >
                                <li
                                    :class="[
                                        active
                                            ? 'bg-indigo-600 text-white outline-none'
                                            : 'text-gray-900',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                        >{{ position.label }}</span
                                    >
                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                        ]"
                                    >
                                        <CheckIcon class="size-5" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </transition>
                </div>
            </Listbox>

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

            <Listbox as="div" v-model="config.angle">
                <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.x_tick_layout")
                }}</ListboxLabel>
                <div class="relative mt-2">
                    <ListboxButton
                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    >
                        <span class="col-start-1 row-start-1 truncate pr-6">{{
                            config.angle
                        }}</span>
                        <ChevronUpDownIcon
                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                            aria-hidden="true"
                        />
                    </ListboxButton>

                    <transition
                        leave-active-class="transition ease-in duration-100"
                        leave-from-class="opacity-100"
                        leave-to-class="opacity-0"
                    >
                        <ListboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                as="template"
                                v-for="(angle, index) in angles"
                                :key="index"
                                :value="angle"
                                v-slot="{ active, selected }"
                                @click="update()"
                            >
                                <li
                                    :class="[
                                        active
                                            ? 'bg-indigo-600 text-white outline-none'
                                            : 'text-gray-900',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                        >{{ angle }}</span
                                    >

                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                        ]"
                                    >
                                        <CheckIcon class="size-5" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </transition>
                </div>
            </Listbox>

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
} from "@/utils/collimato/chartUtils.js";
import { t } from "@/i18n/index.js";
import BaseCheckbox from "@/components/BaseCheckbox.vue";
import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { ChevronUpIcon, CheckIcon } from "@heroicons/vue/20/solid";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";

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
};

const config = reactive({
    ...defaultConfig,
    ...props.configuration,
});

const angles = [0, 45, 90];

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

const selectedSortBars = computed(
    () => sortBarsOptions.value.find((s) => s.value == config.sortBars)?.label ?? config.sortBars,
);

const selectedXAxisField = computed(
    () => xAxisFields.value.find((f) => f.value == config.xAxisField)?.label ?? config.xAxisField,
);

const legendPositions = computed(() =>
    Object.keys(LEGEND_POSITIONS).map((value) => ({
        value: value,
        label: t.value(
            `collimato.charts.new_chart.customize.legend_position.${value.toLowerCase()}`,
        ),
    })),
);

const selectedLegendPosition = computed(
    () =>
        legendPositions.value.find((p) => p.value == config.legendPosition)?.label ??
        config.legendPosition,
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

    emit("update:configuration", payload);
}
</script>
