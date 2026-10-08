<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="px-3 pb-2">
        <div class="flex gap-x-2 overflow-x-auto pt-2 pb-1">
            <div
                v-for="(filePreview, index) in props.previews"
                :key="index"
                class="relative shrink-0 flex items-center gap-x-2 rounded-lg border border-gray-200 bg-gray-50 p-2"
                :class="filePreview.kind === 'gifv' ? 'w-auto' : 'w-52'"
            >
                <!-- Remove button -->
                <button
                    @click="emits('remove', index)"
                    class="absolute -top-1.5 -right-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-gray-500 text-white text-xs leading-none hover:bg-red-500 transition-colors shadow-sm"
                >
                    ×
                </button>

                <!-- GIF preview -->
                <video
                    v-if="filePreview.kind === 'gifv'"
                    :src="filePreview.url"
                    autoplay
                    loop
                    muted
                    playsinline
                    class="h-10 w-10 rounded-md object-cover shrink-0"
                />

                <!-- Image preview -->
                <img
                    v-else-if="filePreview.type.includes('image') && filePreview.url"
                    :src="filePreview.url"
                    alt="File preview"
                    class="h-10 w-10 rounded-md object-cover shrink-0"
                />

                <!-- File type icon -->
                <component
                    v-else-if="filePreview.url == null"
                    :is="fileTypes[filePreview.type]?.icon ?? fileTypes['not-found'].icon"
                    :class="
                        'h-10 w-10 shrink-0 ' +
                        (fileTypes[filePreview.type]?.color ?? fileTypes['not-found'].color)
                    "
                />

                <!-- File info -->
                <div
                    v-if="filePreview.kind !== 'gifv'"
                    class="flex flex-col min-w-0 flex-1 gap-y-0.5"
                >
                    <span class="text-xs font-medium text-gray-700 truncate">
                        {{ filePreview.name }}
                    </span>
                    <div v-if="filePreview.progress >= 100" class="text-xs text-gray-400">
                        {{ convertSize(filePreview.size) }}
                    </div>
                    <div v-else class="h-1.5 w-full overflow-hidden rounded-full bg-gray-200">
                        <div
                            class="h-full bg-indigo-500 transition-all duration-300"
                            :style="{
                                width: filePreview.progress + '%',
                            }"
                        />
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import fileTypes from "@/constants/fileTypes";
import { convertSize } from "@/utils/utils";

const props = defineProps({
    previews: {
        type: Array,
        required: true,
    },
});

const emits = defineEmits(["remove"]);
</script>
