<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ul
        v-if="links.length"
        role="list"
        class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200"
    >
        <li
            v-for="link in links"
            :key="link.id"
            class="flex items-center justify-between gap-x-4 px-3 py-2"
        >
            <div class="flex min-w-0 items-center gap-x-3">
                <span
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-gray-500"
                >
                    <LinkIcon class="h-5 w-5 text-white" aria-hidden="true" />
                </span>
                <div class="min-w-0">
                    <p class="truncate text-sm font-medium text-gray-900">
                        {{ accessLabel(link) }}
                    </p>
                    <p class="truncate text-xs text-gray-500">
                        {{ link.name }}
                    </p>
                </div>
            </div>
            <div class="flex shrink-0 items-center gap-x-1">
                <button
                    type="button"
                    @click="emit('edit', link)"
                    class="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                >
                    <PencilSquareIcon class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    @click="emit('copy', link.id)"
                    class="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                >
                    <ClipboardDocumentIcon class="h-5 w-5" aria-hidden="true" />
                </button>
                <button
                    type="button"
                    @click="emit('remove', link.id)"
                    class="rounded-md p-1.5 text-gray-400 hover:bg-gray-100 hover:text-red-600"
                >
                    <TrashIcon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>
        </li>
    </ul>

    <div v-else class="py-8 text-center">
        <LinkIcon class="mx-auto h-8 w-8 text-gray-300" aria-hidden="true" />
        <p class="mt-2 text-sm text-gray-500">
            {{ t("files.share_dialog.link.empty") }}
        </p>
    </div>
    <button
        type="button"
        @click="emit('create')"
        class="mt-4 flex w-full items-center justify-center gap-2 rounded-md border border-dashed border-gray-300 px-3 py-2 text-sm font-medium text-gray-600 hover:border-indigo-400 hover:text-indigo-600"
    >
        <PlusIcon class="h-5 w-5" aria-hidden="true" />
        {{ t("files.share_dialog.link.create") }}
    </button>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import {
    LinkIcon,
    ClipboardDocumentIcon,
    PencilSquareIcon,
    PlusIcon,
    TrashIcon,
} from "@heroicons/vue/24/outline";

defineProps({
    links: {
        type: Array,
        default: () => [],
    },
    accessLabel: {
        type: Function,
        required: true,
    },
});

const emit = defineEmits(["edit", "copy", "remove", "create"]);
</script>
