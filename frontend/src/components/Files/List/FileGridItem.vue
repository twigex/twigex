<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <li
        @click="handleFileClick"
        @contextmenu="handleFileClick"
        class="group relative flex h-44 flex-col rounded-xl border border-gray-200 bg-white p-2.5 transition-all hover:border-gray-300 hover:shadow-md sm:h-48"
    >
        <span data-drag-handle class="absolute inset-0 z-10" />

        <div
            class="flex min-h-0 flex-1 items-center justify-center overflow-hidden rounded-lg bg-gray-50"
        >
            <img
                v-if="fileType === 'image'"
                class="h-full w-full object-cover"
                :src="`/api/files/thumbnails/${file.id}`"
                alt=""
            />
            <component v-else :is="icon" :class="'h-14 w-14 ' + color" />
        </div>

        <StarIcon
            v-if="file.favourite"
            class="absolute start-3 top-3 z-20 h-5 w-5 text-yellow-400 drop-shadow"
        />

        <div class="mt-2.5 shrink-0 px-0.5">
            <p class="truncate text-sm font-medium leading-5 text-gray-900" :title="file.name">
                {{ file.name }}
            </p>
            <p class="mt-1 truncate text-xs text-gray-500">
                {{ useDateOperations().getDateAndTime(file.modified) }}
            </p>
        </div>

        <div
            class="absolute end-2 top-2 z-20 flex shrink-0 items-center rounded-lg bg-white p-0.5 opacity-0 shadow-sm ring-1 ring-gray-200 transition-opacity duration-150 group-hover:opacity-100"
        >
            <slot name="actions" :item="file"></slot>
        </div>
    </li>
</template>

<script setup>
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
