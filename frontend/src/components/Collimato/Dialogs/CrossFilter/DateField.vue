<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <input
            :value="modelValue"
            @click="open = true"
            type="text"
            readonly
            class="block w-full cursor-pointer rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
            :class="
                error ? 'ring-red-600 focus:ring-red-600' : 'ring-gray-300 focus:ring-indigo-600'
            "
            :placeholder="label"
        />
        <p v-if="error" class="mt-2 text-sm text-red-600">
            {{ t("collimato.dashboard.filter.required") }}
        </p>
    </div>

    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-50" @close="cancel">
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="text-center sm:mt-5">
                                <DialogTitle
                                    as="h3"
                                    class="text-base text-left font-semibold leading-6 text-gray-900"
                                    >{{ title }}</DialogTitle
                                >
                                <div class="mx-8 mt-5">
                                    <DatePicker v-model="draft" />
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:ml-3 sm:w-auto"
                                    @click="apply"
                                >
                                    {{ t("common.button.save") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="cancel"
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
import { ref, watch } from "vue";
import { t } from "@/i18n/index.js";
import DatePicker from "@/components/DatePicker/DatePicker.vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const props = defineProps({
    modelValue: {
        type: [String, Date, null],
        default: null,
    },
    label: {
        type: String,
        required: true,
    },
    title: {
        type: String,
        required: true,
    },
    error: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:modelValue"]);

const open = ref(false);
const draft = ref(props.modelValue);

watch(open, (isOpen) => {
    if (isOpen) {
        draft.value = props.modelValue;
    }
});

function apply() {
    emit("update:modelValue", draft.value);
    open.value = false;
}

function cancel() {
    open.value = false;
}
</script>
