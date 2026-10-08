<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col items-center justify-center text-center h-full">
        <div v-if="route.name == 'file'">
            <template v-if="can('upload_files') || can('create_files')">
                <input
                    type="file"
                    ref="file"
                    style="display: none"
                    @change="onFilesSelected"
                    multiple
                />
                <svg
                    class="mx-auto h-12 w-12 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    aria-hidden="true"
                >
                    <path
                        vector-effect="non-scaling-stroke"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                    />
                </svg>
                <h3 class="mt-2 text-sm font-semibold text-gray-900">
                    {{ t("files.empty_state.title") }}
                </h3>
                <p class="mt-1 text-sm text-gray-500">
                    {{ t("files.empty_state.description") }}
                </p>
                <div v-if="can('upload_files')" class="mt-6">
                    <button
                        @click="$refs.file.click()"
                        type="button"
                        class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    >
                        <PlusIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
                        {{ t("files.empty_state.upload_button") }}
                    </button>
                </div>
            </template>
            <p v-else class="text-sm text-gray-400">
                {{ t("files.empty_state.no_permission") }}
            </p>
        </div>
        <div v-if="route.name == 'shared'">
            <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                class="mx-auto h-12 w-12 text-gray-400"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M7.217 10.907a2.25 2.25 0 1 0 0 2.186m0-2.186c.18.324.283.696.283 1.093s-.103.77-.283 1.093m0-2.186 9.566-5.314m-9.566 7.5 9.566 5.314m0 0a2.25 2.25 0 1 0 3.935 2.186 2.25 2.25 0 0 0-3.935-2.186Zm0-12.814a2.25 2.25 0 1 0 3.933-2.185 2.25 2.25 0 0 0-3.933 2.185Z"
                />
            </svg>
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("files.empty_state.shared_title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("files.empty_state.shared_description") }}
            </p>
        </div>
        <div v-if="route.name == 'recents'">
            <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                class="mx-auto h-12 w-12 text-gray-400"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                />
            </svg>
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("files.empty_state.recents_title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("files.empty_state.recents_description") }}
            </p>
        </div>
        <div v-if="route.name == 'favorites'">
            <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                class="mx-auto h-12 w-12 text-gray-400"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="M11.48 3.499a.562.562 0 0 1 1.04 0l2.125 5.111a.563.563 0 0 0 .475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 0 0-.182.557l1.285 5.385a.562.562 0 0 1-.84.61l-4.725-2.885a.562.562 0 0 0-.586 0L6.982 20.54a.562.562 0 0 1-.84-.61l1.285-5.386a.562.562 0 0 0-.182-.557l-4.204-3.602a.562.562 0 0 1 .321-.988l5.518-.442a.563.563 0 0 0 .475-.345L11.48 3.5Z"
                />
            </svg>
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("files.empty_state.favorites_title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("files.empty_state.favorites_description") }}
            </p>
        </div>
        <div v-if="route.name == 'deleted'">
            <svg
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                class="mx-auto h-12 w-12 text-gray-400"
            >
                <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
                />
            </svg>
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("files.empty_state.deleted_title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("files.empty_state.deleted_description") }}
            </p>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useRoute } from "vue-router";
import { PlusIcon } from "@heroicons/vue/20/solid";
import { usePermissions } from "@/composables/usePermissions";

const emits = defineEmits(["filesSelected"]);
const route = useRoute();
const { can } = usePermissions();

function onFilesSelected(event) {
    const selectedFiles = event.target.files;

    if (selectedFiles && selectedFiles.length) {
        emits("filesSelected", { files: selectedFiles, type: "file" });
    }
}
</script>
