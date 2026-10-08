<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="mt-4 grid grid-cols-2 gap-1 rounded-lg bg-gray-100 p-1">
        <button
            type="button"
            @click="emit('update:activeTab', 'people')"
            :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                activeTab === 'people'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700',
            ]"
        >
            {{ t("files.share_dialog.tab_people") }}
        </button>
        <button
            type="button"
            @click="emit('update:activeTab', 'link')"
            :class="[
                'rounded-md px-3 py-1.5 text-sm font-medium transition-colors',
                activeTab === 'link'
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-500 hover:text-gray-700',
            ]"
        >
            {{ t("files.share_dialog.tab_link") }}
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

defineProps({
    activeTab: {
        type: String,
        required: true,
    },
});

const emit = defineEmits(["update:activeTab"]);
</script>
