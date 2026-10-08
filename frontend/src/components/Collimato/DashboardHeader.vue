<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="shrink-0 flex items-center justify-between border-b px-4 py-2">
        <div class="flex items-center gap-x-3 min-w-0">
            <button
                type="button"
                @click="router.push({ name: 'dashboards' })"
                class="shrink-0 rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            >
                <ChevronLeftIcon class="h-5 w-5" aria-hidden="true" />
            </button>
            <h1 class="truncate text-base font-semibold text-gray-900">
                {{ title }}
            </h1>
        </div>
        <div v-if="collimatoStore.hasPermissionToEditDashboards" class="flex items-center gap-x-2">
            <button
                v-if="edit"
                type="button"
                @click="emit('cancel')"
                class="inline-flex items-center gap-x-1.5 rounded-md bg-white px-3 py-1.5 text-sm font-semibold text-gray-700 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
            >
                <XMarkIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                {{ t("common.button.cancel") }}
            </button>
            <button
                type="button"
                :disabled="saving"
                @click="edit ? emit('save') : emit('startEdit')"
                :class="[
                    edit
                        ? 'bg-indigo-600 hover:bg-indigo-500 text-white'
                        : 'bg-white text-gray-700 ring-1 ring-inset ring-gray-300 hover:bg-gray-50',
                    'inline-flex items-center gap-x-1.5 rounded-md px-3 py-1.5 text-sm font-semibold shadow-sm disabled:opacity-50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600',
                ]"
            >
                <CheckIcon v-if="edit" class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                <PencilSquareIcon v-else class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                {{
                    edit
                        ? t("collimato.dashboard.button.save")
                        : t("collimato.dashboard.button.edit")
                }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { useRouter } from "vue-router";
import { t } from "@/i18n/index.js";
import { useCollimatoStore } from "@/store/collimato";
import { CheckIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { ChevronLeftIcon, PencilSquareIcon } from "@heroicons/vue/24/outline";

defineProps({
    title: {
        type: String,
        default: "",
    },
    edit: {
        type: Boolean,
        default: false,
    },
    saving: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["cancel", "save", "startEdit"]);

const collimatoStore = useCollimatoStore();
const router = useRouter();
</script>
