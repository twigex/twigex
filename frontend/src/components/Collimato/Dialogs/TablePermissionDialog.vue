<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog as="div" class="relative z-50" @close="close">
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
                            class="relative transform overflow-hidden rounded-lg bg-white text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-md"
                        >
                            <!-- Header -->
                            <div class="px-6 pt-6 pb-4">
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ t("collimato.roles.table_permission_dialog.title") }}
                                </DialogTitle>
                                <p class="mt-1 text-sm text-gray-500">
                                    {{ t("collimato.roles.table_permission_dialog.description") }}
                                </p>

                                <!-- Search -->
                                <div class="relative mt-3">
                                    <MagnifyingGlassIcon
                                        class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                        aria-hidden="true"
                                    />
                                    <input
                                        v-model="searchQuery"
                                        type="search"
                                        :placeholder="
                                            t(
                                                'collimato.roles.table_permission_dialog.search_placeholder',
                                            )
                                        "
                                        class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                    />
                                </div>

                                <!-- Selected chips -->
                                <div
                                    v-if="selectedModels.length > 0"
                                    class="mt-3 flex flex-wrap gap-1.5"
                                >
                                    <span
                                        v-for="model in selectedModels"
                                        :key="model.name"
                                        class="inline-flex items-center gap-x-1 rounded-full bg-indigo-100 px-2.5 py-1 text-xs font-medium text-indigo-700"
                                    >
                                        {{ model.name }}
                                        <button
                                            type="button"
                                            @click="deselect(model)"
                                            class="rounded-full p-0.5 hover:bg-indigo-200"
                                        >
                                            <XMarkIcon class="h-3 w-3" aria-hidden="true" />
                                        </button>
                                    </span>
                                </div>
                            </div>

                            <!-- Model list -->
                            <div class="border-t max-h-64 overflow-y-auto">
                                <div
                                    v-if="filteredModels.length === 0"
                                    class="flex flex-col items-center justify-center py-8 px-4 text-center"
                                >
                                    <CircleStackIcon
                                        class="h-8 w-8 text-gray-300"
                                        aria-hidden="true"
                                    />
                                    <p class="mt-2 text-sm text-gray-500">
                                        {{ t("common.label.no_results") }}
                                    </p>
                                </div>
                                <ul v-else role="list" class="divide-y divide-gray-100">
                                    <li
                                        v-for="model in filteredModels"
                                        :key="model.name"
                                        class="flex items-center gap-x-3 px-4 py-3 hover:bg-gray-50 cursor-pointer"
                                        @click="toggleModel(model)"
                                    >
                                        <div
                                            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full"
                                            :class="
                                                isSelected(model) ? 'bg-indigo-100' : 'bg-gray-100'
                                            "
                                        >
                                            <CircleStackIcon
                                                class="h-4 w-4"
                                                :class="
                                                    isSelected(model)
                                                        ? 'text-indigo-600'
                                                        : 'text-gray-400'
                                                "
                                                aria-hidden="true"
                                            />
                                        </div>
                                        <span
                                            class="flex-1 text-sm font-medium text-gray-900 truncate"
                                            >{{ model.name }}</span
                                        >
                                        <CheckIcon
                                            v-if="isSelected(model)"
                                            class="h-4 w-4 shrink-0 text-indigo-600"
                                            aria-hidden="true"
                                        />
                                    </li>
                                </ul>
                            </div>

                            <!-- Footer -->
                            <div
                                class="flex flex-row-reverse gap-x-3 border-t bg-gray-50 px-6 py-4"
                            >
                                <button
                                    type="button"
                                    @click="add"
                                    :disabled="selectedModels.length === 0"
                                    class="inline-flex items-center rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                    :class="
                                        selectedModels.length === 0
                                            ? 'bg-indigo-300 cursor-not-allowed'
                                            : 'bg-indigo-600 hover:bg-indigo-500'
                                    "
                                >
                                    {{ t("common.button.add") }}
                                    <span
                                        v-if="selectedModels.length > 0"
                                        class="ml-1.5 inline-flex items-center rounded-full bg-indigo-500 px-1.5 py-0.5 text-xs font-semibold"
                                    >
                                        {{ selectedModels.length }}
                                    </span>
                                </button>
                                <button
                                    type="button"
                                    @click="close"
                                    class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
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
import { ref, computed } from "vue";
import { t } from "@/i18n/index.js";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, MagnifyingGlassIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { CircleStackIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    modelValue: { type: Boolean, required: true },
    dataModels: { type: Array, required: true },
});

const emit = defineEmits(["add", "update:modelValue"]);

const selectedModels = ref([]);
const searchQuery = ref("");

const filteredModels = computed(() => {
    const q = searchQuery.value.trim().toLowerCase();

    if (!q) return props.dataModels;

    return props.dataModels.filter((m) => m.name.toLowerCase().includes(q));
});

function isSelected(model) {
    return selectedModels.value.some((m) => m.name === model.name);
}

function toggleModel(model) {
    const idx = selectedModels.value.findIndex((m) => m.name === model.name);

    if (idx >= 0) {
        selectedModels.value.splice(idx, 1);
    } else {
        selectedModels.value.push(model);
    }
}

function deselect(model) {
    selectedModels.value = selectedModels.value.filter((m) => m.name !== model.name);
}

function add() {
    if (selectedModels.value.length === 0) return;
    emit("add", selectedModels.value);
    close();
}

function close() {
    selectedModels.value = [];
    searchQuery.value = "";
    emit("update:modelValue", false);
}
</script>
