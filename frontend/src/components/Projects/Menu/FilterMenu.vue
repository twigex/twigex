<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="inline-flex items-stretch">
        <Popover v-slot="{ close }" class="relative flex">
            <PopoverButton
                ref="button"
                :as="BaseButton"
                size="small"
                variant="secondary"
                :prepend-icon="FunnelIcon"
                :class="[
                    'relative focus:outline-none focus-visible:outline-none',
                    filterActive ? '!rounded-r-none' : '',
                ]"
            >
                <span
                    v-if="filterActive"
                    class="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-green-500 ring-2 ring-white"
                />
                <span class="max-w-[12rem] truncate">{{ buttonLabel }}</span>
            </PopoverButton>

            <Teleport to="body">
                <transition
                    enter-active-class="transition ease-out duration-100"
                    enter-from-class="transform opacity-0 scale-95"
                    enter-to-class="transform opacity-100 scale-100"
                    leave-active-class="transition ease-in duration-75"
                    leave-from-class="transform opacity-100 scale-100"
                    leave-to-class="transform opacity-0 scale-95"
                >
                    <PopoverPanel
                        ref="floating"
                        @vue:unmounted="creating = false"
                        :style="floatingStyles"
                        class="z-40 w-[750px] max-w-[calc(100vw-16px)] overflow-y-auto rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                    >
                        <div class="space-y-4 border-b border-gray-200 px-4 py-4">
                            <div>
                                <p
                                    class="text-xs font-semibold uppercase tracking-wider text-gray-500"
                                >
                                    {{ t("projects.filter_menu.quick") }}
                                </p>
                                <span class="mt-2 isolate inline-flex rounded-md shadow-sm">
                                    <button
                                        v-for="(option, index) in defaultFilterOptions"
                                        :key="option.value"
                                        type="button"
                                        :class="[
                                            isPresetActive(option.value)
                                                ? 'z-10 bg-indigo-50 text-indigo-700 ring-indigo-300'
                                                : 'bg-white text-gray-900 ring-gray-300 hover:bg-gray-50',
                                            index === 0 ? 'rounded-l-md' : '-ml-px',
                                            index === defaultFilterOptions.length - 1
                                                ? 'rounded-r-md'
                                                : '',
                                            'relative inline-flex items-center px-3 py-1.5 text-sm font-medium ring-1 ring-inset focus:z-10',
                                        ]"
                                        @click="pickPreset(option)"
                                    >
                                        {{ option.label }}
                                    </button>
                                </span>
                            </div>

                            <div>
                                <div class="flex items-center justify-between">
                                    <p
                                        class="text-xs font-semibold uppercase tracking-wider text-gray-500"
                                    >
                                        {{ t("projects.menu.task_filter_menu.saved_filters") }}
                                    </p>
                                    <BaseButton
                                        type="button"
                                        size="small"
                                        variant="soft"
                                        :prepend-icon="PlusIcon"
                                        @click="startNewFilter"
                                    >
                                        {{ t("projects.filter_menu.new_filter") }}
                                    </BaseButton>
                                </div>
                                <p
                                    v-if="savedFilters.length === 0"
                                    class="mt-2 text-sm text-gray-500"
                                >
                                    {{ t("projects.filter_menu.no_saved_filters") }}
                                </p>
                                <ul
                                    v-else
                                    class="mt-2 divide-y divide-gray-100 rounded-md ring-1 ring-inset ring-gray-200"
                                >
                                    <li
                                        v-for="filter in savedFilters"
                                        :key="filter.id"
                                        class="group flex items-center gap-x-3 px-3 py-2 hover:bg-gray-50"
                                    >
                                        <button
                                            type="button"
                                            class="flex min-w-0 flex-1 items-center gap-x-3 text-left text-sm"
                                            @click="pickSavedFilter(filter)"
                                        >
                                            <span
                                                :class="[
                                                    isSavedActive(filter)
                                                        ? 'border-indigo-600 bg-indigo-600'
                                                        : 'border-gray-300 bg-white',
                                                    'flex h-4 w-4 flex-none items-center justify-center rounded-full border',
                                                ]"
                                                aria-hidden="true"
                                            >
                                                <span
                                                    v-if="isSavedActive(filter)"
                                                    class="h-1.5 w-1.5 rounded-full bg-white"
                                                />
                                            </span>
                                            <span
                                                :class="[
                                                    isSavedActive(filter)
                                                        ? 'font-medium text-gray-900'
                                                        : 'text-gray-700',
                                                    'truncate',
                                                ]"
                                            >
                                                {{ filter.name }}
                                            </span>
                                            <span
                                                v-if="filter.is_active"
                                                class="inline-flex flex-none items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                            >
                                                {{ t("projects.filter_menu.default") }}
                                            </span>
                                            <span
                                                v-if="filter.is_private"
                                                class="flex-none text-xs text-gray-400"
                                            >
                                                {{ t("projects.menu.task_filter_menu.private") }}
                                            </span>
                                        </button>
                                        <button
                                            v-if="canSetDefault && !filter.is_active"
                                            type="button"
                                            class="flex-none rounded px-2 py-1 text-xs font-medium text-indigo-600 opacity-0 hover:bg-indigo-50 focus:opacity-100 group-hover:opacity-100"
                                            @click="toggleDefault(filter)"
                                        >
                                            {{ t("projects.filter_menu.make_default") }}
                                        </button>
                                        <button
                                            v-if="canManage(filter)"
                                            type="button"
                                            class="flex-none rounded p-1 text-gray-400 opacity-0 hover:bg-gray-100 hover:text-gray-600 focus:opacity-100 group-hover:opacity-100"
                                            @click="
                                                toggleItemMenuPopover(
                                                    $event.currentTarget,
                                                    filter,
                                                    close,
                                                )
                                            "
                                        >
                                            <span class="sr-only">{{
                                                t("projects.folder_view.actions")
                                            }}</span>
                                            <EllipsisHorizontalIcon
                                                class="h-4 w-4"
                                                aria-hidden="true"
                                            />
                                        </button>
                                    </li>
                                </ul>
                            </div>
                        </div>

                        <FilterBuilder
                            :key="builderKey"
                            :creating="creating"
                            :title="
                                creating
                                    ? t('projects.filter_menu.new_filter')
                                    : t('projects.filter_menu.conditions')
                            "
                            @apply="applyConditions($event, close)"
                            @close="close"
                        />
                    </PopoverPanel>
                </transition>
            </Teleport>
        </Popover>

        <BaseButton
            v-if="filterActive"
            type="button"
            size="small"
            variant="secondary"
            class="-ml-px !rounded-l-none focus:outline-none focus-visible:outline-none"
            :prepend-icon="XMarkIcon"
            :aria-label="t('projects.filter_menu.clear')"
            :title="t('projects.filter_menu.clear')"
            @click="clearFilter"
        />

        <Teleport to="body">
            <SavedFilterMenu
                :visible="itemPopoverVisible"
                :anchor="itemPopoverAnchor"
                :filter="activePopoverItem"
                :close-parent-menu="filterMenuClose"
                @close="handlePopoverClose"
                @remove-default="toggleDefault"
            />
        </Teleport>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, shallowRef, computed, watch } from "vue";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { FunnelIcon, PlusIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { EllipsisHorizontalIcon } from "@heroicons/vue/24/outline";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useRoute } from "vue-router";
