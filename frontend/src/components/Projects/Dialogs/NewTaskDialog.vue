<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <FormDialog
        :open="open"
        :title="t('projects.new_task.title')"
        :confirm-label="t('common.button.create')"
        :secondary-label="t('projects.new_task.create_and_open')"
        :confirm-disabled="!canCreate"
        :loading="saving"
        @confirm="create(false)"
        @secondary="create(true)"
        @close="close"
    >
        <div class="space-y-4">
            <div>
                <label
                    for="new-task-name"
                    class="block text-sm font-medium leading-6 text-gray-900"
                >
                    {{ t("projects.new_task.name") }}
                </label>
                <input
                    id="new-task-name"
                    ref="nameInput"
                    v-model="name"
                    type="text"
                    class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                    :placeholder="t('projects.new_task.name_placeholder')"
                />
            </div>

            <p v-if="columnPreset" class="text-sm text-gray-500">
                {{ t("projects.new_task.column", { name: preset.columnName }) }}
            </p>

            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Listbox v-if="statusOptions.length" v-model="status" as="div">
                    <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">
                        {{ t("projects.new_task.status") }}
                    </ListboxLabel>
                    <div class="relative mt-1">
                        <ListboxButton
                            class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                        >
                            <span class="flex items-center gap-x-2 truncate">
                                <span
                                    v-if="selectedStatus?.color"
                                    class="h-2 w-2 flex-none rounded-full"
                                    :style="{ backgroundColor: selectedStatus.color }"
                                />
                                {{ selectedStatus?.name || t("projects.new_task.no_status") }}
                            </span>
                            <span
                                class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                            >
                                <ChevronUpDownIcon
                                    class="h-5 w-5 text-gray-400"
                                    aria-hidden="true"
                                />
                            </span>
                        </ListboxButton>
                        <ListboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                v-for="option in [
                                    { id: '', name: t('projects.new_task.no_status') },
                                    ...statusOptions,
                                ]"
                                :key="option.id"
                                v-slot="{ active, selected }"
                                :value="option.id"
                                as="template"
                            >
                                <li
                                    :class="[
                                        active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        'relative flex cursor-default select-none items-center gap-x-2 py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        v-if="option.color"
                                        class="h-2 w-2 flex-none rounded-full"
                                        :style="{ backgroundColor: option.color }"
                                    />
                                    <span
                                        :class="[
                                            selected ? 'font-semibold' : 'font-normal',
                                            'truncate',
                                        ]"
                                    >
                                        {{ option.name }}
                                    </span>
                                    <span
                                        v-if="selected"
                                        :class="[
                                            active ? 'text-white' : 'text-indigo-600',
                                            'absolute inset-y-0 right-0 flex items-center pr-3',
                                        ]"
                                    >
                                        <CheckIcon class="h-5 w-5" aria-hidden="true" />
                                    </span>
                                </li>
                            </ListboxOption>
                        </ListboxOptions>
                    </div>
                </Listbox>

                <Combobox v-model="assignee" as="div" nullable>
                    <ComboboxLabel class="block text-sm font-medium leading-6 text-gray-900">
                        {{ t("projects.new_task.assignee") }}
                    </ComboboxLabel>
                    <div class="relative mt-1">
                        <ComboboxInput
                            class="w-full rounded-md border-0 bg-white py-1.5 pl-3 pr-10 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            :display-value="(person) => personName(person)"
                            :placeholder="t('projects.new_task.unassigned')"
                            @change="searchMembers($event.target.value)"
                        />
                        <ComboboxButton
                            class="absolute inset-y-0 right-0 flex items-center rounded-r-md px-2 focus:outline-none"
                        >
                            <ChevronUpDownIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                        </ComboboxButton>
                        <ComboboxOptions
                            class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <li v-if="!members.length" class="px-3 py-2 text-sm text-gray-500">
                                {{ t("projects.new_task.no_members") }}
                            </li>
                            <ComboboxOption
                                v-for="person in members"
                                :key="person.id"
                                v-slot="{ active, selected }"
                                :value="person"
                                as="template"
                            >
                                <li
                                    :class="[
                                        active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        'relative flex cursor-default select-none items-center gap-x-3 py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span class="inline-block h-6 w-6 flex-none">
                                        <UserAvatar :user="person" />
                                    </span>
                                    <span :class="[selected ? 'font-semibold' : '', 'truncate']">
                                        {{ personName(person) }}
                                    </span>
                                </li>
                            </ComboboxOption>
                        </ComboboxOptions>
                    </div>
                </Combobox>

                <div>
                    <label
                        for="new-task-start"
                        class="block text-sm font-medium leading-6 text-gray-900"
                    >
                        {{ t("projects.new_task.start_date") }}
                    </label>
                    <DatePickerField
                        id="new-task-start"
                        v-model="startDate"
                        class="mt-1"
                        :placeholder="t('projects.new_task.pick_date')"
                    />
                </div>
                <div>
                    <label
                        for="new-task-due"
                        class="block text-sm font-medium leading-6 text-gray-900"
                    >
                        {{ t("projects.new_task.due_date") }}
                    </label>
                    <DatePickerField
                        id="new-task-due"
                        v-model="dueDate"
                        class="mt-1"
                        :placeholder="t('projects.new_task.pick_date')"
                    />
                </div>
            </div>
            <p v-if="datesReversed" class="text-sm text-red-600">
                {{ t("projects.new_task.dates_reversed") }}
            </p>

            <div>
                <label
                    for="new-task-description"
                    class="block text-sm font-medium leading-6 text-gray-900"
                >
                    {{ t("projects.new_task.description") }}
                </label>
                <textarea
                    id="new-task-description"
                    v-model="description"
                    rows="3"
                    class="mt-1 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                />
            </div>
        </div>
    </FormDialog>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
    Combobox,
    ComboboxButton,
    ComboboxInput,
    ComboboxLabel,
    ComboboxOption,
    ComboboxOptions,
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import FormDialog from "@/components/FormDialog.vue";
import DatePickerField from "@/components/DatePicker/DatePickerField.vue";
import UserAvatar from "@/components/UserAvatar.vue";
import useDateOperations from "@/composables/useDateOperations.js";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { useWorkspaceStore } from "@/store/workspaces";
import { extractErrorMessage } from "@/utils/errors";

