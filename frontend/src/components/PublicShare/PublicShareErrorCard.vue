<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full max-w-md rounded-xl bg-white p-8 text-center shadow-sm ring-1 ring-gray-200">
        <div
            class="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100"
        >
            <ExclamationTriangleIcon class="h-6 w-6 text-red-600" aria-hidden="true" />
        </div>
        <h2 class="text-lg font-semibold text-gray-900">
            {{ t("public_share.error_title") }}
        </h2>
        <p class="mt-2 text-sm text-gray-500">{{ message }}</p>
    </div>
</template>

<script setup>
import { ExclamationTriangleIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n";

defineProps({
    message: {
        type: String,
        default: "",
    },
});
</script>
