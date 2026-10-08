<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex min-h-full flex-col justify-center px-6 py-12 lg:px-8">
        <div class="sm:mx-auto sm:w-full sm:max-w-sm">
            <img class="mx-auto h-10 w-auto" :src="'/logo.png'" alt="Your Company" />
            <h2 class="mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900">
                Reset your password
            </h2>
        </div>

        <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
            <form class="space-y-6" @submit.prevent="resetPassword">
                <div>
                    <label for="password" class="block text-sm font-medium leading-6 text-gray-900"
                        >New password</label
                    >
                    <div class="mt-2">
                        <input
                            v-model="password"
                            id="password"
                            name="password"
                            type="password"
                            autocomplete="new-password"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        />
                        <p v-if="v$.password.$error" class="text-red-500 text-sm mt-1">
                            {{ v$.password.$errors[0].$message }}
                        </p>
                    </div>
                </div>

                <div>
                    <label
                        for="confirm-password"
                        class="block text-sm font-medium leading-6 text-gray-900"
                        >Confirm password</label
                    >
                    <div class="mt-2">
                        <input
                            v-model="confirmPassword"
                            id="confirm-password"
                            name="confirm-password"
                            type="password"
                            autocomplete="new-password"
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        />
                        <p v-if="v$.confirmPassword.$error" class="text-red-500 text-sm mt-1">
                            {{ v$.confirmPassword.$errors[0].$message }}
                        </p>
                    </div>
                </div>

                <div class="rounded-md bg-gray-100 p-4">
                    <div class="flex">
                        <div class="shrink-0">
                            <InformationCircleIcon
                                class="size-5 text-gray-400"
                                aria-hidden="true"
                            />
                        </div>
                        <div class="ml-3 flex-1 md:flex md:justify-between">
                            <p class="text-sm text-gray-700">
                                Password requirements: <br />
                                {{ passwordPolicy.PasswordMinLength }}
                                characters,
                                {{ passwordPolicy.NumericCharacters ? "1 number," : "" }}
                                {{ passwordPolicy.SpecialCharacters ? "1 special character," : "" }}
                                {{
                                    passwordPolicy.UpperLowerCharacters
                                        ? "1 uppercase and 1 lowercase letter."
                                        : ""
                                }}
                            </p>
                        </div>
                    </div>
                </div>
                <div>
                    <button
                        type="submit"
                        class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    >
                        Update password
                    </button>
                </div>
            </form>
        </div>
    </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import authService from "@/services/authService";
import { useVuelidate } from "@vuelidate/core";
import { required, minLength, helpers, sameAs } from "@vuelidate/validators";
import { InformationCircleIcon } from "@heroicons/vue/20/solid";

const route = useRoute();
const router = useRouter();
const password = ref("");
const confirmPassword = ref("");
const passwordPolicy = reactive({
    PasswordMinLength: 8,
    NumericCharacters: true,
    SpecialCharacters: true,
    UpperLowerCharacters: true,
});

const containsNumber = helpers.withMessage("Must contain at least one number", (value) =>
    /\d/.test(value),
);
const containsSpecial = helpers.withMessage(
    "Must contain at least one special character",
    (value) => /[!@#$%^&*(),.?":{}|<>]/.test(value),
);
const containsUpperLower = helpers.withMessage(
    "Must contain both upper and lower case letters",
    (value) => /[a-z]/.test(value) && /[A-Z]/.test(value),
);

const rules = computed(() => ({
    password: {
        required: helpers.withMessage("Password is required", required),
        minLength: helpers.withMessage(
            `Password must be at least ${passwordPolicy.PasswordMinLength} characters`,
            minLength(passwordPolicy.PasswordMinLength),
        ),
        ...(passwordPolicy.NumericCharacters ? { containsNumber } : {}),
        ...(passwordPolicy.SpecialCharacters ? { containsSpecial } : {}),
        ...(passwordPolicy.UpperLowerCharacters ? { containsUpperLower } : {}),
    },
    confirmPassword: {
        required: helpers.withMessage("Confirm Password is required", required),
        sameAsPassword: helpers.withMessage("Passwords must match", sameAs(password)),
    },
}));

const v$ = useVuelidate(rules, { password, confirmPassword });

function resetPassword() {
    v$.value.$validate();
    if (v$.value.$error) return;

    authService
        .updatePassword({
            password: password.value,
            confirmPassword: confirmPassword.value,
            token: route.params.token,
        })
        .then(() => {
            router.push({ name: "login" });
        });
}

onMounted(() => {
    authService.passwordPolicy().then((response) => {
        Object.assign(passwordPolicy, response.data);
    });
});
</script>
