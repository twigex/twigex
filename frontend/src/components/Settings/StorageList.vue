<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex-grow w-full flex flex-col min-h-0">
        <!-- Fixed header -->
        <div class="flex-shrink-0 flex flex-row justify-center px-2 pt-2">
            <div class="w-full md:w-3/4">
                <div
                    v-if="!storages.some((s) => s.is_primary)"
                    class="mb-4 rounded-md bg-yellow-50 p-4 ring-1 ring-inset ring-yellow-700/10"
                >
                    <div class="flex">
                        <ExclamationTriangleIcon
                            class="h-5 w-5 text-yellow-500 shrink-0 mt-0.5"
                            aria-hidden="true"
                        />
                        <div class="ml-3">
                            <p class="text-sm text-yellow-700">
                                {{ t("settings.storage.no_primary_warning") }}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="mb-4 rounded-md bg-blue-50 p-4 ring-1 ring-inset ring-blue-700/10">
                    <div class="flex">
                        <InformationCircleIcon
                            class="h-5 w-5 text-blue-400 shrink-0 mt-0.5"
                            aria-hidden="true"
                        />
                        <div class="ml-3">
                            <p class="text-sm text-blue-700">
                                {{ t("settings.storage.primary_info") }}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="flex items-center justify-between pb-4 border-b border-gray-900/10">
                    <div>
                        <h2 class="text-base font-semibold leading-7 text-gray-900">
                            {{ t("settings.storage.configured_storages") }}
                        </h2>
                        <p class="mt-0.5 text-sm text-gray-500">
                            {{ storages.length }}
                            {{ t("settings.storage.storages_configured") }}
                        </p>
                    </div>
                    <div class="flex flex-col items-end gap-y-1">
                        <button
                            type="button"
                            class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                            :class="{
                                'opacity-50 pointer-events-none': !hasMultipleStoragesLicense,
                            }"
                            @click="emit('create')"
                        >
                            <LockClosedIcon
                                v-if="!hasMultipleStoragesLicense"
                                class="-ml-0.5 mr-1.5 h-4 w-4"
                                aria-hidden="true"
                            />
                            <PlusIcon v-else class="-ml-0.5 mr-1.5 h-4 w-4" aria-hidden="true" />
                            {{ t("common.button.new") }}
                        </button>
                        <span
                            v-if="!hasMultipleStoragesLicense"
                            class="inline-flex items-center gap-x-1 rounded-md bg-amber-50 px-1.5 py-0.5 text-xs font-medium text-amber-800 ring-1 ring-inset ring-amber-600/20"
                        >
                            <LockClosedIcon class="h-3 w-3" />
                            {{ t("settings.license.business_plan") }}
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <!-- Scrollable list -->
        <div class="flex-grow overflow-y-auto">
            <div class="flex flex-row justify-center px-2 pb-2">
                <div class="w-full md:w-3/4">
                    <ul
                        role="list"
                        class="mt-4 divide-y divide-gray-100 overflow-hidden bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl"
                    >
                        <li
                            v-for="s in storages"
                            :key="s.id"
                            class="flex items-center justify-between gap-x-6 px-4 py-4 hover:bg-gray-50 sm:px-6"
                        >
                            <div class="min-w-0 flex-auto">
                                <div class="flex items-center gap-x-2">
                                    <p class="text-sm font-semibold leading-6 text-gray-900">
                                        {{ s.label }}
                                    </p>
                                    <span
                                        v-if="s.is_primary"
                                        class="inline-flex items-center rounded-md bg-indigo-50 px-2 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-700/10"
                                    >
                                        {{ t("settings.storage.primary") }}
                                    </span>
                                </div>
                                <div class="mt-1 flex items-center gap-x-2">
                                    <span
                                        :class="typeBadgeClass(s.type)"
                                        class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset"
                                    >
                                        {{
                                            s.type === 2
                                                ? t("settings.storage.type.s3")
                                                : t("settings.storage.type.local")
                                        }}
                                    </span>
                                    <span class="text-xs text-gray-400">
                                        {{ s.type === 2 ? s.endpoint : s.directory }}
                                    </span>
                                </div>
                            </div>
                            <div class="flex shrink-0 items-center gap-x-1">
                                <button
                                    v-if="!s.is_primary"
                                    type="button"
                                    class="rounded-md px-2 py-1 text-xs font-medium text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 ring-1 ring-inset ring-gray-300 hover:ring-indigo-300"
                                    @click="emit('setPrimary', s)"
                                >
                                    {{ t("settings.storage.set_primary") }}
                                </button>
                                <button
                                    type="button"
                                    class="rounded-md p-1.5 text-gray-400 hover:text-indigo-600 hover:bg-indigo-50"
                                    @click="emit('edit', s)"
                                >
                                    <PencilIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                                <button
                                    type="button"
                                    class="rounded-md p-1.5 text-gray-400 hover:text-red-600 hover:bg-red-50"
                                    @click="emit('delete', s)"
                                >
                                    <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                </button>
                            </div>
                        </li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { PlusIcon } from "@heroicons/vue/20/solid";
import {
    TrashIcon,
    ExclamationTriangleIcon,
    InformationCircleIcon,
    PencilIcon,
    LockClosedIcon,
} from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

defineProps({
    storages: {
        type: Array,
        default: () => [],
    },
    hasMultipleStoragesLicense: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["create", "edit", "delete", "setPrimary"]);

function typeBadgeClass(type) {
    return type === 2
        ? "bg-blue-50 text-blue-700 ring-blue-700/10"
        : "bg-gray-50 text-gray-600 ring-gray-500/10";
}
</script>
