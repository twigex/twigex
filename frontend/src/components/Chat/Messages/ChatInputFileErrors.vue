<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="props.sizeErrors.length > 0" class="mt-1 flex items-start gap-x-2">
        <p class="text-xs text-red-500">
            {{ t("channels.input.error_size_exceeded", { size: props.sizeLimit }) }}
            {{ props.sizeErrors.join(", ") }}
        </p>
        <button
            type="button"
            class="shrink-0 text-red-400 hover:text-red-600"
            :title="t('common.button.dismiss')"
            :aria-label="t('common.button.dismiss')"
            @click="emits('dismissSize')"
        >
            <XMarkIcon class="size-4" aria-hidden="true" />
        </button>
    </div>

    <div v-if="props.uploadErrors.length > 0" class="mt-1 flex items-start gap-x-2">
        <p class="text-xs text-red-500">
            {{ t("channels.input.error_upload_failed") }}
            {{ props.uploadErrors.join(", ") }}
        </p>
        <button
            type="button"
            class="shrink-0 text-red-400 hover:text-red-600"
            :title="t('common.button.dismiss')"
            :aria-label="t('common.button.dismiss')"
            @click="emits('dismissUpload')"
        >
            <XMarkIcon class="size-4" aria-hidden="true" />
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { XMarkIcon } from "@heroicons/vue/24/solid";

const props = defineProps({
    sizeErrors: {
        type: Array,
        required: true,
    },
    uploadErrors: {
        type: Array,
        required: true,
    },
    sizeLimit: {
        type: String,
        required: true,
    },
});

const emits = defineEmits(["dismissSize", "dismissUpload"]);
</script>
