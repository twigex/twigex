<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SwitchGroup as="div" class="flex items-center justify-between gap-x-3">
        <SwitchLabel
            v-if="label"
            as="span"
            class="min-w-0 flex-1 truncate text-sm text-gray-900"
            passive
        >
            {{ label }}
        </SwitchLabel>
        <Switch
            :model-value="modelValue"
            :disabled="disabled"
            :class="[
                modelValue ? 'bg-indigo-600' : 'bg-gray-200',
                disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
                'relative inline-flex h-5 w-9 flex-shrink-0 rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2',
            ]"
            @update:model-value="emit('update:modelValue', $event)"
        >
            <span v-if="!label" class="sr-only">{{ srLabel }}</span>
            <span
                aria-hidden="true"
                :class="[
                    modelValue ? 'translate-x-4' : 'translate-x-0',
                    'pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                ]"
            />
        </Switch>
    </SwitchGroup>
</template>

<script setup>
import { Switch, SwitchGroup, SwitchLabel } from "@headlessui/vue";

defineProps({
    modelValue: { type: Boolean, default: false },
    label: { type: String, default: "" },
    srLabel: { type: String, default: "" },
    disabled: { type: Boolean, default: false },
});

const emit = defineEmits(["update:modelValue"]);
</script>
