<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex shrink-0 items-center justify-center gap-4 border-t border-gray-200 py-3 text-sm text-gray-600"
    >
        <button
            type="button"
            :disabled="page <= 1"
            @click="emit('update:page', page - 1)"
            class="rounded-md px-3 py-1.5 font-medium hover:bg-gray-100 disabled:opacity-40"
        >
            {{ t("common.button.previous") }}
        </button>
        <span>{{ page }} / {{ pageCount }}</span>
        <button
            type="button"
            :disabled="page >= pageCount"
            @click="emit('update:page', page + 1)"
            class="rounded-md px-3 py-1.5 font-medium hover:bg-gray-100 disabled:opacity-40"
        >
            {{ t("common.button.next") }}
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n";

defineProps({
    page: {
        type: Number,
        required: true,
    },
    pageCount: {
        type: Number,
        required: true,
    },
});

const emit = defineEmits(["update:page"]);
</script>
