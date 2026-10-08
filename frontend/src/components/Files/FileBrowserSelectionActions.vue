<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <transition
        enter-active-class="transition-opacity duration-300 ease-in-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-0 ease-in-out"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
    >
        <div
            v-if="hasSelection && route.name !== 'deleted'"
            key="selected-actions"
            class="flex flex-row gap-x-1 items-center"
        >
            <BaseTooltip position="bottom">
                <button
                    type="button"
                    class="rounded-full p-1 text-gray-500 shadow-sm hover:bg-gray-100 hover:text-gray-700"
                    @click="dialogStore.openMoveDialog"
                >
                    <ArrowRightStartOnRectangleIcon class="size-5" aria-hidden="true" />
                </button>
                <template v-slot:text>{{ t("common.button.move") }}</template>
            </BaseTooltip>

            <BaseTooltip position="bottom">
                <button
                    @click="emit('selectAll')"
                    type="button"
                    class="rounded-full p-1 text-gray-500 shadow-sm hover:bg-gray-100 hover:text-gray-700"
                >
                    <Square2StackIcon class="size-5" aria-hidden="true" />
                </button>
                <template v-slot:text>{{ t("files.toolbar.select_all") }}</template>
            </BaseTooltip>

            <BaseTooltip position="bottom">
                <button
                    @click="dialogStore.openDeleteDialog"
                    type="button"
                    class="rounded-full p-1 text-red-500 shadow-sm hover:bg-gray-100"
                >
                    <TrashIcon class="size-5" aria-hidden="true" />
                </button>
                <template v-slot:text>{{ t("common.button.delete") }}</template>
            </BaseTooltip>
        </div>
    </transition>
    <BaseTooltip position="bottom">
        <BaseButton
            v-if="route.name === 'deleted'"
            rounded="full"
            :color="'text-red-500 hover:bg-gray-100 '"
            :is-disabled="filesStore.getFiles.length <= 0"
            :is-loading="deleting"
            :size="'small'"
            @click="emit('emptyTrash')"
        >
            <TrashIcon class="size-5" aria-hidden="true" />
        </BaseButton>
        <template v-slot:text>{{ t("files.toolbar.empty_trash") }}</template>
    </BaseTooltip>
</template>

<script setup>
import BaseButton from "@/components/BaseButton.vue";
import BaseTooltip from "@/components/BaseTooltip.vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useDialogStore } from "@/store/dialogs";
import { useFilesStore } from "@/store/files";
import {
    TrashIcon,
    Square2StackIcon,
    ArrowRightStartOnRectangleIcon,
} from "@heroicons/vue/24/outline";

defineProps({
    hasSelection: {
        type: Boolean,
        default: false,
    },
    deleting: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["selectAll", "emptyTrash"]);

const route = useRoute();
const dialogStore = useDialogStore();
const filesStore = useFilesStore();
</script>
