<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pt-5 pb-4 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-indigo-100 sm:mx-0 sm:size-10"
                                >
                                    <PencilIcon class="size-6 text-indigo-600" aria-hidden="true" />
                                </div>
                                <div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >{{
                                            t("channels.rename_channel_dialog.title")
                                        }}</DialogTitle
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="name"
                                            name="name"
                                            id="name"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                            :placeholder="
                                                t('channels.rename_channel_dialog.placeholder')
                                            "
                                            :class="
                                                v$.name.$error
                                                    ? 'ring-red-600 focus:ring-red-600'
                                                    : 'ring-gray-300 focus:ring-indigo-600'
                                            "
                                        />
                                        <p v-if="v$.name.$error" class="mt-2 text-sm text-red-600">
                                            {{ t("common.error.required_field") }}
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-xs hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                    @click="rename"
                                >
                                    {{ t("common.button.save") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-xs ring-1 ring-gray-300 ring-inset hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="close"
                                    ref="cancelButtonRef"
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

import { watch, ref } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { PencilIcon } from "@heroicons/vue/24/outline";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";

const props = defineProps({
    modelValue: {
        type: Boolean,
        default: false,
    },
    channelName: {
        type: String,
        default: "",
    },
});

const emits = defineEmits(["update:modelValue", "rename"]);

const name = ref(props.channelName || "");

const rules = {
    name: { required },
};
const v$ = useVuelidate(rules, { name: name });

async function rename() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    emits("rename", name.value);
}

function close() {
    //reset  validation
    v$.value.$reset();

    emits("update:modelValue", false);
}

watch(
    () => props.channelName,
    (newValue) => {
        if (newValue) {
            name.value = newValue || "";
        }
    },
);
</script>
