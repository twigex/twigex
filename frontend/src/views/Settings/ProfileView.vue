<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <PhotoCropperDialog
        v-model="showCropperDialog"
        :file="cropperFile"
        @uploaded="onPhotoUploaded"
    />

    <div class="flex flex-col h-full overflow-hidden">
        <div v-if="loaded" class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="space-y-1 w-full md:w-3/4">
                <div
                    v-if="userSettings.user.auth_service != ''"
                    class="rounded-md bg-blue-50 p-4 mt-4"
                >
                    <div class="flex">
                        <div class="shrink-0">
                            <InformationCircleIcon
                                class="h-5 w-5 text-blue-400"
                                aria-hidden="true"
                            />
                        </div>
                        <div class="ml-3">
                            <h3 class="text-sm font-medium text-blue-800">
                                {{ t("settings.profile_info.sso_managed_title") }}
                            </h3>
                            <p class="mt-1 text-sm text-blue-700">
                                {{ t("settings.profile_info.sso_managed_description") }}
                            </p>
                        </div>
                    </div>
                </div>

                <div class="pb-2 border-b border-gray-900/10">
                    <div class="mt-10 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                        <div class="col-span-full">
                            <label class="block text-sm font-medium leading-6 text-gray-900">{{
                                t("settings.profile_info.photo")
                            }}</label>
                            <div class="mt-2 flex items-center gap-x-4">
                                <button
                                    type="button"
                                    class="relative h-20 w-20 rounded-full group focus:outline-none focus-visible:ring-2 focus-visible:ring-indigo-600 focus-visible:ring-offset-2"
                                    @click="$refs.file.click()"
                                    :title="t('settings.profile_info.change_photo')"
                                >
                                    <UserAvatar :user="userSettings.user" />
                                    <span
                                        class="absolute inset-0 rounded-full flex items-center justify-center bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        <CameraIcon class="h-6 w-6 text-white" />
                                    </span>
                                </button>
                                <p class="text-xs text-gray-500">
                                    {{ t("settings.profile_info.change_photo") }}
                                </p>
                                <button
                                    v-if="userSettings.user.photo"
                                    type="button"
                                    class="text-xs font-medium text-gray-700 hover:text-gray-900"
                                    @click="onPhotoRemoved"
                                >
                                    {{ t("settings.profile_info.remove_photo") }}
                                </button>
                                <input
                                    type="file"
                                    ref="file"
                                    style="display: none"
                                    @change="onFileSelected"
                                    accept="image/png, image/gif, image/jpeg"
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <div class="pt-2">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.profile_info.title") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.profile_info.description") }}
                    </p>

                    <div class="mt-10 grid grid-cols-1 gap-x-6 gap-y-8 sm:grid-cols-6">
                        <div class="sm:col-span-3">
                            <label
                                for="first-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.profile_info.first_name") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="userSettings.user.name"
                                    type="text"
                                    name="first-name"
                                    id="first-name"
                                    autocomplete="given-name"
                                    :disabled="userSettings.user.auth_service != ''"
                                    class="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        userSettings.user.auth_service != ''
                                            ? 'bg-gray-50 text-gray-500 ring-gray-200 cursor-not-allowed focus:ring-gray-200'
                                            : v$.user.name.$error
                                              ? 'text-gray-900 ring-red-600 focus:ring-red-600'
                                              : 'text-gray-900 ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.user.name.$error" class="mt-2 text-sm text-red-600">
                                    {{ t("settings.profile_info.error.name_required") }}
                                </p>
                            </div>
                        </div>
                        <div class="sm:col-span-3">
                            <label
                                for="last-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.profile_info.last_name") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="userSettings.user.lastname"
                                    type="text"
                                    name="last-name"
                                    id="last-name"
                                    autocomplete="family-name"
                                    :disabled="userSettings.user.auth_service != ''"
                                    class="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        userSettings.user.auth_service != ''
                                            ? 'bg-gray-50 text-gray-500 ring-gray-200 cursor-not-allowed focus:ring-gray-200'
                                            : v$.user.lastname.$error
                                              ? 'text-gray-900 ring-red-600 focus:ring-red-600'
                                              : 'text-gray-900 ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.user.lastname.$error" class="mt-2 text-sm text-red-600">
                                    {{ t("settings.profile_info.error.lastname_required") }}
                                </p>
                            </div>
                        </div>
                        <div class="sm:col-span-4">
                            <label
                                for="email"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.profile_info.email") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="userSettings.user.email"
                                    id="email"
                                    name="email"
                                    type="email"
                                    autocomplete="email"
                                    :disabled="userSettings.user.auth_service != ''"
                                    class="block w-full rounded-md border-0 py-1.5 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        userSettings.user.auth_service != ''
                                            ? 'bg-gray-50 text-gray-500 ring-gray-200 cursor-not-allowed focus:ring-gray-200'
                                            : v$.user.email.$error
                                              ? 'text-gray-900 ring-red-600 focus:ring-red-600'
                                              : 'text-gray-900 ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.user.email.$error" class="mt-2 text-sm text-red-600">
                                    {{ t("settings.profile_info.error.email_required") }}
                                </p>
                            </div>
                        </div>
                        <div class="sm:col-span-4">
                            <div class="mt-2">
                                <Listbox
                                    as="div"
                                    v-model="selected"
                                    :disabled="
                                        userSettings.user.timezone.useAutomaticTimezone === 'true'
                                            ? true
                                            : false
                                    "
                                >
                                    <ListboxLabel
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                        >{{ t("settings.profile_info.timezone") }}</ListboxLabel
                                    >
                                    <div class="relative flex gap-x-3 pt-2">
                                        <div class="flex h-6 items-center">
                                            <input
                                                v-model="
                                                    userSettings.user.timezone.useAutomaticTimezone
                                                "
                                                @click="
                                                    userSettings.user.timezone.automaticTimezone =
                                                        Intl.DateTimeFormat().resolvedOptions().timeZone
                                                "
                                                id="autotz"
                                                name="autotz"
                                                type="checkbox"
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                        </div>
                                        <div class="text-sm leading-6">
                                            <label for="autotz" class="font-medium text-gray-900">
                                                {{ t("settings.profile_info.automatic_timezone") }}
                                            </label>
                                            <p class="text-gray-500">
                                                {{
                                                    Intl.DateTimeFormat().resolvedOptions().timeZone
                                                }}
                                            </p>
                                        </div>
                                    </div>
                                    <div class="relative mt-2">
                                        <ListboxButton
                                            :disabled="
                                                userSettings.user.timezone.useAutomaticTimezone
                                            "
                                            class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                        >
                                            <span
                                                class="block truncate"
                                                :class="
                                                    userSettings.user.timezone.useAutomaticTimezone
                                                        ? 'text-gray-400'
                                                        : ''
                                                "
                                            >
                                                {{
                                                    selected == ""
                                                        ? t(
                                                              "settings.profile_info.selecte_timezone",
                                                          )
                                                        : selected
                                                }}
                                            </span>
                                            <span
                                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                            >
                                                <ChevronUpDownIcon
                                                    class="h-5 w-5 text-gray-400"
                                                    aria-hidden="true"
                                                />
                                            </span>
                                        </ListboxButton>

                                        <transition
                                            leave-active-class="transition ease-in duration-100"
                                            leave-from-class="opacity-100"
                                            leave-to-class="opacity-0"
                                        >
                                            <ListboxOptions
                                                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                            >
                                                <ListboxOption
                                                    as="template"
                                                    v-for="timezone in timezoneList"
                                                    :key="timezone"
                                                    :value="timezone"
                                                    v-slot="{ active, selected: isSelected }"
                                                >
                                                    <li
                                                        :class="[
                                                            active
                                                                ? 'bg-indigo-600 text-white'
                                                                : 'text-gray-900',
                                                            'relative cursor-default select-none py-2 pl-8 pr-4',
                                                        ]"
                                                    >
                                                        <span
                                                            :class="[
                                                                isSelected
                                                                    ? 'font-semibold'
                                                                    : 'font-normal',
                                                                'block truncate',
                                                            ]"
                                                        >
                                                            {{ timezone }}
                                                        </span>
                                                        <span
                                                            v-if="isSelected"
                                                            :class="[
                                                                active
                                                                    ? 'text-white'
                                                                    : 'text-indigo-600',
                                                                'absolute inset-y-0 left-0 flex items-center pl-1.5',
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
                                <div class="relative flex gap-x-3 pt-2">
                                    <div class="flex h-6 items-center">
                                        <input
                                            v-model="userSettings.user.clockDisplay24h"
                                            id="clock24h"
                                            name="clock24h"
                                            type="checkbox"
                                            class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        />
                                    </div>
                                    <div class="text-sm leading-6">
                                        <label for="clock24h" class="font-medium text-gray-900">
                                            {{ t("settings.profile_info.use_24h_format") }}
                                        </label>
                                        <p class="text-gray-500">
                                            {{
                                                t(
                                                    "settings.profile_info.use_24h_format_description",
                                                )
                                            }}
                                        </p>
                                    </div>
                                </div>
                                <Listbox
                                    class="mt-1"
                                    as="div"
                                    v-model="lang"
                                    :disabled="!settingsStore.config?.AllowUserLanguageOverride"
                                >
                                    <ListboxLabel class="block text-sm/6 font-medium text-gray-900">
                                        {{ t("settings.profile_info.language") }}
                                    </ListboxLabel>
                                    <p
                                        v-if="!settingsStore.config?.AllowUserLanguageOverride"
                                        class="mt-1 text-sm text-gray-500"
                                    >
                                        {{ t("settings.profile_info.language_managed_by_admin") }}
                                    </p>
                                    <div class="relative mt-2">
                                        <ListboxButton
                                            :disabled="
                                                !settingsStore.config?.AllowUserLanguageOverride
                                            "
                                            :class="[
                                                'grid w-full cursor-default grid-cols-1 rounded-md py-1.5 pl-3 pr-2 text-left outline outline-1 -outline-offset-1 sm:text-sm/6',
                                                !settingsStore.config?.AllowUserLanguageOverride
                                                    ? 'bg-gray-50 text-gray-400 outline-gray-200 cursor-not-allowed'
                                                    : 'bg-white text-gray-900 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600',
                                            ]"
                                        >
                                            <span class="col-start-1 row-start-1 truncate pr-6">
                                                {{
                                                    lang != null
                                                        ? lang.nativeName
                                                        : t("settings.profile_info.select_language")
                                                }}
                                            </span>
                                            <ChevronUpDownIcon
                                                class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                                                aria-hidden="true"
                                            />
                                        </ListboxButton>

                                        <transition
                                            leave-active-class="transition ease-in duration-100"
                                            leave-from-class=""
                                            leave-to-class="opacity-0"
                                        >
                                            <ListboxOptions
                                                class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg outline outline-1 outline-black/5 sm:text-sm"
                                            >
                                                <ListboxOption
                                                    as="template"
                                                    v-for="l in languageList"
                                                    :key="l.code"
                                                    :value="l"
                                                    v-slot="{ active, selected: isSelected }"
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
                                                                isSelected
                                                                    ? 'font-semibold'
                                                                    : 'font-normal',
                                                                'block truncate',
                                                            ]"
                                                        >
                                                            {{ l.nativeName }}
                                                        </span>
                                                        <span
                                                            v-if="isSelected"
                                                            :class="[
                                                                active
                                                                    ? 'text-white'
                                                                    : 'text-indigo-600',
                                                                'absolute inset-y-0 right-0 flex items-center pr-4',
                                                            ]"
                                                        >
                                                            <CheckIcon
                                                                class="size-5"
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
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="disableSave()"
                @click.prevent="save"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useSettingsStore } from "@/store/settings";
