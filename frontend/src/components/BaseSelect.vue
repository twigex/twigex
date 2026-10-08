<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Listbox
        as="div"
        :model-value="modelValue"
        :disabled="disabled"
        @update:model-value="$emit('update:modelValue', $event)"
    >
        <ListboxLabel v-if="label" class="block text-sm font-medium leading-6 text-gray-900">
            {{ label }}
        </ListboxLabel>
        <div class="relative" :class="label ? 'mt-2' : ''">
            <ListboxButton
                ref="button"
                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset focus:outline-none focus:ring-2 disabled:bg-gray-50 disabled:text-gray-400 sm:leading-6"
                :class="
                    error
                        ? 'ring-red-600 focus:ring-red-600'
                        : 'ring-gray-300 focus:ring-indigo-600'
                "
            >
                <span class="block truncate" :class="selectedOption ? '' : 'text-gray-400'">
                    <slot name="selected" :option="selectedOption">{{
                        selectedLabel || "\u00A0"
                    }}</slot>
                </span>
                <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                    <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                </span>
            </ListboxButton>

            <Portal>
                <transition
                    leave-active-class="transition ease-in duration-100"
                    leave-from-class="opacity-100"
                    leave-to-class="opacity-0"
                >
                    <ListboxOptions
                        ref="floating"
                        :style="floatingStyles"
                        class="z-[60] max-h-60 overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                    >
                        <ListboxOption
                            as="template"
                            v-for="option in normalised"
                            :key="option.value"
                            :value="option.value"
                            v-slot="{ active, selected }"
                        >
                            <li
                                :class="[
                                    active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                    'relative cursor-default select-none py-2 pl-3 pr-9',
                                ]"
                            >
                                <slot
                                    name="option"
                                    :option="option"
                                    :active="active"
                                    :selected="selected"
                                >
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'block truncate',
                                        ]"
                                        >{{ option.label }}</span
                                    >
                                </slot>
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
            </Portal>
        </div>
    </Listbox>
</template>

<script setup>
import { computed, ref } from "vue";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    Portal,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";

const props = defineProps({
    modelValue: {
        type: [String, Number],
        default: "",
    },
    options: {
        type: Array,
        default: () => [],
    },
    label: {
        type: String,
        default: "",
    },
    placeholder: {
        type: String,
        default: "",
    },
    disabled: {
        type: Boolean,
        default: false,
    },
    error: {
        type: Boolean,
        default: false,
    },
});

defineEmits(["update:modelValue"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({ matchWidth: true, anchor: button });

const normalised = computed(() =>
    props.options.map((option) =>
        typeof option === "object" ? option : { value: option, label: String(option) },
    ),
);

const selectedOption = computed(() =>
    normalised.value.find((option) => option.value === props.modelValue),
);

const selectedLabel = computed(() => selectedOption.value?.label ?? props.placeholder);
</script>
