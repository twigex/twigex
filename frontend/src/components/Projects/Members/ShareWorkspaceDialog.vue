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
                            class="relative transform rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div class="mt-3 text-center sm:ml-0 sm:mt-0 sm:text-left w-full">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold leading-6 text-gray-900"
                                    >
                                        {{ t("projects.share_workspace_dialog.assign_task") }}
                                    </DialogTitle>

                                    <div class="mt-2">
                                        <div
                                            class="w-full flex items-center rounded-md bg-white text-gray-900 ring-1 ring-inset ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600 shadow-sm sm:text-sm sm:leading-6"
                                        >
                                            <div
                                                v-if="selectedPerson && !isOpen"
                                                class="ml-3 h-6 w-6 shrink-0"
                                            >
                                                <UserAvatar :user="selectedPerson" />
                                            </div>
                                            <input
                                                ref="searchInput"
                                                v-model="query"
                                                type="text"
                                                class="flex-1 bg-transparent border-none outline-none focus:outline-none focus:ring-0 py-1.5 px-3 shadow-none text-sm"
                                                :placeholder="
                                                    selectedPerson && !isOpen
                                                        ? `${selectedPerson.name} ${selectedPerson.lastname}`
                                                        : t(
                                                              'projects.share_workspace_dialog.search_placeholder',
                                                          )
                                                "
                                                @click="openList"
                                                @input="onQueryInput"
                                            />
                                            <button
                                                v-if="selectedPerson && !isOpen"
                                                type="button"
                                                class="pr-1 pl-1 text-gray-400 hover:text-gray-600"
                                                @click.stop="selectedPerson = null"
                                            >
                                                <span class="sr-only">{{
                                                    t("common.button.clear")
                                                }}</span>
                                                <XMarkIcon class="h-4 w-4" aria-hidden="true" />
                                            </button>
                                            <button
                                                class="pr-2 pl-1 text-gray-400 hover:text-gray-600"
                                                @click.stop="toggleList"
                                            >
                                                <ChevronUpDownIcon
                                                    class="h-5 w-5"
                                                    aria-hidden="true"
                                                />
                                            </button>
                                        </div>

                                        <!-- Results list, normal flow so overflow-hidden doesn't clip it -->
                                        <div
                                            v-if="isOpen"
                                            ref="listEl"
                                            class="mt-1 max-h-56 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 sm:text-sm"
                                        >
                                            <ul>
                                                <li
                                                    v-for="person in results"
                                                    :key="person.id"
                                                    class="relative cursor-default select-none py-2 pl-3 pr-9 hover:bg-indigo-600 hover:text-white text-gray-900"
                                                    :class="
                                                        selectedPerson?.id === person.id
                                                            ? 'bg-indigo-50 font-semibold'
                                                            : ''
                                                    "
                                                    @click="selectPerson(person)"
                                                >
                                                    <div class="flex items-center">
                                                        <div class="h-6 w-6 shrink-0">
                                                            <UserAvatar :user="person" />
                                                        </div>
                                                        <span class="ml-3 truncate"
                                                            >{{ person.name }}
                                                            {{ person.lastname }}</span
                                                        >
                                                    </div>
                                                    <span
                                                        v-if="selectedPerson?.id === person.id"
                                                        class="absolute inset-y-0 right-0 flex items-center pr-4 text-indigo-600"
                                                    >
                                                        <CheckIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
                                                    </span>
                                                </li>

                                                <li ref="sentinel" class="h-1" />

                                                <li
                                                    v-if="loading"
                                                    class="py-2 text-center text-xs text-gray-400"
                                                >
                                                    {{ t("common.label.loading") }}
                                                </li>
                                                <li
                                                    v-else-if="!results.length"
                                                    class="py-2 text-center text-xs text-gray-400"
                                                >
                                                    {{ t("common.label.no_users") }}
                                                </li>
                                            </ul>
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="mt-5 sm:mt-4 sm:flex sm:items-center sm:justify-between">
                                <div class="min-w-[150px]">
                                    <button
                                        v-if="props.selectedUserID != null"
                                        type="button"
                                        class="inline-flex justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:w-auto"
                                        @click="assignTask(true)"
                                    >
                                        {{ t("projects.share_workspace_dialog.remove_assignee") }}
                                    </button>
                                </div>
                                <div class="flex sm:flex-row-reverse space-x-2 space-x-reverse">
                                    <button
                                        type="button"
                                        class="inline-flex justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 sm:w-auto disabled:opacity-50"
                                        :disabled="!selectedPerson"
                                        @click="assignTask(false)"
                                    >
                                        {{ t("projects.share_workspace_dialog.assign") }}
                                    </button>
                                    <button
                                        type="button"
                                        class="inline-flex justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:w-auto"
                                        @click="closeDialog"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </button>
                                </div>
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
import { ref, computed, toRefs, watch, onBeforeUnmount, nextTick } from "vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { CheckIcon, ChevronUpDownIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";

const PAGE = 20;

const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();

