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
                        {{ t("settings.office.title") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.office.description") }}
                    </p>

                    <EnvManagedNotice v-if="settings.new.env_locked" class="mt-4" />

                    <div class="mt-6 grid grid-cols-1 gap-x-6 gap-y-4 sm:grid-cols-6">
                        <div class="sm:col-span-full">
                            <div class="relative flex gap-x-3">
                                <div class="flex h-6 items-center">
                                    <input
                                        v-model="settings.new.Enable"
                                        id="office-enable"
                                        name="office-enable"
                                        type="checkbox"
                                        :disabled="settings.new.env_locked"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        for="office-enable"
                                        class="font-medium text-gray-900 cursor-pointer"
                                        >{{ t("settings.office.enable") }}</label
                                    >
                                    <p class="text-gray-500">
                                        {{ t("settings.office.enable_description") }}
                                    </p>
                                </div>
                            </div>
                        </div>

                        <div class="sm:col-span-4">
                            <label
                                for="office-type"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.office.type") }}</label
                            >
                            <div class="mt-2">
                                <select
                                    v-model="settings.new.Type"
                                    id="office-type"
                                    name="office-type"
                                    :disabled="settings.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                >
                                    <option value="collabora">
                                        {{ t("settings.office.type_collabora") }}
                                    </option>
                                    <option value="eurooffice">
                                        {{ t("settings.office.type_eurooffice") }}
                                    </option>
                                </select>
                            </div>
                        </div>

                        <div class="sm:col-span-4">
                            <label
                                for="office-host"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.office.host") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="settings.new.Host"
                                    id="office-host"
                                    name="office-host"
                                    type="text"
                                    :disabled="settings.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                    :class="
                                        v$.settings.new.Host.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p
                                    v-if="v$.settings.new.Host.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div class="sm:col-span-3">
                            <label
                                for="office-secret"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.office.secret") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="settings.new.Secret"
                                    id="office-secret"
                                    name="office-secret"
                                    type="password"
                                    autocomplete="new-password"
                                    :disabled="settings.new.env_locked"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="disableSave() || settings.new.env_locked"
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
import { required } from "@vuelidate/validators";

import settingsService from "@/services/settingsService";

onMounted(() => {
    settingsService.getOfficeSettings().then((res) => {
        settings.old = JSON.parse(JSON.stringify(res.data));
        settings.new = JSON.parse(JSON.stringify(res.data));

        loaded.value = true;
    });
});

const loaded = ref(false);
const settings = reactive({ old: null, new: null });

const rules = {
    settings: {
        new: {
            Host: { required },
        },
    },
};
const v$ = useVuelidate(rules, { settings });

async function save() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    settingsService
        .updateOfficeSettings(settings.new)
        .then(() => {
            useAlertStore().showSuccess(t.value("common.success.settings_updated"));
            settings.old = JSON.parse(JSON.stringify(settings.new));
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function disableSave() {
    return JSON.stringify(settings.old) === JSON.stringify(settings.new);
}
</script>
