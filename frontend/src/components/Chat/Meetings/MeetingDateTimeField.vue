<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover as="div" class="relative">
        <PopoverButton
            ref="reference"
            class="relative w-full cursor-default rounded-md bg-white py-2 pl-3 pr-10 text-left text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600"
        >
            <span class="block truncate">{{ label }}</span>
            <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                <CalendarDaysIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </span>
        </PopoverButton>
        <Teleport to="body">
            <PopoverPanel
                ref="floating"
                :style="floatingStyles"
                class="z-[100] rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5"
            >
                <DatePicker
                    :model-value="modelValue"
                    with-time
                    @update:model-value="(v) => emit('update:modelValue', v)"
                />
            </PopoverPanel>
        </Teleport>
    </Popover>
</template>

<script setup>
import { ref, computed } from "vue";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { CalendarDaysIcon } from "@heroicons/vue/24/outline";
import { useFloating, flip, shift, offset, autoUpdate } from "@floating-ui/vue";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import { t } from "@/i18n/index";

const { getDateAndTime } = useDateOperations();

const props = defineProps({
    // "YYYY-MM-DDTHH:MM"
    modelValue: {
        type: String,
        default: "",
    },
    // Optional override; falls back to a translated default when empty.
    placeholder: {
        type: String,
        default: "",
    },
});
const emit = defineEmits(["update:modelValue"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    // Prefer below, then above; only when neither vertical fits, fall back to
    // the side (right, then left). shift() keeps it within the viewport.
    placement: "bottom-start",
    strategy: "fixed",
    middleware: [
        offset(4),
        flip({
            fallbackPlacements: ["top-start", "right-start", "left-start"],
        }),
        shift({ padding: 8 }),
    ],
    whileElementsMounted: autoUpdate,
});

const label = computed(() => {
    if (!props.modelValue)
        return props.placeholder || t.value("meetings.schedule.datetime_placeholder");
    const ms = new Date(props.modelValue).getTime();

    if (Number.isNaN(ms)) return props.modelValue;

    return getDateAndTime(Math.floor(ms / 1000));
});
</script>
