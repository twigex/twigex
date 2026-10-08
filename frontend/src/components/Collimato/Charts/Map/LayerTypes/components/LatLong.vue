<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <h2 class="text-base font-semibold leading-7 text-gray-900">
        {{ title || t("collimato.chart.new_chart.map.lat_long.table_columns") }}
    </h2>
    <p class="text-sm leading-6 text-gray-500">
        {{ description || t("collimato.chart.new_chart.map.lat_long.table_columns_description") }}
    </p>

    <Listbox
        as="div"
        :model-value="selectedLat"
        @update:model-value="$emit('update:coordinates', { lat: $event ? $event.name : null })"
    >
        <label class="mt-3 block text-xs font-medium text-gray-700">
            {{ t("collimato.chart.new_chart.map.lat_long.select_latitude") }}
        </label>
        <div class="relative mt-1">
            <ListboxButton
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
            >
                <span class="block truncate">{{
                    selectedLat
                        ? selectedLat.title
                        : t("collimato.chart.new_chart.map.lat_long.select_latitude")
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
        as="div"
        :model-value="selectedLong"
        @update:model-value="$emit('update:coordinates', { long: $event ? $event.name : null })"
    >
        <label class="mt-3 block text-xs font-medium text-gray-700">
            {{ t("collimato.chart.new_chart.map.lat_long.select_longitude") }}
        </label>
        <div class="relative mt-1">
            <ListboxButton
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
            >
                <span class="block truncate">{{
                    selectedLong
                        ? selectedLong.title
                        : t("collimato.chart.new_chart.map.lat_long.select_longitude")
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
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";
import { CheckIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    cubes: {
        type: Array,
        default: () => [],
    },
    coordinates: {
        type: Object,
        required: true,
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

defineEmits(["update:coordinates"]);

const dimensions = computed(() => props.cubes.flatMap((cube) => cube.dimensions));

const selectedLat = computed(
    () => dimensions.value.find((d) => d.name === props.coordinates.lat) ?? null,
);

const selectedLong = computed(
    () => dimensions.value.find((d) => d.name === props.coordinates.long) ?? null,
);
</script>