import SavedFilterMenu from "@/components/Projects/Menu/SavedFilterMenu.vue";
import FilterBuilder from "@/components/Projects/Filters/FilterBuilder.vue";
import BaseButton from "@/components/BaseButton.vue";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { PRESET_LABELS, useTaskReportFilters } from "@/composables/projects/useTaskReportFilters";
import { useTableFilters } from "@/composables/projects/useTableFilters";
import { filterConditionsKey } from "@/utils/projects/filterCombine";

const emit = defineEmits(["apply"]);

const button = ref(null);
const { floating, floatingStyles } = useAnchoredPopup({
    placement: "bottom-start",
    fitHeight: true,
    anchor: button,
});

// Remounting the conditions makes them read what a preset or saved filter
// has just put in the store.
const builderKey = ref(0);

// A new filter starts from no conditions, without touching what is applied
// until it is applied or saved.
const creating = ref(false);
let creatingFrom = null;

function startNewFilter() {
    creating.value = true;
    creatingFrom = fullFilter.value?.id ?? null;
    builderKey.value++;
}

const tableData = computed({
    get() {
        return route.name === "detailed-task-report"
            ? workspaceStore.getTaskReportData
            : workspaceStore.getTableData;
    },
    set(value) {
        if (route.name === "detailed-task-report") {
            workspaceStore.setTaskReportData(value);
        } else {
            workspaceStore.setTableData(value);
        }
    },
});

