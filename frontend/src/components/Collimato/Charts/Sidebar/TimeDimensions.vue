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
                <span class="mr-2">{{ t(`${TIME}title`) }}</span>
            </div>
            <ChevronUpIcon
                :class="open ? 'rotate-180 transform' : ''"
                class="h-5 w-5 text-purple-500"
            />
        </DisclosureButton>
        <DisclosurePanel class="px-4 pb-2 pt-4 text-sm text-gray-500">
            <div>
                <div class="flex flex-row justify-between">
                    <h4 class="text-small font-semibold leading-4 text-gray-400">
                        {{ t(`${TIME}title`) }}
                    </h4>

                    <button
                        @click="clearTime()"
                        type="button"
                        class="font-semibold text-indigo-600 hover:text-indigo-500"
                    >
                        {{ t("common.button.clear") }}
                    </button>
                </div>

                <div
                    class="mt-2 divide-y divide-gray-100 border-t border-gray-200 text-sm leading-6"
                >
                    <Listbox
                        as="div"
                        :model-value="time.dimension"
                        @update:model-value="patch({ dimension: $event })"
                    >
                        <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">{{
                            t(`${TIME}dimension`)
                        }}</ListboxLabel>
                        <div class="relative mt-2">
                            <ListboxButton
                                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            >
                                <span class="block truncate">{{
                                    time.dimension
                                        ? time.dimension.title
                                        : t(`${TIME}select_dimension`)
                                }}</span>
                                <span
                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                >
                                    <ChevronUpDownIcon
                                        class="h-5 w-5 text-gray-400"
                                        aria-hidden="true"
                                    />
                                </span>
                            </ListboxButton>

                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                >
                                    <!-- Group Iteration -->
                                    <template v-for="(d, index) in cubes" :key="index">
                                        <li
                                            class="text-gray-900 text-left cursor-default select-none relative py-2 pl-3 pr-9 font-semibold bg-gray-100"
                                        >
                                            {{ d.table }}
                                        </li>
                                        <!-- Person Iteration -->
                                        <ListboxOption
                                            as="li"
                                            v-for="dimension in d.dimensions"
                                            :key="dimension.name"
                                            :value="dimension"
                                            v-slot="{ active, selected }"
                                        >
                                            <li
                                                v-if="dimension.type == 'time'"
                                                :class="[
                                                    active
                                                        ? 'bg-indigo-600 text-white'
                                                        : 'text-gray-900',
                                                    'relative cursor-default select-none py-2 pl-3 pr-9',
                                                ]"
                                            >
                                                <span
                                                    :class="[
                                                        selected ? 'font-semibold' : 'font-normal',
                                                        'block truncate ml-5',
                                                    ]"
                                                    >{{ dimension.title }}</span
                                                >

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
                                    </template>
                                </ListboxOptions>
                            </transition>
                        </div>
                    </Listbox>

                    <Listbox
                        as="div"
                        :model-value="time.dateRange"
                        @update:model-value="patch({ dateRange: $event })"
                    >
                        <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">{{
                            t(`${TIME}date_range`)
                        }}</ListboxLabel>
                        <div class="relative mt-2">
                            <ListboxButton
                                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            >
                                <span class="block truncate">{{
                                    time.dateRange == ""
                                        ? t(`${TIME}select_date_range`)
                                        : selectedRangeLabel
                                }}</span>
                                <span
                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                >
                                    <ChevronUpDownIcon
                                        class="h-5 w-5 text-gray-400"
                                        aria-hidden="true"
                                    />
                                </span>
                            </ListboxButton>

                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                >
                                    <ListboxOption
                                        as="template"
                                        v-for="range in timeOptions"
                                        :key="range.value"
                                        :value="range.value"
                                        v-slot="{ active, selected }"
                                    >
                                        <li
                                            :class="[
                                                active
                                                    ? 'bg-indigo-600 text-white'
                                                    : 'text-gray-900',
                                                'relative cursor-default select-none py-2 pl-3 pr-9',
                                            ]"
                                        >
                                            <span
                                                :class="[
                                                    selected ? 'font-semibold' : 'font-normal',
                                                    'block truncate',
                                                ]"
                                                >{{ range.label }}</span
                                            >

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
                        </div>
                    </Listbox>

                    <div v-if="time.dateRange == 'custom'">
                        <div>
                            <label
                                for="start_time"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t(`${TIME}start_time`) }}</label
                            >
                            <div class="mt-2">
                                <input
                                    :value="time.startTime"
                                    @input="patch({ startTime: $event.target.value })"
                                    @click="startDialog = true"
                                    type="text"
                                    name="start_time"
                                    id="start_time"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                    :placeholder="
                                        t(
                                            'collimato.charts.new_chart.filters.filter_dialog.select_start_date',
                                        )
                                    "
                                />
                            </div>
                        </div>

                        <TransitionRoot as="template" :show="startDialog">
                            <Dialog as="div" class="relative z-50" @close="startDialog = false">
                                <TransitionChild
                                    as="template"
                                    enter="ease-out duration-300"
                                    enter-from="opacity-0"
                                    enter-to="opacity-100"
                                    leave="ease-in duration-200"
                                    leave-from="opacity-100"
                                    leave-to="opacity-0"
                                >
                                    <div
                                        class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
                                    />
                                </TransitionChild>

                                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                                    <div
                                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                                    >
                                        <TransitionChild
                                            as="template"
                                            enter="ease-out duration-300"
                                            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                                            enter-to="opacity-100 translate-y-0 sm:scale-100"
                                            leave="ease-in duration-200"
                                            leave-from="opacity-100 translate-y-0 sm:scale-100"
                                            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                                        >
                                            <DialogPanel
                                                class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                                            >
                                                <div>
                                                    <div class="text-center sm:mt-5">
                                                        <DialogTitle
                                                            as="h3"
                                                            class="text-base text-left font-semibold leading-6 text-gray-900"
                                                            >{{
                                                                t(`${TIME}choose_start_date`)
                                                            }}</DialogTitle
                                                        >
                                                        <div class="mx-8 mt-5">
                                                            <DatePicker
                                                                :model-value="time.startTime"
                                                                @update:model-value="
                                                                    patch({
                                                                        startTime: $event,
                                                                    })
                                                                "
                                                            />
                                                        </div>
                                                    </div>
                                                </div>
                                                <div
                                                    class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse"
                                                >
                                                    <button
                                                        type="button"
                                                        class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                                        @click="startDialog = false"
                                                    >
                                                        {{ t("common.button.save") }}
                                                    </button>
                                                    <button
                                                        type="button"
                                                        class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                                        @click="startDialog = false"
                                                    >
                                                        {{ t("common.button.cancel") }}
                                                    </button>
                                                </div>
                                            </DialogPanel>
                                        </TransitionChild>
                                    </div>
                                </div>
                            </Dialog>
                        </TransitionRoot>

                        <div>
                            <label
                                for="end_time"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t(`${TIME}end_time`) }}</label
                            >
                            <div class="mt-2">
                                <input
                                    :value="time.endTime"
                                    @input="patch({ endTime: $event.target.value })"
                                    @click="endDialog = true"
                                    type="text"
                                    name="end_time"
                                    id="end_time"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                    :placeholder="
                                        t(
                                            'collimato.charts.new_chart.filters.filter_dialog.select_end_date',
                                        )
                                    "
                                />
                            </div>
                        </div>

                        <TransitionRoot as="template" :show="endDialog">
                            <Dialog as="div" class="relative z-50" @close="endDialog = false">
                                <TransitionChild
                                    as="template"
                                    enter="ease-out duration-300"
                                    enter-from="opacity-0"
                                    enter-to="opacity-100"
                                    leave="ease-in duration-200"
                                    leave-from="opacity-100"
                                    leave-to="opacity-0"
                                >
                                    <div
                                        class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
                                    />
                                </TransitionChild>

                                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                                    <div
                                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                                    >
                                        <TransitionChild
                                            as="template"
                                            enter="ease-out duration-300"
                                            enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                                            enter-to="opacity-100 translate-y-0 sm:scale-100"
                                            leave="ease-in duration-200"
                                            leave-from="opacity-100 translate-y-0 sm:scale-100"
                                            leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                                        >
                                            <DialogPanel
                                                class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                                            >
                                                <div>
                                                    <div class="text-center sm:mt-5">
                                                        <DialogTitle
                                                            as="h3"
                                                            class="text-base text-left font-semibold leading-6 text-gray-900"
                                                            >{{
                                                                t(`${TIME}choose_end_date`)
                                                            }}</DialogTitle
                                                        >
                                                        <div class="mx-8 mt-5">
                                                            <DatePicker
                                                                :model-value="time.endTime"
                                                                @update:model-value="
                                                                    patch({
                                                                        endTime: $event,
                                                                    })
                                                                "
                                                            />
                                                        </div>
                                                    </div>
                                                </div>
                                                <div
                                                    class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse"
                                                >
                                                    <button
                                                        type="button"
                                                        class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                                        @click="endDialog = false"
                                                    >
                                                        {{ t("common.button.save") }}
                                                    </button>
                                                    <button
                                                        type="button"
                                                        class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                                        @click="endDialog = false"
                                                    >
                                                        {{ t("common.button.cancel") }}
                                                    </button>
                                                </div>
                                            </DialogPanel>
                                        </TransitionChild>
                                    </div>
                                </div>
                            </Dialog>
                        </TransitionRoot>
                    </div>
                    <Listbox
                        as="div"
                        :model-value="time.granularity"
                        @update:model-value="patch({ granularity: $event })"
                    >
                        <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">{{
                            t(`${TIME}granularity`)
                        }}</ListboxLabel>
                        <div class="relative mt-2">
                            <ListboxButton
                                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            >
                                <span class="block truncate">{{
                                    time.granularity == "" || time.granularity == null
                                        ? t(`${TIME}select_granularity`)
                                        : selectedGranularityLabel
                                }}</span>
                                <span
                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                >
                                    <ChevronUpDownIcon
                                        class="h-5 w-5 text-gray-400"
                                        aria-hidden="true"
                                    />
                                </span>
                            </ListboxButton>

                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                >
                                    <ListboxOption
                                        as="template"
                                        v-for="g in granularityOptions"
                                        :key="g.value"
                                        :value="g.value"
                                        v-slot="{ active, selected }"
                                    >
                                        <li
                                            :class="[
                                                active
                                                    ? 'bg-indigo-600 text-white'
                                                    : 'text-gray-900',
                                                'relative cursor-default select-none py-2 pl-3 pr-9',
                                            ]"
                                        >
                                            <span
                                                :class="[
                                                    selected ? 'font-semibold' : 'font-normal',
                                                    'block truncate',
                                                ]"
                                                >{{ g.label }}</span
                                            >

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
                        </div>
                    </Listbox>
                </div>
            </div>
        </DisclosurePanel>
    </Disclosure>
