<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<!-- LAST -->
<template>
    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-50" @close="closeDialog()">
            <!-- Overlay -->
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-900/40 backdrop-blur-sm" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 flex items-center justify-center p-4">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-300"
                    enter-from="opacity-0 translate-y-4 scale-95"
                    enter-to="opacity-100 translate-y-0 scale-100"
                    leave="ease-in duration-200"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <DialogPanel
                        class="w-full max-w-lg overflow-hidden rounded-xl bg-white shadow-2xl ring-1 ring-black/5"
                    >
                        <!-- Header -->
                        <div class="border-b px-5 py-4">
                            <DialogTitle class="text-base font-semibold text-gray-900 truncate">
                                {{ t("files.move_dialog.title") }}
                                {{
                                    files.length > 1
                                        ? t("common.dialog.selected_files")
                                        : files[0].name
                                }}
                            </DialogTitle>
                            <p class="mt-1 text-sm text-gray-500">
                                {{ t("files.move_dialog.select_folder") }}
                            </p>
                        </div>

                        <!-- Content -->
                        <div class="px-5 py-4">
                            <div
                                class="min-h-72 max-h-72 overflow-y-auto rounded-lg border bg-white"
                            >
                                <ul
                                    v-if="state.files.length > 0"
                                    role="list"
                                    class="divide-y divide-gray-200"
                                >
                                    <li
                                        v-for="file in state.files"
                                        :key="file.id"
                                        @dblclick="loadChildren(file)"
                                        @click="setActive(file)"
                                        class="group flex items-center justify-between px-4 py-3 cursor-pointer"
                                        :class="{
                                            'bg-indigo-600 text-white': state.activeFile === file,
                                            'hover:bg-gray-100': state.activeFile !== file,
                                        }"
                                    >
                                        <div class="flex items-center gap-3 min-w-0">
                                            <FolderIcon
                                                class="h-5 w-5 flex-shrink-0"
                                                :class="
                                                    state.activeFile === file
                                                        ? 'text-white'
                                                        : 'text-gray-400'
                                                "
                                            />
                                            <p
                                                class="truncate text-sm font-medium"
                                                :class="
                                                    state.activeFile === file
                                                        ? 'text-white'
                                                        : 'text-gray-900'
                                                "
                                            >
                                                {{ file.name }}
                                            </p>
                                        </div>

                                        <div class="flex items-center gap-2">
                                            <ShareIcon
                                                v-if="file.shared"
                                                class="h-4 w-4 text-indigo-400 group-hover:text-white"
                                            />
                                            <button
                                                @click.stop="loadChildren(file)"
                                                class="rounded-md p-1 hover:bg-indigo-500/20"
                                            >
                                                <ChevronRightIcon
                                                    class="h-4 w-4"
                                                    :class="
                                                        state.activeFile === file
                                                            ? 'text-white'
                                                            : 'text-gray-400'
                                                    "
                                                />
                                            </button>
                                        </div>
                                    </li>
                                </ul>

                                <!-- Empty -->
                                <div
                                    v-else
                                    class="flex h-72 flex-col items-center justify-center text-center"
                                >
                                    <FolderOpenIcon class="h-14 w-14 text-gray-300" />
                                    <p class="mt-3 text-sm font-medium text-gray-700">
                                        {{ t("files.move_dialog.folder_empty") }}
                                    </p>
                                </div>
                            </div>

                            <!-- Breadcrumbs -->
                            <div
                                v-if="state.breadcrumbs.length > 1"
                                class="mt-3 text-xs text-gray-500"
                            >
                                <nav class="flex items-center gap-1">
                                    <span
                                        v-for="(page, index) in state.breadcrumbs"
                                        :key="index"
                                        class="flex items-center gap-1 cursor-pointer hover:text-gray-700"
                                        @click="breadcrumbClick(page, index)"
                                    >
                                        <span class="truncate max-w-24">
                                            {{ page.name }}
                                        </span>
                                        <ChevronRightIcon
                                            v-if="index !== state.breadcrumbs.length - 1"
                                            class="h-3 w-3"
                                        />
                                    </span>
                                </nav>
                            </div>
                        </div>

                        <!-- Warning -->
                        <div v-if="isDifferentDisk" class="px-5 pb-3">
                            <p
                                class="text-xs text-amber-600 bg-amber-50 border border-amber-200 rounded-md px-3 py-2"
                            >
                                {{ t("files.move_dialog.warning") }}
                            </p>
                        </div>

                        <!-- Footer -->
                        <div
                            class="flex items-center justify-end gap-2 border-t bg-white px-5 py-3"
                        >
                            <button
                                type="button"
                                @click="closeDialog"
                                class="rounded-md px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
                            >
                                {{ t("common.button.cancel") }}
                            </button>

                            <BaseButton
                                :isDisabled="state.activeFile === null && state.location === null"
                                :isLoading="state.moving"
                                @click="moveFile"
                            >
                                {{
                                    state.moving
                                        ? t("files.move_dialog.button.moving")
                                        : t("common.button.move")
                                }}
                            </BaseButton>
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { FolderIcon, ChevronRightIcon } from "@heroicons/vue/24/outline";
import { FolderOpenIcon } from "@heroicons/vue/24/solid";
import BaseButton from "@/components/BaseButton.vue";
import { ShareIcon } from "@heroicons/vue/20/solid";
import { nextTick, ref, reactive, watch, computed } from "vue";
import fileService from "@/services/fileService";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },

    files: {
        type: Object,
        default: () => ({}),
    },
});

const isDifferentDisk = computed(() => {
    if (props.files.length === 0) {
        return false;
    }

    if (state.location === null) {
        return false;
    }

    return props.files[0].storage !== state.location.storage;
});

const emit = defineEmits(["close", "move"]);
const name = ref("");
const state = reactive({
    files: [],
    activeFile: null,
    location: null, //currently opened folder
    breadcrumbs: [],
    moving: false,
});

watch(
    () => props.open,
    async () => {
        if (props.open) {
            await nextTick();
            await fileService.getDrive().then((res) => {
                state.files = res.data;
            });
            await fileService.sharedFiles().then((res) => {
                let sharedFolders = res.data.filter((item) => item.isFolder);

                state.files.push(...sharedFolders);
            });
        }
    },
);

function loadChildren(file) {
    state.activeFile = null;
    state.location = file;
    fileService.getFilesFromFolder(file.id).then((res) => {
        state.files = res.data.filter(
            (file) => file.isFolder && !props.files.some((f) => f.id === file.id),
        );
    });

    state.breadcrumbs.push(file);
}

function setActive(file) {
    if (state.activeFile === file) {
        state.activeFile = null;

        return;
    }

    state.activeFile = file;
}

function closeDialog() {
    emit("close");

    name.value = "";
    state.activeFile = null;
    state.location = null;
    state.breadcrumbs = [];
    state.moving = false;
}

function moveFile() {
    if (state.activeFile === null && state.location === null) {
        return;
    }

    let location;

    if (state.activeFile !== null) {
        location = state.activeFile.id;
    } else {
        location = state.location.id;
    }

    state.moving = true;

    emit("move", { location: location });
}

function breadcrumbClick(page, index) {
    state.breadcrumbs = state.breadcrumbs.slice(0, index);
    loadChildren(page);
}

defineExpose({
    closeDialog,
});
</script>