const selectedFilters = ref([]);
const { applyPreset, applySavedFilter } = useTaskReportFilters();
const { loadTablePreset, storeTablePreset, setViewDefaultFilter } = useTableFilters();

// The report loads its pages from the server with the saved filters, which
// selectSavedFilter has just stored, so this only asks for the first page.
function handleApplyFiltersForAllProjects({ groups = [], flatFilters = [] }) {
    selectedFilters.value = { groups, flatFilters };
    workspaceStore.bumpGridReloadToken();
}

// The report's presets: the filters come from the shared composable, which
// the report's own setup panel uses too.
async function reportDefFilter(option) {
    pendingOpenTasksRoute = null;
    if (!["open_tasks", "my_tasks", "all_tasks", "late_tasks"].includes(option.value)) return;
    try {
        if (!(await applyPreset(option.value))) return;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));

        return;
    }

    selectedLabel.value = workspaceStore.getFastFilterLabel || selectedLabel.value;
    workspaceStore.bumpGridReloadToken();
}

const filterActive = computed({
    get() {
        return workspaceStore.getFilterActive;
    },
    set(value) {
        workspaceStore.setFilterActive(value);
    },
});

const quickFilter = computed({
    get() {
        return workspaceStore.getQuickFilterActive;
    },
    set(value) {
        workspaceStore.setQuickFilterActive(value);
    },
});

const viewFilters = computed({
    get() {
        return workspaceStore.getFilters;
    },
    set(value) {
        workspaceStore.setFilters(value);
    },
});

const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();
const route = useRoute();

const itemPopoverVisible = ref(false);
const itemPopoverAnchor = shallowRef(null);
const activePopoverItem = ref(null);
const filterMenuClose = ref(null);

const canManage = (filter) => filter.is_private || rolesStore.hasPermissionToUpdateWorkspaceView;

function handlePopoverClose() {
    itemPopoverVisible.value = false;
    activePopoverItem.value = null;
    filterMenuClose.value = null;
}

function toggleItemMenuPopover(anchorEl, filter, closePanel) {
    if (itemPopoverVisible.value && activePopoverItem.value === filter) {
        itemPopoverVisible.value = false;
        activePopoverItem.value = null;
    } else {
        filterMenuClose.value = closePanel;
        itemPopoverVisible.value = true;
        activePopoverItem.value = filter;
        itemPopoverAnchor.value = anchorEl;
    }
}

const selectedLabel = ref(t.value("projects.menu.task_filter_menu.filter_options.all_tasks"));

const buttonLabel = computed(() =>
    filterActive.value ? selectedLabel.value : t.value("projects.top_navigation.filter"),
);

// Must be declared before the watch with { immediate: true } to avoid TDZ error
const PAGINATED_ROUTES = ["grid-view", "kanban-view", "calendar-view", "gantt-view"];

let pendingOpenTasksRoute = null;

watch(
    () => [route.name, route.params.tid, route.params.fid],
    async ([name]) => {
        if (
            name === "detailed-task-report" ||
            name === "grid-view" ||
            name === "gantt-view" ||
            name === "assigned-to-me"
        ) {
            // Restore persisted fast filter label
            const savedLabel = workspaceStore.getFastFilterLabel;

            selectedLabel.value =
                savedLabel || t.value("projects.menu.task_filter_menu.filter_options.all_tasks");

            if (name === "assigned-to-me") return;

            // The grid applies its default filter itself as it loads.
            if (name === "grid-view") {
                pendingOpenTasksRoute = null;
            } else if (PAGINATED_ROUTES.includes(name)) {
                // Always set. Fires on every navigation, including same-name table/view switches
                pendingOpenTasksRoute = name;
            } else if (name === "detailed-task-report") {
                // The report runs from its own setup panel, with the filter chosen there.
                pendingOpenTasksRoute = null;
            } else {
                pendingOpenTasksRoute = null;
            }
        } else {
            pendingOpenTasksRoute = null;
        }
    },
    { immediate: true },
);

