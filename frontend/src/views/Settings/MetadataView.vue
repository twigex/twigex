<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="w-full h-full flex flex-col items-center">
        <!-- Has metadata + license -->
        <div v-if="settingsStore.metadata.length > 0" class="w-full flex-1 overflow-y-auto">
            <div class="px-6 py-6">
                <DataTable v-model:sort="sort" :columns="columns" :items="metadataRows">
                    <template #toolbar>
                        <div class="flex items-center justify-between gap-x-4">
                            <div>
                                <h2 class="text-sm font-semibold text-gray-900">
                                    {{ t("settings.metadata.your_defined_metadata") }}
                                </h2>
                                <p class="mt-0.5 text-sm text-gray-500">
                                    {{ settingsStore.metadata.length }}
                                    {{ t("settings.metadata.items_defined") }}
                                </p>
                            </div>
                            <BaseButton :prepend-icon="PlusIcon" @click="open = true">
                                {{ t("common.button.new") }}
                            </BaseButton>
                        </div>
                    </template>

                    <template #title="{ item }">
                        <span class="font-medium text-gray-900">{{ item.title }}</span>
                    </template>

                    <template #type="{ item }">
                        <span
                            :class="typeBadgeClass(item.type)"
                            class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset"
                        >
                            {{ item.type }}
                        </span>
                    </template>

                    <template #options_count="{ item }">
                        {{ item.type === "selection" ? item.options_count : "" }}
                    </template>

                    <template #created="{ item }">
                        {{ getDateAndTime(item.created) }}
                    </template>

                    <template #actions="{ item }">
                        <div class="flex items-center justify-end gap-x-4 font-medium">
                            <button
                                type="button"
                                class="text-indigo-600 hover:text-indigo-900"
                                @click="openEditDialog(item.source)"
                            >
                                {{ t("common.button.edit") }}
                                <span class="sr-only">, {{ item.title }}</span>
                            </button>
                            <button
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click="openDeleteDialog(item.source)"
                            >
                                {{ t("common.button.delete") }}
                                <span class="sr-only">, {{ item.title }}</span>
                            </button>
                        </div>
                    </template>
                </DataTable>
            </div>
        </div>

        <!-- Empty state with license -->
        <div v-else class="flex-grow h-full flex justify-center items-center">
            <div class="text-center">
                <svg
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                    class="mx-auto h-12 w-12 text-gray-400"
                    aria-hidden="true"
                >
                    <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z"
                    />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 6h.008v.008H6V6Z" />
                </svg>
                <p class="mt-4 text-lg font-semibold text-gray-900">
                    {{ t("settings.metadata.no_metadata") }}
                </p>
                <p class="mt-1 text-sm text-gray-500 max-w-sm mx-auto">
                    {{ t("settings.metadata.description") }}
                </p>
                <div class="mt-6">
                    <button
                        type="button"
                        class="inline-flex items-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                        @click="open = true"
                    >
                        <PlusIcon class="-ml-0.5 mr-1.5 h-5 w-5" aria-hidden="true" />
                        {{ t("settings.metadata.button.new_metadata") }}
                    </button>
                </div>
            </div>
        </div>

        <!-- Create / Edit dialog -->
        <TransitionRoot as="template" :show="open">
            <Dialog as="div" class="relative z-50" @close="open = false">
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
                        class="flex min-h-full items-end justify-center p-4 sm:items-center sm:p-0"
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
                                class="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-md sm:p-6"
                            >
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{
                                        edit
                                            ? t("settings.metadata.title.edit_metadata")
                                            : t("settings.metadata.title.new_metadata")
                                    }}
                                </DialogTitle>

                                <div class="mt-4 space-y-4">
                                    <div>
                                        <label
                                            for="meta-name"
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.metadata.label.metadata_name") }}
                                        </label>
                                        <div class="mt-2">
                                            <input
                                                v-model="name"
                                                type="text"
                                                id="meta-name"
                                                name="meta-name"
                                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:text-sm sm:leading-6"
                                                :class="
                                                    v$.name.$error
                                                        ? 'ring-red-600 focus:ring-red-600'
                                                        : 'ring-gray-300 focus:ring-indigo-600'
                                                "
                                                :placeholder="
                                                    t('settings.metadata.label.metadata_name')
                                                "
                                            />
                                            <p
                                                v-if="v$.name.$error"
                                                class="mt-1.5 text-sm text-red-600"
                                            >
                                                {{ t("common.error.required_field") }}
                                            </p>
                                        </div>
                                    </div>

                                    <div>
                                        <label
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{ t("settings.metadata.label.metadata_type") }}
                                        </label>
                                        <div class="mt-2">
                                            <Listbox as="div" v-model="selected">
                                                <div class="relative">
                                                    <ListboxButton
                                                        class="grid w-full cursor-default grid-cols-1 rounded-md bg-white py-1.5 pl-3 pr-2 text-left text-gray-900 shadow-sm outline outline-1 -outline-offset-1 outline-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-indigo-600 sm:text-sm/6"
                                                    >
                                                        <span
                                                            class="col-start-1 row-start-1 truncate pr-6"
                                                            >{{ selected }}</span
                                                        >
                                                        <ChevronUpDownIcon
                                                            class="col-start-1 row-start-1 size-5 self-center justify-self-end text-gray-500 sm:size-4"
                                                            aria-hidden="true"
                                                        />
                                                    </ListboxButton>

                                                    <transition
                                                        leave-active-class="transition ease-in duration-100"
                                                        leave-from-class="opacity-100"
                                                        leave-to-class="opacity-0"
                                                    >
                                                        <ListboxOptions
                                                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg outline outline-1 outline-black/5 sm:text-sm"
                                                        >
                                                            <ListboxOption
                                                                as="template"
                                                                v-for="type in types"
                                                                :key="type"
                                                                :value="type"
                                                                v-slot="{
                                                                    active,
                                                                    selected: isSelected,
                                                                }"
                                                            >
                                                                <li
                                                                    :class="[
                                                                        active
                                                                            ? 'bg-indigo-600 text-white outline-none'
                                                                            : 'text-gray-900',
                                                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                                                    ]"
                                                                >
                                                                    <span
                                                                        :class="[
                                                                            isSelected
                                                                                ? 'font-semibold'
                                                                                : 'font-normal',
                                                                            'block truncate',
                                                                        ]"
                                                                        >{{ type }}</span
                                                                    >
                                                                    <span
                                                                        v-if="isSelected"
                                                                        :class="[
                                                                            active
                                                                                ? 'text-white'
                                                                                : 'text-indigo-600',
                                                                            'absolute inset-y-0 right-0 flex items-center pr-4',
                                                                        ]"
                                                                    >
                                                                        <CheckIcon
                                                                            class="h-5 w-5"
                                                                            aria-hidden="true"
                                                                        />
                                                                    </span>
                                                                </li>
                                                            </ListboxOption>
                                                        </ListboxOptions>
                                                    </transition>
                                                </div>
                                            </Listbox>
                                        </div>
                                    </div>

                                    <div v-if="selected === types[1]">
                                        <label
                                            class="block text-sm font-medium leading-6 text-gray-900"
                                        >
                                            {{
                                                t("settings.metadata.label.enter_selection_values")
                                            }}
                                        </label>
                                        <div class="mt-2 flex rounded-md shadow-sm">
                                            <div
                                                class="relative flex flex-grow items-stretch focus-within:z-10"
                                            >
                                                <input
                                                    v-model="fieldValue"
                                                    type="text"
                                                    name="field"
                                                    id="field"
                                                    @keyup.enter="addField"
                                                    class="block w-full rounded-none rounded-l-md border-0 py-1.5 pl-3 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                                    :placeholder="
                                                        t('settings.metadata.enter_value')
                                                    "
                                                />
                                            </div>
                                            <button
                                                type="button"
                                                class="relative -ml-px inline-flex items-center gap-x-1.5 rounded-r-md px-3 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                                @click="addField"
                                            >
                                                {{ t("common.button.add") }}
                                            </button>
                                        </div>

                                        <ul
                                            v-if="fields.length > 0"
                                            role="list"
                                            class="mt-3 divide-y divide-gray-100 rounded-md ring-1 ring-gray-200"
                                        >
                                            <li
                                                v-for="(f, i) in fields"
                                                :key="i"
                                                class="flex items-center justify-between px-3 py-2"
                                            >
                                                <span class="text-sm text-gray-900">{{ f }}</span>
                                                <button
                                                    type="button"
                                                    class="rounded p-0.5 text-gray-400 hover:text-red-600 hover:bg-red-50"
                                                    @click="deleteField(f)"
                                                >
                                                    <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                                </button>
                                            </li>
                                        </ul>
                                    </div>
                                </div>

                                <div class="mt-6 flex gap-x-3 sm:flex-row-reverse">
                                    <button
                                        type="button"
                                        class="flex-1 inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                                        @click="create"
                                    >
                                        {{
                                            edit
                                                ? t("common.button.update")
                                                : t("common.button.create")
                                        }}
                                    </button>
                                    <button
                                        type="button"
                                        class="flex-1 inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                        @click="
                                            open = false;
                                            clear();
                                        "
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

        <!-- Delete confirmation dialog -->
        <TransitionRoot as="template" :show="deleteDialog">
            <Dialog as="div" class="relative z-10" @close="((deleteDialog = false), clear)">
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
                                    <div
                                        class="mx-auto flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:h-10 sm:w-10"
                                    >
                                        <ExclamationTriangleIcon
                                            class="h-6 w-6 text-red-600"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold leading-6 text-gray-900"
                                            >{{
                                                t("settings.metadata.dialog.delete_title")
                                            }}</DialogTitle
                                        >
                                        <div class="mt-2">
                                            <p class="text-sm text-gray-500">
                                                {{ t("settings.metadata.dialog.delete_confirm") }}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                                <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                    <button
                                        type="button"
                                        class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto"
                                        @click="deleteMetadata"
                                    >
                                        {{ t("common.button.delete") }}
                                    </button>
                                    <button
                                        type="button"
                                        class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                        @click="deleteDialog = false"
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
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { computed, ref, onMounted } from "vue";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { PlusIcon, CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { TrashIcon, ExclamationTriangleIcon } from "@heroicons/vue/24/outline";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import settingsService from "@/services/settingsService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useSettingsStore } from "@/store/settings";

onMounted(() => {
    settingsService.metadata().then((res) => {
        settingsStore.metadata = res.data;
        loaded.value = true;
    });
});

const settingsStore = useSettingsStore();
const { getDateAndTime } = useDateOperations();
const loaded = ref(false);
const sort = ref({ key: "title", desc: false });

const columns = computed(() => [
    { key: "title", label: t.value("data_table.name"), sortable: true },
    { key: "type", label: t.value("data_table.type"), sortable: true },
    {
        key: "options_count",
        label: t.value("data_table.options"),
        sortable: true,
        hiddenBelow: "sm",
    },
    { key: "created", label: t.value("data_table.created"), sortable: true, hiddenBelow: "md" },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const metadataRows = computed(() =>
    settingsStore.metadata.map((m) => ({
        ...m,
        options_count: m.type === "selection" ? (m.fields?.length ?? 0) : 0,
        source: m,
    })),
);
const open = ref(false);
const deleteDialog = ref(false);
const types = ["text", "selection", "date", "color"];
const colors = [
    { color: "#673AB7", name: "Personal" },
    { color: "#4CAF50", name: "Todo" },
    { color: "#B71C1C", name: "Work" },
    { color: "#FF5722", name: "Important" },
];
const selected = ref(types[0]);
const editItem = ref(null);
const name = ref("");
const fields = ref([]);
const fieldValue = ref("");
const edit = ref(false);

const rules = {
    name: { required },
};
const v$ = useVuelidate(rules, { name });

const typeBadgeClass = (type) => {
    const map = {
        text: "bg-gray-50 text-gray-600 ring-gray-500/10",
        selection: "bg-blue-50 text-blue-700 ring-blue-700/10",
        date: "bg-green-50 text-green-700 ring-green-600/20",
        color: "bg-purple-50 text-purple-700 ring-purple-700/10",
    };

    return map[type] ?? "bg-gray-50 text-gray-600 ring-gray-500/10";
};

async function create() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    if (edit.value == true) {
        settingsService
            .updateMetadata(editItem.value.id, {
                name: name.value,
                type: selected.value,
                fields: fields.value,
            })
            .then(() => {
                let newItem = editItem.value;

                newItem.title = name.value;
                newItem.type = selected.value;
                newItem.fields = fields.value;

                settingsStore.metadata[settingsStore.metadata.indexOf(editItem.value)] = newItem;
                open.value = false;

                clear();
            });

        return;
    }

    settingsService
        .createMetadata({
            name: name.value,
            type: selected.value,
            fields: selected.value == "color" ? colors : fields.value,
        })
        .then((response) => {
            settingsStore.metadata.push(response.data);
            open.value = false;

            clear();
        });
}

function addField() {
    if (!fieldValue.value.trim()) return;
    fields.value.push(fieldValue.value);
    fieldValue.value = "";
}

function deleteField(field) {
    fields.value.splice(fields.value.indexOf(field), 1);
}

function deleteMetadata() {
    settingsService.deleteMetadata(editItem.value.id).then(() => {
        settingsStore.metadata.splice(settingsStore.metadata.indexOf(editItem.value), 1);

        deleteDialog.value = false;

        clear();
    });
}

function openDeleteDialog(item) {
    deleteDialog.value = true;
    editItem.value = item;
}

function openEditDialog(item) {
    name.value = item.title;
    selected.value = item.type;
    fields.value = item.fields;

    open.value = true;
    edit.value = true;
    editItem.value = item;
}

function clear() {
    name.value = "";
    selected.value = types[0];
    fields.value = [];
    fieldValue.value = "";
    edit.value = false;
    editItem.value = null;

    v$.value.$reset();
}
</script>
