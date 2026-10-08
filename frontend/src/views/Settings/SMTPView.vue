<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="flex flex-col h-full overflow-hidden">
        <div class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="w-full md:w-3/4 space-y-6">
                <div class="pb-6 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.smtp.title") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.smtp.description") }}
                    </p>

                    <EnvManagedNotice v-if="server.new.env_locked" class="mt-4" />

                    <div class="mt-6 relative flex gap-x-3">
                        <div class="flex h-6 items-center">
                            <input
                                v-model="server.new.EnableEmail"
                                id="email-enabled"
                                name="email-enabled"
                                type="checkbox"
                                :disabled="server.new.env_locked"
                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                            />
                        </div>
                        <div class="text-sm leading-6">
                            <label
                                for="email-enabled"
                                class="font-medium text-gray-900 cursor-pointer"
                            >
                                {{ t("settings.smtp.enable_email") }}
                            </label>
                            <p class="text-gray-500">{{ t("settings.smtp.enable_email_hint") }}</p>
                        </div>
                    </div>

                    <div class="mt-6 grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-6">
                        <div class="sm:col-span-3">
                            <label
                                for="connection-security"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.connection") }}</label
                            >
                            <div class="mt-2">
                                <Listbox
                                    as="div"
                                    v-model="server.new.ConnectionSecurity"
                                    :disabled="server.new.env_locked"
                                >
                                    <div class="relative">
                                        <ListboxButton
                                            class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6"
                                        >
                                            <span class="col-start-1 row-start-1 truncate pr-6">{{
                                                server.new.ConnectionSecurity
                                            }}</span>
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
                                                    v-for="option in options"
                                                    :key="option"
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
                                                                selected
                                                                    ? 'font-semibold'
                                                                    : 'font-normal',
                                                                'block truncate',
                                                            ]"
                                                            >{{ option }}</span
                                                        >
                                                        <span
                                                            v-if="selected"
                                                            :class="[
                                                                active
                                                                    ? 'text-white'
                                                                    : 'text-indigo-600',
                                                                'absolute inset-y-0 right-0 flex items-center pr-4',
                                                            ]"
                                                        >
                                                            <CheckIcon
                                                                class="h-5 w-5"
                                                                aria-hidden="true"
                                                            />
                                                        </span>
                                                    </li>
                                                </ListboxOption>
                                            </ListboxOptions>
                                        </transition>
                                    </div>
                                </Listbox>
                            </div>
                        </div>

                        <div class="sm:col-span-3 flex items-end pb-1">
                            <div class="relative flex gap-x-3">
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="server.new.EnableSMTPAuth"
                                        id="smtp-auth"
                                        name="smtp-auth"
                                        type="checkbox"
                                        :disabled="server.new.env_locked"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        for="smtp-auth"
                                        class="font-medium text-gray-900 cursor-pointer"
                                        >{{ t("settings.smtp.enable_smtp_authentication") }}</label
                                    >
                                </div>
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="smtp-server"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.server") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.SMTPServer"
                                    id="smtp-server"
                                    name="smtp-server"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.server.new.SMTPServer.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p
                                    v-if="v$.server.new.SMTPServer.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="sm:col-span-2">
                            <label
                                for="smtp-port"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.port") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.SMTPPort"
                                    id="smtp-port"
                                    name="smtp-port"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.server.new.SMTPPort.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p
                                    v-if="v$.server.new.SMTPPort.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="smtp-username"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.server_username") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.SMTPUsername"
                                    id="smtp-username"
                                    name="smtp-username"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.server.new.SMTPUsername.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p
                                    v-if="v$.server.new.SMTPUsername.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="sm:col-span-2">
                            <label
                                for="smtp-password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.server_password") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.SMTPPassword"
                                    id="smtp-password"
                                    name="smtp-password"
                                    type="password"
                                    autocomplete="new-password"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                />
                                <p v-if="!server.new.env_locked" class="mt-2 text-sm text-gray-500">
                                    {{ t("settings.smtp.password_unchanged_hint") }}
                                </p>
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="smtp-from-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.from_name") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.FromName"
                                    id="smtp-from-name"
                                    name="smtp-from-name"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                />
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="smtp-from-address"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.from_address") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.FromAddress"
                                    id="smtp-from-address"
                                    name="smtp-from-address"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                />
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="smtp-reply-to"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.smtp.reply_to") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="server.new.ReplyToAddress"
                                    id="smtp-reply-to"
                                    name="smtp-reply-to"
                                    type="text"
                                    :disabled="server.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                />
                            </div>
                        </div>

                        <div class="sm:col-span-full flex items-center gap-x-4">
                            <button
                                type="button"
                                class="rounded-md bg-green-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-green-600"
                                @click="testConnection"
                            >
                                {{ t("settings.smtp.test_connection") }}
                            </button>
                        </div>

                        <div v-if="messages.success || messages.error" class="sm:col-span-full">
                            <div v-if="messages.success" class="rounded-md bg-green-50 p-4">
                                <div class="flex">
                                    <div class="flex-shrink-0">
                                        <CheckCircleIcon
                                            class="h-5 w-5 text-green-400"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <div class="ml-3">
                                        <p class="text-sm font-medium text-green-800">
                                            {{ messages.success }}
                                        </p>
                                    </div>
                                    <div class="ml-auto pl-3">
                                        <button
                                            type="button"
                                            class="inline-flex rounded-md bg-green-50 p-1.5 text-green-500 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-green-600 focus:ring-offset-2 focus:ring-offset-green-50"
                                            @click="messages.success = ''"
                                        >
                                            <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                                        </button>
                                    </div>
                                </div>
                            </div>
                            <div v-if="messages.error" class="rounded-md bg-red-50 p-4">
                                <div class="flex">
                                    <div class="flex-shrink-0">
                                        <XMarkIcon
                                            class="h-5 w-5 text-red-400"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <div class="ml-3">
                                        <p class="text-sm font-medium text-red-800">
                                            {{ messages.error }}
                                        </p>
                                    </div>
                                    <div class="ml-auto pl-3">
                                        <button
                                            type="button"
                                            class="inline-flex rounded-md bg-red-50 p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-red-50"
                                            @click="messages.error = ''"
                                        >
                                            <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="disableSave() || server.new.env_locked"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                @click="save"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { ref, reactive, onMounted } from "vue";
