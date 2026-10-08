<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex items-center justify-between border-t border-gray-200 bg-white px-4 py-3 sm:px-6"
    >
        <!-- Mobile Pagination -->
        <div class="flex flex-1 justify-between sm:hidden">
            <a
                href="#"
                @click.prevent="changePage(currentPage - 1)"
                class="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                :class="{
                    'cursor-not-allowed text-gray-300': currentPage <= 1,
                }"
            >
                {{ t("pagination.previous") }}
            </a>
            <a
                href="#"
                @click.prevent="changePage(currentPage + 1)"
                class="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
                :class="{
                    'cursor-not-allowed text-gray-300': currentPage >= totalPages,
                }"
            >
                {{ t("pagination.next") }}
            </a>
        </div>

        <!-- Desktop Pagination -->
        <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
            <div>
                <p class="text-sm text-gray-700">
                    {{ t("pagination.showing") }}
                    <span class="font-medium">{{ startItem }}</span>
                    {{ t("pagination.to") }}
                    <span class="font-medium">{{ endItem }}</span>
                    {{ t("pagination.of") }}
                    <span class="font-medium">{{ totalItems }}</span>
                    {{ t("pagination.results") }}
                </p>
            </div>

            <div>
                <nav
                    class="isolate inline-flex -space-x-px rounded-md shadow-sm"
                    aria-label="Pagination"
                >
                    <!-- Previous Button -->
                    <a
                        href="#"
                        @click.prevent="changePage(currentPage - 1)"
                        class="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20"
                        :class="{
                            'cursor-not-allowed text-gray-300': currentPage <= 1,
                        }"
                    >
                        <span class="sr-only">{{ t("pagination.previous") }}</span>
                        <ChevronLeftIcon class="size-5" aria-hidden="true" />
                    </a>

                    <!-- Page Numbers with "..." for Large Pages -->
                    <template v-for="page in pageNumbers" :key="page">
                        <span
                            v-if="page === '...'"
                            class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-700 ring-1 ring-inset ring-gray-300 focus:outline-offset-0"
                        >
                            ...
                        </span>
                        <a
                            v-else
                            href="#"
                            @click.prevent="changePage(page)"
                            class="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20"
                            :class="{
                                'z-10 bg-indigo-600 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600':
                                    page === currentPage,
                            }"
                        >
                            {{ page }}
                        </a>
                    </template>

                    <!-- Next Button -->
                    <a
                        href="#"
                        @click.prevent="changePage(currentPage + 1)"
                        class="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus:z-20"
                        :class="{
                            'cursor-not-allowed text-gray-300': currentPage >= totalPages,
                        }"
                    >
                        <span class="sr-only">{{ t("pagination.next") }}</span>
                        <ChevronRightIcon class="size-5" aria-hidden="true" />
                    </a>
                </nav>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed } from "vue";
import { ChevronLeftIcon, ChevronRightIcon } from "@heroicons/vue/20/solid";

// Define props
const props = defineProps({
    currentPage: {
        type: Number,
        required: true,
    },
    totalItems: {
        type: Number,
        required: true,
    },
    itemsPerPage: {
        type: Number,
        required: true,
    },
});

// Define emits
const emit = defineEmits(["update:currentPage"]);

// Computed properties
const totalPages = computed(() => Math.ceil(props.totalItems / props.itemsPerPage));
const startItem = computed(() => (props.currentPage - 1) * props.itemsPerPage + 1);
const endItem = computed(() => Math.min(props.currentPage * props.itemsPerPage, props.totalItems));

// Page numbers logic (same as Tailwind UI with "..." for large pages)
const pageNumbers = computed(() => {
    const windowSize = 5;
    const halfWindow = Math.floor(windowSize / 2);
    let start = Math.max(1, props.currentPage - halfWindow);
    let end = Math.min(totalPages.value, props.currentPage + halfWindow);

    let numbers = [];

    if (start > 1) numbers.push(1);
    if (start > 2) numbers.push("...");
    for (let i = start; i <= end; i++) {
        numbers.push(i);
    }

    if (end < totalPages.value - 1) numbers.push("...");
    if (end < totalPages.value) numbers.push(totalPages.value);

    return numbers;
});

// Change page method
const changePage = (page) => {
    if (page >= 1 && page <= totalPages.value && page !== "...") {
        emit("update:currentPage", page);
    }
};
</script>
