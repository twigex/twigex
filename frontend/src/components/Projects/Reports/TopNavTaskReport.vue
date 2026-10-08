<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="tnr-row" ref="listboxWrapper">
        <!-- Left: icon + title + description -->
        <span class="tnr-avatar">
            <ChartBarIcon class="h-5 w-5" aria-hidden="true" />
        </span>
        <div class="flex flex-col min-w-0">
            <span class="tnr-title">{{ t("projects.detailed_task_report.title") }}</span>
            <span class="tnr-desc">{{ t("projects.top_navigation_reports.description") }}</span>
        </div>

        <!-- Right: buttons pushed to the edge -->
        <div class="tnr-actions">
            <template v-if="workspaceStore.taskReportReady">
                <!-- Workspace selector -->
                <Listbox
                    as="div"
                    class="relative"
                    multiple
                    :model-value="selectedWorkspaces.map((w) => w.value)"
                    @update:model-value="onWorkspacesPicked"
                >
                    <ListboxButton
                        :as="BaseButton"
                        size="small"
                        variant="secondary"
                        :prepend-icon="BuildingOffice2Icon"
                        :append-icon="ChevronUpDownIcon"
                    >
                        <span class="tnr-btn-label--ws">
                            {{
                                selectedWorkspaces.length === 1 &&
                                selectedWorkspaces[0].value === "all_projects"
                                    ? t("projects.top_navigation_reports.all_workspaces")
                                    : selectedWorkspaces.length === 1
                                      ? selectedWorkspaces[0].label
                                      : t("projects.top_navigation_reports.selected_workspaces", {
                                            count: selectedWorkspaces.length,
                                        })
                            }}
                        </span>
                    </ListboxButton>
                    <transition
                        leave-active-class="transition ease-in duration-100"
                        leave-from-class="opacity-100"
                        leave-to-class="opacity-0"
                    >
                        <ListboxOptions
                            class="absolute right-0 z-20 mt-2 max-h-72 w-64 overflow-auto rounded-md bg-white py-1 text-base shadow-lg ring-1 ring-black/5 focus:outline-none sm:text-sm"
                        >
                            <ListboxOption
                                v-for="(workspace, i) in [ALL_PROJECTS, ...workspaceOptions]"
                                :key="workspace.value"
                                v-slot="{ active }"
                                :value="workspace.value"
                                as="template"
                            >
                                <li
                                    :class="[
                                        active ? 'bg-indigo-600 text-white' : 'text-gray-900',
                                        i === 0 ? 'border-b border-gray-100' : '',
                                        'relative cursor-default select-none py-2 pl-3 pr-9',
                                    ]"
                                >
                                    <span
                                        :class="[
                                            isWorkspaceSelected(workspace)
                                                ? 'font-medium'
                                                : 'font-normal',
                                            'block truncate',
                                        ]"
                                    >
                                        {{ workspace.label }}
                                    </span>
                                    <span
                                        v-if="isWorkspaceSelected(workspace)"
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
                    </transition>
                </Listbox>

                <FilterMenu @apply="handleApplyFilters" />

                <!-- Sort -->
                <Popover v-slot="{ close }" style="display: inline-flex; align-items: center">
                    <PopoverButton
                        :as="BaseButton"
                        size="small"
                        variant="secondary"
                        :prepend-icon="AdjustmentsHorizontalIcon"
                        class="relative focus:outline-none focus-visible:outline-none"
                    >
                        <span
                            v-if="sortActive"
                            class="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-green-500 ring-2 ring-white"
                        />
                        <span class="tnr-btn-label">{{ t("projects.task_sort.sort") }}</span>
                    </PopoverButton>
                    <transition
                        enter-active-class="transition ease-out duration-100"
                        enter-from-class="transform opacity-0 scale-95"
                        enter-to-class="transform opacity-100 scale-100"
                        leave-active-class="transition ease-in duration-75"
                        leave-from-class="transform opacity-100 scale-100"
                        leave-to-class="transform opacity-0 scale-95"
                    >
                        <teleport to="body">
                            <PopoverPanel
                                class="fixed top-[100px] left-[calc(50%-250px)] z-40 w-[500px] rounded-md bg-white shadow-lg ring-1 ring-black/5"
                            >
                                <TaskSort @close="close" />
                            </PopoverPanel>
                        </teleport>
                    </transition>
                </Popover>

                <!-- CSV -->
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="TableCellsIcon"
                    @click="downloadCSV"
                >
                    <span class="tnr-btn-label">CSV</span>
                </BaseButton>
            </template>
        </div>
        <!-- /tnr-actions -->
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed, ref, onMounted, onBeforeUnmount, watch, toRaw } from "vue";
import { useWorkspaceStore } from "@/store/workspaces";
import FilterMenu from "../Menu/FilterMenu.vue";
import TaskSort from "@/components/Projects/Filters/TaskSort.vue";
import BaseButton from "@/components/BaseButton.vue";
import { useUserStore } from "@/store/user";
import { withoutFilterOptions } from "@/utils/projects/filterCombine";
import {
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
    Popover,
    PopoverButton,
    PopoverPanel,
} from "@headlessui/vue";
import {
    ChevronUpDownIcon,
    CheckIcon,
    AdjustmentsHorizontalIcon,
    ChartBarIcon,
    TableCellsIcon,
    BuildingOffice2Icon,
} from "@heroicons/vue/24/outline";

