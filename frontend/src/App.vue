<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <VersionBanner />
        <div class="min-h-0 flex-1">
            <RouterView v-if="loaded" />
        </div>
        <AppAlert />
    </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref } from "vue";
import { RouterView } from "vue-router";
import AppAlert from "@/components/Notifications/AppAlert.vue";
import VersionBanner from "@/components/Notifications/VersionBanner.vue";
import configService from "@/services/configService";
import { useSettingsStore } from "@/store/settings";
import { setLocale } from "@/i18n/index.js";

const settingsStore = useSettingsStore();
const loaded = ref(false);

// A tab left open across a deploy makes no requests on its own, so re-check on focus.
function onVisibilityChange() {
    if (document.visibilityState === "visible") {
        configService.config().catch(() => {});
    }
}

onMounted(() => {
    document.addEventListener("visibilitychange", onVisibilityChange);

    configService
        .config()
        .then((response) => {
            settingsStore.setConfig(response.data);
            setLocale(response.data?.DefaultLocale || "en");
            loaded.value = true;
        })
        .catch(() => {
            loaded.value = true;
        });
});

onBeforeUnmount(() => {
    document.removeEventListener("visibilitychange", onVisibilityChange);
});
</script>
