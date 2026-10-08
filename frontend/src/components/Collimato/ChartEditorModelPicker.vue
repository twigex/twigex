<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="border-t border-gray-200 px-4 pt-4 pb-2">
        <p class="text-xs font-semibold uppercase tracking-wider text-gray-400">
            {{ t("collimato.charts.new_chart.data_model") }}
        </p>
    </div>
    <Listbox
        as="div"
        :model-value="selectedModel"
        @update:model-value="emit('select', $event)"
        class="px-4 pb-4"
    >
        <div class="relative">
            <ListboxButton
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
                <span class="block truncate">{{
                    selectedModel
                        ? selectedModel.table
                        : t("collimato.charts.new_chart.select_model")
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
                    class="absolute z-10 mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-sm shadow-lg ring-1 ring-black/5 focus:outline-none"
                >
                    <ListboxOption
                        as="template"
                        v-for="c in dataModels"
                        :key="c"
                        :value="c"
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
                                >{{ c.table }}</span
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
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";

defineProps({
    dataModels: {
        type: Array,
        default: () => [],
    },
    selectedModel: {
        type: Object,
        default: null,
    },
});

const emit = defineEmits(["select"]);
</script>
