<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        :class="[
            'items-center justify-between gap-x-4 border-t border-gray-200 bg-white pt-1 pb-2',
            totalPages > 1 ? 'flex' : 'hidden sm:flex',
        ]"
    >
        <!-- Mobile Pagination -->
        <div class="flex flex-1 justify-between sm:hidden">
            <BaseButton
                type="button"
                variant="secondary"
                :is-disabled="currentPage <= 1"
                @click="changePage(currentPage - 1)"
            >
                {{ t("pagination.previous") }}
            </BaseButton>
            <BaseButton
                type="button"
                variant="secondary"
                :is-disabled="currentPage >= totalPages"
                @click="changePage(currentPage + 1)"
            >
                {{ t("pagination.next") }}
            </BaseButton>
        </div>

        <!-- Desktop Pagination -->
        <div class="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
            <p v-if="showSummary" class="text-sm text-gray-700">
                {{ t("pagination.showing") }}
                <span class="font-medium">{{ startItem }}</span>
                {{ t("pagination.to") }}
                <span class="font-medium">{{ endItem }}</span>
                {{ t("pagination.of") }}
                <span class="font-medium">{{ totalItems }}</span>
                {{ t("pagination.entries") }}
            </p>
            <span v-else />

            <nav class="flex items-center gap-x-1" :aria-label="t('pagination.label')">
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronDoubleLeftIcon"
                    :aria-label="t('pagination.first')"
                    :is-disabled="currentPage <= 1"
                    @click="changePage(1)"
                />
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronLeftIcon"
                    :aria-label="t('pagination.previous')"
                    :is-disabled="currentPage <= 1"
                    @click="changePage(currentPage - 1)"
                />
                <BaseButton
                    v-for="page in pageNumbers"
                    :key="page"
                    type="button"
                    size="small"
                    class="min-w-[1.75rem]"
                    :variant="page === currentPage ? 'primary' : 'secondary'"
                    :aria-current="page === currentPage ? 'page' : undefined"
                    @click="changePage(page)"
                >
                    {{ page }}
                </BaseButton>
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronRightIcon"
                    :aria-label="t('pagination.next')"
                    :is-disabled="currentPage >= totalPages"
                    @click="changePage(currentPage + 1)"
                />
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronDoubleRightIcon"
                    :aria-label="t('pagination.last')"
                    :is-disabled="currentPage >= totalPages"
                    @click="changePage(totalPages)"
                />
            </nav>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";
import { computed } from "vue";
import {
    ChevronLeftIcon,
    ChevronRightIcon,
    ChevronDoubleLeftIcon,
    ChevronDoubleRightIcon,
} from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";

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
    showSummary: {
        type: Boolean,
        default: true,
    },
});

const emit = defineEmits(["update:currentPage"]);

const totalPages = computed(() => Math.ceil(props.totalItems / props.itemsPerPage));
const startItem = computed(() => (props.currentPage - 1) * props.itemsPerPage + 1);
const endItem = computed(() => Math.min(props.currentPage * props.itemsPerPage, props.totalItems));

// At most five page numbers, around the current one.
const pageNumbers = computed(() => {
    const windowSize = 5;
    const halfWindow = Math.floor(windowSize / 2);
    let start = Math.max(1, props.currentPage - halfWindow);
    let end = Math.min(totalPages.value, props.currentPage + halfWindow);

    if (end - start + 1 < windowSize) {
        if (start === 1) {
            end = Math.min(start + windowSize - 1, totalPages.value);
        } else if (end === totalPages.value) {
            start = Math.max(1, end - windowSize + 1);
        }
    }

    const numbers = [];

    for (let i = start; i <= end; i++) {
        numbers.push(i);
    }

    return numbers;
});

const changePage = (page) => {
    if (page >= 1 && page <= totalPages.value) {
        emit("update:currentPage", page);
    }
};
</script>
