<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }">
        <DisclosureButton
            class="flex w-full justify-between px-4 py-2 text-left text-sm font-medium text-gray-700"
            :class="open ? 'bg-indigo-50 hover:bg-indigo-100' : 'bg-gray-50 hover:bg-gray-100'"
        >
            <div>
                <span class="mr-2">{{ t("collimato.charts.new_chart.filters.title") }}</span>
                <span
                    v-if="filters.length > 0"
                    class="ml-auto w-9 min-w-max whitespace-nowrap rounded-full bg-white px-2.5 py-0.5 text-center text-xs font-medium leading-5 text-gray-600 ring-1 ring-inset ring-gray-200"
                    aria-hidden="true"
                    >{{ filters.length }}</span
                >
            </div>
            <ChevronUpIcon
                :class="open ? 'rotate-180 transform' : ''"
                class="h-5 w-5 text-purple-500"
            />
        </DisclosureButton>
        <DisclosurePanel class="px-4 pb-2 pt-4 text-sm text-gray-500">
            <FilterList
                :filters="filters"
                :cubes="cubes"
                :title="t('collimato.charts.new_chart.filters.title')"
                @add-filter="$emit('add-filter', $event)"
                @update-filter="(member, filter) => $emit('update-filter', member, filter)"
                @remove-filter="$emit('remove-filter', $event)"
            />
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import FilterList from "@/components/Collimato/Charts/Sidebar/FilterList.vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import { ChevronUpIcon } from "@heroicons/vue/20/solid";

defineProps({
    filters: {
        type: Array,
        default: () => [],
    },
    cubes: {
        type: Array,
        default: () => [],
    },
});

defineEmits(["add-filter", "update-filter", "remove-filter"]);
</script>
