<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="w-full max-w-md rounded-xl border-2 border-dashed bg-white p-8 text-center transition-colors"
        :class="dragOver ? 'border-indigo-400 bg-indigo-50/40' : 'border-gray-300'"
        @dragover.prevent
        @dragenter.prevent="emit('update:dragOver', true)"
        @dragleave.self="emit('update:dragOver', false)"
        @drop.prevent="emit('drop', $event)"
    >
        <ArrowUpTrayIcon class="mx-auto h-6 w-6 text-gray-400" aria-hidden="true" />
        <p class="mt-2 text-sm font-medium text-gray-900">
            {{ t("public_share.upload_only_title") }}
        </p>
        <p class="mt-1 text-sm text-gray-500">
            {{ t("public_share.upload_only_prompt") }}
        </p>
        <button
            type="button"
            @click="emit('upload')"
            :disabled="uploading"
            class="mt-4 rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
        >
            {{ uploadLabel }}
        </button>
    </div>
</template>

<script setup>
import { ArrowUpTrayIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n";

defineProps({
    dragOver: {
        type: Boolean,
        default: false,
    },
    uploading: {
        type: Boolean,
        default: false,
    },
    uploadLabel: {
        type: String,
        default: "",
    },
});

const emit = defineEmits(["update:dragOver", "upload", "drop"]);
</script>
