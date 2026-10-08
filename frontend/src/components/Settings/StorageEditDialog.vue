<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="editOpen">
        <Dialog as="div" class="relative z-50" @close="closeEditDialog">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div class="flex min-h-full items-end justify-center p-4 sm:items-center sm:p-0">
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
                            class="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <DialogTitle
                                as="h3"
                                class="text-base font-semibold leading-6 text-gray-900"
                            >
                                {{ t("settings.storage.dialog.edit_storage") }}
                            </DialogTitle>

                            <div class="mt-4 space-y-4">
                                <!-- Name -->
                                <div>
                                    <label
                                        for="edit-storage-name"
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.label.name") }}
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="editForm.label"
                                            id="edit-storage-name"
                                            type="text"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                            :class="
                                                ve$.editForm.label.$error
                                                    ? 'ring-red-600 focus:ring-red-600'
                                                    : 'ring-gray-300 focus:ring-indigo-600'
                                            "
                                        />
                                        <p
                                            v-if="ve$.editForm.label.$error"
                                            class="mt-1.5 text-sm text-red-600"
                                        >
                                            {{ t("common.error.required_field") }}
                                        </p>
                                    </div>
                                </div>

                                <!-- Description -->
                                <div>
                                    <label
                                        for="edit-storage-desc"
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.label.description") }}
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="editForm.description"
                                            id="edit-storage-desc"
                                            type="text"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                        />
                                    </div>
                                </div>

                                <!-- S3 fields -->
                                <template v-if="editForm.type === 2">
                                    <div>
                                        <label
                                            for="edit-storage-endpoint"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.endpoint") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="editForm.endpoint"
                                                id="edit-storage-endpoint"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    ve$.editForm.endpoint.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                                placeholder="s3.amazonaws.com"
                                            />
                                            <p
                                                v-if="ve$.editForm.endpoint.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="edit-storage-bucket"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.bucket") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="editForm.bucket"
                                                id="edit-storage-bucket"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    ve$.editForm.bucket.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                                placeholder="my-bucket"
                                            />
                                            <p
                                                v-if="ve$.editForm.bucket.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="edit-storage-access"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.access_key") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="editForm.access_key"
                                                id="edit-storage-access"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    ve$.editForm.access_key.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                            />
                                            <p
                                                v-if="ve$.editForm.access_key.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="edit-storage-secret"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.secret_key") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="editForm.secret_key"
                                                id="edit-storage-secret"
                                                type="password"
                                                autocomplete="new-password"
                                                :placeholder="
                                                    t(
                                                        'settings.storage.label.secret_key_placeholder',
                                                    )
                                                "
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                            />
                                            <p class="mt-1 text-xs text-gray-400">
                                                {{ t("settings.storage.label.secret_key_hint") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div class="flex items-center gap-x-3">
                                        <input
                                            v-model="editForm.ssl"
                                            id="edit-storage-ssl"
                                            type="checkbox"
                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        />
                                        <label
                                            for="edit-storage-ssl"
                                            class="text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.ssl") }}
                                        </label>
                                    </div>
                                </template>
                            </div>

                            <div class="mt-6 flex gap-x-3 sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="flex-1 inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                    @click="update"
                                >
                                    {{ t("common.button.save") }}
                                </button>
                                <button
                                    type="button"
                                    class="flex-1 inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    @click="closeEditDialog"
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
import { useVuelidate } from "@vuelidate/core";
import { required, requiredIf } from "@vuelidate/validators";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { t } from "@/i18n/index.js";
import settingsService from "@/services/settingsService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const emit = defineEmits(["updated"]);

const editOpen = ref(false);
const editTarget = ref(null);

const editForm = ref({
    type: 1,
    label: "",
    description: "",
    endpoint: "",
    bucket: "",
    access_key: "",
    secret_key: "",
    ssl: true,
});

const editRules = {
    editForm: {
        label: { required },
        endpoint: { required: requiredIf(() => editForm.value.type === 2) },
        bucket: { required: requiredIf(() => editForm.value.type === 2) },
        access_key: { required: requiredIf(() => editForm.value.type === 2) },
    },
};
const ve$ = useVuelidate(editRules, { editForm });

function openEditDialog(storage) {
    editTarget.value = storage;
    editForm.value = {
        type: storage.type,
        label: storage.label,
        description: storage.description,
        endpoint: storage.endpoint || "",
        bucket: storage.bucket || "",
        access_key: storage.access_key || "",
        secret_key: "",
        ssl: storage.ssl ?? true,
    };
    ve$.value.$reset();
    editOpen.value = true;
}

function closeEditDialog() {
    editOpen.value = false;
    editTarget.value = null;
    ve$.value.$reset();
}

async function update() {
    const isValid = await ve$.value.$validate();

    if (!isValid) return;

    settingsService
        .updateStorage(editTarget.value.id, editForm.value)
        .then((res) => {
            emit("updated", editTarget.value.id, res.data);
            closeEditDialog();
            useAlertStore().showSuccess(t.value("settings.storage.updated_success"));
        })
        .catch((err) => {
            useAlertStore().showError(extractErrorMessage(err));
        });
}

defineExpose({ open: openEditDialog });
</script>