const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();
const isListboxOpen = ref(false);
const listboxWrapper = ref(null);
const showFilterPopover = ref(false);
const sortActive = computed(() => workspaceStore.getSortActive);
const selectedFilters = ref([]);

const ALL_PROJECTS = {
    label: t.value("projects.top_navigation_reports.all_workspaces"),
    value: "all_projects",
};

const workspaceOptions = computed(() => {
    return workspaceStore.getWorkspaces.map((w) => ({
        label: w.title,
        value: w.id,
    }));
});

const selectedWorkspaces = ref([ALL_PROJECTS]);

// The Listbox reports every value chosen; the one clicked is whichever was
// added or removed, and what choosing it means is toggleWorkspaceSelection's.
function onWorkspacesPicked(values) {
    const before = selectedWorkspaces.value.map((w) => w.value);
    const clicked =
        values.find((v) => !before.includes(v)) ?? before.find((v) => !values.includes(v));
    const option = [ALL_PROJECTS, ...workspaceOptions.value].find((w) => w.value === clicked);

    if (option) toggleWorkspaceSelection(option);
}

const isWorkspaceSelected = (workspace) => {
    if (
        selectedWorkspaces.value.some((w) => w.value === "all_projects") &&
        workspace.value !== "all_projects"
    ) {
        return true;
    }

    return selectedWorkspaces.value.some((w) => w.value === workspace.value);
};

const selectedWorkspaceIds = computed(() => {
    const isAllSelected = selectedWorkspaces.value.some((w) => w.value === "all_projects");

    return isAllSelected
        ? workspaceStore.getWorkspaces.map((w) => w.id)
        : selectedWorkspaces.value.map((w) => w.value);
});

// The server streams the file and the browser saves it straight to disk, so
// an export of any size never passes through the page.
function downloadCSV() {
    const params = new URLSearchParams();

    params.set(
        "filters",
        JSON.stringify(
            withoutFilterOptions({
                groups: toRaw(workspaceStore.getSavedGroups) || [],
                flatFilters: toRaw(workspaceStore.getSavedFlatFilters) || [],
                sort: toRaw(workspaceStore.getSortOptions)?.filter((s) => s.field) || [],
            }),
        ),
    );
    const workspaceIds = selectedWorkspacesForFilter.value;

    if (workspaceIds?.length) params.set("workspace_ids", workspaceIds.join(","));
    if (userStore.getTimezone) params.set("tz", userStore.getTimezone);
    params.set("clock", userStore.getClockDisplay);
    const columns = [
        "name",
        "start_date",
        "due_date",
        "assignee",
        "status",
        "updated_at",
        "workspace_name",
        "table_name",
        "link_to_table",
    ];

    params.set(
        "labels",
        JSON.stringify(
            Object.fromEntries(columns.map((c) => [c, t.value(`projects.assigned_to_me.${c}`)])),
        ),
    );

    const a = document.createElement("a");

    a.href = `/api/workspaces/all-workspace-tasks/export?${params}`;
    a.download = "";
    document.body.appendChild(a);
    a.click();
    a.remove();
}

