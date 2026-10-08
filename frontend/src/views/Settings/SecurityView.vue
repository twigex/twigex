<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <div class="flex flex-row justify-center flex-1 overflow-y-auto p-2">
            <div class="w-full md:w-3/4 space-y-6">
                <div class="pb-6 border-b border-gray-900/10">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.security.session_timeout") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.security.session_timeout_description") }}
                    </p>

                    <div class="mt-6 max-w-sm">
                        <label
                            for="timeoutPeriod"
                            class="block text-sm font-medium leading-6 text-gray-900"
                        >
                            {{ t("settings.security.timeout_period") }}
                        </label>
                        <div class="mt-2">
                            <div
                                class="flex rounded-md shadow-sm ring-1 ring-inset focus-within:ring-2 focus-within:ring-inset focus-within:ring-indigo-600"
                                :class="
                                    v$.state.new.timeoutValue.$error
                                        ? 'ring-red-600'
                                        : 'ring-gray-300'
                                "
                            >
                                <input
                                    id="timeoutPeriod"
                                    type="text"
                                    autocomplete="off"
                                    class="block flex-1 border-0 bg-transparent py-1.5 px-3 text-gray-900 placeholder:text-gray-400 focus:ring-0 sm:text-sm sm:leading-6"
                                    placeholder="timeout period"
                                    v-model="state.new.timeoutValue"
                                />
                            </div>
                            <p
                                v-if="v$.state.new.timeoutValue.$error"
                                class="mt-2 text-sm text-red-600"
                            >
                                <span
                                    v-if="
                                        v$.state.new.timeoutValue.$errors.some(
                                            (e) => e.$validator === 'required',
                                        )
                                    "
                                >
                                    {{ t("common.error.required_field") }}
                                </span>
                                <span
                                    v-else-if="
                                        v$.state.new.timeoutValue.$errors.some(
                                            (e) => e.$validator === 'minValue',
                                        )
                                    "
                                >
                                    {{ t("settings.security.error.value_greater_than_zero") }}
                                </span>
                            </p>
                        </div>
                    </div>
                </div>

                <div class="pb-6">
                    <h2 class="text-base font-semibold leading-7 text-gray-900">
                        {{ t("settings.security.password_policy") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.security.password_policy_description") }}
                    </p>

                    <div class="mt-6 space-y-6 max-w-sm">
                        <div>
                            <label class="block text-sm font-medium leading-6 text-gray-900">
                                {{ t("settings.security.password_length") }}
                            </label>
                            <div class="mt-2">
                                <Listbox as="div" v-model="state.new.passwordMinLength">
                                    <div class="relative">
                                        <ListboxButton
                                            class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6"
                                        >
                                            <span class="col-start-1 row-start-1 truncate pr-6">{{
                                                state.new.passwordMinLength
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
                                                    v-for="(option, index) in options"
                                                    :key="index"
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

                        <div class="space-y-4">
                            <div class="relative flex gap-x-3">
                                <div class="flex h-6 items-center">
                                    <input
                                        id="uppercaseLowercase"
                                        name="uppercaseLowercase"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        v-model="state.new.uppercaseLowercase"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        for="uppercaseLowercase"
                                        class="font-medium text-gray-900 cursor-pointer"
                                        >{{ t("settings.security.enforce_upper_and_lower") }}</label
                                    >
                                </div>
                            </div>

                            <div class="relative flex gap-x-3">
                                <div class="flex h-6 items-center">
                                    <input
                                        id="numericCharacters"
                                        name="numericCharacters"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        v-model="state.new.numericCharacters"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        for="numericCharacters"
                                        class="font-medium text-gray-900 cursor-pointer"
                                        >{{ t("settings.security.enforce_numeric") }}</label
                                    >
                                </div>
                            </div>

                            <div class="relative flex gap-x-3">
                                <div class="flex h-6 items-center">
                                    <input
                                        id="specialCharacters"
                                        name="specialCharacters"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                        v-model="state.new.specialCharacters"
                                    />
                                </div>
                                <div class="text-sm leading-6">
                                    <label
                                        for="specialCharacters"
                                        class="font-medium text-gray-900 cursor-pointer"
                                        >{{
                                            t("settings.security.enforce_special_characters")
                                        }}</label
                                    >
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                :disabled="disableSave()"
                @click="save"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { watch, onMounted, reactive } from "vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useVuelidate } from "@vuelidate/core";
import { required, minValue } from "@vuelidate/validators";
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import settingsService from "@/services/settingsService";

onMounted(() => {
    settingsService.passwordPolicy().then((response) => {
        state.new = {
            timeoutValue: response.data.SessionLength,
            passwordMinLength: response.data.PasswordSettings.PasswordMinLength,
            uppercaseLowercase: response.data.PasswordSettings.UpperLowerCharacters,
            numericCharacters: response.data.PasswordSettings.NumericCharacters,
            specialCharacters: response.data.PasswordSettings.SpecialCharacters,
        };

        state.old = JSON.parse(JSON.stringify(state.new));
    });
});

const options = [
    8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
];

const state = reactive({
    old: {},
    new: {
        uppercaseLowercase: false,
        numericCharacters: false,
        specialCharacters: false,
        timeoutValue: "",
        passwordMinLength: options[0],
    },
});

const rules = {
    state: {
        new: {
            timeoutValue: { required, minValue: minValue(1) },
        },
    },
};
const v$ = useVuelidate(rules, { state });

watch(
    () => state.new.timeoutValue,
    (newValue) => {
        state.new.timeoutValue = newValue.replace(/[^0-9]/g, "");
    },
);

async function save() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    settingsService
        .updatePasswordPolicy({
            SessionLength: state.new.timeoutValue,
            PasswordSettings: {
                PasswordMinLength: state.new.passwordMinLength,
                UpperLowerCharacters: state.new.uppercaseLowercase,
                NumericCharacters: state.new.numericCharacters,
                SpecialCharacters: state.new.specialCharacters,
            },
        })
        .then(() => {
            state.old = JSON.parse(JSON.stringify(state.new));
            useAlertStore().showSuccess(t.value("common.success.settings_updated"));
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function disableSave() {
    return JSON.stringify(state.old) === JSON.stringify(state.new);
}
</script>
