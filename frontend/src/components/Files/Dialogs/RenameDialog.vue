<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-50" @close="closeDialog()">
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left w-full">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                        >{{ t("files.rename_dialog.title") }}:
                                        {{ file.name }}</DialogTitle
                                    >
                                    <div class="mt-2">
                                        <input
                                            v-model="name"
                                            name="name"
                                            id="name"
                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                            placeholder="enter new name"
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
                                    class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                    @click="renameFile"
                                >
                                    {{ t("common.button.rename") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="closeDialog"
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

import { nextTick, ref, watch } from "vue";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },

    file: {
        type: Object,
        default: () => ({}),
    },
});

const emit = defineEmits(["close", "rename"]);
const name = ref("");

const rules = {
    name: { required },
};
const v$ = useVuelidate(rules, { name });

watch(
    () => props.open,
    async () => {
        name.value = props.file.name;
        if (props.open) {
            await nextTick();
            const inputElement = document.getElementById("name");

            if (inputElement) {
                inputElement.focus();
                inputElement.setSelectionRange(0, props.file.name.indexOf("."));
            }
        }
    },
);

function closeDialog() {
    name.value = "";
    emit("close");
}

async function renameFile() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    emit("rename", { name: name.value, file: props.file });
    closeDialog();
}
</script>
