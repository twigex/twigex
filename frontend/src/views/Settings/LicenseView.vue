<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-row justify-center h-full overflow-y-auto p-2">
        <div v-if="loaded" class="w-full md:w-3/4 space-y-6">
            <input type="file" ref="fileInput" class="hidden" @change="handleFileUpload" />

            <div v-if="license != null" class="space-y-6">
                <div
                    v-if="seatsOverLimit"
                    class="flex items-start gap-x-2 rounded-md border border-amber-200 bg-amber-50 px-4 py-3"
                >
                    <LockClosedIcon
                        class="mt-0.5 h-4 w-4 shrink-0 text-amber-700"
                        aria-hidden="true"
                    />
                    <div class="min-w-0">
                        <p class="text-sm font-medium text-amber-800">
                            {{
                                t("settings.license.seats_over.title", {
                                    used: activeUsers,
                                    licensed: seatLimit,
                                })
                            }}
                        </p>
                        <p class="mt-0.5 text-xs text-amber-700">
                            {{
                                t("settings.license.seats_over.description", {
                                    hard: seatHardLimit,
                                })
                            }}
                        </p>
                    </div>
                </div>

                <div class="pb-6 border-b border-gray-900/10">
                    <p class="text-sm text-gray-500">
                        {{ t("settings.license.enterpise_edition") }}
                    </p>
                    <h2 class="mt-1 text-2xl font-bold text-gray-900">
                        {{ license.sku_name }}
                    </h2>
                    <p class="mt-2 text-sm leading-6 text-gray-600">
                        {{
                            t("settings.license.description", {
                                name: license.sku_name,
                            })
                        }}
                    </p>
                </div>

                <div class="pb-6 border-b border-gray-900/10">
                    <h3 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.license.license_details") }}
                    </h3>
                    <dl class="mt-4 space-y-3">
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.starts_from") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ getDate(license.starts_at) }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.expires_at") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ getDate(license.expires_at) }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.issued_at") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ getDateAndTime(license.issued_at) }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.licensed_seats") }}
                            </dt>
                            <dd
                                class="tabular-nums"
                                :class="
                                    seatsOverLimit
                                        ? 'font-semibold text-amber-700'
                                        : 'text-gray-900'
                                "
                            >
                                {{
                                    t("settings.license.seats_in_use", {
                                        used: activeUsers,
                                        licensed: license.features.users,
                                    })
                                }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.edition") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ license.sku_name }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.name") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ license.customer.name }}
                            </dd>
                        </div>
                        <div class="flex justify-between text-sm">
                            <dt class="font-medium text-gray-500">
                                {{ t("settings.license.company") }}
                            </dt>
                            <dd class="text-gray-900">
                                {{ license.customer.company }}
                            </dd>
                        </div>
                    </dl>
                </div>

                <EnvManagedNotice v-if="envLocked" class="mb-4" />

                <div class="flex items-center justify-between pb-6">
                    <button
                        :disabled="envLocked"
                        @click="$refs.fileInput.click()"
                        class="rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        :class="
                            envLocked
                                ? 'bg-gray-300 cursor-not-allowed'
                                : 'bg-indigo-600 hover:bg-indigo-500'
                        "
                    >
                        {{ t("settings.license.button.add_new_license") }}
                    </button>

                    <button
                        type="button"
                        :disabled="envLocked"
                        @click="open = true"
                        class="text-sm font-semibold"
                        :class="
                            envLocked
                                ? 'text-gray-400 cursor-not-allowed'
                                : 'text-red-600 hover:text-red-500'
                        "
                    >
                        {{ t("settings.license.button.remove_license") }}
                    </button>

                    <TransitionRoot as="template" :show="open">
                        <Dialog class="relative z-10" @close="open = false">
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
                                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
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
                                                <div
                                                    class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left"
                                                >
                                                    <DialogTitle
                                                        as="h3"
                                                        class="text-base font-semibold text-gray-900"
                                                        >{{
                                                            t(
                                                                "settings.license.remove_fialog.title",
                                                            )
                                                        }}</DialogTitle
                                                    >
                                                    <div class="mt-2">
                                                        <p class="text-sm text-gray-500">
                                                            {{
                                                                t(
                                                                    "settings.license.remove_fialog.confirm",
                                                                )
                                                            }}
                                                        </p>
                                                    </div>
                                                </div>
                                            </div>
                                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                                <button
                                                    type="button"
                                                    class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-red-500 sm:ml-3 sm:w-auto"
                                                    @click="removeLicense()"
                                                >
                                                    {{ t("common.button.remove") }}
                                                </button>
                                                <button
                                                    type="button"
                                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-xs ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                                    @click="open = false"
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
                </div>
            </div>

            <div v-else class="pb-6">
                <p class="text-sm text-gray-500">
                    {{ t("settings.license.enterpise_edition") }}
                </p>
                <h2 class="mt-1 text-2xl font-bold text-gray-900">Free</h2>
                <p class="mt-2 text-sm leading-6 text-gray-600">
                    {{ t("settings.license.enterpise_edition.purchase") }}
                </p>

                <div class="mt-6 border border-gray-200 rounded-lg p-4">
                    <h3 class="text-sm font-semibold text-gray-900">
                        {{ t("settings.license.license_details") }}
                    </h3>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("settings.license.see_license_details") }}
                    </p>
                </div>

                <EnvManagedNotice v-if="envLocked" class="mt-4" />

                <div class="mt-6">
                    <button
                        :disabled="envLocked"
                        @click="$refs.fileInput.click()"
                        class="rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        :class="
                            envLocked
                                ? 'bg-gray-300 cursor-not-allowed'
                                : 'bg-indigo-600 hover:bg-indigo-500'
                        "
                    >
                        {{ t("settings.license.button.add_new_license") }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { computed, ref, onMounted } from "vue";
import EnvManagedNotice from "@/components/Settings/EnvManagedNotice.vue";
import settingsService from "@/services/settingsService";
import { useSettingsStore } from "@/store/settings";
import useDateOperations from "@/composables/useDateOperations";
import { LockClosedIcon, ExclamationTriangleIcon } from "@heroicons/vue/24/outline";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const { getDate, getDateAndTime } = useDateOperations();
const settingsStore = useSettingsStore();

const seatsOverLimit = computed(() => settingsStore.getLicenseFeature("seats_over_limit"));
const activeUsers = computed(() => settingsStore.getLicenseFeature("active_users"));
const seatLimit = computed(() => settingsStore.getLicenseFeature("seat_limit"));
const seatHardLimit = computed(() => settingsStore.getLicenseFeature("seat_hard_limit"));
const loaded = ref(false);
const license = ref(null);
const envLocked = ref(false);
const open = ref(false);

const handleFileUpload = (event) => {
    const file = event.target.files[0];

    settingsService.uploadLicense(file).then((response) => {
        license.value = response.data.license;
        settingsStore.setLicense(response.data.map);
    });
};

function removeLicense() {
    settingsService.removeActiveLicense().then(() => {
        license.value = null;
        settingsStore.setLicense(null);
    });

    open.value = false;
}

onMounted(() => {
    settingsService.getActiveLicense().then((response) => {
        license.value = response.data.license;
        envLocked.value = response.data.env_locked;

        loaded.value = true;
    });
});
</script>
