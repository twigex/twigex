<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <LatLong
        :coordinates="config.source"
        :cubes="filteredCubes"
        :title="t('collimato.charts.new_chart.map.arc.source')"
        :description="t('collimato.charts.new_chart.map.arc.source_description')"
        @update:coordinates="$emit('update:fields', { source: $event })"
    />

    <LatLong
        class="mt-5"
        :coordinates="config.target"
        :cubes="filteredCubes"
        :title="t('collimato.charts.new_chart.map.arc.target')"
        :description="t('collimato.charts.new_chart.map.arc.target_description')"
        @update:coordinates="$emit('update:fields', { target: $event })"
    />

    <FillColor
        :config="config"
        :cubes="valueCubes"
        color-key="source_color"
        field-key="source_color_field"
        scheme-key="source_color_scheme"
        :title="t('collimato.charts.new_chart.map.arc.source_color')"
        :description="t('collimato.charts.new_chart.map.arc.source_color_description')"
        @update="patch"
    />

    <FillColor
        :config="config"
        :cubes="valueCubes"
        color-key="target_color"
        field-key="target_color_field"
        scheme-key="target_color_scheme"
        :title="t('collimato.charts.new_chart.map.arc.target_color')"
        :description="t('collimato.charts.new_chart.map.arc.target_color_description')"
        @update="patch"
    />

    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.charts.new_chart.map.arc.width") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.charts.new_chart.map.arc.width_description") }}
    </p>

    <Listbox
        as="div"
        class="mt-2"
        :model-value="widthField"
        @update:model-value="patch({ width_field: $event ? $event.name : null })"
    >
        <label class="block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.arc.width_field") }}
        </label>
        <div class="relative mt-1">
            <div
                class="flex rounded-md shadow-sm ring-1 ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600"
            >
                <ListboxButton
                    class="grow relative w-full cursor-default rounded-l-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 focus:outline-none sm:text-sm"
                >
                    <span class="block truncate">{{
                        widthField
                            ? widthField.title
                            : t("collimato.charts.new_chart.map.arc.width_fixed")
                    }}</span>
                </ListboxButton>
                <button
                    type="button"
                    class="flex shrink-0 items-center rounded-r-md border-l bg-white px-3 py-2 text-sm text-gray-900 hover:bg-gray-50"
                    @click="patch({ width_field: null })"
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
                            v-for="field in d.dimensions"
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
                            </li>
                        </ListboxOption>
                    </template>
                </ListboxOptions>
            </transition>
        </div>
    </Listbox>

    <div v-if="!config.width_field" class="mt-2">
        <label class="block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.arc.width") }}
        </label>
        <BaseSlider
            :model-value="config.width"
            :min="1"
            :max="20"
            @input="patch({ width: Number($event) })"
        />
    </div>

    <div class="mt-3 flex gap-4">
        <div class="flex-1">
            <label class="block text-xs font-medium text-gray-700">
                {{ t("collimato.charts.new_chart.map.arc.min_width") }}
            </label>
            <input
                type="number"
                min="0"
                class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm"
                :value="config.min_width"
                @input="patch({ min_width: Number($event.target.value) })"
            />
        </div>
        <div class="flex-1">
            <label class="block text-xs font-medium text-gray-700">
                {{ t("collimato.charts.new_chart.map.arc.max_width") }}
            </label>
            <input
                type="number"
                min="0"
                class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 ring-1 ring-inset ring-gray-300 sm:text-sm"
                :value="config.max_width"
                @input="patch({ max_width: Number($event.target.value) })"
            />
        </div>
    </div>

    <h2 class="mt-5 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.charts.new_chart.map.arc.shape") }}
    </h2>

    <div class="mt-2">
        <label class="block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.arc.height") }}
        </label>
        <BaseSlider
            :model-value="config.height"
            :min="0"
            :max="5"
            :step="0.1"
            @input="patch({ height: Number($event) })"
        />
    </div>

    <div class="mt-3">
        <label class="block text-xs font-medium text-gray-700">
            {{ t("collimato.charts.new_chart.map.arc.tilt") }}
        </label>
        <BaseSlider
            :model-value="config.tilt"
            :min="-90"
            :max="90"
            @input="patch({ tilt: Number($event) })"
        />
    </div>

    <BaseCheckbox
        class="mt-4"
        :model-value="config.great_circle"
        :label="t('collimato.charts.new_chart.map.arc.great_circle')"
        :description="t('collimato.charts.new_chart.map.arc.great_circle_description')"
        @update:model-value="patch({ great_circle: $event })"
    />
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { withCountMeasure } from "@/utils/collimato/mapLayerTypes.js";
import LatLong from "@/components/Collimato/Charts/Map/LayerTypes/components/LatLong.vue";
import BaseCheckbox from "@/components/BaseCheckbox.vue";
import FillColor from "@/components/Collimato/Charts/Map/LayerTypes/components/FillColor.vue";
import BaseSlider from "@/components/BaseSlider.vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/20/solid";

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

const widthField = computed(
    () =>
        valueCubes.value
            .flatMap((cube) => cube.dimensions)
            .find((d) => d.name === props.config.width_field) ?? null,
);

function patch(changes) {
    emit("update:config", { ...props.config, ...changes });
}
</script>
