<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="bg-white border-b text-white p-1 flex items-center justify-between">
        <div class="flex flex-row gap-x-2">
            <ToolbarDropdown @files-selected="handleFilesSelected" />
            <nav class="hidden sm:flex border-l pl-2" aria-label="Breadcrumb">
                <ol role="list" class="flex items-center">
                    <template v-if="pages.length <= 3">
                        <li
                            v-for="(page, index) in pages"
                            :key="page.name + index"
                            class="cursor-pointer"
                            @click="pageClicked(page)"
                        >
                            <div class="flex items-center truncate">
                                <ChevronRightIcon
                                    v-if="index > 0"
                                    class="size-5 shrink-0 text-gray-400"
                                    aria-hidden="true"
                                />
                                <a class="text-sm font-medium text-gray-500 hover:text-gray-700">{{
                                    page.name
                                }}</a>
                            </div>
                        </li>
                    </template>
                    <div class="flex flex-row items-center" v-else>
                        <li class="cursor-pointer" @click="pageClicked(pages[0])">
                            <div class="flex items-center truncate">
                                <a class="text-sm font-medium text-gray-500 hover:text-gray-700">{{
                                    pages[0].name
                                }}</a>

                                <ChevronRightIcon
                                    class="size-5 shrink-0 text-gray-400"
                                    aria-hidden="true"
                                />
                            </div>
                        </li>

                        <DropdownMenu
                            :items="middlePages"
                            size="md"
                            @select="(item) => pageClicked(item.page)"
                        >
                            <template #button>
                                <div
                                    class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-3 py-1 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    ...
                                </div>
                            </template>
                        </DropdownMenu>

                        <li class="cursor-pointer" @click="pageClicked(pages[pages.length - 1])">
                            <div class="flex items-center truncate">
                                <ChevronRightIcon
                                    class="size-5 shrink-0 text-gray-400"
                                    aria-hidden="true"
                                />
                                <a class="text-sm font-medium text-gray-500 hover:text-gray-700">{{
                                    pages[pages.length - 1].name
                                }}</a>
                            </div>
                        </li>
                    </div>
                </ol>
            </nav>
            <nav
                v-if="pages.length"
                class="flex min-w-0 items-center border-l pl-2 sm:hidden"
                aria-label="Breadcrumb"
            >
                <button
                    v-if="pages.length > 1"
                    type="button"
                    class="-ml-1 mr-1 flex shrink-0 items-center justify-center rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
                    @click="pageClicked(pages[pages.length - 2])"
                >
                    <span class="sr-only">Back</span>
                    <ChevronLeftIcon class="size-5" aria-hidden="true" />
                </button>
                <span class="truncate text-sm font-medium text-gray-700">
                    {{ pages[pages.length - 1].name }}
                </span>
            </nav>
        </div>
        <div class="flex max-h-9 flex-row items-center gap-x-1 align-middle">
            <slot name="actions"></slot>
            <div class="mx-1 flex rounded-md bg-gray-100 p-0.5">
                <button
                    type="button"
                    @click="emits('update:viewMode', 'grid')"
                    :class="[
                        'rounded p-1',
                        viewMode === 'grid'
                            ? 'bg-white text-indigo-600 shadow-sm'
                            : 'text-gray-500 hover:text-gray-700',
                    ]"
                >
                    <Squares2X2Icon class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    @click="emits('update:viewMode', 'list')"
                    :class="[
                        'rounded p-1',
                        viewMode === 'list'
                            ? 'bg-white text-indigo-600 shadow-sm'
                            : 'text-gray-500 hover:text-gray-700',
                    ]"
                >
                    <Bars3Icon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>
            <SortDropdown class="mr-1" />
            <button
                @click="closeDetails"
                type="button"
                class="rounded-full p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
            >
                <InformationCircleIcon class="h-5 w-5" aria-hidden="true" />
            </button>
        </div>
    </div>
</template>

<script setup>
import { computed } from "vue";
import ToolbarDropdown from "@/components/Files/Toolbar/ToolbarDropdown.vue";
import SortDropdown from "@/components/Files/Toolbar/SortDropdown.vue";
import DropdownMenu from "@/components/DropdownMenu.vue";
import { useDetailsStore } from "@/store/details";
import {
    InformationCircleIcon,
    ChevronRightIcon,
    ChevronLeftIcon,
    Squares2X2Icon,
    Bars3Icon,
} from "@heroicons/vue/20/solid";

const props = defineProps({
    pages: {
        type: Array,
        default: () => [],
    },
    viewMode: {
        type: String,
        default: "list",
    },
});

const emits = defineEmits(["filesSelected", "pageClicked", "update:viewMode"]);
const detailsStore = useDetailsStore();

const middlePages = computed(() =>
    props.pages.slice(1, props.pages.length - 1).map((page) => ({ label: page.name, page })),
);

function closeDetails() {
    detailsStore.open = !detailsStore.open;
    detailsStore.mobileOpen = !detailsStore.mobileOpen;
}

function handleFilesSelected(files) {
    emits("filesSelected", files);
}

function pageClicked(page) {
    emits("pageClicked", page);
}
</script>
