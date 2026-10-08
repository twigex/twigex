<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto bg-white px-4 py-10 sm:px-6">
        <div class="mx-auto max-w-4xl rounded-lg bg-white shadow-sm ring-1 ring-gray-900/5">
            <div class="px-4 py-6 sm:p-8">
                <h2 class="text-base font-semibold leading-7 text-gray-900">
                    {{ t("projects.detailed_task_report.title") }}
                </h2>
                <p class="mt-1 text-sm leading-6 text-gray-600">
                    {{ t("projects.task_report_setup.description") }}
                </p>

                <div class="mt-8 space-y-8">
                    <Listbox
                        :model-value="workspacePickerValue"
                        multiple
                        as="div"
                        @update:model-value="pickWorkspaces"
                    >
                        <ListboxLabel class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("projects.task_report_setup.workspaces") }}
                        </ListboxLabel>
                        <div class="relative mt-2">
                            <ListboxButton
                                class="relative w-full cursor-default rounded-md bg-white py-1.5 pl-3 pr-10 text-left text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:outline-none focus:ring-2 focus:ring-indigo-600 sm:text-sm sm:leading-6"
                            >
                                <span class="block truncate">{{ workspaceSummary }}</span>
                                <span
                                    class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2"
                                >
                                    <ChevronUpDownIcon
                                        class="h-5 w-5 text-gray-400"
                                        aria-hidden="true"
                                    />
                                </span>
                            </ListboxButton>

                            <transition
                                leave-active-class="transition ease-in duration-100"
                                leave-from-class="opacity-100"
                                leave-to-class="opacity-0"
                            >
                                <ListboxOptions
                                    class="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                                >
                                    <ListboxOption
                                        v-for="workspace in [allWorkspacesOption, ...workspaces]"
                                        :key="workspace.id"
                                        v-slot="{ active, selected }"
                                        :value="workspace.id"
                                        as="template"
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
                                                    selected ? 'font-semibold' : 'font-normal',
                                                    'block truncate',
                                                ]"
                                            >
                                                {{ workspace.title }}
                                            </span>
                                            <span
                                                v-if="selected"
                                                :class="[
                                                    active ? 'text-white' : 'text-indigo-600',
                                                    'absolute inset-y-0 right-0 flex items-center pr-4',
                                                ]"
                                            >
                                                <CheckIcon class="h-5 w-5" aria-hidden="true" />
                                            </span>
                                        </li>
                                    </ListboxOption>
                                </ListboxOptions>
                            </transition>
                        </div>
                        <p class="mt-2 text-sm text-gray-500">
                            {{ t("projects.task_report_setup.workspaces_hint") }}
                        </p>
                    </Listbox>

                    <div>
                        <p class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("projects.top_navigation.filter") }}
                        </p>

                        <div class="mt-2 flex flex-wrap items-end gap-x-6 gap-y-3">
                            <span class="isolate inline-flex rounded-md shadow-sm">
                                <button
                                    v-for="(option, index) in presets"
                                    :key="option.value"
                                    type="button"
                                    :class="[
                                        preset === option.value && !savedFilterId
                                            ? 'z-10 bg-indigo-50 text-indigo-700 ring-indigo-300'
                                            : 'bg-white text-gray-900 ring-gray-300 hover:bg-gray-50',
                                        index === 0 ? 'rounded-l-md' : '-ml-px',
                                        index === presets.length - 1 ? 'rounded-r-md' : '',
                                        'relative inline-flex items-center px-3 py-1.5 text-sm font-medium ring-1 ring-inset focus:z-10',
                                    ]"
                                    @click="pickPreset(option.value)"
                                >
                                    {{ t(option.label) }}
                                </button>
                            </span>

                            <BaseSelect
                                v-if="savedFilters.length"
                                class="min-w-[14rem] flex-1"
                                :model-value="savedFilterId"
                                :options="savedFilterOptions"
                                @update:model-value="pickSavedFilter"
                            />
                        </div>
                        <p v-if="savedFilters.length" class="mt-2 text-sm text-gray-500">
                            {{ t("projects.task_report_setup.saved_filter_hint") }}
                        </p>

                        <div class="mt-4 rounded-md p-4 ring-1 ring-inset ring-gray-200">
                            <FilterBuilder
                                v-if="ready"
                                :key="builderKey"
                                embedded
                                :workspace-ids="selectedWorkspaceIds"
                                :title="t('projects.filter_menu.conditions')"
                                @loaded="onConditionsLoaded"
                                @change="conditions = $event"
                            />
                        </div>
                    </div>

                    <div>
                        <p class="block text-sm font-medium leading-6 text-gray-900">
                            {{ t("projects.task_sort.sort") }}
                        </p>
                        <div class="mt-2">
                            <TaskSort v-if="ready" embedded />
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="flex items-center justify-end gap-x-6 border-t border-gray-900/10 px-4 py-4 sm:px-8"
            >
                <BaseButton type="button" :is-disabled="!ready || loadedKey === ''" @click="run">
                    {{ t("projects.task_report_setup.show_report") }}
                </BaseButton>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { t } from "@/i18n/index.js";
import {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
} from "@headlessui/vue";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import FilterBuilder from "@/components/Projects/Filters/FilterBuilder.vue";
import TaskSort from "@/components/Projects/Filters/TaskSort.vue";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useTaskReportFilters } from "@/composables/projects/useTaskReportFilters";
import { filterConditionsKey } from "@/utils/projects/filterCombine";

