<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="flex min-h-full flex-1 flex-col justify-center px-6 py-12 lg:px-8">
        <div class="sm:mx-auto sm:w-full sm:max-w-sm">
            <img class="mx-auto h-10 w-auto" :src="'/logo.png'" alt="Your Company" />
            <h2 class="mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900">
                {{ t("login.label.sign_in_account") }}
            </h2>
        </div>

        <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
            <!-- No login methods at all -->
            <div v-if="!hasAnyLoginMethod" class="rounded-md bg-red-50 border border-red-200 p-4">
                <p class="text-sm text-red-800 text-center">
                    {{ t("login.error.no_methods_available") }}
                </p>
            </div>

            <!-- OTP step -->
            <div v-else-if="showOtp" class="flex flex-col">
                <label class="block text-sm text-center font-medium leading-6 text-gray-900 pb-3">{{
                    t("login.label.enter_otp_code")
                }}</label>
                <OtpInput
                    :length="6"
                    @otp-complete="
                        token = $event;
                        login();
                    "
                />
            </div>

            <!-- Normal login -->
            <div v-else>
                <!-- SSO section (rendered first if oidcFirst) -->
                <div v-if="showSSO && authConfig.oidcFirst">
                    <SSOButtons :providers="authConfig.oidcProviders" />

                    <div v-if="showLocalForm" class="relative mt-10" aria-hidden="true">
                        <div class="absolute inset-0 flex items-center">
                            <div class="w-full border-t border-gray-200" />
                        </div>
                        <div class="relative flex justify-center text-sm font-medium leading-6">
                            <span class="bg-white px-6 text-gray-900">
                                {{ t("login.label.or_sign_in_with_password") }}
                            </span>
                        </div>
                    </div>
                </div>

                <!-- Password form -->
                <form
                    v-if="showLocalForm"
                    class="space-y-6"
                    :class="{ 'mt-6': showSSO && authConfig.oidcFirst }"
                    action="#"
                    method="POST"
                >
                    <div>
                        <label
                            for="login-id"
                            class="block text-sm font-medium leading-6 text-gray-900"
                            >{{ t("login.label.login_id") }}</label
                        >
                        <div class="mt-2">
                            <input
                                v-model="loginID"
                                id="login-id"
                                name="login-id"
                                type="text"
                                autocomplete="username"
                                required
                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            />
                        </div>
                    </div>

                    <div>
                        <div class="flex items-center justify-between">
                            <label
                                for="password"
                                class="block text-sm font-medium leading-6 text-gray-900"
                                >{{ t("common.label.password") }}</label
                            >
                            <div v-if="authConfig.passwordResetEnabled" class="text-sm">
                                <a
                                    @click.prevent="router.push({ name: 'password-forgot' })"
                                    class="font-semibold text-indigo-600 hover:text-indigo-500 cursor-pointer"
                                    >{{ t("login.button.forgot_password") }}</a
                                >
                            </div>
                        </div>
                        <div class="mt-2">
                            <input
                                v-model="password"
                                id="password"
                                name="password"
                                type="password"
                                autocomplete="current-password"
                                required
                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            />
                        </div>
                    </div>

                    <div>
                        <button
                            @click.prevent="login"
                            class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        >
                            {{ t("login.button.sign_in") }}
                        </button>
                    </div>
                </form>

                <!-- SSO section (rendered last if not oidcFirst) -->
                <div v-if="showSSO && !authConfig.oidcFirst">
                    <div v-if="showLocalForm" class="relative mt-10" aria-hidden="true">
                        <div class="absolute inset-0 flex items-center">
                            <div class="w-full border-t border-gray-200" />
                        </div>
                        <div class="relative flex justify-center text-sm font-medium leading-6">
                            <span class="bg-white px-6 text-gray-900">{{
                                t("login.label.continue_with")
                            }}</span>
                        </div>
                    </div>

                    <SSOButtons
                        :providers="authConfig.oidcProviders"
                        :class="{ 'mt-6': showLocalForm }"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index";
import { ref, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useUserStore } from "@/store/user";
import { useAuthStore } from "@/store/auth";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import authService from "@/services/authService";
import settingsService from "@/services/settingsService";
import { getCsrfFromCookie } from "@/utils/utils";
import OtpInput from "@/components/Login/OtpInput.vue";
import SSOButtons from "@/components/Login/SSOButtons.vue";

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const authStore = useAuthStore();

const loginID = ref("");
const password = ref("");
const token = ref("");
const showOtp = ref(false);
const loaded = ref(false);

const authConfig = ref({
    localAuthEnabled: false,
    ldapEnabled: false,
    oidcFirst: false,
    passwordResetEnabled: true,
    oidcProviders: [],
});

const showLocalForm = computed(
    () => authConfig.value.localAuthEnabled || authConfig.value.ldapEnabled,
);

const showSSO = computed(() => authConfig.value.oidcProviders.length > 0);

const hasAnyLoginMethod = computed(() => showLocalForm.value || showSSO.value);

onMounted(async () => {
    const params = new URLSearchParams(window.location.search);
    const error = params.get("error");

    if (error) {
        useAlertStore().showError(t.value(`login.error.${error}`));
        // Clean the URL so a refresh doesn't re-show the error
        window.history.replaceState({}, "", window.location.pathname);
    }

    try {
        const response = await settingsService.publicAuthSettings();

        authConfig.value = {
            ...authConfig.value,
            ...response.data,
            oidcProviders: response.data.oidcProviders || [],
        };
    } catch (e) {
        useAlertStore().showError(e.response?.data.error || t.value("login.error.load_failed"));
    } finally {
        loaded.value = true;
    }
});

function login() {
    authService
        .logIn(loginID.value, password.value, token.value)
        .then((response) => {
            userStore.setUser(response.data);
            authStore.setAuthenticated(true);
            authStore.setCSRF(getCsrfFromCookie());

            // Return the user to ?redirect set by the auth guard, or files by default.
            const redirect = route.query.redirect;

            router.push(typeof redirect === "string" ? redirect : { name: "files" });
        })
        .catch((error) => {
            if ("provide_mfa_token" == error.response.data.error) {
                showOtp.value = true;

                return;
            }

            useAlertStore().showError(extractErrorMessage(error));
        });
}
</script>
