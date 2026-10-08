<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover as="div" class="flex-1 relative">
        <PopoverButton
            ref="reference"
            class="w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
        >
            <span class="block truncate">{{ modelValue || "Select date" }}</span>
            <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                <CalendarDaysIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </span>
        </PopoverButton>
        <Teleport to="body">
            <PopoverPanel
                ref="floating"
                :style="floatingStyles"
                class="z-[100] bg-white rounded-xl shadow-lg ring-1 ring-black/5 p-3"
                v-slot="{ close }"
            >
                <DatePicker
                    :modelValue="modelValue"
                    @update:modelValue="
                        (val) => {
                            emit('update:modelValue', val);
                            close();
                        }
                    "
                />
            </PopoverPanel>
        </Teleport>
    </Popover>
</template>

<script setup>
import { ref } from "vue";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { CalendarDaysIcon } from "@heroicons/vue/24/outline";
import { useFloating, flip, shift, offset, autoUpdate } from "@floating-ui/vue";
import DatePicker from "@/components/DatePicker/DatePicker.vue";

defineProps({
    modelValue: String,
});
const emit = defineEmits(["update:modelValue"]);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    middleware: [offset(4), flip(), shift({ padding: 8 })],
    whileElementsMounted: autoUpdate,
});
</script>
