<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Listbox as="div" :model-value="modelValue" @update:model-value="pick($event)">
        <ListboxLabel class="block text-sm/6 font-medium text-gray-900">{{ label }}</ListboxLabel>
        <p v-if="description" class="text-sm leading-6 text-gray-500">
            {{ description }}
        </p>
        <div class="relative mt-2">
            <ListboxButton
                class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1 pl-3 pr-2 text-left text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600 sm:text-sm/6"
            >
                <span class="col-start-1 row-start-1 truncate pr-6">{{ selectedLabel }}</span>
                <ChevronUpDownIcon
                    class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                    aria-hidden="true"
                />
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
                        v-for="option in options"
                        :key="option.value"
                        :value="option.value"
                        v-slot="{ active, selected }"
                    >
                        <li
                            :class="[
                                active ? 'bg-indigo-600 text-white outline-none' : 'text-gray-900',
                                'relative cursor-default select-none py-2 pl-3 pr-9',
                            ]"
                        >
                            <span
                                :class="[
                                    selected ? 'font-semibold' : 'font-normal',
                                    'block truncate',
                                ]"
                                >{{ option.label }}</span
                            >
                            <span
                                v-if="selected"
                                :class="[
                                    active ? 'text-white' : 'text-indigo-600',
                                    'absolute inset-y-0 right-0 flex items-center pr-4',
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
import { computed } from "vue";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon } from "@heroicons/vue/20/solid";
import { ChevronUpDownIcon } from "@heroicons/vue/16/solid";

const props = defineProps({
    modelValue: {
        type: [String, Number],
        default: "",
    },
    label: {
        type: String,
        required: true,
    },
    description: {
        type: String,
        default: "",
    },
    options: {
        type: Array,
        required: true,
    },
});

const emit = defineEmits(["update:modelValue", "change"]);

const selectedLabel = computed(() => {
    const option = props.options.find((o) => o.value == props.modelValue);

    if (!option) {
        return props.modelValue;
    }

    return option.label;
});

function pick(value) {
    emit("update:modelValue", value);
    emit("change", value);
}
</script>