const props = defineProps({
    open: { type: Boolean, default: false },
    selectedWorkspace: { type: Object, default: () => ({}) },
    selectedTask: { type: Object, default: () => ({}) },
    assignFieldName: { type: String, default: "" },
    selectedAssigneTable: { type: Object, default: () => ({}) },
    isDialog: { type: Boolean, default: false },
    selectedUserID: { type: String, default: null },
});

const emit = defineEmits(["close", "create", "updateSelectedItem", "update:assignee"]);
const { selectedWorkspace, selectedTask, assignFieldName, selectedAssigneTable } = toRefs(props);

const query = ref("");
const selectedPerson = ref(null);
const isOpen = ref(false);
const results = ref([]);
const loading = ref(false);
const offset = ref(0);
const hasMore = ref(true);
const listEl = ref(null);
const sentinel = ref(null);
const searchInput = ref(null);

function openList() {
    if (isOpen.value) return;
    isOpen.value = true;
    resetAndSearch();
}

function closeList() {
    isOpen.value = false;
    query.value = "";
    results.value = [];
    offset.value = 0;
    hasMore.value = true;
}

function toggleList() {
    if (isOpen.value) closeList();
    else openList();
}

let debounceTimer = null;

function onQueryInput() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(resetAndSearch, 250);
}

function resetAndSearch() {
    results.value = [];
    offset.value = 0;
    hasMore.value = true;
    fetchPage(true);
}

let latestRequest = 0;

async function fetchPage(reset = false) {
    if (!reset && (loading.value || !hasMore.value)) return;
    const wsId = selectedWorkspace.value?.id;

    if (!wsId) return;

    const seq = ++latestRequest;

    loading.value = true;
    try {
        const res = await workspaceService.searchWorkspaceMembers(
            wsId,
            query.value,
            PAGE,
            offset.value,
        );

        if (seq !== latestRequest) return;
        const batch = res.data ?? [];

        results.value.push(...batch);
        offset.value += batch.length;
        hasMore.value = batch.length === PAGE;
    } catch {
        if (seq !== latestRequest) return;
        hasMore.value = false;
    } finally {
        if (seq === latestRequest) loading.value = false;
    }
}

let observer = null;

watch(results, async () => {
    await nextTick();
    if (!sentinel.value || !listEl.value) return;
    if (observer) observer.disconnect();
    observer = new IntersectionObserver(
        (entries) => {
            if (entries[0].isIntersecting) fetchPage();
        },
        { root: listEl.value, threshold: 0.1 },
    );
    observer.observe(sentinel.value);
});

watch(
    () => props.open,
    (open) => {
        if (!open) return;
        selectedPerson.value = null;
        isOpen.value = false;
        query.value = "";
        results.value = [];
        offset.value = 0;
        hasMore.value = true;
        showCurrentAssignee(props.selectedUserID);
        nextTick(() => searchInput.value?.focus());
    },
);

async function showCurrentAssignee(id) {
    if (!id) return;

    await userStore.ensureUsers([id]);

    if (props.open && props.selectedUserID === id && !selectedPerson.value) {
        selectedPerson.value = userStore.getUserById(id) ?? null;
    }
}

onBeforeUnmount(() => observer?.disconnect());

function selectPerson(person) {
    selectedPerson.value = person;
    closeList();
}

const tableData = computed({
    get: () => workspaceStore.getTableData,
    set: (v) => workspaceStore.setTableData(v),
});
const getSelectedItem = computed({
    get: () => workspaceStore.getSelectedItem,
    set: (v) => workspaceStore.setSelectedItem(v),
});

const assignTask = (remove) => {
    const userID = remove ? "" : (selectedPerson.value?.id ?? "");

    // Ensure assigned user is in cache so grid/kanban can show their name/photo
    if (!remove && selectedPerson.value) {
        userStore.addUsers([selectedPerson.value]);
    }

    workspaceService
        .addMemberToTask({
            workspace_id: selectedWorkspace.value.id,
            table_id: selectedAssigneTable.value.id,
            task_id: selectedTask.value.id,
            user_id: userID,
            field_name: assignFieldName.value,
        })
        .then(() => {
            const nowMs = Date.now();

            tableData.value = tableData.value.map((row) =>
                row.id === selectedTask.value.id
                    ? { ...row, [assignFieldName.value]: userID, updated_at: nowMs }
                    : row,
            );
            if (selectedTask.value.id === getSelectedItem.value.id) {
                getSelectedItem.value = {
                    ...getSelectedItem.value,
                    [assignFieldName.value]: userID,
                    updated_at: nowMs,
                };
            }

            emit("update:assignee", {
                workspace_id: selectedWorkspace.value.id,
                table_id: selectedAssigneTable.value.id,
                task_id: selectedTask.value.id,
                user_id: userID,
                field_name: assignFieldName.value,
            });
            closeDialog();
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
};

const closeDialog = () => {
    selectedPerson.value = null;
    closeList();
    emit("close");
};
</script>
