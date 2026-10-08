<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex shrink-0 gap-x-4 opacity-0 group-hover:opacity-100 transition-opacity duration-150"
        :class="
            detailsStore.file != null && detailsStore.file.id == item.id
                ? 'opacity-100'
                : 'opacity-0'
        "
    >
        <button
            v-if="view != 'deleted'"
            @click.stop="dialogStore.openShareDialog(item)"
            @dblclick.stop
            type="button"
            class="z-20 rounded-full p-1 text-gray-500 hover:bg-gray-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 transition duration-300 ease-in-out"
        >
            <ShareIcon class="size-5" aria-hidden="true" />
        </button>
        <button
            v-if="view != 'deleted'"
            @click.stop="emit('download', item)"
            @dblclick.stop
            type="button"
            class="z-20 rounded-full p-1 text-gray-500 hover:bg-gray-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 transition duration-300 ease-in-out"
        >
            <ArrowDownOnSquareIcon class="size-5" aria-hidden="true" />
        </button>
        <button
            v-if="view == 'deleted'"
            @click.stop="emit('restore', item)"
            @dblclick.stop
            :disabled="deleting"
            type="button"
            class="z-20 rounded-full p-1 text-red-500 hover:bg-gray-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 transition duration-300 ease-in-out"
        >
            <ArrowUturnLeftIcon class="size-5" aria-hidden="true" />
        </button>
        <button
            v-if="view == 'deleted'"
            @click.stop="emit('deletePermanently', item)"
            @dblclick.stop
            :disabled="deleting"
            type="button"
            class="z-20 rounded-full p-1 text-red-500 hover:bg-gray-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-red-600 transition duration-300 ease-in-out"
        >
            <TrashIcon class="size-5" aria-hidden="true" />
        </button>
    </div>
</template>

<script setup>
import { useDialogStore } from "@/store/dialogs";
import { useDetailsStore } from "@/store/details";
import {
    ShareIcon,
    ArrowDownOnSquareIcon,
    ArrowUturnLeftIcon,
    TrashIcon,
} from "@heroicons/vue/24/outline";

defineProps({
    item: {
        type: Object,
        required: true,
    },
    view: {
        type: String,
        default: undefined,
    },
    deleting: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["download", "restore", "deletePermanently"]);

const dialogStore = useDialogStore();
const detailsStore = useDetailsStore();
</script>