const props = defineProps({
    workspaces: { type: Array, default: () => [] },
    savedFilters: { type: Array, default: () => [] },
    initial: { type: Object, default: () => ({}) },
});

const emit = defineEmits(["run"]);

const workspaceStore = useWorkspaceStore();
const { applyPreset, applySavedFilter } = useTaskReportFilters();

const presets = [
    { value: "all_tasks", label: "projects.menu.task_filter_menu.filter_options.all_tasks" },
    { value: "open_tasks", label: "projects.menu.task_filter_menu.filter_options.open_tasks" },
    { value: "my_tasks", label: "projects.menu.task_filter_menu.filter_options.my_tasks" },
    { value: "late_tasks", label: "projects.menu.task_filter_menu.filter_options.late_tasks" },
];

const selectedWorkspaceIds = ref([...(props.initial.workspaceIds || [])]);

// None chosen means every workspace, which the list shows as its own option.
const ALL_WORKSPACES = "__all_workspaces";
const allWorkspacesOption = computed(() => ({
    id: ALL_WORKSPACES,
    title: t.value("projects.top_navigation_reports.all_workspaces"),
}));
const workspacePickerValue = computed(() =>
    selectedWorkspaceIds.value.length ? selectedWorkspaceIds.value : [ALL_WORKSPACES],
);

function pickWorkspaces(values) {
    const hadAll = selectedWorkspaceIds.value.length === 0;

    if (values.includes(ALL_WORKSPACES) && !hadAll) {
        selectedWorkspaceIds.value = [];

        return;
    }

    selectedWorkspaceIds.value = values.filter((id) => id !== ALL_WORKSPACES);
}

const preset = ref(
    presets.some((p) => p.value === props.initial.preset) ? props.initial.preset : "",
);
const savedFilterId = ref(
    props.savedFilters.some((f) => f.id === props.initial.savedFilterId)
        ? props.initial.savedFilterId
        : "",
);

// The filter lives in the store, as the toolbar's does: a preset or saved
// filter puts its conditions there and the conditions below are loaded from
// it, to be edited before the report runs.
const ready = ref(false);
const builderKey = ref(0);
const conditions = ref({ flatFilters: [], groups: [] });
const loadedKey = ref("");
const editedAtLoad = ref(false);

const edited = computed(
    () =>
        editedAtLoad.value ||
        (loadedKey.value !== "" && filterConditionsKey(conditions.value) !== loadedKey.value),
);

function onConditionsLoaded(payload) {
    conditions.value = payload;
    loadedKey.value = filterConditionsKey(payload);
}

function reloadConditions() {
    loadedKey.value = "";
    builderKey.value++;
}

async function pickPreset(value) {
    savedFilterId.value = "";
    preset.value = value;
    editedAtLoad.value = false;
    try {
        if (await applyPreset(value)) reloadConditions();
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

function pickSavedFilter(id) {
    const filter = props.savedFilters.find((f) => f.id === id);

    if (!filter) return pickPreset(preset.value || "open_tasks");

    savedFilterId.value = id;
    preset.value = "";
    editedAtLoad.value = false;
    applySavedFilter(filter);
    reloadConditions();
}

// The last choice comes back as it was, conditions changed by hand included.
onMounted(async () => {
    const sort = props.initial.sort?.length
        ? props.initial.sort
        : [{ field: "", direction: "asc" }];

    workspaceStore.setSortOptions(JSON.parse(JSON.stringify(sort)));

    const saved = props.savedFilters.find((f) => f.id === savedFilterId.value);

    if (saved) {
        applySavedFilter(saved);
    } else {
        if (!props.initial.edited && !preset.value) preset.value = "open_tasks";
        try {
            await applyPreset(preset.value || "open_tasks");
        } catch (error) {
            useAlertStore().showError(extractErrorMessage(error));
        }
    }

    if (props.initial.edited && props.initial.conditions) {
        workspaceStore.setSavedFlatFilters(props.initial.conditions.flatFilters || []);
        workspaceStore.setSavedGroups(props.initial.conditions.groups || []);
        editedAtLoad.value = true;
    }

    ready.value = true;
});

// The workspaces arrive on their own, so a remembered choice is only checked
// against them when it is used. None chosen means all of them.
const chosenWorkspaces = computed(() => {
    const chosen = new Set(selectedWorkspaceIds.value);

    return props.workspaces.filter((w) => chosen.has(w.id));
});

const workspaceSummary = computed(() => {
    const count = chosenWorkspaces.value.length;

    if (count === 0 || count === props.workspaces.length) {
        return t.value("projects.top_navigation_reports.all_workspaces");
    }

    if (count === 1) return chosenWorkspaces.value[0].title;

    return t.value("projects.top_navigation_reports.selected_workspaces", { count });
});

const savedFilterOptions = computed(() => [
    { value: "", label: t.value("projects.task_report_setup.no_saved_filter") },
    ...props.savedFilters.map((f) => ({ value: f.id, label: f.name })),
]);

function run() {
    const ids = chosenWorkspaces.value.map((w) => w.id);

    emit("run", {
        workspaceIds: ids.length === props.workspaces.length ? [] : ids,
        preset: preset.value,
        savedFilterId: savedFilterId.value,
        conditions: conditions.value,
        edited: edited.value,
        sort: (workspaceStore.getSortOptions || []).filter((s) => s.field),
    });
}
</script>
