<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ul role="list" class="divide-y divide-gray-100 rounded-lg bg-white">
        <li
            v-for="child in items"
            :key="child.ID"
            class="group flex items-center justify-between gap-x-4 px-4 py-2.5 hover:bg-gray-50"
        >
            <button
                type="button"
                class="flex min-w-0 items-center gap-x-3 text-left"
                @click="emit('childClick', child)"
                @dblclick="emit('open', child)"
            >
                <img
                    v-if="isImage(child.Type) && !failedThumbs.has(child.ID)"
                    :src="publicShareService.thumbnailUrl(token, child.ID)"
                    class="h-9 w-9 shrink-0 rounded bg-gray-100 object-cover"
                    alt=""
                    @error="emit('thumbFailed', child.ID)"
                />
                <component
                    :is="typeMeta(child).icon"
                    v-else
                    :class="['h-6 w-6 shrink-0', typeMeta(child).color]"
                    aria-hidden="true"
                />
                <span class="min-w-0">
                    <span
                        class="block truncate text-sm font-medium text-gray-900 group-hover:text-indigo-600"
                    >
                        {{ child.Name }}
                    </span>
                    <span class="block truncate text-xs text-gray-500">
                        {{ child.IsFolder ? t("public_share.folder") : convertSize(child.Size) }}
                    </span>
                </span>
            </button>
            <a
                v-if="!child.IsFolder && child.AllowDownload"
                :href="publicShareService.downloadUrl(token, child.ID)"
                class="shrink-0 text-sm font-medium text-indigo-600 opacity-0 transition-opacity hover:text-indigo-500 group-hover:opacity-100"
                @click.stop
            >
                {{ t("public_share.download") }}
            </a>
        </li>
    </ul>
</template>

<script setup>
import { t } from "@/i18n";
import { convertSize } from "@/utils/utils";
import { isImage, typeMeta } from "@/utils/files/publicShareTypes";
import publicShareService from "@/services/publicShareService";

defineProps({
    items: {
        type: Array,
        default: () => [],
    },
    token: {
        type: String,
        required: true,
    },
    failedThumbs: {
        type: Set,
        required: true,
    },
});

const emit = defineEmits(["childClick", "open", "thumbFailed"]);
</script>
