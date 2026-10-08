<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="shrink-0 flex items-center justify-between gap-x-4 border-b border-gray-200 bg-white px-4 py-3"
    >
        <div class="relative w-64">
            <input
                v-model="name"
                type="text"
                name="name"
                id="name"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                :class="
                    nameError
                        ? 'ring-red-300 focus:ring-red-500'
                        : 'ring-gray-300 focus:ring-indigo-600'
                "
                :placeholder="t('collimato.charts.new_chart.enter_chart_name')"
            />
            <div
                v-if="nameError"
                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3"
            >
                <ExclamationCircleIcon class="h-5 w-5 text-red-400" aria-hidden="true" />
            </div>
        </div>

        <div class="flex items-center gap-x-3">
            <BaseButton
                v-if="showSqlButton"
                @click="sqlPreviewDialog = true"
                color="bg-white ring-1 ring-inset ring-gray-300 text-gray-700 hover:bg-gray-50"
            >
                {{ t("collimato.charts.new_chart.button.sql_query") }}
            </BaseButton>
            <SqlPreviewDialog v-model="sqlPreviewDialog" :message="sqlPreview" />
            <BaseButton v-if="showSaveButton" @click="emit('save')">
                {{ isEdit ? t("common.button.update") : t("common.button.create") }}
            </BaseButton>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { t } from "@/i18n/index.js";
import BaseButton from "@/components/BaseButton.vue";
import SqlPreviewDialog from "@/components/Collimato/Dialogs/SqlPreviewDialog.vue";
import { ExclamationCircleIcon } from "@heroicons/vue/24/outline";

defineProps({
    nameError: {
        type: Boolean,
        default: false,
    },
    showSqlButton: {
        type: Boolean,
        default: true,
    },
    sqlPreview: {
        type: String,
        default: "",
    },
    isEdit: {
        type: Boolean,
        default: false,
    },
    showSaveButton: {
        type: Boolean,
        default: true,
    },
});

const emit = defineEmits(["save"]);

const name = defineModel("name", {
    type: String,
    default: "",
});

const sqlPreviewDialog = ref(false);
</script>
