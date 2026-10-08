<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="-translate-y-full"
        enter-to-class="translate-y-0"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
    >
        <div
            v-if="versionStore.showBanner"
            :role="versionStore.updateRequired ? 'alert' : 'status'"
            :aria-live="versionStore.updateRequired ? 'assertive' : 'polite'"
            class="w-full shadow-sm"
            :class="versionStore.updateRequired ? 'bg-red-600' : 'bg-indigo-600'"
        >
            <div
                class="mx-auto flex max-w-7xl flex-wrap items-center justify-center gap-x-4 gap-y-2 px-4 py-2 text-center text-white"
            >
                <div class="flex items-center gap-x-2">
                    <ArrowPathIcon class="h-5 w-5 flex-shrink-0" aria-hidden="true" />
                    <p class="text-sm font-medium">
                        {{
                            versionStore.updateRequired
                                ? t("version.update_required.message")
                                : t("version.update_available.message")
                        }}
                    </p>
                </div>
                <div class="flex items-center gap-x-2">
                    <button
                        type="button"
                        @click="reload"
                        class="rounded-md bg-white/20 px-3 py-1 text-sm font-semibold hover:bg-white/30 focus:outline-none focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-transparent"
                    >
                        {{ t("version.button.refresh") }}
                    </button>
                    <button
                        v-if="!versionStore.updateRequired"
                        type="button"
                        @click="versionStore.dismiss()"
                        class="rounded-md p-1 hover:bg-white/20 focus:outline-none focus:ring-2 focus:ring-white"
                    >
                        <span class="sr-only">{{ t("common.button.close") }}</span>
                        <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                    </button>
                </div>
            </div>
        </div>
    </transition>
</template>

<script setup>
import { ArrowPathIcon } from "@heroicons/vue/24/outline";
import { XMarkIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index";
import { useVersionStore } from "@/store/version";

const versionStore = useVersionStore();

function reload() {
    window.location.reload();
}
</script>