const emit = defineEmits(["created"]);

const memberPage = 20;

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const { getTimestampFromDateString } = useDateOperations();

const name = ref("");
const status = ref("");
const assignee = ref(null);
const startDate = ref("");
const dueDate = ref("");
const description = ref("");
const saving = ref(false);
const statusOptions = ref([]);
const members = ref([]);
const nameInput = ref(null);

const preset = computed(() => workspaceStore.newTaskDialog || {});
const open = computed(() => workspaceStore.newTaskDialog !== null);

// A column of a field other than status stays as the Kanban column; a status
// column is just the status the dialog starts with, and can be changed.
const columnPreset = computed(() => !!preset.value.section && preset.value.section !== "status");

const selectedStatus = computed(() =>
    statusOptions.value.find((o) => String(o.id) === String(status.value)),
);
const datesReversed = computed(
    () => !!startDate.value && !!dueDate.value && startDate.value > dueDate.value,
);
const canCreate = computed(() => name.value.trim() !== "" && !datesReversed.value);

let membersRequest = 0;
let searchTimer = null;

watch(open, async (isOpen) => {
    if (!isOpen) return;

    name.value = "";
    status.value =
        preset.value.section === "status" && preset.value.singleSelect !== "0"
            ? preset.value.singleSelect || ""
            : "";
    assignee.value = null;
    startDate.value = "";
    dueDate.value = "";
    description.value = "";

    loadStatusOptions();
    searchMembers("");
    await nextTick();
    nameInput.value?.focus();
});

async function loadStatusOptions() {
    try {
        const res = await workspaceService.getTableStatusTypes(route.params.id, route.params.tid);

        statusOptions.value = res?.data?.options || [];
    } catch {
        statusOptions.value = [];
    }
}

function searchMembers(query) {
    clearTimeout(searchTimer);
    searchTimer = setTimeout(async () => {
        const seq = ++membersRequest;

        try {
            const res = await workspaceService.searchWorkspaceMembers(
                route.params.id,
                query,
                memberPage,
                0,
            );

            if (seq === membersRequest) members.value = res.data ?? [];
        } catch {
            if (seq === membersRequest) members.value = [];
        }
    }, 250);
}

function personName(person) {
    if (!person) return "";

    return [person.name, person.lastname].filter(Boolean).join(" ") || person.username || "";
}

function close() {
    workspaceStore.closeNewTaskDialog();
}

async function create(openAfter) {
    if (!canCreate.value) return;
    saving.value = true;

    const payload = {
        workspace_id: route.params.id,
        table_id: route.params.tid,
        name: name.value.trim(),
        fields: {
            status: status.value || "",
            assignee: assignee.value?.id || "",
            start_date: getTimestampFromDateString(startDate.value) || 0,
            due_date: getTimestampFromDateString(dueDate.value) || 0,
            description: description.value.trim(),
        },
    };

    if (columnPreset.value) {
        payload.section = preset.value.section;
        payload.single_select = preset.value.singleSelect;
    }

    let created;

    try {
        created = await workspaceService.createNewTask(payload);
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
        saving.value = false;

        return;
    }

    // The task exists from here on, so the dialog closes whatever follows:
    // left open, it would make a second one. Unread, the view reloads to
    // show it, as this tab is not sent its own changes.
    try {
        const res = await workspaceService.getItemForTableByID({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            task_id: created.data.id,
        });

        emit("created", { item: res.data.item, headers: res.data.headers, open: openAfter });
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
        workspaceStore.bumpGridReloadToken();
    } finally {
        saving.value = false;
        close();
    }
}
</script>