const toggleWorkspaceSelection = (workspace) => {
    if (workspace.value === "all_projects") {
        selectedWorkspaces.value = [ALL_PROJECTS];

        selectedWorkspacesForFilter.value = selectedWorkspaceIds.value;

        return;
    }

    // Remove "All Workspaces" if it's currently selected
    selectedWorkspaces.value = selectedWorkspaces.value.filter((w) => w.value !== "all_projects");

    const index = selectedWorkspaces.value.findIndex((w) => w.value === workspace.value);

    if (index > -1) {
        selectedWorkspaces.value.splice(index, 1);
    } else {
        selectedWorkspaces.value.push(workspace);
    }

    selectedWorkspacesForFilter.value = selectedWorkspaceIds.value;
};

const selectedWorkspacesForFilter = computed({
    get() {
        return workspaceStore.getSelectedWorkspaces;
    },
    set(value) {
        workspaceStore.setSelectedWorkspaces(value);
    },
});

// The report's setup panel chooses the workspaces too; this picker follows.
watch(selectedWorkspacesForFilter, (ids) => {
    if (
        !Array.isArray(ids) ||
        ids.length === 0 ||
        ids.length === workspaceStore.getWorkspaces.length
    ) {
        selectedWorkspaces.value = [ALL_PROJECTS];

        return;
    }

    const options = new Map(workspaceOptions.value.map((w) => [w.value, w]));

    selectedWorkspaces.value = ids.map((id) => options.get(id)).filter(Boolean);
});

// The report loads its pages from the server with the saved filters, so
// applying filters only stores them and asks for the first page again.
function handleApplyFilters(filters) {
    selectedFilters.value = filters;
    showFilterPopover.value = false;
    workspaceStore.setSavedGroups(filters.groups || []);
    workspaceStore.setSavedFlatFilters(filters.flatFilters || []);
    workspaceStore.setFilterActive(
        (filters.groups?.length || 0) + (filters.flatFilters?.length || 0) > 0,
    );
    workspaceStore.bumpGridReloadToken();
}

function handleClickOutside(event) {
    if (listboxWrapper.value && !listboxWrapper.value.contains(event.target)) {
        isListboxOpen.value = false;
    }
}

onMounted(() => {
    document.addEventListener("click", handleClickOutside);
});

onBeforeUnmount(() => {
    document.removeEventListener("click", handleClickOutside);
});
</script>

<style scoped>
.tnr-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 8px;
    width: 100%;
    container-type: inline-size;
}
@container (max-width: 680px) {
    .tnr-btn-label {
        display: none;
    }
}
.tnr-btn-label--ws {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    line-height: 20px;
}
.tnr-avatar {
    display: inline-flex;
    height: 32px;
    width: 32px;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    background-color: #3b82f6;
    color: #fff;
    flex-shrink: 0;
}
/* Tailwind UI's heading with description: text-base font-semibold over
   text-sm text-gray-500. */
.tnr-title {
    font-size: 16px;
    line-height: 24px;
    font-weight: 600;
    color: #111827;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.tnr-desc {
    font-size: 14px;
    line-height: 20px;
    color: #6b7280;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.tnr-actions {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-left: auto;
    flex-shrink: 0;
}
</style>
