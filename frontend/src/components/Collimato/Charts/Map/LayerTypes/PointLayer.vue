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

    <!-- Radius -->
    <h2 class="mt-3 text-base font-semibold leading-7 text-gray-900">
        {{ t("collimato.charts.new_chart.map.point.radius") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ t("collimato.charts.new_chart.map.point.calculate_radius_based_on") }}
    </p>

    <Listbox as="div" :model-value="radiusField" @update:model-value="setRadiusField">
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
                                radiusField
                                    ? radiusField.title
                                    : t("collimato.charts.new_chart.map.point.select_field")
                            }}
                        </span>
                    </ListboxButton>
                </div>
                <button
                    type="button"
                    @click="setRadiusField(null)"
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

    <div v-if="config.radius_field" class="mt-3 flex items-center justify-between gap-x-3">
        <div>
            <label for="min_radius" class="block text-sm/6 font-medium text-gray-900">
                {{ t("collimato.charts.new_chart.map.point.min") }}
            </label>
            <div class="mt-2">
                <input
                    :value="config.min_radius"
                    type="number"
                    name="min_radius"
                    id="min_radius"
                    class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    placeholder="1"
                    @input="patch({ min_radius: Number($event.target.value) })"
                />
            </div>
        </div>
        <div>
            <label for="max_radius" class="block text-sm/6 font-medium text-gray-900">
                {{ t("collimato.charts.new_chart.map.point.max") }}
            </label>
            <div class="mt-2">
                <input
                    :value="config.max_radius"
                    type="number"
                    name="max_radius"
                    id="max_radius"
                    class="block w-full rounded-md bg-white px-3 py-1.5 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 placeholder:text-gray-400 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
                    placeholder="10"
                    @input="patch({ max_radius: Number($event.target.value) })"
                />
            </div>
        </div>
    </div>

    <div v-else class="mt-4">
        <div class="mb-2 flex items-center justify-between">
            <label class="block text-sm font-medium leading-6 text-gray-900">
                {{ t("collimato.charts.new_chart.map.point.radius") }}
            </label>
            <span
                class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
            >
                {{ config.radius }}
            </span>
        </div>
        <BaseSlider :model-value="config.radius" @input="patch({ radius: $event })" />
    </div>

    <FillColor :config="config" :cubes="valueCubes" @update="updateColor" />

    <!-- Tooltip fields -->
    <div class="mt-3">
        <h2 class="text-base font-semibold leading-7 text-gray-900">
            {{ t("collimato.charts.new_chart.map.tooltip") }}
        </h2>
        <p class="text-sm leading-6 text-gray-500">
            {{ t("collimato.charts.new_chart.map.tooltip_description") }}
        </p>

        <Listbox
            multiple
            as="div"
            :model-value="tooltipFields"
            @update:model-value="setTooltipFields"
        >
            <div class="relative mt-2">
                <ListboxButton
                    class="relative flex min-h-[40px] w-full flex-wrap gap-2 cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                >
                    <template v-if="tooltipFields.length > 0">
                        <span
                            v-for="(field, index) in tooltipFields"
                            :key="index"
                            class="inline-flex items-center gap-x-1 rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600"
                        >
                            {{ field.name }}
                            <button
                                type="button"
                                @click.stop="removeTooltipField(index)"
                                class="group relative -mr-1 size-3.5 rounded-xs hover:bg-gray-500/20"
                            >
                                <svg
                                    viewBox="0 0 14 14"
                                    class="size-3.5 stroke-gray-700/50 group-hover:stroke-gray-700/75"
                                >
                                    <path d="M4 4l6 6m0-6l-6 6" />
                                </svg>
                                <span class="absolute -inset-1" />
                            </button>
                        </span>
                    </template>
                    <span v-else class="block truncate text-gray-400">
                        {{ t("collimato.charts.new_chart.map.select_tooltip_fields") }}
                    </span>
                    <span
                        class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                    >
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
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { joinableDataModels } from "@/utils/collimato/chartQuery.js";
import { withCountMeasure } from "@/utils/collimato/mapLayerTypes.js";
import LatLong from "@/components/Collimato/Charts/Map/LayerTypes/components/LatLong.vue";
import FillColor from "@/components/Collimato/Charts/Map/LayerTypes/components/FillColor.vue";
import BaseSlider from "@/components/BaseSlider.vue";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";
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

const radiusField = computed(
    () =>
        valueCubes.value
            .flatMap((cube) => cube.dimensions)
            .find((d) => d.name === props.config.radius_field) ?? null,
);

const tooltipFields = computed(() =>
    (props.config.popover?.fields ?? [])
        .map((field) =>
            valueCubes.value.flatMap((cube) => cube.dimensions).find((d) => d.name === field),
        )
        .filter(Boolean),
);

function patch(changes) {
    emit("update:config", changes);
}

function setRadiusField(field) {
    patch({ radius_field: field ? field.name : null });
}

function setTooltipFields(fields) {
    patch({
        popover: { fields: fields.map((field) => field.name) },
    });
}

function removeTooltipField(index) {
    setTooltipFields(tooltipFields.value.filter((_, i) => i !== index));
}

function updateColor(value) {
    patch(value);
}
</script>
