<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Popover class="relative">
        <PopoverButton
            class="flex items-center gap-x-1 text-sm font-semibold leading-6 text-gray-900"
        >
            <button
                type="button"
                class="rounded-full p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
            >
                <BarsArrowDownIcon class="h-5 w-5" aria-hidden="true" />
            </button>
        </PopoverButton>

        <transition
            enter-active-class="transition ease-out duration-200"
            enter-from-class="opacity-0 translate-y-1"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition ease-in duration-150"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 translate-y-1"
        >
            <PopoverPanel class="absolute -left-20 z-50 mt-5 flex w-64 -translate-x-1/2 px-4">
                <div
                    class="w-screen max-w-md flex-auto overflow-hidden rounded-md bg-white text-sm leading-6 shadow-lg ring-1 ring-gray-900/5"
                >
                    <div class="p-2">
                        <div
                            v-for="item in state.sortOptions"
                            :key="item.name"
                            class="group relative flex items-center gap-x-6 rounded-lg p-3"
                            :class="item.order == 'none' ? 'hover:bg-gray-50' : 'bg-indigo-50'"
                        >
                            <div
                                class="mt-1 flex h-5 w-5 flex-none items-center justify-center rounded-lg bg-gray-50 group-hover:bg-white"
                            >
                                <component
                                    :is="item.icon"
                                    class="h-6 w-6 text-gray-600 group-hover:text-indigo-600"
                                    aria-hidden="true"
                                />
                            </div>
                            <div>
                                <a href="#" class="text-gray-900" @click="setSort(item)">
                                    {{ item.text }}
                                    <span class="absolute inset-0" />
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </PopoverPanel>
        </transition>
    </Popover>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { reactive } from "vue";
import { useFilesStore } from "@/store/files";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { BarsArrowDownIcon } from "@heroicons/vue/20/solid";
import { ArrowUpIcon, ArrowDownIcon } from "@heroicons/vue/24/outline";

const fileStore = useFilesStore();

const state = reactive({
    sortOptions: [
        {
            type: "date",
            order: "none",
            text: t.value("files.sort.sort_by_created"),
            icon: ArrowDownIcon,
            defaultIcon: ArrowDownIcon,
        },
        {
            type: "modified",
            order: "none",
            text: t.value("files.sort.sort_by_modified"),
            icon: ArrowDownIcon,
            defaultIcon: ArrowDownIcon,
        },
        {
            type: "size",
            order: "none",
            text: t.value("files.sort.sort_by_filesize"),
            icon: ArrowDownIcon,
            defaultIcon: ArrowDownIcon,
        },
        {
            type: "name",
            order: "none",
            text: t.value("files.sort.sort_by_filename"),
            icon: ArrowDownIcon,
            defaultIcon: ArrowDownIcon,
        },
        {
            type: "type",
            order: "none",
            text: t.value("files.sort.sort_by_filetype"),
            icon: ArrowDownIcon,
            defaultIcon: ArrowDownIcon,
        },
    ],
});

function setSort(item) {
    for (let i = 0; i < state.sortOptions.length; i++) {
        if (item.type != state.sortOptions[i].type) {
            state.sortOptions[i].icon = state.sortOptions[i].defaultIcon;
            state.sortOptions[i].order = "none";
        }
    }

    switch (item.order) {
        case "none":
            item.order = "desc";
            item.icon = ArrowDownIcon;
            break;
        case "desc":
            item.order = "asc";
            item.icon = ArrowUpIcon;
            break;
        case "asc":
            item.order = "none";
            item.icon = ArrowDownIcon;
            break;

        default:
            break;
    }

    fileStore.setFileOrder(item);
}
</script>