// Applies the view's saved filter, or open tasks, once per navigation.
async function applyDefaultFilter() {
    if (!pendingOpenTasksRoute) return;
    // DTR always resets to open_tasks, skip the filterActive check.
    // The view already has a saved default filter, don't override it.
    if (workspaceStore.getFilterActive && pendingOpenTasksRoute !== "detailed-task-report") {
        pendingOpenTasksRoute = null;
        const activeSaved = savedFilters.value.find((f) => f.is_active);

        if (activeSaved) {
            selectedLabel.value = activeSaved.name;
        } else {
            selectedLabel.value = t.value(
                "projects.menu.task_filter_menu.filter_options.all_tasks",
            );
        }

        workspaceStore.bumpGridReloadToken();

        return;
    }

    const targetRoute = pendingOpenTasksRoute;

    pendingOpenTasksRoute = null;
    if (targetRoute === "detailed-task-report") {
        await reportDefFilter({ value: "open_tasks" });
    } else if (PAGINATED_ROUTES.includes(targetRoute)) {
        await selectOptionFast({ value: "open_tasks" });
    } else {
        await selectOption({ value: "open_tasks" });
    }
}

watch(
    () => tableData.value,
    async (newData, oldData) => {
        const prevLen = Array.isArray(oldData) ? oldData.length : 0;
        const newLen = Array.isArray(newData) ? newData.length : 0;

        if (newLen > 0 && prevLen === 0) await applyDefaultFilter();
    },
);

const savedFilters = computed(() => {
    if (route.name === "detailed-task-report") {
        const filtered = (viewFilters.value || []).filter(
            (f) => f.table_id === "detailed-task-report" && f.view_id === "detailed-task-report",
        );

        return filtered;
    }

    const currentTableId = route.params.tid;
    const currentViewId = route.params.fid;

    return (viewFilters.value || []).filter(
        (f) => f.table_id === currentTableId && f.view_id === currentViewId,
    );
});

watch(
    () => workspaceStore.getFilterActive,
    (newVal) => {
        if (!newVal && route.name !== "detailed-task-report") {
            selectedLabel.value = t.value(
                "projects.menu.task_filter_menu.filter_options.all_tasks",
            );
        }
    },
);

// Starring a filter changes the list without applying the filter, so the
// label stays as it is while that happens.
let settingDefault = false;

const canSetDefault = computed(
    () => route.name !== "detailed-task-report" && rolesStore.hasPermissionToUpdateWorkspaceView,
);

async function toggleDefault(filter) {
    settingDefault = true;
    try {
        await setViewDefaultFilter(filter, !filter.is_active, {
            workspaceId: route.params.id,
            tableId: route.params.tid,
            viewId: route.params.fid,
        });
    } finally {
        settingDefault = false;
    }
}

watch(
    () => savedFilters.value,
    (filters) => {
        // The report has no default filter; it applies what its setup chose.
        if (settingDefault || route.name === "detailed-task-report") return;
        const activeSaved = filters.find((f) => f.is_active);

        if (activeSaved) {
            selectedLabel.value = activeSaved.name;
        }
    },
    { deep: true, immediate: true },
);

// A table preset is loaded before it is applied; only the latest choice, on
// the table it was made on, may be.
let tableChoice = 0;

async function selectOptionFast(option) {
    const mine = ++tableChoice;
    const labelKey = PRESET_LABELS[option.value];

    if (labelKey) {
        const tableId = route.params.tid;
        let preset;

        try {
            preset = await loadTablePreset(option.value, route.params.id, tableId);
        } catch (error) {
            useAlertStore().showError(extractErrorMessage(error));

            return;
        }

        if (mine !== tableChoice || route.params.tid !== tableId) return;

        selectedLabel.value = t.value(labelKey);
        storeTablePreset(preset);
        if (option.value === "all_tasks") quickFilter.value = true;
        workspaceStore.bumpGridReloadToken();

        return;
    }

    // default_filter
    workspaceStore.setSavedFlatFilters(workspaceStore.getDefaultFlatFilters || []);
    workspaceStore.setSavedGroups(workspaceStore.getDefaultGroupFilters || []);
    workspaceStore.setFilterActive(true);
    workspaceStore.bumpGridReloadToken();
}

