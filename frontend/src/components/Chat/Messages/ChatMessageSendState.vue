<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="showSending"
        class="absolute right-3 flex size-4 overflow-hidden"
        :class="collapsed ? 'top-1' : 'top-3'"
    >
        <BaseSpinner size="sm" class="text-indigo-600" :label="t('channels.message.sending')" />
    </div>

    <div
        v-else-if="sendState === 'failed'"
        class="absolute right-3 flex items-center gap-x-2 text-xs"
        :class="collapsed ? 'top-1' : 'top-3'"
    >
        <ExclamationCircleIcon class="size-4 text-red-500" aria-hidden="true" />
        <span class="text-red-600">{{ t("channels.message.not_sent") }}</span>
        <button
            type="button"
            class="font-medium text-indigo-600 hover:text-indigo-500"
            @click.stop="emit('retry')"
        >
            {{ t("channels.message.retry") }}
        </button>
        <button
            type="button"
            class="font-medium text-gray-500 hover:text-gray-700"
            @click.stop="emit('discard')"
        >
            {{ t("common.button.delete") }}
        </button>
    </div>
</template>

<script setup>
import { ExclamationCircleIcon } from "@heroicons/vue/20/solid";
import BaseSpinner from "@/components/BaseSpinner.vue";
import { t } from "@/i18n/index.js";

defineProps({
    sendState: {
        type: String,
        default: null,
    },
    showSending: {
        type: Boolean,
        default: false,
    },
    collapsed: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["retry", "discard"]);
</script>
