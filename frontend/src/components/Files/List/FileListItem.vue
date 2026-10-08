<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li
        @click="handleFileClick"
        @contextmenu="handleFileClick"
        class="relative flex justify-between gap-x-6 px-2 py-1 transition-all sm:px-2"
    >
        <div class="flex min-w-0 gap-x-4 relative">
            <div class="relative">
                <component
                    v-if="fileType != 'image'"
                    :is="icon"
                    :class="'h-10 w-10 flex-none ' + color"
                />

                <img
                    v-else
                    class="h-10 w-10 flex-none object-cover"
                    :src="`/api/files/thumbnails/${file.id}`"
                />
                <StarIcon
                    v-if="file.favourite"
                    class="absolute top-6 end-0 h-5 w-5 text-yellow-500"
                />
            </div>

            <div class="min-w-0 flex-auto">
                <p class="text-sm font-semibold leading-6 text-gray-900 truncate">
                    <a class="cursor-pointer">
                        <span
                            data-drag-handle
                            class="absolute z-20 w-64 inset-x-0 -top-px bottom-0"
                        />
                        {{ file.name }}
                    </a>
                </p>
                <p class="flex text-xs leading-5 text-gray-500">
                    <a class="relative truncate cursor-pointer pr-1"
                        >{{ t("files.fileitem.last_modified") }}
                    </a>
                    {{ useDateOperations().getDateAndTime(file.modified) }}
                </p>
            </div>
        </div>
        <div class="flex shrink-0 items-center gap-x-4">
            <slot name="actions" :item="file"></slot>
        </div>
    </li>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useDetailsStore } from "@/store/details";
import fileTypes from "@/constants/fileTypes";
import useDateOperations from "@/composables/useDateOperations.js";
import { StarIcon } from "@heroicons/vue/24/solid";

const props = defineProps({
    file: {
        type: Object,
        default: () => ({}),
    },
});

const detailsStore = useDetailsStore();

const icon = fileTypes[props.file.type]?.icon ?? fileTypes["not-found"].icon;

const color = fileTypes[props.file.type]?.color ?? fileTypes["not-found"].color;

const fileType = fileTypes[props.file.type]?.fileType ?? fileTypes["not-found"].fileType;

function handleFileClick() {
    // On touch a tap opens the item; details are opened explicitly from the context menu.
    if (window.matchMedia("(max-width: 1023px)").matches) {
        return;
    }

    detailsStore.loadDetails(props.file);
}
</script>
