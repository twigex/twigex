<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="deleteOpen">
        <Dialog as="div" class="relative z-50" @close="closeDeleteDialog">
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
                                    class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10"
                                >
                                    <ExclamationTriangleIcon
                                        class="h-6 w-6 text-red-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.dialog.delete_title") }}
                                    </DialogTitle>
                                    <div class="mt-2 space-y-2">
                                        <p class="text-sm text-gray-500">
                                            {{ t("settings.storage.dialog.delete_warning") }}
                                        </p>
                                        <p class="text-sm font-medium text-gray-700">
                                            {{ t("settings.storage.dialog.delete_files_warning") }}
                                        </p>
                                        <p class="text-sm text-gray-500">
                                            {{
                                                t("settings.storage.dialog.delete_type_confirm", {
                                                    name: deleteTarget?.label,
                                                })
                                            }}
                                        </p>
                                        <input
                                            v-model="deleteConfirmText"
                                            type="text"
                                            class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-red-600 sm:text-sm sm:leading-6"
                                            :placeholder="deleteTarget?.label"
                                        />
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    :disabled="deleteConfirmText !== deleteTarget?.label"
                                    class="inline-flex w-full justify-center rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm sm:ml-3 sm:w-auto disabled:opacity-40 disabled:cursor-not-allowed"
                                    :class="
                                        deleteConfirmText === deleteTarget?.label
                                            ? 'bg-red-600 hover:bg-red-500'
                                            : 'bg-red-300'
                                    "
                                    @click="confirmDelete"
                                >
                                    {{ t("common.button.delete") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="closeDeleteDialog"
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
import { ref } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { ExclamationTriangleIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";
import settingsService from "@/services/settingsService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const emit = defineEmits(["deleted"]);

const deleteOpen = ref(false);
const deleteTarget = ref(null);
const deleteConfirmText = ref("");

function openDeleteDialog(storage) {
    deleteTarget.value = storage;
    deleteConfirmText.value = "";
    deleteOpen.value = true;
}

function closeDeleteDialog() {
    deleteOpen.value = false;
    deleteTarget.value = null;
    deleteConfirmText.value = "";
}

function confirmDelete() {
    if (deleteConfirmText.value !== deleteTarget.value?.label) return;

    settingsService
        .deleteStorage(deleteTarget.value.id)
        .then(() => {
            emit("deleted", deleteTarget.value.id);
            closeDeleteDialog();
            useAlertStore().showSuccess(t.value("settings.storage.deleted_success"));
        })
        .catch((err) => {
            useAlertStore().showError(extractErrorMessage(err));
        });
}

defineExpose({ open: openDeleteDialog });
</script>
