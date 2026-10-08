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
                        {{ t("settings.password.title") }}
                    </h2>
                    <p class="mt-1 text-sm leading-6 text-gray-600">
                        {{ t("settings.password.description") }}
                    </p>

                    <div class="mt-6 space-y-4 max-w-sm">
                        <div>
                            <label
                                for="old-password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.password.old_password") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="password.old"
                                    type="password"
                                    name="old-password"
                                    id="old-password"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.password.old.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.password.old.$error" class="mt-2 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>

                        <div>
                            <label
                                for="new-password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.password.new_password") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="password.new"
                                    type="password"
                                    name="new-password"
                                    id="new-password"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.password.new.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p v-if="v$.password.new.$error" class="mt-2 text-sm text-red-600">
                                    {{ v$.password.new.$errors[0].$message }}
                                </p>
                            </div>
                        </div>

                        <div>
                            <label
                                for="confirm-password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("settings.password.confirm_new_password") }}</label
                            >
                            <div class="mt-2">
                                <input
                                    v-model="password.confirm"
                                    id="confirm-password"
                                    name="confirm-password"
                                    type="password"
                                    class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                    :class="
                                        v$.password.confirm.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                />
                                <p
                                    v-if="v$.password.confirm.$error"
                                    class="mt-2 text-sm text-red-600"
                                >
                                    <span
                                        v-if="
                                            v$.password.confirm.$errors.some(
                                                (e) => e.$validator === 'required',
                                            )
                                        "
                                    >
                                        {{ t("common.error.required_field") }}
                                    </span>
                                    <span
                                        v-else-if="
                                            v$.password.confirm.$errors.some(
                                                (e) => e.$validator === 'sameAs',
                                            )
                                        "
                                    >
                                        {{ t("settings.password.error.no_match") }}
                                    </span>
                                </p>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4">
            <button
                @click.prevent="save"
                class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                {{ t("common.button.save") }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";

import { ref, reactive, computed, onMounted } from "vue";
import { useVuelidate } from "@vuelidate/core";
import { required, minLength, helpers, sameAs } from "@vuelidate/validators";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import userService from "@/services/userService";
import authService from "@/services/authService";

const password = ref({
    old: "",
    new: "",
    confirm: "",
});

const passwordPolicy = reactive({
    PasswordMinLength: 8,
    NumericCharacters: true,
    SpecialCharacters: true,
    UpperLowerCharacters: true,
});

const containsNumber = helpers.withMessage(t.value("settings.password.rules.number"), (value) =>
    /\d/.test(value),
);
const containsSpecial = helpers.withMessage(
    t.value("settings.password.rules.special_char"),
    (value) => /[!@#$%^&*(),.?":{}|<>]/.test(value),
);
const containsUpperLower = helpers.withMessage(
    t.value("settings.password.rules.uppe_lower"),
    (value) => /[a-z]/.test(value) && /[A-Z]/.test(value),
);

const rules = computed(() => ({
    password: {
        old: { required },
        new: {
            required: helpers.withMessage(t.value("settings.password.rules.required"), required),
            minLength: helpers.withMessage(
                t.value("settings.password.rules.min_length", {
                    min: passwordPolicy.PasswordMinLength,
                }),
                minLength(passwordPolicy.PasswordMinLength),
            ),
            ...(passwordPolicy.NumericCharacters ? { containsNumber } : {}),
            ...(passwordPolicy.SpecialCharacters ? { containsSpecial } : {}),
            ...(passwordPolicy.UpperLowerCharacters ? { containsUpperLower } : {}),
        },
        confirm: {
            required,
            sameAs: sameAs(computed(() => password.value.new)),
        },
    },
}));
const v$ = useVuelidate(rules, { password });

async function save() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    userService
        .updatePassword({
            oldPassword: password.value.old,
            Password: password.value.new,
            Confirmed: password.value.confirm,
        })
        .then(() => {
            password.value = {
                old: "",
                new: "",
                confirm: "",
            };

            useAlertStore().showSuccess(t.value("settings.password.success.updated"));

            v$.value.$reset();
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

onMounted(() => {
    authService.passwordPolicy().then((response) => {
        Object.assign(passwordPolicy, response.data);
    });
});
</script>
