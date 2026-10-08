<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="open">
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
                            class="relative transform rounded-xl bg-white text-left shadow-xl transition-all w-full sm:my-8 sm:max-w-xl"
                        >
                            <!-- Header -->
                            <div
                                class="flex items-start justify-between gap-x-4 border-b border-gray-200 px-5 py-4"
                            >
                                <div class="min-w-0">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                    >
                                        {{ t("files.details.edit_metadata") }}
                                    </DialogTitle>
                                    <p
                                        v-if="detailsStore.file"
                                        class="mt-0.5 text-xs text-gray-500 truncate"
                                        :title="detailsStore.file.name"
                                    >
                                        {{ detailsStore.file.name }}
                                    </p>
                                </div>
                                <button
                                    type="button"
                                    class="shrink-0 rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors cursor-pointer"
                                    @click="close"
                                >
                                    <span class="sr-only">{{
                                        t("files.metadata_dialog.close")
                                    }}</span>
                                    <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                                </button>
                            </div>

                            <!-- Empty configuration -->
                            <p
                                v-if="!settingsStore.metadata?.length"
                                class="px-5 py-10 text-sm text-center text-gray-500"
                            >
                                {{ t("files.metadata_dialog.no_fields_configured") }}
                            </p>

                            <template v-else>
                                <!-- Search -->
                                <div class="px-5 pt-4">
                                    <div class="relative">
                                        <MagnifyingGlassIcon
                                            class="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        <input
                                            v-model="search"
                                            type="text"
                                            :placeholder="
                                                t('files.metadata_dialog.search_placeholder')
                                            "
                                            class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                        />
                                    </div>
                                </div>

                                <!-- Lists -->
                                <div class="px-5 py-4 max-h-[28rem] overflow-y-auto space-y-5">
                                    <!-- Applied -->
                                    <section v-if="appliedMetadata.length">
                                        <h4
                                            class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2"
                                        >
                                            {{ t("files.metadata_dialog.applied") }}
                                            <span class="text-gray-400 font-normal"
                                                >({{ appliedMetadata.length }})</span
                                            >
                                        </h4>
                                        <ul
                                            role="list"
                                            class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200 bg-white"
                                        >
                                            <li
                                                v-for="m in appliedMetadata"
                                                :key="m.id"
                                                class="px-4 py-3"
                                            >
                                                <div
                                                    class="flex items-center justify-between gap-x-3"
                                                >
                                                    <div class="min-w-0">
                                                        <p
                                                            class="text-sm font-semibold text-gray-900 truncate"
                                                        >
                                                            {{ m.title }}
                                                        </p>
                                                        <p class="text-xs text-gray-400 capitalize">
                                                            {{ m.type }}
                                                        </p>
                                                    </div>
                                                    <div class="flex items-center gap-x-2 shrink-0">
                                                        <span
                                                            v-if="pending.has(m.id)"
                                                            class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-indigo-600"
                                                        />
                                                        <button
                                                            type="button"
                                                            class="rounded-md p-1 text-gray-400 hover:text-red-600 hover:bg-red-50 transition-colors cursor-pointer"
                                                            :disabled="pending.has(m.id)"
                                                            @click="remove(m)"
                                                            :title="
                                                                t(
                                                                    'files.details.menu.remove_access',
                                                                )
                                                            "
                                                        >
                                                            <TrashIcon
                                                                class="h-4 w-4"
                                                                aria-hidden="true"
                                                            />
                                                        </button>
                                                    </div>
                                                </div>

                                                <div class="mt-2">
                                                    <!-- text: auto-save on blur or Enter -->
                                                    <template v-if="m.type === 'text'">
                                                        <input
                                                            v-model="localValues[m.id]"
                                                            type="text"
                                                            class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6 disabled:opacity-50"
                                                            :placeholder="
                                                                t(
                                                                    'files.metadata_dialog.enter_value',
                                                                )
                                                            "
                                                            :disabled="pending.has(m.id)"
                                                            @blur="hasChanges(m) && save(m)"
                                                            @keyup.enter="hasChanges(m) && save(m)"
                                                        />
                                                    </template>

                                                    <!-- date: auto-save on pick -->
                                                    <template v-else-if="m.type === 'date'">
                                                        <MetadataDateField
                                                            :modelValue="localValues[m.id]"
                                                            @update:modelValue="
                                                                (val) => {
                                                                    localValues[m.id] = val;
                                                                    save(m);
                                                                }
                                                            "
                                                        />
                                                    </template>

                                                    <!-- selection / color: auto-save on pick -->
                                                    <MetadataSelectField
                                                        v-else-if="
                                                            m.type === 'selection' ||
                                                            m.type === 'color'
                                                        "
                                                        :type="m.type"
                                                        :fields="m.fields"
                                                        :modelValue="localValues[m.id]"
                                                        @update:modelValue="
                                                            (val) => {
                                                                localValues[m.id] = val;
                                                                save(m);
                                                            }
                                                        "
                                                    />
                                                </div>
                                            </li>
                                        </ul>
                                    </section>

                                    <!-- Available -->
                                    <section v-if="availableMetadata.length">
                                        <h4
                                            class="text-xs font-semibold uppercase tracking-wide text-gray-500 mb-2"
                                        >
                                            {{ t("files.metadata_dialog.available") }}
                                            <span class="text-gray-400 font-normal"
                                                >({{ availableMetadata.length }})</span
                                            >
                                        </h4>
                                        <ul
                                            role="list"
                                            class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200 bg-white overflow-hidden"
                                        >
                                            <li v-for="m in availableMetadata" :key="m.id">
                                                <button
                                                    type="button"
                                                    class="w-full flex items-center justify-between gap-x-3 px-4 py-3 text-left hover:bg-gray-50 transition-colors disabled:opacity-50 cursor-pointer"
                                                    :disabled="pending.has(m.id)"
                                                    @click="add(m)"
                                                >
                                                    <div class="min-w-0">
                                                        <p
                                                            class="text-sm font-medium text-gray-900 truncate"
                                                        >
                                                            {{ m.title }}
                                                        </p>
                                                        <p class="text-xs text-gray-400 capitalize">
                                                            {{ m.type }}
                                                        </p>
                                                    </div>
                                                    <span
                                                        v-if="pending.has(m.id)"
                                                        class="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-indigo-600 shrink-0"
                                                    />
                                                    <PlusIcon
                                                        v-else
                                                        class="h-4 w-4 text-gray-400 shrink-0"
                                                        aria-hidden="true"
                                                    />
                                                </button>
                                            </li>
                                        </ul>
                                    </section>

                                    <!-- No matches -->
                                    <p
                                        v-if="!appliedMetadata.length && !availableMetadata.length"
                                        class="text-sm text-center text-gray-500 py-6"
                                    >
                                        {{
                                            t("files.metadata_dialog.no_matches", { query: search })
                                        }}
                                    </p>
                                </div>
                            </template>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { reactive, ref, computed, watch } from "vue";