import userService from "@/services/userService";
import PhotoCropperDialog from "@/components/Settings/PhotoCropperDialog.vue";
import moment from "moment-timezone";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import {
    CameraIcon,
    CheckIcon,
    ChevronUpDownIcon,
    InformationCircleIcon,
} from "@heroicons/vue/20/solid";

import { languages } from "../../i18n/languages";
import { setLocale } from "../../i18n";
import { t } from "@/i18n/index.js";

onMounted(() => {
    timezoneList.value = moment.tz.names();
    selected.value = userStore.user.timezone.manualTimezone;

    let user = userStore.user;

    user.clockDisplay24h =
        userStore.getClockDisplay === "24h" || userStore.getClockDisplay === "" ? true : false;

    userSettings.user = JSON.parse(JSON.stringify(user));
    userSettings.old = JSON.parse(JSON.stringify(userStore.user));
    const userLangCode = userStore.getLanguage || settingsStore.config?.DefaultLocale || "en";

    const langList = Object.values(languages);
    const resolvedLang =
        langList.find((l) => l.code === userLangCode) || langList.find((l) => l.code === "en");

    lang.value = resolvedLang ?? null;
    oldLang.value = resolvedLang ?? null;

    loaded.value = true;
});

const settingsStore = useSettingsStore();
const userStore = useUserStore();
const timezoneList = ref([]);
const userSettings = reactive({
    user: null,
    old: null,
});
const selected = ref("");
const loaded = ref(false);
const lang = ref(null);
const oldLang = ref(null);
const showCropperDialog = ref(false);
const cropperFile = ref(null);
const languageList = computed(() => Object.values(languages));