async function selectOption(option) {
    pendingOpenTasksRoute = null;
    if (route.name === "detailed-task-report") {
        await reportDefFilter(option);

        return;
    }

    if (!route.params.tid) return;
    await selectOptionFast(option);
}

const defaultFilterOptions = [
    {
        label: t.value("projects.menu.task_filter_menu.filter_options.all_tasks"),
        value: "all_tasks",
    },
    {
        label: t.value("projects.menu.task_filter_menu.filter_options.open_tasks"),
        value: "open_tasks",
    },
    {
        label: t.value("projects.menu.task_filter_menu.filter_options.my_tasks"),
        value: "my_tasks",
    },
    {
        label: t.value("projects.menu.task_filter_menu.filter_options.late_tasks"),
        value: "late_tasks",
    },
];

const fullFilter = computed({
    get() {
        return workspaceStore.getFilter;
    },
    set(value) {
        workspaceStore.setFilter(value);
    },
});

async function selectSavedFilter(filter) {
    tableChoice++;
    selectedLabel.value = filter.name;

    if (route.name === "detailed-task-report") {
        applySavedFilter(filter);
        await handleApplyFiltersForAllProjects({
            groups: filter.filters?.groups || [],
            flatFilters: filter.filters?.flatFilters || [],
        });
    } else {
        fullFilter.value = filter;
        workspaceStore.setSavedFlatFilters(filter.filters?.flatFilters || []);
        workspaceStore.setSavedGroups(filter.filters?.groups || []);
        workspaceStore.setFilterActive(true);
        workspaceStore.setDefaultFlatFilters(filter.filters?.flatFilters || []);
        workspaceStore.setDefaultGroupFilters(filter.filters?.groups || []);
        workspaceStore.setFastFilter("saved_filter", filter.name);
        workspaceStore.bumpGridReloadToken();
    }
}

function isPresetActive(value) {
    if (fullFilter.value?.id) return false;
    const option = workspaceStore.fastFilterOption || "all_tasks";

    return option === value;
}

function isSavedActive(filter) {
    return fullFilter.value?.id === filter.id;
}

async function pickPreset(option) {
    creating.value = false;
    await selectOption(option);
    builderKey.value++;
}

async function pickSavedFilter(filter) {
    creating.value = false;
    await selectSavedFilter(filter);
    builderKey.value++;
}

// Conditions set by hand name themselves by how many there are, unless they
// are a saved filter's, which keeps its name, marked when they were changed.
function applyConditions(payload, closePanel) {
    tableChoice++;
    // Applied without being saved, a new filter is not the one it started from.
    if (creating.value && fullFilter.value?.id === creatingFrom) fullFilter.value = null;

    const count =
        (payload.flatFilters?.length || 0) +
        (payload.groups || []).reduce((sum, group) => sum + (group.filters?.length || 0), 0);

    let option = "custom";
    let label = t.value("projects.filter_menu.conditions_count", { count });

    if (count === 0) {
        option = "all_tasks";
        label = t.value(PRESET_LABELS.all_tasks);
        fullFilter.value = null;
    } else if (fullFilter.value?.name) {
        option = "saved_filter";
        label =
            filterConditionsKey(payload) === filterConditionsKey(fullFilter.value.filters)
                ? fullFilter.value.name
                : t.value("projects.filter_menu.edited", { name: fullFilter.value.name });
    }

    selectedLabel.value = label;
    workspaceStore.setFastFilter(option, label);

    creating.value = false;
    emit("apply", payload);
    closePanel();
}

async function clearFilter() {
    creating.value = false;
    await selectOption({ value: "all_tasks" });
    builderKey.value++;
}

// The report's setup panel and the grid's own load pick a preset or saved
// filter too; the label here follows them.
watch(
    () => workspaceStore.fastFilterVersion,
    () => {
        const label = workspaceStore.getFastFilterLabel;

        if (label && (route.name === "detailed-task-report" || route.name === "grid-view")) {
            selectedLabel.value = label;
        }
    },
);
</script>
