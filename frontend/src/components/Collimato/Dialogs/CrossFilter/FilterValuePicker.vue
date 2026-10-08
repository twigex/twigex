<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Combobox
        as="div"
        @update:modelValue="search = ''"
        multiple
        class="mt-2"
        :model-value="modelValue"
        @update:model-value="$emit('update:modelValue', $event)"
    >
        <ComboboxLabel class="block text-sm/6 font-medium text-gray-900">{{
            t("collimato.dashboard.filter.values")
        }}</ComboboxLabel>
        <div class="relative mt-2">
            <!-- Input area with badges -->
            <div
                class="w-full rounded-md bg-white py-1 pl-3 pr-12 text-base text-gray-900 placeholder:text-gray-400 sm:text-sm/6 flex flex-wrap items-center gap-1 ring-1 ring-inset ring-gray-300"
                :class="
                    error
                        ? 'ring-red-600 focus:ring-red-600'
                        : 'ring-gray-300 focus:ring-indigo-600'
                "
            >
                <!-- Render badges for selected people -->
                <template v-for="(value, index) in modelValue" :key="index">
                    <span
                        class="inline-flex items-center rounded bg-gray-100 px-2 py-1 text-sm font-medium text-gray-700"
                    >
                        {{ value }}
                        <button
                            type="button"
                            class="ml-1 rounded-full bg-gray-200 p-0.5 text-gray-600 hover:bg-gray-300 focus:outline-none"
                            @click.stop="remove(value)"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                class="h-4 w-4"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                stroke-width="2"
                            >
                                <path
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                    d="M6 18L18 6M6 6l12 12"
                                />
                            </svg>
                        </button>
                    </span>
                </template>

                <ComboboxInput
                    class="flex-1 border-none py-1.5 pl-1 text-sm leading-5 text-gray-900 focus:ring-0 min-w-[50px]"
                    @change="search = $event.target.value"
                    @blur="search = ''"
                    :placeholder="modelValue.length === 0 ? 'Select values' : ''"
                />
            </div>

            <!-- Dropdown button -->
            <ComboboxButton
                class="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none"
            >
                <ChevronUpDownIcon class="size-5 text-gray-400" aria-hidden="true" />
            </ComboboxButton>

            <!-- Options list -->
            <ComboboxOptions
                v-if="available.length > 0"
                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
            >
                <ComboboxOption
                    v-for="(value, index) in available"
                    :key="index"
                    :value="value"
                    as="template"
                    v-slot="{ active, selected }"
                >
                    <li
                        :class="[
                            'relative cursor-default select-none py-2 pl-3 pr-9',
                            active ? 'bg-indigo-600 text-white outline-none' : 'text-gray-900',
                        ]"
                    >
                        <span :class="['block truncate', selected && 'font-semibold']">
                            {{ value }}
                        </span>
                        <span
                            v-if="selected"
                            :class="[
                                'absolute inset-y-0 right-0 flex items-center pr-4',
                                active ? 'text-white' : 'text-indigo-600',
                            ]"
                        >
                            <CheckIcon class="size-5" aria-hidden="true" />
                        </span>
                    </li>
                </ComboboxOption>
            </ComboboxOptions>
        </div>

        <p v-if="error" class="mt-2 text-sm text-red-600">This field is required</p>
    </Combobox>
</template>

<script setup>
import { computed, ref } from "vue";
import { t } from "@/i18n/index.js";
import {
    Combobox,
    ComboboxButton,
    ComboboxInput,
    ComboboxLabel,
    ComboboxOption,
    ComboboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    modelValue: {
        type: Array,
        default: () => [],
    },
    values: {
        type: Array,
        default: () => [],
    },
    error: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:modelValue"]);

const search = ref("");

const available = computed(() =>
    props.values.filter(
        (value) =>
            !props.modelValue.includes(value) &&
            (search.value === "" ||
                String(value).toLowerCase().includes(search.value.toLowerCase())),
    ),
);

function remove(value) {
    emit(
        "update:modelValue",
        props.modelValue.filter((selected) => selected !== value),
    );
}
</script>