const rules = {
    user: {
        name: { required },
        lastname: { required },
        email: { required },
    },
};
const v$ = useVuelidate(rules, userSettings);

async function save() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    userService
        .updateProfile({
            name: userSettings.user.name,
            lastname: userSettings.user.lastname,
            email: userSettings.user.email,
            useAutomaticTimezone: userSettings.user.timezone.useAutomaticTimezone,
            automaticTimezone: userSettings.user.timezone.automaticTimezone,
            manualTimezone: selected.value,
            clockDisplay: userSettings.user.clockDisplay24h,
            language: lang.value != null ? lang.value.code : "en",
        })
        .then(() => {
            userSettings.old = JSON.parse(JSON.stringify(userSettings.user));
            userStore.user = JSON.parse(JSON.stringify(userSettings.user));
            const newLangCode = lang.value != null ? lang.value.code : "en";
            const langChanged = newLangCode !== (oldLang.value?.code ?? "en");

            oldLang.value = lang.value;
            const displaySettings = userStore.preferences?.DisplaySettings;

            if (Array.isArray(displaySettings)) {
                const langPref = displaySettings.find((item) => item.name === "language");

                if (langPref) {
                    langPref.value = newLangCode;
                }

                const clockValue = userSettings.user.clockDisplay24h ? "24h" : "12h";
                const clockPref = displaySettings.find((item) => item.name === "clock_display");

                if (clockPref) {
                    clockPref.value = clockValue;
                } else {
                    displaySettings.push({
                        name: "clock_display",
                        value: clockValue,
                    });
                }
            }

            setLocale(newLangCode);
            if (langChanged) {
                window.location.reload();
            } else {
                useAlertStore().showSuccess(t.value("settings.profile_info.updated"));
            }
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function onFileSelected(event) {
    const file = event.target.files[0];

    if (!file) return;
    cropperFile.value = file;
    showCropperDialog.value = true;
    event.target.value = "";
}

function onPhotoUploaded(newPhoto) {
    userStore.setPhoto(userStore.user.id, newPhoto);
    userSettings.user.photo = newPhoto;
    userSettings.old = JSON.parse(JSON.stringify(userSettings.user));
    useAlertStore().showSuccess(t.value("settings.profile_info.updated"));
}

async function onPhotoRemoved() {
    try {
        await userService.deletePhoto();
    } catch {
        useAlertStore().showError(t.value("settings.profile_info.photo_remove_failed"));

        return;
    }

    onPhotoUploaded("");
}

function disableSave() {
    return (
        JSON.stringify(userSettings.user) === JSON.stringify(userSettings.old) &&
        lang.value?.code === oldLang.value?.code
    );
}
</script>
