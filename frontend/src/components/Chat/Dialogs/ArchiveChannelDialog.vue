<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-10" @close="$emit('update:modelValue', false)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to=""
                leave="ease-in duration-200"
                leave-from=""
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-amber-100 sm:mx-0 sm:size-10"
                                >
                                    <ArchiveBoxArrowDownIcon
                                        class="size-6 text-amber-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >{{ t(prefix + ".title") }}</DialogTitle
                                    >
                                    <div class="mt-2">
                                        <p class="text-sm text-gray-500">
                                            {{
                                                t(prefix + ".description", { channel: channelName })
                                            }}
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-amber-600 px-3 py-2 text-sm font-semibold text-white hover:bg-amber-500 sm:ml-3 sm:w-auto"
                                    @click="confirm"
                                >
                                    {{ t(prefix + ".confirm_button") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="$emit('update:modelValue', false)"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { computed } from "vue";

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { ArchiveBoxArrowDownIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    channelName: {
        type: String,
        default: "",
    },
    unarchive: {
        type: Boolean,
        default: false,
    },
});

const prefix = computed(() =>
    props.unarchive ? "channels.unarchive_channel_dialog" : "channels.archive_channel_dialog",
);

const emit = defineEmits(["update:modelValue", "confirm"]);

function confirm() {
    emit("confirm");
    emit("update:modelValue", false);
}
</script>
