<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <div v-if="loaded" class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="space-y-6 w-full md:w-3/4">
                <div class="pb-4 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.language.default_language") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.language.default_language_description") }}
                    </p>

                    <div class="mt-6 w-72">
                        <Listbox as="div" v-model="selectedLanguage">
                            <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">
                                {{ t("settings.language.server_language") }}
                            </ListboxLabel>
                            <div class="relative mt-2">
                                <ListboxButton
                                    class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6"
                                >
                                    <span class="col-start-1 row-start-1 truncate pr-6">
                                        {{
                                            selectedLanguage
                                                ? selectedLanguage.nativeName
                                                : t("settings.language.select_language")
                                        }}
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
                                            v-for="lang in languageList"
                                            :key="lang.code"
                                            :value="lang"
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
                                                    {{ lang.nativeName }}
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
                    </div>
                </div>

                <div class="pb-4 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.language.user_language") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.language.user_language_description") }}
                    </p>

                    <div class="mt-6 space-y-4">
                        <div class="relative flex gap-x-3">
                            <div class="flex h-6 items-center">
                                <input
                                    v-model="allowUserOverride"
                                    id="allow-user-override"
                                    type="checkbox"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                            </div>
                            <div class="text-sm leading-6">
                                <label for="allow-user-override" class="font-medium text-gray-900">
                                    {{ t("settings.language.allow_user_override") }}
                                </label>
                                <p class="text-gray-500">
                                    {{ t("settings.language.allow_user_override_description") }}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="!hasChanges || saving"
                @click.prevent="save"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useSettingsStore } from "@/store/settings";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import settingsService from "@/services/settingsService";
import { languages } from "../../i18n/languages";
import { t } from "@/i18n/index.js";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";

const settingsStore = useSettingsStore();
const loaded = ref(false);
const saving = ref(false);

const languageList = computed(() => Object.values(languages));

const selectedLanguage = ref(null);
const allowUserOverride = ref(true);

const savedState = ref(null);

const hasChanges = computed(() => {
    if (!savedState.value) return false;

    return (
        selectedLanguage.value?.code !== savedState.value.defaultLocale ||
        allowUserOverride.value !== savedState.value.allowUserOverride
    );
});

onMounted(async () => {
    try {
        const response = await settingsService.getLanguageSettings();

        applySettings(response.data);
    } catch {
        const defaultLocale = settingsStore.config?.DefaultLocale || "en";

        applySettings({ defaultLocale, allowUserOverride: true });
    }

    loaded.value = true;
});

function applySettings(data) {
    const langCode = data.defaultLocale || settingsStore.config?.DefaultLocale || "en";

    selectedLanguage.value =
        languageList.value.find((l) => l.code === langCode) ||
        languageList.value.find((l) => l.code === "en");
    allowUserOverride.value = data.allowUserOverride ?? true;
    savedState.value = {
        defaultLocale: selectedLanguage.value?.code,
        allowUserOverride: allowUserOverride.value,
    };
}

async function save() {
    saving.value = true;
    try {
        await settingsService.updateLanguageSettings({
            defaultLocale: selectedLanguage.value?.code ?? "en",
            allowUserOverride: allowUserOverride.value,
        });
        savedState.value = {
            defaultLocale: selectedLanguage.value?.code,
            allowUserOverride: allowUserOverride.value,
        };
        settingsStore.config.DefaultLocale = selectedLanguage.value?.code ?? "en";
        useAlertStore().showSuccess(t.value("common.success.settings_updated"));
    } catch (e) {
        useAlertStore().showError(extractErrorMessage(e));
    } finally {
        saving.value = false;
    }
}
</script>
