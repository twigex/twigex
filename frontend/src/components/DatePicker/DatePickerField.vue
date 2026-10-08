<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover as="div" class="relative">
        <PopoverButton
            :id="id"
            ref="reference"
            class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-16 text-left shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
        >
            <span :class="[modelValue ? 'text-gray-900' : 'text-gray-400', 'block truncate']">
                {{ shown || placeholder }}
            </span>
            <span class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                <CalendarDaysIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
            </span>
        </PopoverButton>
        <button
            v-if="modelValue"
            type="button"
            class="absolute inset-y-0 right-8 flex items-center px-1 text-gray-400 hover:text-gray-600"
            @click="emit('update:modelValue', '')"
        >
            <span class="sr-only">{{ t("common.button.clear") }}</span>
            <XMarkIcon class="h-4 w-4" aria-hidden="true" />
        </button>
        <Teleport to="body">
            <PopoverPanel
                ref="floating"
                v-slot="{ close }"
                :style="floatingStyles"
                class="z-[100] rounded-xl bg-white p-3 shadow-lg ring-1 ring-black/5"
            >
                <DatePicker
                    :model-value="modelValue"
                    @update:model-value="
                        (value) => {
                            emit('update:modelValue', value);
                            close();
                        }
                    "
                />
            </PopoverPanel>
        </Teleport>
    </Popover>
</template>

<script setup>
import { computed, ref } from "vue";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { CalendarDaysIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { autoUpdate, flip, offset, shift, useFloating } from "@floating-ui/vue";
import { t } from "@/i18n/index.js";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import useDateOperations from "@/composables/useDateOperations.js";

const props = defineProps({
    modelValue: { type: String, default: "" },
    placeholder: { type: String, default: "" },
    id: { type: String, default: undefined },
});
const emit = defineEmits(["update:modelValue"]);

const { getDate, getTimestampFromDateString } = useDateOperations();

const shown = computed(() =>
    props.modelValue ? getDate(getTimestampFromDateString(props.modelValue)) : "",
);

const reference = ref(null);
const floating = ref(null);
const { floatingStyles } = useFloating(reference, floating, {
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift({ padding: 8 })],
    whileElementsMounted: autoUpdate,
});
</script>
