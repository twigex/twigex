<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="isMap" class="shrink-0 border-t border-gray-200 p-4">
        <BaseButton
            v-if="layersLoading"
            @click="emit('cancelLayers')"
            class="w-full"
            color="bg-slate-600 hover:bg-slate-700 text-white"
            :title="t('common.button.cancel')"
        >
            <span
                class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"
            ></span>
            {{ t("collimato.query.running") }}
        </BaseButton>
        <BaseButton
            v-else
            @click="emit('reloadLayers')"
            class="w-full"
            color="bg-indigo-600 hover:bg-indigo-500 text-white"
            :disabled="!hasLayers"
        >
            {{ t("collimato.charts.new_chart.map.reload_layers") }}
        </BaseButton>
    </div>

    <div v-else class="shrink-0 border-t border-gray-200 p-4">
        <BaseButton
            v-if="loading"
            @click="emit('cancelLoad')"
            class="w-full"
            color="bg-slate-600 hover:bg-slate-700 text-white"
            :title="t('common.button.cancel')"
        >
            <span
                class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin mr-2"
            ></span>
            {{ t("collimato.query.running") }}
            <svg
                class="h-4 w-4 ml-2 opacity-75"
                viewBox="0 0 20 20"
                fill="currentColor"
                aria-hidden="true"
            >
                <path
                    d="M6.28 5.22a.75.75 0 0 0-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 1 0 1.06 1.06L10 11.06l3.72 3.72a.75.75 0 1 0 1.06-1.06L11.06 10l3.72-3.72a.75.75 0 0 0-1.06-1.06L10 8.94 6.28 5.22Z"
                />
            </svg>
        </BaseButton>
        <BaseButton
            v-else
            @click="emit('load')"
            class="w-full"
            color="bg-indigo-600 hover:bg-indigo-500 text-white"
        >
            {{ t("collimato.charts.new_chart.button.run_query") }}
        </BaseButton>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import BaseButton from "@/components/BaseButton.vue";

defineProps({
    isMap: {
        type: Boolean,
        default: false,
    },
    loading: {
        type: Boolean,
        default: false,
    },
    layersLoading: {
        type: Boolean,
        default: false,
    },
    hasLayers: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["cancelLoad", "load", "cancelLayers", "reloadLayers"]);
</script>
