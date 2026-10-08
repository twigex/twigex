<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex w-full h-full items-center justify-center">
        <div class="flex flex-row w-full items-center justify-center gap-x-10 text-left">
            <div class="flex flex-col w-full">
                <div class="flex flex-col mb-7">
                    <LockClosedIcon class="h-20 w-20 shrink-0 text-indigo-600" aria-hidden="true" />
                </div>

                <div class="text-2xl text-gray-600 mb-5">
                    {{ t("license_purchase.title") }}
                </div>

                <div class="text-sm text-gray-500">
                    {{ t(`license_purchase.feature.${feature}`) }}
                </div>

                <div>
                    <div class="mt-10 flex items-center gap-x-6">
                        <a
                            href="https://twigex.com/contact-us"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="rounded-md bg-indigo-600 px-3.5 py-2.5 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 dark:bg-indigo-500 dark:shadow-none dark:hover:bg-indigo-400 dark:focus-visible:outline-indigo-500"
                            >{{ t("license_purchase.button.contact_sales") }}</a
                        >
                        <a
                            href="https://twigex.com"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="text-sm/6 font-semibold text-gray-600 hover:text-gray-500"
                            >{{ t("license_purchase.button.learn_more") }}
                            <span aria-hidden="true">→</span></a
                        >
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { LockClosedIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

defineProps({
    feature: {
        type: String,
        required: true,
    },
});
</script>
