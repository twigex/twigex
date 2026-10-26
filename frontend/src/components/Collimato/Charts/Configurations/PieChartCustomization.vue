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
            <div class="flex flex-row justify-between items-center mb-2 mt-5">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.legend") }}
                </h4>
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
                v-model="config.legendPosition"
                :label="t('collimato.charts.new_chart.customize.legend_position')"
                :options="legendPositions"
                @change="update()"
            />

            <div class="flex flex-row justify-between items-center mb-2 mt-5">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.labels") }}
                </h4>
            </div>

            <div>
                <label for="bottom" class="block text-sm/6 font-medium text-gray-900">{{
                    t("collimato.charts.new_chart.customize.percentage_threshold")
                }}</label>
                <div class="mt-2">
                    <input
                        v-model="config.threshold"
                        type="number"
                        name="yAxisNameGap"
                        class="block w-full rounded-md border-0 py-1 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        placeholder="50"
                        @input="removeNonDigitsThreshold($event)"
                        @change="update()"
                    />
                </div>
            </div>

            <SettingSelect
                v-model="config.labelType"
                :label="t('collimato.charts.new_chart.customize.label_type')"
                :options="labelTypes"
                @change="update()"
            />

            <BaseCheckbox
                class="mt-2 mb-5"
                :model-value="config.showLabels"
                :label="t('collimato.charts.new_chart.customize.show_labels')"
                :description="t('collimato.charts.new_chart.customize.show_labels_description')"
                @update:model-value="
                    config.showLabels = $event;
                    update();
                "
            />

            <BaseCheckbox
                class="mt-2 mb-5"
                :model-value="config.labelPosition"
                :label="t('collimato.charts.new_chart.customize.labels_outside')"
                :description="t('collimato.charts.new_chart.customize.labels_outside_description')"
                @update:model-value="
                    config.labelPosition = $event;
                    update();
                "
            />

            <div class="flex flex-row justify-between items-center mb-2 mt-5">
                <h4 class="text-small font-semibold leading-4 text-gray-400">
                    {{ t("collimato.charts.new_chart.customize.pie_shape") }}
                </h4>
            </div>

            <div>
                <div class="text-sm/6">
                    <label class="font-medium text-gray-900">{{
                        t("collimato.charts.new_chart.customize.outer_radius")
                    }}</label>
                </div>

                <BaseSlider v-model="config.outerRadius" @input="update()" />
            </div>

            <div>
                <div class="text-sm/6">
                    <label class="font-medium text-gray-900">{{
                        t("collimato.charts.new_chart.customize.inner_radius")
                    }}</label>
                </div>

                <BaseSlider v-model="config.innerRadius" @input="update()" />
            </div>
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { reactive, computed, onMounted } from "vue";
import { t } from "@/i18n/index.js";
import {
    LEGEND_POSITIONS,
    DEFAULT_LEGEND_POSITION,
} from "@/utils/collimato/charts/chartSettings.js";

const CUSTOMIZE = "collimato.charts.new_chart.customize.";

import BaseSlider from "@/components/BaseSlider.vue";
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
});
const emit = defineEmits(["update:configuration"]);

const defaultConfig = {
    legend: true,
    legendPosition: DEFAULT_LEGEND_POSITION,
    threshold: 1,
    labelType: "category",
    labelPosition: true,
    outerRadius: 55,
    innerRadius: 0,
    showLabels: true,
};

// Create a reactive configuration by merging defaults with any incoming configuration
const config = reactive({
    ...defaultConfig,
    ...props.configuration,
});

//superset category values
const labelTypes = computed(() => [
    { value: "value", label: t.value(`${CUSTOMIZE}label_type.value`) },
    {
        value: "percentage",
        label: t.value(`${CUSTOMIZE}label_type.percentage`),
    },
    { value: "category", label: t.value(`${CUSTOMIZE}label_type.category`) },
    {
        value: "category, value",
        label: t.value(`${CUSTOMIZE}label_type.category_value`),
    },
    {
        value: "category, percentage",
        label: t.value(`${CUSTOMIZE}label_type.category_percentage`),
    },
    {
        value: "category, value, percentage",
        label: t.value(`${CUSTOMIZE}label_type.category_value_percentage`),
    },
]);

const legendPositions = computed(() =>
    Object.keys(LEGEND_POSITIONS).map((value) => ({
        value: value,
        label: t.value(`${CUSTOMIZE}legend_position.${value.toLowerCase()}`),
    })),
);

function removeNonDigitsThreshold(event) {
    let text = removeNonDigits(event);

    config.threshold = Number(text);

    update();
}

const removeNonDigits = (event) => {
    event.target.value = event.target.value.replace(/\D/g, "");

    return event.target.value;
};

onMounted(() => update());

function update() {
    emit("update:configuration", {
        legend: config.legend,
        threshold: config.threshold,
        labelType: config.labelType,
        labelPosition: config.labelPosition,
        outerRadius: config.outerRadius,
        innerRadius: config.innerRadius,
        legendPosition: config.legendPosition,
        showLabels: config.showLabels,
    });
}
</script>
