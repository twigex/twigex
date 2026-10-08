<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Listbox as="div" class="mt-3" v-model="localValue">
        <ListboxLabel class="block text-sm/6 font-medium text-gray-900">
            {{ t("collimato.chart.new_chart.map.color_selector.color_scheme") }}
        </ListboxLabel>

        <div class="relative mt-2">
            <ListboxButton
                class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
            >
                <div class="flex flex-row">
                    <div
                        v-for="(color, idx) in getColorScheme(localValue)"
                        :key="idx"
                        class="w-4 h-4"
                        :style="{ backgroundColor: color }"
                    ></div>

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
                                    class="w-4 h-4"
                                    :style="{ backgroundColor: color }"
                                ></div>
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
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, watch } from "vue";
import {
    Listbox,
    ListboxLabel,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
} from "@headlessui/vue";

import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";
import { CheckIcon } from "@heroicons/vue/20/solid";

import * as d3 from "d3";

const props = defineProps({
    modelValue: {
        type: String,
        required: true,
    },
});
const emit = defineEmits(["update:modelValue"]);

const localValue = ref(props.modelValue);

watch(localValue, (newVal) => {
    emit("update:modelValue", newVal);
});

watch(
    () => props.modelValue,
    (newVal) => {
        localValue.value = newVal;
    },
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
    if (colorScheme == undefined) {
        return;
    }

    const color = d3.scaleSequential(d3[colorScheme]).domain([0, 20]);

    return Array.from({ length: 20 }, (_, i) => color(i));
}
</script>
