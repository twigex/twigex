<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
        <div
            v-for="child in items"
            :key="child.ID"
            class="group flex h-44 flex-col rounded-lg border border-gray-200 bg-white p-2 transition-shadow hover:border-gray-300 hover:shadow-sm sm:h-52"
        >
            <button
                type="button"
                class="flex min-h-0 flex-1 items-center justify-center overflow-hidden rounded-md bg-gray-50"
                @click="emit('childClick', child)"
                @dblclick="emit('open', child)"
            >
                <img
                    v-if="isImage(child.Type) && !failedThumbs.has(child.ID)"
                    :src="publicShareService.thumbnailUrl(token, child.ID)"
                    class="h-full w-full object-cover"
                    alt=""
                    @error="emit('thumbFailed', child.ID)"
                />
                <component
                    :is="typeMeta(child).icon"
                    v-else
                    :class="['h-12 w-12', typeMeta(child).color]"
                    aria-hidden="true"
                />
            </button>
            <div class="mt-2 shrink-0 px-0.5">
                <button
                    type="button"
                    class="block w-full truncate text-left text-sm font-medium text-gray-900 hover:text-indigo-600"
                    :title="child.Name"
                    @click="emit('childClick', child)"
                    @dblclick="emit('open', child)"
                >
                    {{ child.Name }}
                </button>
                <div class="mt-0.5 flex items-center justify-between gap-2">
                    <span class="truncate text-xs text-gray-500">
                        {{ child.IsFolder ? t("public_share.folder") : convertSize(child.Size) }}
                    </span>
                    <a
                        v-if="!child.IsFolder && child.AllowDownload"
                        :href="publicShareService.downloadUrl(token, child.ID)"
                        class="shrink-0 text-xs font-medium text-indigo-600 opacity-0 transition-opacity hover:text-indigo-500 group-hover:opacity-100"
                        @click.stop
                    >
                        {{ t("public_share.download") }}
                    </a>
                </div>
            </div>
        </div>
    </div>
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
