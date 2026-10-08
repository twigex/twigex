<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <h2 class="mt-3 text-base font-semibold leading-7 text-gray-900">
        {{ title || t("collimato.chart.new_chart.map.fill_color.fill_color") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ description || t("collimato.chart.new_chart.map.fill_color.fill_color_description") }}
    </p>

    <Listbox as="div" :model-value="colorField" @update:model-value="setColorField">
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
                                colorField
                                    ? colorField.title
                                    : t("collimato.chart.new_chart.map.fill_color.select_field")
                            }}
                        </span>
                    </ListboxButton>
                </div>
                <button
                    type="button"
                    @click="setColorField(null)"
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
                    <template v-for="(d, index) in cubes" :key="index">
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

    <Listbox
        v-if="colorField"
        as="div"
        class="mt-3"
        :model-value="selectedColorScheme"
        @update:model-value="setColorScheme"
    >
        <ListboxLabel class="block text-sm/6 font-medium text-gray-900">
            {{ t("collimato.chart.new_chart.map.fill_color.color_scheme") }}
        </ListboxLabel>
        <div class="relative mt-2">
            <ListboxButton
                class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
            >
                <div class="flex flex-row">
                    <div
                        v-for="(color, idx) in getColorScheme(selectedColorScheme)"
                        :key="idx"
                        class="h-4 w-4"
                        :style="{ backgroundColor: color }"
                    />
                    <ChevronUpDownIcon
                        class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                        aria-hidden="true"
                    />
                </div>
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
                        v-for="(scheme, index) in colorSchemes"
                        :key="index"
                        :value="scheme"
                        v-slot="{ active, selected }"
                    >
                        <li
                            :class="[
                                active ? 'bg-indigo-600 text-white outline-none' : 'text-gray-900',
                                'relative cursor-default select-none py-2 pl-8 pr-4',
                            ]"
                        >
                            <div class="flex flex-row">
                                <div
                                    v-for="(color, idx) in getColorScheme(scheme)"
                                    :key="idx"
                                    class="h-4 w-4"
                                    :style="{ backgroundColor: color }"
                                />
                            </div>
                            <span
                                v-if="selected"
                                :class="[
                                    active ? 'text-white' : 'text-indigo-600',
                                    'absolute inset-y-0 left-0 flex items-center pl-1.5',
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

    <div v-else class="mt-3">
        <label :for="inputId" class="block text-sm font-medium leading-6 text-gray-900">
            {{ title || t("collimato.chart.new_chart.map.fill_color.fill_color") }}
        </label>
        <div class="mt-2">
            <input
                :value="fillColor"
                type="color"
                :name="inputId"
                :id="inputId"
                class="block h-10 w-14 cursor-pointer rounded-lg border border-gray-200 bg-white p-1 disabled:pointer-events-none disabled:opacity-50"
                :title="t('collimato.chart.new_chart.map.fill_color.fill_color')"
                @input="setFillColor($event.target.value)"
            />
        </div>
    </div>
</template>

<script setup>
import { computed, useId } from "vue";
import { t } from "@/i18n/index.js";
import * as d3 from "d3";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";
import { CheckIcon, XMarkIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    config: {
        type: Object,
        required: true,
    },
    cubes: {
        type: Array,
        required: true,
    },
    colorKey: {
        type: String,
        default: "fill_color",
    },
    fieldKey: {
        type: String,
        default: "fill_color_field",
    },
    schemeKey: {
        type: String,
        default: "color_scheme",
    },
    title: {
        type: String,
        default: "",
    },
    description: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["update"]);

const inputId = useId();

const fillColor = computed(() => props.config[props.colorKey]);

const colorField = computed(
    () =>
        props.cubes
            .flatMap((cube) => cube.dimensions)
            .find((d) => d.name === props.config[props.fieldKey]) ?? null,
);

const selectedColorScheme = computed(
    () => props.config[props.schemeKey] ?? props.config.color_scheme,
);

const colorSchemes = [
    "interpolateViridis",
    "interpolateInferno",
    "interpolateMagma",
    "interpolatePlasma",
    "interpolateWarm",
    "interpolateCool",
    "interpolateCubehelixDefault",
    "interpolateTurbo",
    "interpolateSinebow",
    "interpolateBlues",
    "interpolateGreens",
    "interpolateGreys",
    "interpolateOranges",
    "interpolatePurples",
    "interpolateReds",
    "interpolateBrBG",
    "interpolatePRGn",
    "interpolatePiYG",
    "interpolatePuOr",
    "interpolateRdBu",
    "interpolateRdGy",
    "interpolateRdYlBu",
    "interpolateRdYlGn",
    "interpolateSpectral",
    "interpolateRainbow",
];

function getColorScheme(colorScheme) {
    const color = d3.scaleSequential(d3[colorScheme]).domain([0, 20]);

    return Array.from({ length: 20 }, (_, i) => color(i));
}

function emitColor(changes) {
    emit("update", {
        [props.fieldKey]: colorField.value ? colorField.value.name : null,
        [props.schemeKey]: selectedColorScheme.value,
        [props.colorKey]: fillColor.value,
        ...changes,
    });
}

function setColorField(field) {
    emitColor({ [props.fieldKey]: field ? field.name : null });
}

function setColorScheme(scheme) {
    emitColor({ [props.schemeKey]: scheme });
}

function setFillColor(color) {
    emitColor({ [props.colorKey]: color });
}
</script>
