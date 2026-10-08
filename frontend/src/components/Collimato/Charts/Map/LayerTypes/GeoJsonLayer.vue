<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <h2 class="text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.charts.new_chart.map.geojson.geometry") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.charts.new_chart.map.geojson.geometry_description") }}
    </p>

    <Listbox
        as="div"
        class="mt-2"
        :model-value="geometryField"
        @update:model-value="
            $emit('update:fields', {
                geojson_field: $event ? $event.name : null,
            })
        "
    >
        <div class="relative">
            <ListboxButton
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
            >
                <span class="block truncate">{{
                    geometryField
                        ? geometryField.title
                        : t("collimato.charts.new_chart.map.geojson.select_geometry")
                }}</span>
                <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                    <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                </span>
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
                        v-for="field in dimensions"
                        :key="field.name"
                        :value="field"
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
                                    'block truncate',
                                ]"
                                >{{ field.title }}</span
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
                </ListboxOptions>
            </transition>
        </div>
    </Listbox>

    <div class="mt-4 space-y-3">
        <BaseCheckbox
            :model-value="config.filled"
            :label="t('collimato.charts.new_chart.map.geojson.filled')"
            @update:model-value="patch({ filled: $event })"
        />
        <BaseCheckbox
            :model-value="config.stroked"
            :label="t('collimato.charts.new_chart.map.geojson.stroked')"
            @update:model-value="patch({ stroked: $event })"
        />
        <BaseCheckbox
            :model-value="config.extruded"
            :label="t('collimato.charts.new_chart.map.geojson.extruded')"
            :description="t('collimato.charts.new_chart.map.geojson.extruded_description')"
            @update:model-value="patch({ extruded: $event })"
        />
        <BaseCheckbox
            v-if="config.extruded"
            :model-value="config.wireframe"
            :label="t('collimato.charts.new_chart.map.geojson.wireframe')"
            @update:model-value="patch({ wireframe: $event })"
        />
    </div>

    <FillColor v-if="config.filled" :config="config" :cubes="valueCubes" @update="patch" />

    <div v-if="config.stroked" class="mt-4">
        <label :for="lineColorId" class="block text-sm font-medium leading-6 text-gray-900">
            {{ t("collimato.charts.new_chart.map.geojson.line_color") }}
        </label>
        <input
            :id="lineColorId"
            :value="config.line_color"
            type="color"
            class="mt-2 block h-10 w-14 cursor-pointer rounded-lg border border-gray-200 bg-white p-1"
            @input="patch({ line_color: $event.target.value })"
        />

        <label class="mt-3 block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.geojson.line_width") }}
        </label>
        <BaseSlider
            :model-value="config.line_width"
            :min="0"
            :max="20"
            @input="patch({ line_width: Number($event) })"
        />
    </div>

    <div v-if="config.extruded" class="mt-4">
        <label class="block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.geojson.elevation") }}
        </label>
        <input
            type="number"
            min="0"
            class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm"
            :value="config.elevation"
            @input="patch({ elevation: Number($event.target.value) })"
        />

        <label class="mt-3 block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.geojson.elevation_scale") }}
        </label>
        <BaseSlider
            :model-value="config.elevation_scale"
            :min="0"
            :max="10"
            :step="0.1"
            @input="patch({ elevation_scale: Number($event) })"
        />
    </div>
</template>

<script setup>
import { computed, useId } from "vue";
import { t } from "@/i18n/index.js";
import { joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { withCountMeasure } from "@/utils/collimato/mapLayerTypes.js";
import FillColor from "@/components/Collimato/Charts/Map/LayerTypes/components/FillColor.vue";
import BaseCheckbox from "@/components/BaseCheckbox.vue";
import BaseSlider from "@/components/BaseSlider.vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";
import { CheckIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    cubes: {
        type: Array,
        default: () => [],
    },
    selectedCube: {
        type: String,
        default: null,
    },
    config: {
        type: Object,
        required: true,
    },
});

const emit = defineEmits(["update:config", "update:fields"]);

const lineColorId = useId();

const filteredCubes = computed(() =>
    joinableDataModels(
        props.cubes,
        props.cubes.find((cube) => cube.table === props.selectedCube) ?? null,
    ),
);

const valueCubes = computed(() =>
    withCountMeasure(
        filteredCubes.value,
        props.selectedCube,
        t.value("collimato.charts.new_chart.map.count"),
    ),
);

const dimensions = computed(() => filteredCubes.value.flatMap((cube) => cube.dimensions));

const geometryField = computed(
    () => dimensions.value.find((d) => d.name === props.config.geojson_field) ?? null,
);

function patch(changes) {
    emit("update:config", { ...props.config, ...changes });
}
</script>
