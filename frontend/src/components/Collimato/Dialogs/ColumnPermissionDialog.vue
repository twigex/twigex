<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="close" :initialFocus="searchInput">
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
                            class="relative transform overflow-hidden rounded-lg bg-white shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-md"
                        >
                            <!-- Header -->
                            <div
                                class="flex items-center justify-between px-6 py-4 border-b border-gray-200"
                            >
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{ t("collimato.roles.column_permission_dialog.title") }}
                                </DialogTitle>
                                <button
                                    @click="close"
                                    type="button"
                                    class="rounded-md text-gray-400 hover:text-gray-500"
                                >
                                    <XMarkIcon class="h-5 w-5" />
                                </button>
                            </div>

                            <!-- Search -->
                            <div class="px-4 pt-3 pb-2">
                                <div class="relative">
                                    <MagnifyingGlassIcon
                                        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                    />
                                    <input
                                        ref="searchInput"
                                        v-model="modelSearch"
                                        type="text"
                                        class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        :placeholder="
                                            t(
                                                'collimato.roles.column_permission_dialog.search_placeholder',
                                            )
                                        "
                                    />
                                </div>
                            </div>

                            <!-- Model list -->
                            <div class="max-h-52 overflow-y-auto px-2 pb-2">
                                <p
                                    class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                >
                                    {{ t("collimato.roles.column_permission_dialog.data_model") }}
                                </p>
                                <button
                                    v-for="model in filteredModels"
                                    :key="model.name"
                                    type="button"
                                    @click="selectModel(model)"
                                    class="flex w-full items-center gap-x-2 rounded-md px-3 py-2 text-sm text-left transition-colors"
                                    :class="
                                        selectedModel?.name === model.name
                                            ? 'bg-indigo-50 text-indigo-700'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    "
                                >
                                    <CircleStackIcon
                                        class="h-4 w-4 shrink-0 opacity-60"
                                        aria-hidden="true"
                                    />
                                    <span class="flex-1 truncate">{{ model.name }}</span>
                                    <CheckIcon
                                        v-if="selectedModel?.name === model.name"
                                        class="h-4 w-4 shrink-0 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </button>
                                <p
                                    v-if="filteredModels.length === 0"
                                    class="py-6 text-center text-sm text-gray-500"
                                >
                                    {{ t("common.label.no_results") }}
                                </p>
                            </div>

                            <!-- Column list (once model selected) -->
                            <div
                                v-if="selectedModel"
                                class="border-t border-gray-200 max-h-52 overflow-y-auto px-2 pb-2"
                            >
                                <p
                                    class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                >
                                    {{ t("collimato.roles.column_permission_dialog.column") }}
                                </p>
                                <button
                                    v-for="col in modelColumns"
                                    :key="col.name"
                                    type="button"
                                    @click="selectedColumn = col"
                                    class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                    :class="
                                        selectedColumn?.name === col.name
                                            ? 'bg-indigo-50 text-indigo-700'
                                            : 'text-gray-700 hover:bg-gray-50'
                                    "
                                >
                                    <span>{{ col.shortTitle }}</span>
                                    <CheckIcon
                                        v-if="selectedColumn?.name === col.name"
                                        class="h-4 w-4 shrink-0 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </button>
                            </div>

                            <!-- Footer -->
                            <div class="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
                                <button
                                    type="button"
                                    @click="close"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    type="button"
                                    @click="add"
                                    :disabled="!selectedModel || !selectedColumn"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white transition-colors"
                                    :class="
                                        !selectedModel || !selectedColumn
                                            ? 'bg-gray-300 cursor-not-allowed'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                >
                                    {{ t("common.button.add") }}
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
import { ref, computed } from "vue";
import { t } from "@/i18n/index.js";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, XMarkIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { CircleStackIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    dataModels: { type: Array, required: true },
});

const emit = defineEmits(["add", "update:modelValue"]);

const searchInput = ref(null);
const selectedModel = ref(null);
const selectedColumn = ref(null);
const modelSearch = ref("");

const filteredModels = computed(() => {
    const q = modelSearch.value.trim().toLowerCase();

    if (!q) return props.dataModels;

    return props.dataModels.filter((m) => m.name.toLowerCase().includes(q));
});

const modelColumns = computed(() => {
    if (!selectedModel.value) return [];

    return (selectedModel.value.dimensions ?? []).concat(selectedModel.value.measures ?? []);
});

function selectModel(model) {
    selectedModel.value = model;
    selectedColumn.value = null;
}

function add() {
    if (!selectedModel.value || !selectedColumn.value) return;
    emit("add", {
        table: selectedModel.value.name,
        field: selectedColumn.value.name,
    });
    close();
}

function close() {
    selectedModel.value = null;
    selectedColumn.value = null;
    modelSearch.value = "";
    emit("update:modelValue", false);
}
</script>
