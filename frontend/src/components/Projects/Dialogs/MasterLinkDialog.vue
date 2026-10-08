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
                            class="relative w-full transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:max-w-lg sm:p-6"
                        >
                            <div>
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{ t("projects.grid_view.types_link_to_table") }}
                                </DialogTitle>
                                <p class="mt-2 text-sm text-gray-500">
                                    <Listbox as="div" v-model="selectedTable">
                                        <div class="relative mt-2">
                                            <!-- Button -->
                                            <ListboxButton
                                                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                            >
                                                <span class="block truncate">{{
                                                    selectedTableName
                                                }}</span>
                                                <span
                                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                                >
                                                    <ChevronUpDownIcon
                                                        class="h-5 w-5 text-gray-400"
                                                        aria-hidden="true"
                                                    />
                                                </span>
                                            </ListboxButton>

                                            <!-- Dropdown List -->
                                            <transition
                                                leave-active-class="transition ease-in duration-100"
                                                leave-from-class="opacity-100"
                                                leave-to="opacity-0"
                                            >
                                                <ListboxOptions
                                                    class="absolute z-50 left-0 top-full mt-1 w-full max-h-60 overflow-y-auto bg-white rounded-md py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                                                >
                                                    <ListboxOption
                                                        as="template"
                                                        v-for="table in linkTables"
                                                        :key="table.id"
                                                        :value="table"
                                                        v-slot="{ active, selected: isSelected }"
                                                    >
                                                        <li
                                                            :class="[
                                                                active
                                                                    ? 'bg-indigo-600 text-white'
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
                                                            >
                                                                {{
                                                                    table.display_name?.Valid &&
                                                                    table.display_name?.String?.trim()
                                                                        ? table.display_name.String
                                                                        : table.name
                                                                }}
                                                            </span>

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
                                </p>
                            </div>

                            <div class="mt-5 sm:mt-6 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 sm:ml-3 sm:w-auto"
                                    @click="saveLinkValue(selectedTable)"
                                >
                                    {{ t("common.button.link") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="closeDialog"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                                <button
                                    v-if="currentTableID"
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md px-3 py-2 text-sm font-semibold text-red-600 hover:bg-red-50 sm:mr-auto sm:mt-0 sm:w-auto"
                                    @click="saveLinkValue(null)"
                                >
                                    {{ t("common.button.remove") }}
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
import { ref, computed, watch } from "vue";
import { t } from "@/i18n/index.js";
import { useRoute } from "vue-router";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import {
    Dialog,
    DialogPanel,
    DialogTitle,
    TransitionChild,
    TransitionRoot,
    Listbox,
    ListboxButton,
    ListboxOptions,
    ListboxOption,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    field: {
        type: String,
        default: "",
    },
    masterLinkTask: {
        type: Object,
        default: () => ({}),
    },
});

const emit = defineEmits(["close"]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();

const linkTables = ref([]);
const selectedTable = ref(null);

const loadTables = async () => {
    try {
        const tableID = route.params.tid;

        linkTables.value = []; // Clear previous tables

        // Fetch workspace details from API
        const response = await workspaceStore.fetchWorkspaceNav(route.params.id);

        if (!response || !response.data) {
            return;
        }

        const workspace = response.data; // Assuming response.data contains workspace details

        if (!workspace || !workspace.id) {
            return;
        }

        // Ensure tables and children exist and are arrays
        const tables = Array.isArray(workspace.tables) ? workspace.tables : [];
        const children = Array.isArray(workspace.children) ? workspace.children : [];

        const allTables = [...tables, ...children];

        // Filter tables: remove linked and single-select tables
        const filteredTables = allTables.filter((table) => {
            if (!table || !table.id) {
                return false;
            }

            const isSingleSelect = table.single_select === true;
            const isLinkedTable = table.parent_table_id !== null && table.parent_table_id !== "";

            if (isSingleSelect || isLinkedTable) {
                return false;
            }

            return table.id !== tableID;
        });

        // Update state
        linkTables.value = filteredTables;
        selectedTable.value =
            linkTables.value.find((table) => table.id === currentTableID.value) ??
            linkTables.value[0] ??
            null;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const getSelectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const tableData = computed({
    get() {
        return workspaceStore.getTableData;
    },
    set(value) {
        workspaceStore.setTableData(value);
    },
});

const currentTableID = computed(() => props.masterLinkTask?.[props.field] || "");

const saveLinkValue = (newTable) => {
    const fieldName = props.field;
    const value = newTable?.id ?? "";
    let tableID = route.params.tid;

    workspaceService
        .updateTask({
            workspace_id: route.params.id,
            table_id: tableID,
            task_id: props.masterLinkTask.id,
            field: fieldName,
            value,
        })
        .then((response) => {
            const updatedValue = response.data[fieldName] ?? value;

            const updatedTableData = tableData.value.map((task) =>
                task.id === props.masterLinkTask.id ? { ...task, [fieldName]: updatedValue } : task,
            );

            tableData.value = updatedTableData;

            if (getSelectedItem.value && getSelectedItem.value.id === props.masterLinkTask.id) {
                getSelectedItem.value = {
                    ...getSelectedItem.value,
                    [fieldName]: updatedValue,
                };
            }
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });

    closeDialog();
};

watch(
    () => props.open,
    (newVal) => {
        if (newVal) {
            loadTables();
        }
    },
);

const selectedTableName = computed(() => {
    const table = selectedTable.value;

    if (!table) return t.value("projects.dialogs.master_link_dialog.select_table");

    const displayName = table.display_name?.String;
    const hasDisplayName = table.display_name?.Valid && displayName?.trim();

    return hasDisplayName ? displayName : table.name;
});

const closeDialog = () => {
    emit("close");
};
</script>
