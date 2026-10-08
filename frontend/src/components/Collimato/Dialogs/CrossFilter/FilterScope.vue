<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <fieldset class="mt-3">
        <p class="block text-sm/6 font-medium text-gray-900">
            {{ t("collimato.dashboard.filter.scope_help") }}
        </p>
        <div class="mt-2 space-y-6 sm:flex sm:items-center sm:space-x-10 sm:space-y-0">
            <div v-for="option in options" :key="option.id" class="flex items-center">
                <input
                    :checked="scope?.id === option.id"
                    @change="$emit('update:scope', option)"
                    :id="option.id"
                    :value="option"
                    name="option"
                    type="radio"
                    class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-600"
                />
                <label :for="option.id" class="ml-3 block text-sm/6 font-medium text-gray-600">{{
                    option.title
                }}</label>
            </div>
        </div>
    </fieldset>

    <Listbox
        v-if="scope && scope.id == 'specific'"
        as="div"
        :model-value="charts"
        @update:model-value="$emit('update:charts', $event)"
        multiple
    >
        <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{
            t("collimato.dashboard.filter.assigned_to")
        }}</ListboxLabel>
        <div class="relative mt-2">
            <ListboxButton
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm/6"
            >
                <span class="block truncate">{{
                    charts.length > 0
                        ? charts.map((chart) => chart.name).join(", ")
                        : "Select charts"
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
                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                >
                    <ListboxOption
                        as="template"
                        v-for="chart in availableCharts"
                        :key="chart.id"
                        :value="chart"
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
                                >{{ chart.name }}</span
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
</template>

<script setup>
import { t } from "@/i18n/index.js";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";

defineProps({
    scope: {
        type: Object,
        default: null,
    },
    charts: {
        type: Array,
        default: () => [],
    },
    options: {
        type: Array,
        default: () => [],
    },
    availableCharts: {
        type: Array,
        default: () => [],
    },
});

defineEmits(["update:scope", "update:charts"]);
</script>
