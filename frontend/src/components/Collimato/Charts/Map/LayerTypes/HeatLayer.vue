<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <LatLong
        :coordinates="config.coordinates"
        :cubes="filteredCubes"
        @update:coordinates="$emit('update:coordinates', $event)"
    />

    <!-- Color scheme -->
    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.chart.new_chart.map.heatmap.color") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.chart.new_chart.map.heatmap.color_description") }}
    </p>
    <ColorSelector
        :model-value="config.colorScheme"
        @update:modelValue="patch({ colorScheme: $event })"
    />

    <!-- Weight field -->
    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.chart.new_chart.map.heatmap.weight") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.chart.new_chart.map.heatmap.weight_description") }}
    </p>

    <Listbox as="div" :model-value="weight" @update:model-value="setWeight">
        <div class="relative mt-2">
            <div
                class="flex rounded-md shadow-sm ring-1 ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600"
            >
                <div class="grow grid grid-cols-1">
                    <ListboxButton
                        class="col-start-1 row-start-1 relative w-full cursor-default rounded-l-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 focus:outline-none sm:text-sm sm:leading-6"
                    >
                        <span class="block truncate">
                            {{
                                weight
                                    ? weight.title
                                    : t("collimato.chart.new_chart.map.heatmap.select_weight_field")
                            }}
                        </span>
                    </ListboxButton>
                </div>
                <button
                    type="button"
                    @click="setWeight(null)"
                    class="flex shrink-0 items-center gap-x-1.5 rounded-r-md border-l bg-white px-3 py-2 text-sm font-semibold text-gray-900 hover:bg-gray-50 focus:outline-none"
                >
                    <XMarkIcon class="size-4 text-gray-400" aria-hidden="true" />
                </button>
            </div>

            <transition
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <ListboxOptions
                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                >
                    <template v-for="(d, index) in valueCubes" :key="index">
                        <li
                            class="cursor-default select-none bg-gray-100 py-2 pl-3 pr-9 text-left font-semibold text-gray-900"
                        >
                            {{ d.table }}
                        </li>
                        <ListboxOption
                            as="template"
                            v-for="dimension in d.dimensions"
                            :key="dimension.name"
                            :value="dimension"
                            v-slot="{ active, selected }"
                        >
                            <li
                                :class="[
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                    'relative cursor-default select-none py-2 pl-3 pr-9',
                                ]"
                            >
                                <span
                                    :class="[
                                        selected ? 'font-semibold' : 'font-normal',
                                        'ml-5 block truncate',
                                    ]"
                                    >{{ dimension.title }}</span
                                >
                                <span
                                    v-if="selected"
                                    :class="[
                                        active ? 'text-white' : 'text-indigo-600',
                                        'absolute inset-y-0 right-0 flex items-center pr-4',
                                    ]"
                                >
                                    <CheckIcon class="h-5 w-5" aria-hidden="true" />
                                </span>
                            </li>
                        </ListboxOption>
                    </template>
                </ListboxOptions>
            </transition>
        </div>
    </Listbox>

    <!-- Radius -->
    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.chart.new_chart.map.heatmap.radius") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.chart.new_chart.map.heatmap.radius_description") }}
    </p>
    <div class="mt-3">
        <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium leading-6 text-gray-900">
                {{ t("collimato.chart.new_chart.map.heatmap.radius") }}
            </label>
            <span
                class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
            >
                {{ config.radiusPixels }}
            </span>
        </div>
        <BaseSlider :model-value="config.radiusPixels" @input="patch({ radiusPixels: $event })" />
    </div>

    <!-- Intensity -->
    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.chart.new_chart.map.heatmap.intensity") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.chart.new_chart.map.heatmap.intensity_description") }}
    </p>
    <div class="mt-3">
        <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium leading-6 text-gray-900">
                {{ t("collimato.chart.new_chart.map.heatmap.intensity") }}
            </label>
            <span
                class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
            >
                {{ config.intensity }}
            </span>
        </div>
        <BaseSlider
            :model-value="config.intensity"
            :min="0"
            :max="5"
            :step="0.1"
            @input="patch({ intensity: $event })"
        />
    </div>

    <!-- Threshold -->
    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.chart.new_chart.map.heatmap.threshold") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.chart.new_chart.map.heatmap.threshold_description") }}
    </p>
    <div class="mt-3">
        <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium leading-6 text-gray-900">
                {{ t("collimato.chart.new_chart.map.heatmap.threshold") }}
            </label>
            <span
                class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
            >
                {{ config.threshold }}
            </span>
        </div>
        <BaseSlider
            :model-value="config.threshold"
            :min="0"
            :max="1"
            :step="0.01"
            @input="patch({ threshold: $event })"
        />
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { withCountMeasure } from "@/utils/collimato/mapLayerTypes.js";
import LatLong from "@/components/Collimato/Charts/Map/LayerTypes/components/LatLong.vue";
import ColorSelector from "@/components/Collimato/Charts/Map/LayerTypes/components/ColorSelector.vue";
import BaseSlider from "@/components/BaseSlider.vue";
import { Listbox, ListboxOptions, ListboxOption, ListboxButton } from "@headlessui/vue";
import { CheckIcon, XMarkIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    cubes: {
        type: Array,
        default: () => [],
    },
    selectedCube: {
        type: String,
        default: null,
    },
    layer: {
        type: Object,
        required: true,
    },
    config: {
        type: Object,
        required: true,
    },
});

const emit = defineEmits(["update:config", "update:coordinates"]);

const selectedCube = computed(
    () => props.cubes.find((c) => c.table === props.selectedCube) ?? null,
);

const filteredCubes = computed(() => joinableDataModels(props.cubes, selectedCube.value));

const valueCubes = computed(() =>
    withCountMeasure(
        filteredCubes.value,
        props.selectedCube,
        t.value("collimato.charts.new_chart.map.count"),
    ),
);

const weight = computed(
    () =>
        valueCubes.value
            .flatMap((cube) => cube.dimensions)
            .find((d) => d.name === props.config.weight) ?? null,
);

function patch(changes) {
    emit("update:config", changes);
}

function setWeight(field) {
    patch({ weight: field ? field.name : null });
}
</script>
