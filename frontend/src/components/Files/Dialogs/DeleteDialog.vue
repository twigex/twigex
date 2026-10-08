<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-50" @close="closeDialog()">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:size-10"
                                >
                                    <ExclamationTriangleIcon
                                        class="size-6 text-red-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                        >{{ t("files.delete_dialog.title") }}
                                    </DialogTitle>
                                    <div class="mt-2">
                                        <div
                                            v-if="isSharedFiles() && hasFolder()"
                                            class="rounded-md bg-red-50 p-4"
                                        >
                                            <div class="flex">
                                                <div class="shrink-0">
                                                    <XCircleIcon
                                                        class="size-5 text-red-400"
                                                        aria-hidden="true"
                                                    />
                                                </div>
                                                <div class="ml-3">
                                                    <h3 class="text-sm font-medium text-red-800">
                                                        Contains shared files
                                                    </h3>
                                                    <div class="mt-2 text-sm text-red-700">
                                                        <ul
                                                            role="list"
                                                            class="list-disc space-y-1 pl-5"
                                                        >
                                                            <li>
                                                                All files will be deleted inside
                                                                folder
                                                            </li>
                                                        </ul>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>

                                        <!-- eslint-disable vue/no-v-html -- tHtml escapes the variables and sanitizes the result -->
                                        <p
                                            v-else-if="hasFiles"
                                            v-html="
                                                tHtml('files.delete_dialog.confirm_delete', {
                                                    text:
                                                        files.length > 1
                                                            ? t('common.dialog.selected_files')
                                                            : files[0].name,
                                                })
                                            "
                                            class="text-sm text-gray-500"
                                        ></p>
                                        <!-- eslint-enable vue/no-v-html -->
                                    </div>
                                </div>
                            </div>

                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-red-500 sm:ml-3 sm:w-auto"
                                    @click="deleteFile"
                                >
                                    {{ t("common.button.delete") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="closeDialog"
                                    ref="cancelButtonRef"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t, tHtml } from "@/i18n/index.js";

import { ref, computed } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XCircleIcon } from "@heroicons/vue/20/solid";
import { ExclamationTriangleIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },

    files: {
        type: Array,
        default: () => [],
    },
});

const hasFiles = computed(() => props.files.length > 0);

const emit = defineEmits(["close", "delete"]);
const name = ref("");

function closeDialog() {
    name.value = "";
    emit("close");
}

function deleteFile() {
    emit("delete");
    closeDialog();
}

function isSharedFiles() {
    let shared = false;

    props.files.forEach((file) => {
        if (file.shared) {
            shared = true;
        }
    });

    return shared;
}

function hasFolder() {
    props.files.forEach((file) => {
        if (file.isFolder) {
            return true;
        }
    });

    return false;
}
</script>
