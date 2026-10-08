<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Dialog :open="open" @close="emit('close')" class="relative z-[70]">
        <div class="fixed inset-0 bg-black/50" />
        <div class="fixed inset-0 flex items-center justify-center p-4">
            <DialogPanel class="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
                <DialogTitle class="text-base font-semibold text-gray-900">
                    {{ t("meetings.leave.title") }}
                </DialogTitle>
                <p class="mt-1 text-sm text-gray-500">
                    {{ t("meetings.leave.subtitle") }}
                </p>
                <div class="mt-5 space-y-2">
                    <button
                        @click="emit('endForAll')"
                        class="w-full rounded-md bg-red-600 px-4 py-2 text-sm font-semibold text-white hover:bg-red-500"
                    >
                        {{ t("meetings.leave.end_all") }}
                    </button>
                    <button
                        @click="emit('keepRunning')"
                        class="w-full rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-500"
                    >
                        {{ t("meetings.leave.keep_running") }}
                    </button>
                    <button
                        @click="emit('close')"
                        class="w-full rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                    >
                        {{ t("common.button.cancel") }}
                    </button>
                </div>
            </DialogPanel>
        </div>
    </Dialog>
</template>

<script setup>
import { t } from "@/i18n";
import { Dialog, DialogPanel, DialogTitle } from "@headlessui/vue";

defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["close", "endForAll", "keepRunning"]);
</script>
