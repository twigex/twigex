<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="createOpen">
        <Dialog as="div" class="relative z-50" @close="closeCreateDialog">
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
                                {{ t("settings.storage.dialog.new_storage") }}
                            </DialogTitle>

                            <div class="mt-4 space-y-4">
                                <!-- Type selector -->
                                <div>
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.label.type") }}
                                    </label>
                                    <div class="mt-2 flex gap-x-3">
                                        <button
                                            type="button"
                                            :class="
                                                form.type === 1
                                                    ? 'bg-indigo-600 text-white ring-indigo-600'
                                                    : 'bg-white text-gray-900 ring-gray-300 hover:bg-gray-50'
                                            "
                                            class="flex-1 rounded-md px-3 py-2 text-sm font-semibold shadow-sm ring-1 ring-inset"
                                            @click="form.type = 1"
                                        >
                                            {{ t("settings.storage.type.local") }}
                                        </button>
                                        <button
                                            type="button"
                                            :class="
                                                form.type === 2
                                                    ? 'bg-indigo-600 text-white ring-indigo-600'
                                                    : 'bg-white text-gray-900 ring-gray-300 hover:bg-gray-50'
                                            "
                                            class="flex-1 rounded-md px-3 py-2 text-sm font-semibold shadow-sm ring-1 ring-inset"
                                            @click="form.type = 2"
                                        >
                                            {{ t("settings.storage.type.s3") }}
                                        </button>
                                    </div>
                                </div>

                                <!-- Name -->
                                <div>
                                    <label
                                        for="storage-name"
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.label.name") }}
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="form.label"
                                            id="storage-name"
                                            type="text"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                            :class="
                                                v$.form.label.$error
                                                    ? 'ring-red-600 focus:ring-red-600'
                                                    : 'ring-gray-300 focus:ring-indigo-600'
                                            "
                                            :placeholder="t('settings.storage.label.name')"
                                        />
                                        <p
                                            v-if="v$.form.label.$error"
                                            class="mt-1.5 text-sm text-red-600"
                                        >
                                            {{ t("common.error.required_field") }}
                                        </p>
                                    </div>
                                </div>

                                <!-- Description -->
                                <div>
                                    <label
                                        for="storage-desc"
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.storage.label.description") }}
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="form.description"
                                            id="storage-desc"
                                            type="text"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                            :placeholder="t('settings.storage.label.description')"
                                        />
                                    </div>
                                </div>

                                <!-- S3 fields -->
                                <template v-if="form.type === 2">
                                    <div>
                                        <label
                                            for="storage-endpoint"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.endpoint") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="form.endpoint"
                                                id="storage-endpoint"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    v$.form.endpoint.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                                placeholder="s3.amazonaws.com"
                                            />
                                            <p
                                                v-if="v$.form.endpoint.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="storage-bucket"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.bucket") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="form.bucket"
                                                id="storage-bucket"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    v$.form.bucket.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                                placeholder="my-bucket"
                                            />
                                            <p
                                                v-if="v$.form.bucket.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="storage-access"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.access_key") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="form.access_key"
                                                id="storage-access"
                                                type="text"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    v$.form.access_key.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                            />
                                            <p
                                                v-if="v$.form.access_key.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div>
                                        <label
                                            for="storage-secret"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.storage.label.secret_key") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="form.secret_key"
                                                id="storage-secret"
                                                type="password"
                                                autocomplete="new-password"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    v$.form.secret_key.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                            />
                                            <p
                                                v-if="v$.form.secret_key.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>
                                    <div class="flex items-center gap-x-3">
                                        <input
                                            v-model="form.ssl"
                                            id="storage-ssl"
                                            type="checkbox"
                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        />
                                        <label
                                            for="storage-ssl"
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
                                    @click="create"
                                >
                                    {{ t("common.button.create") }}
                                </button>
                                <button
                                    type="button"
                                    class="flex-1 inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                    @click="closeCreateDialog"
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

const emit = defineEmits(["created"]);

const createOpen = ref(false);

const form = ref({
    type: 1,
    label: "",
    description: "",
    endpoint: "",
    bucket: "",
    access_key: "",
    secret_key: "",
    ssl: true,
});

const rules = {
    form: {
        label: { required },
        endpoint: { required: requiredIf(() => form.value.type === 2) },
        bucket: { required: requiredIf(() => form.value.type === 2) },
        access_key: { required: requiredIf(() => form.value.type === 2) },
        secret_key: { required: requiredIf(() => form.value.type === 2) },
    },
};
const v$ = useVuelidate(rules, { form });

function openCreateDialog() {
    createOpen.value = true;
}

function closeCreateDialog() {
    createOpen.value = false;
    resetForm();
}

function resetForm() {
    form.value = {
        type: 1,
        label: "",
        description: "",
        endpoint: "",
        bucket: "",
        access_key: "",
        secret_key: "",
        ssl: true,
    };
    v$.value.$reset();
}

async function create() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    settingsService
        .createStorage(form.value)
        .then((res) => {
            emit("created", res.data);
            closeCreateDialog();
            useAlertStore().showSuccess(t.value("settings.storage.created_success"));
        })
        .catch((err) => {
            useAlertStore().showError(extractErrorMessage(err));
        });
}

defineExpose({ open: openCreateDialog });
</script>
