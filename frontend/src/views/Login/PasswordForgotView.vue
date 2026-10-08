<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex min-h-full flex-1 flex-col justify-center px-6 py-12 lg:px-8">
        <div class="sm:mx-auto sm:w-full sm:max-w-sm">
            <img class="mx-auto h-10 w-auto" :src="'/logo.png'" alt="Your Company" />
            <div v-if="!sent">
                <h2
                    class="mt-10 text-center text-2xl font-bold leading-9 tracking-tight text-gray-900"
                >
                    Reset your password
                </h2>
                <p class="text-sm text-center text-gray-500 pt-3">
                    Enter your email, and you will receive a link in your email.
                </p>
            </div>
        </div>

        <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-sm">
            <form v-if="!sent" class="space-y-6" action="#" method="POST">
                <div>
                    <label for="email" class="block text-sm font-medium leading-6 text-gray-900"
                        >Email address</label
                    >
                    <div class="mt-2">
                        <input
                            v-model="email"
                            id="email"
                            name="email"
                            type="email"
                            autocomplete="email"
                            required=""
                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        />
                        <p v-if="error" class="mt-2 text-sm text-red-600" id="email-error">
                            Not a valid email address.
                        </p>
                    </div>
                </div>

                <div>
                    <button
                        @click.prevent="resetPassword"
                        class="flex w-full justify-center rounded-md bg-indigo-600 px-3 py-1.5 text-sm font-semibold leading-6 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    >
                        Reset password
                    </button>
                </div>
            </form>
            <div v-else class="text-center flex flex-col items-center">
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                    class="size-6 text-green-500 h-14 w-14"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"
                    />
                </svg>
                <h3 class="mt-2 text-sm font-semibold text-gray-900">Password reset link sent</h3>
                <p class="mt-1 text-sm text-gray-500">
                    If your email is in our system, you’ll receive an email shortly with a link to
                    reset your password.
                </p>

                <div class="text-sm mt-5">
                    <a
                        @click.prevent="router.push({ name: 'login' })"
                        class="font-semibold text-indigo-600 hover:text-indigo-500 cursor-pointer"
                        >Sign in to your account</a
                    >
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import authService from "@/services/authService";

const router = useRouter();
const email = ref("");
const error = ref(null);
const sent = ref(false);

function resetPassword() {
    if (!isEmailValid()) {
        return;
    }

    authService.resetPassword(email.value).then(() => {
        sent.value = true;
    });
}

function isEmailValid() {
    const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

    if (email.value && emailPattern.test(email.value)) {
        error.value = null;

        return true;
    }

    error.value = "Please enter a valid email address.";

    return false;
}
</script>
