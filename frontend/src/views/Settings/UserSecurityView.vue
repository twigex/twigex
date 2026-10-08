<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-row justify-center h-full overflow-y-auto p-2">
        <div class="w-full md:w-3/4 space-y-6">
            <div class="pb-6 border-b border-gray-900/10">
                <h2 class="text-base font-semibold leading-7 text-gray-900">
                    {{ t("settings.user_security.multi_factor_authentication") }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-gray-600">
                    {{ t("settings.user_security.multi_factor_authentication_description") }}
                </p>
                <div class="mt-4">
                    <button
                        v-show="!userStore.user.mfaActive"
                        type="button"
                        class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        @click="openDialog"
                    >
                        {{ t("settings.user_security.button.enable_mfa") }}
                    </button>
                    <button
                        v-show="userStore.user.mfaActive"
                        type="button"
                        class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        @click="disable"
                    >
                        {{ t("settings.user_security.button.disable_mfa") }}
                    </button>
                </div>

                <TransitionRoot as="template" :show="dialog">
                    <Dialog as="div" class="relative z-50" @close="dialog = false">
                        <TransitionChild
                            as="template"
                            enter="ease-out duration-300"
                            enter-from="opacity-0"
                            enter-to="opacity-100"
                            leave="ease-in duration-200"
                            leave-from="opacity-100"
                            leave-to="opacity-0"
                        >
                            <div
                                class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity"
                            />
                        </TransitionChild>

                        <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                            <div
                                class="flex min-h-full items-end justify-center p-4 text-left sm:items-center sm:p-0"
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
                                        class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-xl sm:p-6"
                                    >
                                        <div>
                                            <div class="mt-3 text-center sm:mt-5">
                                                <DialogTitle
                                                    as="h3"
                                                    class="text-base font-semibold leading-6 text-gray-900"
                                                    >{{
                                                        t("settings.user_security.mfa_setup.title")
                                                    }}</DialogTitle
                                                >
                                                <div class="mt-2">
                                                    <p class="text-sm text-gray-500">
                                                        <strong>{{
                                                            t(
                                                                "settings.user_security.mfa_setup.step",
                                                                { n: 1 },
                                                            )
                                                        }}</strong>
                                                        {{
                                                            t(
                                                                "settings.user_security.mfa_setup.download",
                                                            )
                                                        }}
                                                    </p>
                                                    <p class="mt-1 text-sm text-gray-500">
                                                        <a
                                                            href="https://apps.apple.com/us/app/google-authenticator/id388497605?mt=8%27"
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                            class="font-semibold underline"
                                                            >App Store</a
                                                        >
                                                        ·
                                                        <a
                                                            href="https://play.google.com/store/apps/details?id=com.google.android.apps.authenticator2&hl=en"
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                            class="font-semibold underline"
                                                            >Google Play</a
                                                        >
                                                    </p>
                                                </div>
                                                <div class="mt-5">
                                                    <p class="text-sm text-gray-500">
                                                        <strong>{{
                                                            t(
                                                                "settings.user_security.mfa_setup.step",
                                                                { n: 2 },
                                                            )
                                                        }}</strong>
                                                        {{
                                                            t(
                                                                "settings.user_security.mfa_setup.scan",
                                                            )
                                                        }}
                                                    </p>

                                                    <img
                                                        class="mx-auto block bg-gray-50"
                                                        :src="'data:image/png;base64,' + qr_code"
                                                        alt=""
                                                    />

                                                    <p class="mt-1 text-sm leading-6 text-gray-600">
                                                        {{
                                                            t(
                                                                "settings.user_security.mfa_setup.secret",
                                                            )
                                                        }}
                                                        {{ secret }}
                                                    </p>
                                                </div>
                                                <div class="mt-5">
                                                    <p class="text-sm text-gray-500">
                                                        <strong>{{
                                                            t(
                                                                "settings.user_security.mfa_setup.step",
                                                                { n: 3 },
                                                            )
                                                        }}</strong>
                                                        {{
                                                            t(
                                                                "settings.user_security.mfa_setup.enter_code",
                                                            )
                                                        }}
                                                    </p>

                                                    <div class="mt-2">
                                                        <input
                                                            v-model="token"
                                                            name="mfa"
                                                            id="mfa"
                                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                                            :placeholder="
                                                                t(
                                                                    'settings.user_security.mfa_setup.code_placeholder',
                                                                )
                                                            "
                                                        />
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                        <div class="mt-5 sm:mt-6">
                                            <button
                                                type="button"
                                                class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                                @click="save"
                                            >
                                                {{ t("common.button.save") }}
                                            </button>
                                        </div>
                                    </DialogPanel>
                                </TransitionChild>
                            </div>
                        </div>
                    </Dialog>
                </TransitionRoot>
            </div>

            <div class="pb-6">
                <h2 class="text-base font-semibold leading-7 text-gray-900">
                    {{ t("settings.user_security.devices_and_sessions") }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-gray-600">
                    {{ t("settings.user_security.devices_and_sessions_description") }}
                </p>

                <ul role="list" class="mt-4 space-y-2">
                    <li
                        v-for="session in sortedSessions"
                        :key="session.created"
                        class="relative border border-gray-200 rounded-lg flex justify-between gap-x-6 py-4 px-4 hover:bg-gray-50"
                    >
                        <div class="flex min-w-0 gap-x-4">
                            <div class="min-w-0 flex-auto">
                                <p class="text-sm font-semibold leading-6 text-gray-900">
                                    {{ session.device_id }}
                                    ({{ session.browser }})
                                </p>
                                <p class="mt-1 text-xs leading-5 text-gray-500">
                                    {{ t("settings.user_security.last_ativity") }}
                                    {{ useDateOperations().getDateAndTime(session.last_activity) }}
                                </p>
                            </div>
                        </div>
                        <div class="flex shrink-0 items-center">
                            <button
                                type="button"
                                class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                @click="logout(session)"
                            >
                                {{ t("settings.user_security.logout") }}
                            </button>
                        </div>
                    </li>
                </ul>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, onMounted, computed } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { useUserStore } from "@/store/user";
import userService from "@/services/userService";
import useDateOperations from "@/composables/useDateOperations.js";

onMounted(() => {
    userService.userSessions().then((response) => {
        sessions.value = response.data;
    });
});

const sortedSessions = computed(() => {
    return [...sessions.value].sort((a, b) => {
        return new Date(b.last_activity) - new Date(a.last_activity);
    });
});

const userStore = useUserStore();
const sessions = ref([]);
const dialog = ref(false);
const qr_code = ref("");
const secret = ref("");
const token = ref("");

function logout(session) {
    userService.logoutUserSession(session.id).then(() => {
        sessions.value = sessions.value.filter((s) => s.id !== session.id);
    });
}

function openDialog() {
    userService.generateMfa().then((response) => {
        qr_code.value = response.data.qr_code;
        secret.value = response.data.secret;
    });

    dialog.value = true;
}

function save() {
    userService.enableMfa({ activate: true, token: token.value }).then(() => {
        dialog.value = false;

        userStore.user.mfaActive = true;
    });
}

function disable() {
    userService.enableMfa({ activate: false }).then(() => {
        dialog.value = false;

        userStore.user.mfaActive = false;
    });
}
</script>
