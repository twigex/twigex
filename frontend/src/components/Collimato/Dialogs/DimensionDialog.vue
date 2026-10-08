<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
        <Dialog as="div" class="relative z-10" @close="close" :initialFocus="searchInput">
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
                            class="relative transform rounded-lg bg-white shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-md overflow-hidden"
                        >
                            <div
                                class="flex items-center justify-between px-6 py-4 border-b border-gray-200"
                            >
                                <DialogTitle class="text-base font-semibold text-gray-900">
                                    {{
                                        t(
                                            "collimato.charts.new_chart.chart_query.dimension_dialog.title",
                                        )
                                    }}
                                </DialogTitle>
                                <button
                                    @click="close"
                                    type="button"
                                    class="rounded-md text-gray-400 hover:text-gray-500"
                                >
                                    <XMarkIcon class="h-5 w-5" />
                                </button>
                            </div>

                            <div class="px-4 pt-3 pb-2">
                                <div class="relative">
                                    <MagnifyingGlassIcon
                                        class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                    />
                                    <input
                                        ref="searchInput"
                                        v-model="query"
                                        type="text"
                                        class="block w-full rounded-md border-0 py-1.5 pl-9 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        :placeholder="t('common.placeholder.search')"
                                    />
                                </div>
                            </div>

                            <div class="max-h-72 overflow-y-auto px-2 pb-2">
                                <template v-for="group in filteredData" :key="group.table">
                                    <p
                                        class="sticky top-0 bg-white px-2 py-1.5 text-xs font-semibold uppercase tracking-wider text-gray-400"
                                    >
                                        {{ group.table }}
                                    </p>
                                    <button
                                        v-for="dimension in group.dimensions"
                                        :key="dimension.name"
                                        type="button"
                                        @click="selected = dimension"
                                        class="flex w-full items-center justify-between rounded-md px-3 py-2 text-sm text-left transition-colors"
                                        :class="
                                            selected?.name === dimension.name
                                                ? 'bg-indigo-50 text-indigo-700'
                                                : 'text-gray-700 hover:bg-gray-50'
                                        "
                                    >
                                        <span>{{ dimension.title }}</span>
                                        <CheckIcon
                                            v-if="selected?.name === dimension.name"
                                            class="h-4 w-4 shrink-0 text-indigo-600"
                                        />
                                    </button>
                                </template>
                                <p
                                    v-if="filteredData.length === 0"
                                    class="py-8 text-center text-sm text-gray-500"
                                >
                                    {{ t("common.label.no_results") }}
                                </p>
                            </div>

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
                                    :disabled="!selected"
                                    class="rounded-md px-3 py-2 text-sm font-semibold text-white transition-colors"
                                    :class="
                                        selected
                                            ? 'bg-indigo-600 hover:bg-indigo-500'
                                            : 'bg-gray-300 cursor-not-allowed'
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
import { t } from "@/i18n/index.js";
import { ref, computed } from "vue";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { CheckIcon, MagnifyingGlassIcon, XMarkIcon } from "@heroicons/vue/20/solid";

const props = defineProps({
    open: { type: Boolean, required: true },
    data: { type: Array, default: () => [] },
});

const emit = defineEmits(["close", "add"]);

const selected = ref(null);
const query = ref("");
const searchInput = ref(null);

const filteredData = computed(() => {
    const q = query.value.toLowerCase().replace(/\s+/g, "");

    return props.data
        .map((g) => ({
            ...g,
            dimensions:
                q === ""
                    ? g.dimensions
                    : g.dimensions.filter((d) =>
                          d.title.toLowerCase().replace(/\s+/g, "").includes(q),
                      ),
        }))
        .filter((g) => g.dimensions.length > 0);
});

function add() {
    emit("add", selected.value);
    close();
}

function close() {
    emit("close");
    selected.value = null;
    query.value = "";
}
</script>
