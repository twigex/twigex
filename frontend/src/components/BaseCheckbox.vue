<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex gap-3">
        <div class="flex h-6 shrink-0 items-center">
            <div class="group grid size-4 grid-cols-1">
                <input
                    :id="id"
                    :checked="modelValue"
                    :aria-describedby="description ? `${id}-description` : undefined"
                    type="checkbox"
                    class="col-start-1 row-start-1 appearance-none rounded border border-gray-300 bg-white checked:border-indigo-600 checked:bg-indigo-600 indeterminate:border-indigo-600 indeterminate:bg-indigo-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:border-gray-300 disabled:bg-gray-100 disabled:checked:bg-gray-100 forced-colors:appearance-auto"
                    @change="$emit('update:modelValue', $event.target.checked)"
                />
                <svg
                    class="pointer-events-none col-start-1 row-start-1 size-3.5 self-center justify-self-center stroke-white group-has-[:disabled]:stroke-gray-950/25"
                    viewBox="0 0 14 14"
                    fill="none"
                >
                    <path
                        class="opacity-0 group-has-[:checked]:opacity-100"
                        d="M3 8L6 11L11 3.5"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    />
                    <path
                        class="opacity-0 group-has-[:indeterminate]:opacity-100"
                        d="M3 7H11"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                    />
                </svg>
            </div>
        </div>
        <div class="text-sm/6">
            <label :for="id" class="font-medium text-gray-900">{{ label }}</label>
            <p v-if="description" :id="`${id}-description`" class="text-gray-500">
                {{ description }}
            </p>
        </div>
    </div>
</template>

<script setup>
import { useId } from "vue";

defineProps({
    modelValue: {
        type: Boolean,
        default: false,
    },
    label: {
        type: String,
        required: true,
    },
    description: {
        type: String,
        default: "",
    },
});

defineEmits(["update:modelValue"]);

const id = useId();
</script>
