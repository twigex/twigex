<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col">
        <nav class="flex flex-1 flex-col pt-3" :aria-label="t('graph.title')">
            <ul role="list" class="space-y-1">
                <li>
                    <span
                        class="flex gap-x-3 rounded-md bg-gray-50 p-2 text-sm font-semibold leading-6 text-indigo-600"
                        aria-current="page"
                    >
                        <ShareIcon class="h-6 w-6 shrink-0 text-indigo-600" aria-hidden="true" />
                        {{ t("graph.navigation.explore") }}
                    </span>
                </li>
            </ul>
        </nav>

        <div class="mt-4 rounded-lg p-3 text-sm text-gray-500 ring-1 ring-inset ring-gray-200">
            <p class="font-semibold text-gray-900">{{ t("graph.navigation.types_title") }}</p>
            <p class="mt-1">
                {{
                    t("graph.navigation.types_count", {
                        types: entityTypeCount,
                        relations: relationTypeCount,
                    })
                }}
            </p>
        </div>
    </div>
</template>

<script setup>
import { ShareIcon } from "@heroicons/vue/24/outline";
import { t } from "@/i18n/index.js";

defineProps({
    entityTypeCount: {
        type: Number,
        required: true,
    },
    relationTypeCount: {
        type: Number,
        required: true,
    },
});
</script>
