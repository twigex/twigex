<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div class="grid grid-cols-1 sm:hidden">
            <select
                :aria-label="t('collimato.charts.new_chart.tabs.select')"
                :value="modelValue"
                @change="emit('update:modelValue', $event.target.value)"
                class="col-start-1 row-start-1 w-full appearance-none rounded-md bg-white py-2 pl-3 pr-8 text-base text-gray-900 outline outline-1 -outline-offset-1 outline-gray-300 focus:outline focus:outline-2 focus:-outline-offset-2 focus:outline-indigo-600"
            >
                <option v-for="tab in tabs" :key="tab.value" :value="tab.value">
                    {{ tab.name }}
                </option>
            </select>
            <ChevronDownIcon
                class="pointer-events-none col-start-1 row-start-1 mr-2 size-5 self-center justify-self-end fill-gray-500"
                aria-hidden="true"
            />
        </div>
        <div class="hidden sm:block">
            <div class="border-b border-gray-200">
                <nav class="-mb-px flex" :aria-label="t('collimato.charts.new_chart.tabs.label')">
                    <button
                        v-for="tab in tabs"
                        :key="tab.value"
                        type="button"
                        @click="emit('update:modelValue', tab.value)"
                        :class="[
                            modelValue === tab.value
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                            'w-full border-b-2 px-1 py-4 text-center text-sm font-medium',
                        ]"
                        :aria-current="modelValue === tab.value ? 'page' : undefined"
                    >
                        {{ tab.name }}
                    </button>
                </nav>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed } from "vue";
import { ChevronDownIcon } from "@heroicons/vue/16/solid";

defineProps({
    modelValue: {
        type: String,
        required: true,
    },
});

const emit = defineEmits(["update:modelValue"]);

const tabs = computed(() => [
    { name: t.value("collimato.charts.new_chart.tabs.data"), value: "data" },
    {
        name: t.value("collimato.charts.new_chart.tabs.customize"),
        value: "customize",
    },
]);
</script>
