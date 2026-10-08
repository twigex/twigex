<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <Dialog class="relative z-10" @close="emit('close')">
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <DialogTitle as="h3" class="text-base font-semibold text-gray-900 mb-5">
                                {{
                                    editing
                                        ? t("settings.authorization.oidc_edit_provider")
                                        : t("settings.authorization.oidc_add_identity_provider")
                                }}
                            </DialogTitle>

                            <div class="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-6">
                                <div v-if="form.redirect_url" class="col-span-full">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_redirect_url") }}</label
                                    >
                                    <div class="mt-2 flex items-center gap-x-2">
                                        <input
                                            :value="form.redirect_url"
                                            type="text"
                                            readonly
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-600 bg-gray-50 shadow-sm ring-1 ring-inset ring-gray-300 sm:text-sm sm:leading-6 font-mono text-xs"
                                        />
                                        <button
                                            @click="emit('copy', form.redirect_url)"
                                            type="button"
                                            class="shrink-0 rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                        >
                                            {{ t("settings.authorization.oidc_copy") }}
                                        </button>
                                    </div>
                                    <p class="mt-1 text-xs text-gray-500">
                                        {{ t("settings.authorization.oidc_redirect_url_hint") }}
                                    </p>
                                </div>

                                <div class="col-span-full">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_provider_name") }}</label
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="form.name"
                                            type="text"
                                            :placeholder="
                                                t(
                                                    'settings.authorization.oidc_provider_name_placeholder',
                                                )
                                            "
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                        />
                                    </div>
                                </div>

                                <div class="col-span-full">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.authorization.oidc_discovery_url") }}
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="form.discovery_url"
                                            type="url"
                                            placeholder="https://provider.example.com/.well-known/openid-configuration"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono text-xs"
                                        />
                                    </div>
                                    <p class="mt-1 text-xs text-gray-500">
                                        {{ t("settings.authorization.oidc_discovery_url_hint") }}
                                    </p>
                                </div>

                                <div class="sm:col-span-3">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_client_id") }}</label
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="form.client_id"
                                            type="text"
                                            :placeholder="
                                                t(
                                                    'settings.authorization.oidc_client_id_placeholder',
                                                )
                                            "
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono"
                                        />
                                    </div>
                                </div>

                                <div class="sm:col-span-3">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("settings.authorization.oidc_client_secret") }}
                                        <span v-if="!editing">*</span>
                                        <span
                                            v-else
                                            class="text-xs font-normal text-gray-500 ml-1"
                                            >{{
                                                t("settings.authorization.oidc_client_secret_keep")
                                            }}</span
                                        >
                                    </label>
                                    <div class="mt-2">
                                        <input
                                            v-model="form.client_secret"
                                            type="password"
                                            autocomplete="new-password"
                                            :placeholder="
                                                editing
                                                    ? '••••••••'
                                                    : t(
                                                          'settings.authorization.oidc_client_secret_placeholder',
                                                      )
                                            "
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono"
                                        />
                                    </div>
                                </div>

                                <div class="col-span-full">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_scopes") }}</label
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="form.scopes"
                                            type="text"
                                            placeholder="openid,profile,email"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono"
                                        />
                                    </div>
                                    <p class="mt-1 text-xs text-gray-500">
                                        {{ t("settings.authorization.oidc_scopes_hint") }}
                                    </p>
                                </div>

                                <div class="sm:col-span-3">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_button_text") }}</label
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="form.button_text"
                                            type="text"
                                            :placeholder="
                                                t('login.button.login_with', { name: form.name })
                                            "
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                        />
                                    </div>
                                </div>

                                <div class="sm:col-span-3">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_button_color") }}</label
                                    >
                                    <div class="mt-2 flex items-center gap-x-2">
                                        <input
                                            v-model="form.button_color"
                                            type="color"
                                            class="h-9 w-12 rounded border border-gray-300 cursor-pointer p-0.5"
                                        />
                                        <input
                                            v-model="form.button_color"
                                            type="text"
                                            placeholder="#4f46e5"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono"
                                        />
                                    </div>
                                </div>

                                <!-- Button preview -->
                                <div v-if="form.name" class="col-span-full">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.authorization.oidc_preview") }}</label
                                    >
                                    <div class="mt-2">
                                        <button
                                            type="button"
                                            class="w-full rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm"
                                            :style="{
                                                backgroundColor: form.button_color || '#4f46e5',
                                            }"
                                        >
                                            {{
                                                form.button_text ||
                                                t("login.button.login_with", { name: form.name })
                                            }}
                                        </button>
                                    </div>
                                </div>

                                <div class="col-span-full">
                                    <div class="relative flex gap-x-3">
                                        <div class="flex h-6 items-center">
                                            <input
                                                v-model="form.enabled"
                                                id="provider-enabled"
                                                type="checkbox"
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                        </div>
                                        <div class="text-sm leading-6">
                                            <label
                                                for="provider-enabled"
                                                class="font-medium text-gray-900"
                                                >{{
                                                    t("settings.authorization.oidc_enable_provider")
                                                }}</label
                                            >
                                            <p class="text-gray-500">
                                                {{
                                                    t(
                                                        "settings.authorization.oidc_enable_provider_description",
                                                    )
                                                }}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="mt-6 flex items-center justify-end gap-x-3">
                                <button
                                    type="button"
                                    @click="emit('close')"
                                    class="text-sm font-semibold text-gray-900"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    @click="emit('save')"
                                    :disabled="!valid || saving"
                                    type="button"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm"
                                    :class="
                                        !valid || saving
                                            ? 'bg-gray-300 cursor-not-allowed'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                >
                                    {{
                                        saving
                                            ? t("settings.authorization.oidc_saving")
                                            : editing
                                              ? t("settings.authorization.oidc_save_changes")
                                              : t("settings.authorization.oidc_add_provider_button")
                                    }}
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
import { t } from "@/i18n/index.js";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    editing: {
        type: Boolean,
        default: false,
    },
    saving: {
        type: Boolean,
        default: false,
    },
    valid: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["close", "save", "copy"]);

const form = defineModel("form", { type: Object, required: true });
</script>
