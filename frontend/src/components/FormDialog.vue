<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-50" @close="emit('close')">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
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
                            :class="[
                                'relative w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:p-6',
                                width,
                            ]"
                        >
                            <form @submit.prevent="submit">
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ title }}
                                </DialogTitle>

                                <div class="mt-4">
                                    <slot />
                                </div>

                                <div class="mt-5 sm:mt-6 sm:flex sm:flex-row-reverse">
                                    <BaseButton
                                        type="submit"
                                        class="w-full sm:ml-3 sm:w-auto"
                                        :color="
                                            destructive
                                                ? 'bg-red-600 hover:bg-red-500 text-white'
                                                : ''
                                        "
                                        :is-disabled="confirmDisabled"
                                        :is-loading="loading"
                                    >
                                        {{ confirmLabel || t("common.button.save") }}
                                    </BaseButton>
                                    <BaseButton
                                        v-if="secondaryLabel"
                                        type="button"
                                        variant="secondary"
                                        class="mt-3 w-full sm:ml-3 sm:mt-0 sm:w-auto"
                                        :is-disabled="confirmDisabled || loading"
                                        @click="secondary"
                                    >
                                        {{ secondaryLabel }}
                                    </BaseButton>
                                    <BaseButton
                                        type="button"
                                        variant="secondary"
                                        class="mt-3 w-full sm:mt-0 sm:w-auto"
                                        @click="emit('close')"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </BaseButton>
                                </div>
                            </form>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import BaseButton from "@/components/BaseButton.vue";

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    open: { type: Boolean, default: false },
    title: { type: String, default: "" },
    confirmLabel: { type: String, default: "" },
    confirmDisabled: { type: Boolean, default: false },
    loading: { type: Boolean, default: false },
    destructive: { type: Boolean, default: false },
    width: { type: String, default: "sm:max-w-lg" },
    secondaryLabel: { type: String, default: "" },
});

const emit = defineEmits(["confirm", "close", "secondary"]);

function submit() {
    if (!props.confirmDisabled && !props.loading) emit("confirm");
}

function secondary() {
    if (!props.confirmDisabled && !props.loading) emit("secondary");
}
</script>
