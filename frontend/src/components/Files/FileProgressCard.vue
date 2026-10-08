<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full overflow-hidden rounded-lg bg-white shadow-lg ring-1 ring-black/5">
        <div class="flex items-center justify-between border-b px-4 py-3">
            <p class="text-sm font-semibold text-gray-900">
                {{ title }}
            </p>

            <button
                type="button"
                @click="emit('close')"
                class="rounded-md text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            >
                <span class="sr-only">{{ t("common.button.close") }}</span>
                <XMarkIcon class="h-5 w-5" aria-hidden="true" />
            </button>
        </div>

        <ul role="list" class="max-h-64 space-y-2 overflow-y-auto px-4 py-2">
            <li
                v-for="item in items"
                :key="item.id"
                class="flex items-center justify-between gap-3"
            >
                <div class="flex min-w-0 flex-1 items-center gap-2">
                    <slot name="icon" :item="item" />

                    <div class="min-w-0">
                        <p class="truncate text-sm font-medium text-gray-900" :title="item.name">
                            {{ item.name }}
                        </p>

                        <p
                            v-if="item.error"
                            class="truncate text-xs text-red-600"
                            :title="item.error"
                        >
                            {{ item.error }}
                        </p>
                    </div>
                </div>

                <ProgressCircular
                    class="flex-shrink-0"
                    :size="24"
                    :progress="item.progress"
                    :status="item.status"
                    @cancel="item.cancel()"
                />
            </li>
        </ul>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import ProgressCircular from "@/components/ProgressCircular.vue";
import { XMarkIcon } from "@heroicons/vue/20/solid";

defineProps({
    title: {
        type: String,
        required: true,
    },
    items: {
        type: Array,
        required: true,
    },
});

const emit = defineEmits(["close"]);
</script>
