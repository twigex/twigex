<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="grid grid-cols-1 gap-3">
        <button
            v-for="provider in providers"
            :key="provider.id"
            type="button"
            :style="{ backgroundColor: provider.button_color || '#4f46e5' }"
            class="flex w-full items-center justify-center rounded-md px-3 py-2 text-sm font-semibold leading-6 text-white shadow-sm hover:opacity-90 transition-opacity"
            @click="goToProvider(provider)"
        >
            {{ provider.button_text || t("login.button.login_with", { name: provider.name }) }}
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

defineProps({
    providers: {
        type: Array,
        required: true,
    },
});

function goToProvider(provider) {
    window.location.href = provider.login_url;
}
</script>