</template>

<script setup>
import { ref, computed } from "vue";
import { t } from "@/i18n/index.js";

const TIME = "collimato.charts.new_chart.time.";

import DatePicker from "@/components/DatePicker/DatePicker.vue";
import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
} from "@headlessui/vue";
import { ChevronUpIcon, CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    time: {
        type: Object,
        default: () => ({
            dimension: null,
            dateRange: "",
            granularity: "",
            startTime: "",
            endTime: "",
        }),
    },
    cubes: {
        type: Array,
        default: () => [],
    },
});

const emit = defineEmits(["update:time"]);

const RANGE_VALUES = [
    ["today", "today"],
    ["yesterday", "yesterday"],
    ["this week", "this_week"],
    ["this month", "this_month"],
    ["this quarter", "this_quarter"],
    ["this year", "this_year"],
    ["last quarter", "last_quarter"],
    ["last year", "last_year"],
    ["custom", "custom"],
];

const GRANULARITY_VALUES = ["second", "minute", "hour", "day", "week", "month", "quarter", "year"];

const timeOptions = computed(() =>
    RANGE_VALUES.map(([value, key]) => ({
        value: value,
        label: t.value(`${TIME}range.${key}`),
    })),
);

const granularityOptions = computed(() =>
    GRANULARITY_VALUES.map((value) => ({
        value: value,
        label: t.value(`${TIME}granularity.${value}`),
    })),
);

const selectedRangeLabel = computed(
    () =>
        timeOptions.value.find((o) => o.value == props.time.dateRange)?.label ??
        props.time.dateRange,
);

const selectedGranularityLabel = computed(
    () =>
        granularityOptions.value.find((o) => o.value == props.time.granularity)?.label ??
        props.time.granularity,
);

const startDialog = ref(false);
const endDialog = ref(false);

function patch(changes) {
    emit("update:time", changes);
}

function clearTime() {
    patch({
        dimension: null,
        dateRange: "",
        granularity: "",
        startTime: "",
        endTime: "",
    });
}
</script>
