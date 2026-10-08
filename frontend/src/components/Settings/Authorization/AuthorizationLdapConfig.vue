<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="ml-7">
        <div class="rounded-md bg-gray-50 border border-gray-200 p-4">
            <h3 class="text-sm font-medium text-gray-900 mb-4">
                {{ t("settings.authorization.ldap_configuration") }}
            </h3>
            <EnvManagedNotice v-if="envLocked" class="mb-4" />
            <div class="grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-6">
                <div class="sm:col-span-2">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_host")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.host"
                            type="text"
                            placeholder="ldap.example.com"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-1">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_port")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model.number="ldap.port"
                            type="number"
                            min="1"
                            max="65535"
                            :placeholder="defaultPort"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-1">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_timeout")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model.number="ldap.timeoutSeconds"
                            type="number"
                            min="1"
                            max="120"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-3">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_base_dn")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.baseDN"
                            type="text"
                            placeholder="dc=example,dc=com"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-3">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_bind_dn")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.bindDN"
                            type="text"
                            placeholder="cn=admin,dc=example,dc=com"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-3">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_bind_password")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.bindPassword"
                            type="password"
                            autocomplete="new-password"
                            placeholder="••••••••"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-2">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_user_filter")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.userFilter"
                            type="text"
                            placeholder="(uid={0})"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-2">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_username_attr")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.usernameAttr"
                            type="text"
                            placeholder="uid"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="col-span-full">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_id_attr")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.idAttr"
                            type="text"
                            placeholder="entryUUID"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">
                        {{ t("settings.authorization.ldap_id_attr_hint") }}
                    </p>
                </div>

                <div class="sm:col-span-2">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_email_attr")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.emailAttr"
                            type="text"
                            placeholder="mail"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-3">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_first_name_attr")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.firstNameAttr"
                            type="text"
                            placeholder="givenName"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="sm:col-span-3">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_last_name_attr")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.lastNameAttr"
                            type="text"
                            placeholder="sn"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div class="col-span-full">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_allowed_groups")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model="ldap.allowedGroups"
                            type="text"
                            placeholder="cn=twigex,ou=groups,dc=example,dc=com"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">
                        {{ t("settings.authorization.ldap_allowed_groups_hint") }}
                    </p>
                </div>

                <div class="col-span-full">
                    <div class="relative flex gap-x-3">
                        <div class="flex h-6 items-center">
                            <input
                                v-model="ldap.syncEnabled"
                                id="ldap-sync"
                                type="checkbox"
                                :disabled="envLocked"
                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                            />
                        </div>
                        <div class="text-sm leading-6">
                            <label for="ldap-sync" class="font-medium text-gray-900">
                                {{ t("settings.authorization.ldap_sync") }}
                            </label>
                            <p class="text-gray-500">
                                {{ t("settings.authorization.ldap_sync_hint") }}
                            </p>
                        </div>
                    </div>
                </div>

                <div v-if="ldap.syncEnabled" class="sm:col-span-1">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_sync_interval")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model.number="ldap.syncIntervalMinutes"
                            type="number"
                            min="5"
                            max="1440"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                </div>

                <div v-if="ldap.syncEnabled" class="sm:col-span-2">
                    <label class="block text-sm font-medium leading-6 text-gray-900">{{
                        t("settings.authorization.ldap_sync_max_deactivate")
                    }}</label>
                    <div class="mt-2">
                        <input
                            v-model.number="ldap.syncMaxDeactivatePercent"
                            type="number"
                            min="1"
                            max="100"
                            :disabled="envLocked"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 font-mono disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                        />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">
                        {{ t("settings.authorization.ldap_sync_max_deactivate_hint") }}
                    </p>
                </div>

                <div class="col-span-full">
                    <Listbox as="div" v-model="securityOption" :disabled="envLocked">
                        <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("settings.authorization.ldap_connection_security") }}
                        </ListboxLabel>
                        <div class="relative mt-2">
                            <ListboxButton
                                class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6 disabled:bg-gray-100 disabled:text-gray-500 disabled:cursor-not-allowed"
                            >
                                <span class="col-start-1 row-start-1 truncate pr-6">
                                    {{ t(securityOption.labelKey) }}
                                </span>
                                <ChevronUpDownIcon
                                    class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                                    aria-hidden="true"
                                />
                            </ListboxButton>

                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg outline outline-1 outline-black/5 sm:text-sm"
                                >
                                    <ListboxOption
                                        as="template"
                                        v-for="option in securityOptions"
                                        :key="option.value"
                                        :value="option"
                                        v-slot="{ active, selected }"
                                    >
                                        <li
                                            :class="[
                                                active
                                                    ? 'bg-indigo-600 text-white outline-none'
                                                    : 'text-gray-900',
                                                'relative cursor-default select-none py-2 pl-3 pr-9',
                                            ]"
                                        >
                                            <span
                                                :class="[
                                                    selected ? 'font-semibold' : 'font-normal',
                                                    'block truncate',
                                                ]"
                                            >
                                                {{ t(option.labelKey) }}
                                            </span>
                                            <span
                                                v-if="selected"
                                                :class="[
                                                    active ? 'text-white' : 'text-indigo-600',
                                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                                ]"
                                            >
                                                <CheckIcon class="size-5" aria-hidden="true" />
                                            </span>
                                        </li>
                                    </ListboxOption>
                                </ListboxOptions>
                            </transition>
                        </div>
                    </Listbox>
                    <p class="mt-1 text-xs text-gray-500">
                        {{ t("settings.authorization.ldap_connection_security_hint") }}
                    </p>
                </div>

                <div v-if="usesTls" class="col-span-full">
                    <label
                        for="ldap-ca-cert"
                        class="block text-sm font-medium leading-6 text-gray-900"
                    >
                        {{ t("settings.authorization.ldap_ca_cert") }}
                    </label>
                    <div class="mt-2">
                        <textarea
                            v-model="ldap.caCertificate"
                            id="ldap-ca-cert"
                            rows="4"
                            spellcheck="false"
                            :disabled="envLocked"
                            placeholder="-----BEGIN CERTIFICATE-----"
                            class="block w-full rounded-md border-0 py-1.5 px-3 font-mono text-xs text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 disabled:bg-gray-50 disabled:text-gray-500"
                        />
                    </div>
                    <p class="mt-1 text-xs text-gray-500">
                        {{ t("settings.authorization.ldap_ca_cert_hint") }}
                    </p>
                </div>

                <div v-if="usesTls" class="col-span-full">
                    <div class="relative flex gap-x-3">
                        <div class="flex h-6 items-center">
                            <input
                                v-model="ldap.insecureSkipVerify"
                                id="ldap-skip-tls"
                                type="checkbox"
                                :disabled="envLocked"
                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                            />
                        </div>
                        <div class="text-sm leading-6">
                            <label for="ldap-skip-tls" class="font-medium text-gray-900">
                                {{ t("settings.authorization.ldap_skip_tls") }}
                            </label>
                            <p class="text-gray-500">
                                {{ t("settings.authorization.ldap_skip_tls_description") }}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="col-span-full flex items-center gap-x-4">
                    <button
                        @click="emit('test')"
                        :disabled="testing || envLocked"
                        type="button"
                        class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {{
                            testing
                                ? t("settings.authorization.ldap_testing")
                                : t("settings.smtp.test_connection")
                        }}
                    </button>
                    <p
                        v-if="testResult"
                        class="text-sm"
                        :class="testResult.ok ? 'text-green-600' : 'text-red-600'"
                    >
                        {{ testResult.message }}
                    </p>
                </div>

                <div v-if="ldap.syncEnabled" class="col-span-full flex items-center gap-x-4">
                    <button
                        @click="emit('sync')"
                        :disabled="syncing"
                        type="button"
                        class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {{
                            syncing
                                ? t("settings.authorization.ldap_syncing")
                                : t("settings.authorization.ldap_sync_now")
                        }}
                    </button>
                    <p
                        v-if="syncMessage"
                        class="text-sm"
                        :class="syncMessage.ok ? 'text-green-600' : 'text-red-600'"
                    >
                        {{ syncMessage.message }}
                    </p>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/16/solid";
import EnvManagedNotice from "@/components/Settings/EnvManagedNotice.vue";

defineProps({
    envLocked: {
        type: Boolean,
        default: false,
    },
    defaultPort: {
        type: String,
        default: "",
    },
    securityOptions: {
        type: Array,
        default: () => [],
    },
    usesTls: {
        type: Boolean,
        default: false,
    },
    testing: {
        type: Boolean,
        default: false,
    },
    testResult: {
        type: Object,
        default: null,
    },
    syncing: {
        type: Boolean,
        default: false,
    },
    syncMessage: {
        type: Object,
        default: null,
    },
});

const emit = defineEmits(["test", "sync"]);

const ldap = defineModel("ldap", { type: Object, required: true });

const securityOption = defineModel("securityOption", { type: Object, required: true });
</script>
