<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex-grow h-full flex justify-center items-center">
        <div class="text-center max-w-sm">
            <CircleStackIcon class="mx-auto h-12 w-12 text-gray-400" aria-hidden="true" />
            <p class="mt-4 text-lg font-semibold text-gray-900">
                {{ t("settings.storage.no_storages") }}
            </p>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("settings.storage.no_storages_description") }}
            </p>
            <div
                class="mt-4 rounded-md bg-blue-50 p-3 ring-1 ring-inset ring-blue-700/10 text-left"
            >
                <div class="flex">
                    <InformationCircleIcon
                        class="h-5 w-5 text-blue-400 shrink-0 mt-0.5"
                        aria-hidden="true"
                    />
                    <p class="ml-2 text-sm text-blue-700">
                        {{ t("settings.storage.primary_info") }}
                    </p>
                </div>
            </div>
            <div class="mt-6">
                <button
                    type="button"
                    class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
                    @click="emit('create')"
                >
                    <PlusIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
                    {{ t("settings.storage.button.new_storage") }}
                </button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { PlusIcon } from "@heroicons/vue/20/solid";
import { CircleStackIcon, InformationCircleIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

const emit = defineEmits(["create"]);
</script>
