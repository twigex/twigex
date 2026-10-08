<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="close">
            <TransitionChild
                as="template"
                enter="ease-out duration-200"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-150"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-900/40 backdrop-blur-sm" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 flex items-center justify-center p-4">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-200"
                    enter-from="opacity-0 scale-95"
                    enter-to="opacity-100 scale-100"
                    leave="ease-in duration-150"
                    leave-from="opacity-100 scale-100"
                    leave-to="opacity-0 scale-95"
                >
                    <DialogPanel
                        class="w-full max-w-xl overflow-hidden rounded-xl bg-white text-gray-900 shadow-2xl ring-1 ring-black/5"
                    >
                        <header class="flex items-center justify-between px-5 pt-5">
                            <DialogTitle class="text-lg font-semibold text-gray-900">
                                {{ title }}
                            </DialogTitle>
                            <button
                                type="button"
                                class="rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-700"
                                @click="close"
                            >
                                ×
                            </button>
                        </header>

                        <div v-if="subtitle" class="px-5 pt-2 text-center text-gray-700">
                            <span class="text-lg">{{ subtitle }}</span>
                        </div>

                        <div class="px-5 pt-4 pb-2">
                            <UserPicker
                                v-if="modelValue"
                                v-model="selected"
                                :multiple="multiple"
                                :exclude-ids="excludeIds"
                                :placeholder="placeholder"
                            />
                        </div>

                        <footer class="border-t border-gray-200 bg-white px-5 py-4">
                            <div class="grid grid-cols-2 gap-3">
                                <button
                                    class="rounded-md bg-gray-100 px-3 py-2 text-sm font-semibold hover:bg-gray-200"
                                    @click="close"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-500 disabled:opacity-50"
                                    :disabled="!hasSelection"
                                    @click="confirm"
                                >
                                    {{ confirmLabel || t("common.button.add") }}
                                </button>
                            </div>
                        </footer>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, watch } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import UserPicker from "@/components/UserPicker.vue";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    title: { type: String, default: "" },
    subtitle: { type: String, default: "" },
    multiple: { type: Boolean, default: true },
    excludeIds: { type: [Array, Set], default: () => [] },
    placeholder: { type: String, default: "" },
    confirmLabel: { type: String, default: "" },
});

const emit = defineEmits(["confirm", "update:modelValue"]);

const selected = ref(props.multiple ? [] : null);

watch(
    () => props.modelValue,
    (open) => {
        if (open) selected.value = props.multiple ? [] : null;
    },
);

const hasSelection = computed(() =>
    props.multiple ? selected.value.length > 0 : !!selected.value,
);

function confirm() {
    if (!hasSelection.value) return;
    emit("confirm", selected.value);
    close();
}

function close() {
    emit("update:modelValue", false);
}
</script>