import EnvManagedNotice from "@/components/Settings/EnvManagedNotice.vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useVuelidate } from "@vuelidate/core";
import { required, requiredIf } from "@vuelidate/validators";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon, CheckCircleIcon, XMarkIcon } from "@heroicons/vue/20/solid";

import emailService from "@/services/emailService";

onMounted(() => {
    emailService.emailServer().then((res) => {
        server.old = JSON.parse(JSON.stringify(res.data));
        server.new = JSON.parse(JSON.stringify(res.data));

        loaded.value = true;
    });
});

const messages = reactive({
    success: "",
    error: "",
});

const loaded = ref(false);
const server = reactive({ old: null, new: null });
const options = ["TLS", "STARTTLS", "NONE"];

const requiredWhenEnabled = requiredIf(() => server.new?.EnableEmail);

const rules = {
    server: {
        new: {
            ConnectionSecurity: { requiredWhenEnabled },
            EnableSMTPAuth: { required },
            SMTPServer: { requiredWhenEnabled },
            SMTPPort: { requiredWhenEnabled },
            SMTPUsername: { requiredWhenEnabled },
        },
    },
};
const v$ = useVuelidate(rules, { server });

function testConnection() {
    messages.success = "";
    messages.error = "";

    emailService
        .testConnection(server.new)
        .then(() => {
            messages.success = t.value("settings.smtp.connection_successful");
        })
        .catch((err) => {
            messages.error = extractErrorMessage(err);
        });
}

async function save() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    emailService
        .update(server.new)
        .then(() => {
            useAlertStore().showSuccess(t.value("common.success.settings_updated"));
            server.old = JSON.parse(JSON.stringify(server.new));
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function disableSave() {
    return JSON.stringify(server.old) === JSON.stringify(server.new);
}
</script>