import { t } from "@/i18n/index.js";
import { useSettingsStore } from "@/store/settings";
import { useDetailsStore } from "@/store/details";
import MetadataDateField from "@/components/Files/Metadata/MetadataDateField.vue";
import MetadataSelectField from "@/components/Files/Metadata/MetadataSelectField.vue";
import filesService from "@/services/fileService";

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { PlusIcon, TrashIcon, XMarkIcon, MagnifyingGlassIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    open: Boolean,
});

const emit = defineEmits(["close"]);
const settingsStore = useSettingsStore();
const detailsStore = useDetailsStore();
const localValues = reactive({});
const pending = reactive(new Set());
const search = ref("");

const activeIds = computed(
    () => new Set(detailsStore.file?.metadata?.map((m) => m.metadata_id) ?? []),
);

const filtered = computed(() => {
    const meta = settingsStore.metadata ?? [];
    const q = search.value.trim().toLowerCase();

    if (!q) return meta;

    return meta.filter((m) => m.title?.toLowerCase().includes(q));
});

const appliedMetadata = computed(() => filtered.value.filter((m) => activeIds.value.has(m.id)));

const availableMetadata = computed(() => filtered.value.filter((m) => !activeIds.value.has(m.id)));

watch(
    () => props.open,
    (isOpen) => {
        if (!isOpen) return;
        search.value = "";
        for (const key of Object.keys(localValues)) {
            delete localValues[key];
        }

        for (const m of detailsStore.file?.metadata ?? []) {
            localValues[m.metadata_id] = m.value;
        }
    },
);

function hasChanges(metadata) {
    const record = detailsStore.file.metadata?.find((e) => e.metadata_id === metadata.id);

    return localValues[metadata.id] !== record?.value;
}

function save(metadata) {
    const record = detailsStore.file.metadata?.find((e) => e.metadata_id === metadata.id);

    if (!record) return;
    pending.add(metadata.id);
    filesService
        .updateMetadata(record.id, { value: localValues[metadata.id] })
        .then(() => {
            record.value = localValues[metadata.id];
        })
        .finally(() => {
            pending.delete(metadata.id);
        });
}

function add(metadata) {
    if (pending.has(metadata.id)) return;
    pending.add(metadata.id);
    filesService
        .addMetadata({
            id: detailsStore.file.id,
            metaid: metadata.id,
            value: "",
        })
        .then((response) => {
            if (!detailsStore.file.metadata) {
                detailsStore.file.metadata = [response.data];
            } else {
                detailsStore.file.metadata.push(response.data);
            }

            localValues[metadata.id] = response.data.value ?? "";
        })
        .finally(() => {
            pending.delete(metadata.id);
        });
}

function remove(metadata) {
    const index = detailsStore.file.metadata?.findIndex((e) => e.metadata_id === metadata.id) ?? -1;

    if (index === -1) return;
    pending.add(metadata.id);
    filesService
        .removeMetadata({
            fileID: detailsStore.file.id,
            id: detailsStore.file.metadata[index].id,
        })
        .then(() => {
            detailsStore.file.metadata.splice(index, 1);
            delete localValues[metadata.id];
        })
        .finally(() => {
            pending.delete(metadata.id);
        });
}

function close() {
    emit("close");
}
</script>
