<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full max-w-md rounded-xl bg-white p-8 shadow-sm ring-1 ring-gray-200">
        <div
            class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-indigo-50"
        >
            <LockClosedIcon class="h-6 w-6 text-indigo-600" aria-hidden="true" />
        </div>
        <h2 class="text-center text-lg font-semibold text-gray-900">
            {{ t("public_share.password_title") }}
        </h2>
        <p class="mt-1 text-center text-sm text-gray-500">
            {{ t("public_share.password_subtitle") }}
        </p>
        <form class="mt-6 space-y-3" @submit.prevent="emit('submit')">
            <input
                v-model="passwordModel"
                type="password"
                autocomplete="current-password"
                :placeholder="t('public_share.password_placeholder')"
                class="block w-full rounded-md border-0 py-2 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm"
            />
            <p v-if="error" class="text-sm text-red-600">
                {{ error }}
            </p>
            <button
                type="submit"
                :disabled="unlocking || !password"
                class="w-full rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 disabled:opacity-50"
            >
                {{ t("public_share.unlock") }}
            </button>
        </form>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { LockClosedIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n";

const props = defineProps({
    password: {
        type: String,
        default: "",
    },
    error: {
        type: String,
        default: "",
    },
    unlocking: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:password", "submit"]);

const passwordModel = computed({
    get: () => props.password,
    set: (value) => emit("update:password", value),
});
</script>
