<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Listbox
        :modelValue="modelValue"
        as="div"
        @update:modelValue="(val) => emit('update:modelValue', val)"
    >
        <ListboxButton
            ref="reference"
            class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
        >
            <span v-if="type === 'color'" class="flex items-center gap-x-2">
                <span
                    v-if="modelValue?.color"
                    class="inline-block h-3 w-3 rounded-full shrink-0"
                    :style="{ backgroundColor: modelValue.color }"
                />
                <span class="block truncate">{{
                    modelValue?.name || t("files.metadata_dialog.select_color")
                }}</span>
            </span>
            <span v-else class="block truncate">{{
                modelValue || t("files.metadata_dialog.select_option")
            }}</span>
            <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </span>
        </ListboxButton>
        <Teleport to="body">
            <transition
                leave-active-class="transition ease-in duration-100"
                leave-from-class="opacity-100"
                leave-to-class="opacity-0"
            >
                <ListboxOptions
                    ref="floating"
                    :style="floatingStyles"
                    class="z-[100] max-h-60 overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                >
                    <ListboxOption
                        as="template"
                        v-for="field in fields"
                        :key="type === 'color' ? field.name : field"
                        :value="field"
                        v-slot="{ active: isActive, selected: isSelected }"
                    >
                        <li
                            :class="[
                                isActive ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                'relative cursor-default select-none py-2 pl-3 pr-9',
                            ]"
                        >
                            <span v-if="type === 'color'" class="flex items-center gap-x-2">
                                <span
                                    class="inline-block h-3 w-3 rounded-full shrink-0"
                                    :style="{ backgroundColor: field.color }"
                                />
                                <span
                                    :class="[
                                        isSelected ? 'font-semibold' : 'font-normal',
                                        'block truncate',
                                    ]"
                                    >{{ field.name }}</span
                                >
                            </span>
                            <span
                                v-else
                                :class="[
                                    isSelected ? 'font-semibold' : 'font-normal',
                                    'block truncate',
                                ]"
                                >{{ field }}</span
                            >
                            <span
                                v-if="isSelected"
                                :class="[
                                    isActive ? 'text-white' : 'text-indigo-600',
                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                ]"
                            >
                                <CheckIcon class="h-5 w-5" aria-hidden="true" />
                            </span>
                        </li>
                    </ListboxOption>
                </ListboxOptions>
            </transition>
        </Teleport>
    </Listbox>
</template>

<script setup>
import { ref } from "vue";
import { t } from "@/i18n/index.js";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { useFloating, flip, shift, offset, size, autoUpdate } from "@floating-ui/vue";

defineProps({
    modelValue: {
        type: [String, Object],
        default: null,
    },
    fields: {
        type: Array,
        default: () => [],
    },
    type: {
        type: String,
        required: true,
        validator: (v) => ["selection", "color"].includes(v),
    },
});
const emit = defineEmits(["update:modelValue"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    middleware: [
        offset(4),
        flip(),
        shift({ padding: 8 }),
        size({
            apply({ rects, elements }) {
                elements.floating.style.width = `${rects.reference.width}px`;
            },
        }),
    ],
    whileElementsMounted: autoUpdate,
});
</script>
