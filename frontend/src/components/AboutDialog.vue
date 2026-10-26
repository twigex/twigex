<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="$emit('update:modelValue', false)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
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
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-sm sm:p-6"
                        >
                            <div>
                                <div class="flex flex-row justify-end">
                                    <button
                                        type="button"
                                        class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2"
                                        @click="$emit('update:modelValue', false)"
                                    >
                                        <XMarkIcon class="size-6" aria-hidden="true" />
                                    </button>
                                </div>
                                <div
                                    class="mx-auto flex size-12 items-center justify-center rounded-full"
                                >
                                    <div class="flex h-16 shrink-0 items-center">
                                        <img
                                            class="h-8 w-auto"
                                            :src="'/logo.png'"
                                            alt="Your Company"
                                        />
                                    </div>
                                </div>
                                <div class="mt-3 text-center sm:mt-5">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >Twigex</DialogTitle
                                    >
                                    <div class="mt-2 flex flex-col">
                                        <p class="text-sm text-gray-500">
                                            {{ t("files.about.version") }}
                                            {{ settingsStore.getVersion }}
                                        </p>
                                        <p
                                            v-if="settingsStore.isEnterpriseBuild"
                                            class="text-sm text-gray-500"
                                        >
                                            {{ t("files.about.edition") }}
                                            {{ t("files.about.enterprise") }}
                                        </p>
                                        <p
                                            v-if="settingsStore.getBuildHash"
                                            class="text-sm text-gray-500"
                                        >
                                            {{ t("files.about.commit") }}
                                            {{ settingsStore.getBuildHash }}
                                        </p>
                                        <p class="text-sm text-gray-500">
                                            {{ t("files.about.build_date") }}
                                            {{ settingsStore.getBuildDate }}
                                        </p>
                                    </div>
                                </div>
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

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { useSettingsStore } from "@/store/settings";

defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

defineEmits(["update:modelValue"]);
const settingsStore = useSettingsStore();
</script>
