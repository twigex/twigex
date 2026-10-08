<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full">
        <template v-if="!reportStarted">
            <TaskReportSetup
                v-if="setupReady"
                :workspaces="workspaceStore.getWorkspaces"
                :saved-filters="reportFilters"
                :initial="setupChoice"
                @run="startReport"
            />
        </template>
        <div
            v-else-if="awaitingFilterLoad"
            class="flex h-full items-center justify-center bg-white"
        >
            <div class="flex flex-col items-center gap-y-3">
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>
        </div>
        <div v-else-if="tableLoaded && tableData.length === 0" class="h-full">
            <div
                class="flex h-full w-full flex-col items-center justify-center bg-white text-center"
            >
                <MagnifyingGlassIcon class="mx-auto h-16 w-16 text-gray-300" aria-hidden="true" />
                <h2 class="mt-4 text-lg font-semibold text-gray-900">
                    {{ t("projects.detailed_task_report.no_tasks_title") }}
                </h2>
                <p class="mt-2 max-w-sm text-sm text-gray-500">
                    {{ t("projects.detailed_task_report.no_tasks_description") }}
                </p>
                <BaseButton
                    type="button"
                    variant="secondary"
                    class="mt-6 !px-4 !font-semibold"
                    @click="changeSettings"
                >
                    {{ t("projects.task_report_setup.change_settings") }}
                </BaseButton>
            </div>
        </div>
        <div v-else-if="tableLoaded" class="flex h-full flex-col">
            <div class="table-container" :data-table-id="tableUniqueId">
                <div class="grid-container sticky top-0 z-[2]">
                    <!-- Table header -->
                    <div class="header-row">
                        <ReportHeaderRow
                            :headers="tableHeaders"
                            :column-widths="columnWidths"
                            @resize-start="startResize"
                        />
                    </div>
                </div>

                <div class="table-content">
                    <div class="grid-container">
                        <div
                            class="grid-row"
                            v-for="item in pageSlice"
                            :key="item.id"
                            v-memo="[_dtrRowVersion, userStore.knownUserCount, item.id]"
                        >
                            <div
                                v-for="(header, colIndex) in tableHeaders"
                                :key="colIndex"
                                :class="[
                                    'grid-item-text',
                                    header.name === 'id' || header.name === 'link_to_table'
                                        ? 'hidden-column'
                                        : '',
                                ]"
                                :style="{
                                    width: columnWidths[colIndex] + 'px',
                                }"
                                :data-column-index="colIndex"
                            >
                                <div v-if="tableHeaders[colIndex].name !== 'id'">
                                    <div
                                        class="clickable-area"
                                        style="display: flex; align-items: center; height: 100%"
                                    >
                                        <div v-if="tableHeaders[colIndex].name === 'assignee'">
                                            <template v-if="item[header.name]">
                                                <UserAvatarWithText
                                                    class="gap-2"
                                                    :user-id="item[header.name]"
                                                    avatar-class="h-5 w-5 shrink-0"
                                                    text-class="whitespace-nowrap text-[13px]"
                                                />
                                            </template>
                                            <template v-else>
                                                <UserPlusIcon
                                                    class="h-4 w-4 text-gray-400"
                                                    aria-hidden="true"
                                                />
                                            </template>
                                        </div>

                                        <div
                                            v-else-if="tableHeaders[colIndex].name === 'status'"
                                            class="flex h-full min-w-0 items-center"
                                        >
                                            <span
                                                v-if="item[header.name]?.name"
                                                class="inline-flex max-w-full items-center gap-x-1.5 rounded-full bg-white px-2 py-0.5 text-xs font-medium text-gray-700 ring-1 ring-inset ring-gray-200"
                                            >
                                                <span
                                                    class="h-2 w-2 flex-none rounded-full"
                                                    :class="getStatusStyle(item[header.name]).class"
                                                    :style="getStatusStyle(item[header.name]).style"
                                                    aria-hidden="true"
                                                />
                                                <span class="truncate">{{
                                                    item[header.name].name
                                                }}</span>
                                            </span>

                                            <span
                                                v-else
                                                class="ml-1.5 block h-3 w-3 rounded-full border border-dashed border-gray-400"
                                                aria-hidden="true"
                                            />
                                        </div>

                                        <div
                                            v-else-if="
                                                tableHeaders[colIndex].name === 'link_to_table'
                                            "
                                            class="flex min-w-0 items-center"
                                        >
                                            <button
                                                type="button"
                                                class="inline-flex min-w-0 items-center gap-x-1 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-700 hover:bg-gray-200"
                                                @click="
                                                    goToTable(
                                                        item[header.name].workspace_id,
                                                        item[header.name].table_id,
                                                        item,
                                                        item[header.name],
                                                    )
                                                "
                                            >
                                                <ArrowTopRightOnSquareIcon
                                                    class="h-3 w-3 flex-none text-gray-500"
                                                    aria-hidden="true"
                                                />
                                                <span class="truncate">{{
                                                    t("projects.assigned_to_me.link_to_table")
                                                }}</span>
                                            </button>
                                        </div>
                                        <div
                                            v-else-if="header.name === 'name'"
                                            class="text-left alignLeft text-sm font-medium leading-6 text-gray-600"
                                            style="margin-right: 6px; font-size: 13px"
                                        >
                                            <div
                                                :style="{
                                                    width: columnWidths[colIndex] - 6 + 'px',
                                                    paddingRight: '6px',
                                                    paddingLeft:
                                                        hasChildren(item.id) ||
                                                        item._is_subtask ||
                                                        (item.parent_task_id &&
                                                            item.parent_task_id !== item.id)
                                                            ? (item._indent || 0) * 20 + 'px'
                                                            : '0px',
                                                    overflow: 'hidden',
                                                    whiteSpace: 'nowrap',
                                                    textOverflow: 'ellipsis',
                                                }"
                                                class="flex items-center gap-2"
                                                :title="item[header.name]"
                                            >
                                                <span
                                                    class="w-4 h-4 flex-shrink-0 flex items-center justify-center"
                                                >
                                                    <component
                                                        v-if="hasChildren(item.id)"
                                                        :is="
                                                            expandedTasks.has(item.id)
                                                                ? ChevronDownIcon
                                                                : ChevronRightIcon
                                                        "
                                                        class="w-4 h-4 text-gray-500 cursor-pointer transition-transform"
                                                        @click.stop="toggleExpand(item.id)"
                                                    />
                                                </span>

                                                <span
                                                    v-if="
                                                        item.parent_task_id &&
                                                        item.parent_task_id !== item.id
                                                    "
                                                    class="text-xs rounded"
                                                >
                                                    <Square2StackIcon class="w-4 h-4" />
                                                </span>

                                                <div
                                                    class="truncate text-sm text-gray-800 font-medium flex-1 cursor-pointer hover:text-indigo-600 hover:underline"
                                                    @click="
                                                        showFullView(
                                                            item,
                                                            item.link_to_table.table_id,
                                                            item.link_to_table.workspace_id,
                                                            item.link_to_table.view_id,
                                                        )
                                                    "
                                                >
                                                    {{ item[header.name] || " " }}
                                                </div>

                                                <div
                                                    v-if="getSubtaskCount(item.id) > 0"
                                                    class="flex items-center gap-1 text-gray-400 text-xs flex-shrink-0"
                                                    :title="`${getSubtaskCount(item.id)} ${t('projects.assigned_to_me.subtasks')}`"
                                                >
                                                    <Square2StackIcon class="w-4 h-4" />
                                                    <span>{{ getSubtaskCount(item.id) }}</span>
                                                </div>
                                            </div>
                                        </div>

                                        <div
                                            v-else
                                            class="text-left alignLeft text-sm font-medium leading-6 text-gray-600"
                                            style="margin-right: 6px; font-size: 13px"
                                        >
                                            <div
                                                :style="{
                                                    width: columnWidths[colIndex] - 6 + 'px',
                                                    paddingRight: '6px',
                                                    paddingLeft: '0px',
                                                    overflow: 'hidden',
                                                    whiteSpace: 'nowrap',
                                                    textOverflow: 'ellipsis',
                                                }"
                                                :title="item[header.name]"
                                            >
                                                <template
                                                    v-if="
                                                        [
                                                            'start_date',
                                                            'due_date',
                                                            'created_at',
                                                            'updated_at',
                                                            'deleted_at',
                                                        ].includes(header.name)
                                                    "
                                                >
                                                    <span
                                                        v-if="!item[header.name]"
                                                        class="text-gray-400"
                                                        >–</span
                                                    >
                                                    <template v-else>{{
                                                        getDate(item[header.name])
                                                    }}</template>
                                                </template>

                                                <span
                                                    v-else-if="isBlank(item[header.name])"
                                                    class="text-gray-400"
                                                    >–</span
                                                >
                                                <template v-else>{{ item[header.name] }}</template>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <div
                class="sticky-row flex items-center justify-between bg-white py-2 pl-4 pr-4 border-t border-gray-200"
            >
                <div class="flex items-center gap-x-4">
                    <span class="text-sm text-gray-700 font-medium">
                        {{ t("projects.assigned_to_me.total_tasks") }}
                        {{ rootTotalRows || totalRows }}
                    </span>
                    <button
                        type="button"
                        class="rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                        @click="changeSettings"
                    >
                        {{ t("projects.task_report_setup.change_settings") }}
                    </button>
                </div>

                <BasePagination
                    v-if="totalPages > 1"
                    class="!flex-none !border-0 !bg-transparent !p-0"
                    :current-page="currentPage"
                    :total-items="totalRows"
                    :items-per-page="itemsPerPage"
                    :show-summary="false"
                    @update:current-page="changePage"
                />
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed, onMounted, watch, toRaw, ref, nextTick } from "vue";
import TaskReportSetup from "@/components/Projects/Reports/TaskReportSetup.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import BaseButton from "@/components/BaseButton.vue";
import BasePagination from "@/components/BasePagination.vue";
import { PRESET_LABELS } from "@/composables/projects/useTaskReportFilters";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { useUserStore } from "@/store/user";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { isBlank } from "@/utils/projects/rows";
import { useColumnResize } from "@/composables/projects/useColumnResize";
import { flattenTaskTree } from "@/utils/projects/tree";
import { useRouter, useRoute } from "vue-router";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { withoutFilterOptions } from "@/utils/projects/filterCombine";
import { getStatusStyle } from "@/utils/projects/cells";
import { useSubtaskCounts } from "@/composables/projects/useSubtaskCounts";
import ReportHeaderRow from "@/components/Projects/Reports/ReportHeaderRow.vue";
import { useOpenTask } from "@/composables/projects/useOpenTask";
import { useLatestRequest } from "@/composables/useLatestRequest";
import {
    ArrowTopRightOnSquareIcon,
    UserPlusIcon,
    MagnifyingGlassIcon,
} from "@heroicons/vue/24/outline";

import { ChevronRightIcon, ChevronDownIcon, Square2StackIcon } from "@heroicons/vue/20/solid";

const expandedTasks = ref(new Set());
const { getDate } = useDateOperations();
const route = useRoute();
const router = useRouter();

const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();

const tableHeaders = computed({
    get() {
        return workspaceStore.getTableHeaders;
    },
    set(value) {
        workspaceStore.setTableHeaders(value);
    },
});

const columnWidths = computed({
    get() {
        return workspaceStore.getColumnsWidths;
    },
    set(value) {
        workspaceStore.setColumnsWidths(value);
    },
});

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

const itemsPerPage = 100;
const currentPage = ref(1);
const totalRows = ref(0);
const rootTotalRows = ref(0);

const totalPages = computed(() => Math.max(1, Math.ceil(totalRows.value / itemsPerPage)));

const visibleData = computed(() => flattenTaskTree(tableData.value, expandedTasks.value));

const toggleExpand = (itemId) => {
    const id = String(itemId);
    const next = new Set(expandedTasks.value);

    if (next.has(id)) next.delete(id);
    else next.add(id);
    expandedTasks.value = next;
    // Bump row version so v-memo re-evaluates expand icons and subtask rows.
    _dtrRowVersion.value++;
    // Extend renderedCount to cover newly visible subtask rows.
    _dtrRenderedCount.value = visibleData.value.length;
};

const workspaces = ref([]);
// Use dedicated taskReportData store key. Must NOT share workspaceStore.data with GridView.
// ProjectsView overwrites workspaceStore.setTableData with 46k workspace rows on navigation.
const tableData = computed({
    get() {
        return workspaceStore.getTaskReportData;
    },
    set(value) {
        workspaceStore.setTaskReportData(value);
    },
});
const { hasChildren, getSubtaskCount } = useSubtaskCounts(tableData);

const RENDER_BATCH_DTR = 25;
const _dtrRowVersion = ref(0);
const _dtrRenderedCount = ref(RENDER_BATCH_DTR);
const awaitingFilterLoad = ref(false);

watch(
    tableData,
    () => {
        _dtrRowVersion.value++;
        _dtrRenderedCount.value = RENDER_BATCH_DTR;
        const addBatch = () => {
            _dtrRenderedCount.value = Math.min(
                _dtrRenderedCount.value + RENDER_BATCH_DTR,
                visibleData.value.length,
            );
            if (_dtrRenderedCount.value < visibleData.value.length) requestAnimationFrame(addBatch);
        };

        requestAnimationFrame(addBatch);
        // DO NOT clear awaitingFilterLoad here. fetchPage clears it when filtered data arrives.
        // tableData watch fires in the same sync batch as awaitingFilterLoad=true, so
        // checking it here would immediately clear it before DOM renders the spinner.
    },
    { flush: "pre" },
);

// The task panel saves into the grid's rows; the report's are its own.
watch(
    () => workspaceStore.savedTaskField,
    (saved) => {
        const row = saved && tableData.value?.find((r) => String(r.id) === String(saved.id));

        if (!row) return;

        row[saved.field] = saved.value;
        _dtrRowVersion.value++;
    },
);

const tableUniqueId = `task-report-${Math.random().toString(36).slice(2, 11)}`;

// The rows are memoised, so they draw again at the new widths once the
// drag ends.
const { startResize } = useColumnResize({
    headers: tableHeaders,
    widths: columnWidths,
    tableId: tableUniqueId,
    onResized: () => _dtrRowVersion.value++,
});

const changePage = (page) => {
    const workspaceIds = selectedWorkspacesForFilter.value?.length
        ? selectedWorkspacesForFilter.value
        : null;

    fetchPage(page, workspaceIds, false);
};

const { openTask } = useOpenTask();

async function showFullView(item, table_id, workspace_id) {
    if (!(await openTask({ workspaceId: workspace_id, tableId: table_id, taskId: item.id })))
        return;

    router.push({
        name: "grid-view",
        params: {
            id: workspace_id,
            tid: table_id,
            fid: item.link_to_table.view_id || "",
        },
        query: { task: item.id },
    });
}

function goToTable(workspace_id, table_id, item) {
    showFullView(item, table_id, workspace_id);
    resetState(true);
}

const resetState = (clearHeaders = false) => {
    workspaceStore.clearTableData();
    workspaces.value = [];
    tableData.value = [];
    selectedView.value = null;
    tableLoaded.value = false;
    // Only clear headers on initial load, keep them for sort/filter reloads.
    // so the TaskSort panel and column structure stay intact
    if (clearHeaders) {
        workspaceStore.setTableHeaders([]);
        workspaceStore.setColumnsWidths([]);
        tableHeaders.value = [];
        columnWidths.value = [];

        // The filter and sort in the store are the last table's until the
        // setup panel chooses the report's.
        workspaceStore.setFilter(null);
        workspaceStore.setSavedFlatFilters([]);
        workspaceStore.setSavedGroups([]);
        workspaceStore.setDefaultFlatFilters([]);
        workspaceStore.setDefaultGroupFilters([]);
        workspaceStore.setFilterActive(false);
        workspaceStore.setFastFilter("", "");
        workspaceStore.setSortOptions([{ field: "", direction: "asc" }]);
        workspaceStore.setSortActive(false);
        workspaceStore.setTaskReportReady(false);
    }
};

const isLoading = ref(false);
const tableLoaded = ref(false);
let pendingFilterRetry = false; // set when token bumps while loadData is still running

// The report searches every task table, so it runs only when asked, from
// its setup panel, never just by being opened. The choice made there is
// remembered in this browser to fill the panel in next time.
const SETUP_KEY = "twigex.taskReportSetup";
const reportStarted = ref(false);
const setupReady = ref(false);
const setupChoice = ref(readSetupChoice());

function readSetupChoice() {
    try {
        return JSON.parse(localStorage.getItem(SETUP_KEY) || "{}") || {};
    } catch {
        return {};
    }
}

function rememberSetupChoice(choice) {
    setupChoice.value = choice;
    try {
        localStorage.setItem(SETUP_KEY, JSON.stringify(choice));
    } catch {
        // Without storage the panel just starts empty next time.
    }
}

// The setup has already put its preset or saved filter in the store, and
// its conditions may have been changed since, so they are applied as they
// stand, named for what they came from.
function applySetupFilter({ preset, savedFilterId, conditions, edited }) {
    const flatFilters = conditions?.flatFilters || [];
    const groups = conditions?.groups || [];
    const count = flatFilters.length + groups.reduce((sum, g) => sum + (g.filters?.length || 0), 0);

    workspaceStore.setSavedFlatFilters(flatFilters);
    workspaceStore.setSavedGroups(groups);
    workspaceStore.setDefaultFlatFilters(flatFilters);
    workspaceStore.setDefaultGroupFilters(groups);
    workspaceStore.setFilterActive(count > 0);

    const saved = savedFilterId && reportFilters.value.find((f) => f.id === savedFilterId);

    if (saved) {
        workspaceStore.setFilter(saved);
        workspaceStore.setFastFilter(
            "saved_filter",
            edited ? t.value("projects.filter_menu.edited", { name: saved.name }) : saved.name,
        );

        return;
    }

    workspaceStore.setFilter(null);
    if (preset && !edited) {
        workspaceStore.setFastFilter(preset, t.value(PRESET_LABELS[preset]));
    } else if (count === 0) {
        workspaceStore.setFastFilter("all_tasks", t.value(PRESET_LABELS.all_tasks));
    } else {
        workspaceStore.setFastFilter(
            "custom",
            t.value("projects.filter_menu.conditions_count", { count }),
        );
    }
}

async function startReport(choice) {
    rememberSetupChoice({ ...choice, conditions: withoutFilterOptions(choice.conditions || {}) });
    applySetupFilter(choice);
    workspaceStore.setSortActive((choice.sort || []).length > 0);

    // The workspace watcher sees this change before the report counts as
    // started, so it does not load the report a second time.
    selectedWorkspacesForFilter.value = choice.workspaceIds;
    await nextTick();

    reportStarted.value = true;
    tableLoaded.value = true;
    awaitingFilterLoad.value = true;
    await fetchPage(1, choice.workspaceIds.length ? choice.workspaceIds : null);
    workspaceStore.setTaskReportReady(true);
}

function changeSettings() {
    reportStarted.value = false;
    workspaceStore.setTaskReportReady(false);
}

let loadingData = false;
// The latest page asked for wins; one answered after a newer request, or
// after the report has gone, is dropped.
const pageLoad = useLatestRequest();

const selectedWorkspacesForFilter = computed({
    get() {
        return workspaceStore.getSelectedWorkspaces;
    },
    set(value) {
        workspaceStore.setSelectedWorkspaces(value);
    },
});

// The saved report filters do not depend on the workspaces chosen, so only
// the tasks are loaded again, for page 1 of the new selection.
watch(
    () => selectedWorkspacesForFilter.value,
    () => {
        if (reportStarted.value) workspaceStore.bumpGridReloadToken();
    },
);

const viewFilters = computed({
    get() {
        return workspaceStore.getFilters;
    },
    set(value) {
        workspaceStore.setFilters(value);
    },
});

const reportFilters = computed(() =>
    (viewFilters.value || []).filter(
        (f) => f.table_id === "detailed-task-report" && f.view_id === "detailed-task-report",
    ),
);

// Once the report runs, the toolbar changes it, so the setup panel's
// remembered choice follows the toolbar and shows what was last used.
// Conditions set by hand have no place in the panel, which falls back to
// open tasks for them.
const SETUP_PRESETS = ["open_tasks", "my_tasks", "all_tasks", "late_tasks"];

watch(
    () => [
        selectedWorkspacesForFilter.value,
        workspaceStore.fastFilterOption,
        workspaceStore.getFastFilterLabel,
        workspaceStore.getFilter?.id,
        workspaceStore.getSavedFlatFilters,
        workspaceStore.getSavedGroups,
        workspaceStore.getSortOptions,
    ],
    ([ids, option, label, filterId, flatFilters, groups, sort]) => {
        if (!reportStarted.value) return;

        const all = !ids?.length || ids.length === (workspaceStore.getWorkspaces || []).length;
        const saved =
            option === "saved_filter" && reportFilters.value.find((f) => f.id === filterId);

        rememberSetupChoice({
            workspaceIds: all ? [] : [...ids],
            preset: SETUP_PRESETS.includes(option) ? option : "",
            savedFilterId: saved ? filterId : "",
            conditions: withoutFilterOptions({
                flatFilters: flatFilters || [],
                groups: groups || [],
            }),
            edited: option === "custom" || (!!saved && label !== saved.name),
            sort: (sort || []).filter((s) => s.field),
        });
    },
    { deep: true },
);

const fetchPage = async (page, workspaceIds = null, count = true) => {
    if (loadingData) {
        pendingFilterRetry = true;

        return;
    }

    const req = pageLoad.start();

    isLoading.value = true;
    try {
        const groups = toRaw(workspaceStore.getSavedGroups) || [];
        const flatFilters = toRaw(workspaceStore.getSavedFlatFilters) || [];
        const sort = toRaw(workspaceStore.getSortOptions)?.filter((s) => s.field) || [];
        const body = {
            page,
            limit: itemsPerPage,
            filters: withoutFilterOptions({ groups, flatFilters, sort }),
            skip_count: !count,
        };

        if (workspaceIds?.length) body.workspace_ids = workspaceIds;
        const res = await workspaceService.getAllWorkspaceTasks(body);

        if (!req.isCurrent()) return;
        const data = res.data || {};
        const base = data.data_base || [];

        if (data.total !== undefined) totalRows.value = data.total;
        if (data.root_total !== undefined) rootTotalRows.value = data.root_total;
        currentPage.value = page;
        tableData.value = base.map((item) => ({
            ...item,
            parent_task_id:
                typeof item.parent_task_id === "string" &&
                item.parent_task_id.trim() !== "" &&
                item.parent_task_id !== item.id
                    ? item.parent_task_id.trim()
                    : "",
        }));
        const assigneeIds = [...new Set(tableData.value.map((r) => r.assignee).filter(Boolean))];

        if (assigneeIds.length) await userStore.ensureUsers(assigneeIds);

        if (awaitingFilterLoad.value) {
            awaitingFilterLoad.value = false;
            tableLoaded.value = true;
        }
    } catch (error) {
        if (!req.isCurrent()) return;
        useAlertStore().showError(extractErrorMessage(error, "Failed to load data"));
        if (awaitingFilterLoad.value) {
            awaitingFilterLoad.value = false;
            tableLoaded.value = true;
        }
    } finally {
        if (req.isCurrent()) isLoading.value = false;
    }
};

const loadData = async () => {
    if (loadingData) return;
    loadingData = true;
    pageLoad.cancel();
    isLoading.value = true;
    try {
        // Only the saved report filters are needed here: the tasks come from
        // fetchPage(1) once the filter menu has applied the open tasks filter.
        const req = workspaceService.getAllWorkspaceTasks({ filters_only: true });

        resetState();

        const data = (await req).data || {};

        // The store may still hold the last table's filters, which are
        // not the report's, so an empty answer empties it too.
        viewFilters.value = Array.from(
            new Map((data.filters || []).map((f) => [f.filter_id, f])).values(),
        );
        if (!tableHeaders.value || tableHeaders.value.length === 0) {
            const cols = [
                { name: "id", w: 0 },
                { name: "name", w: 320 },
                { name: "start_date", w: 160 },
                { name: "due_date", w: 160 },
                { name: "assignee", w: 160, header_usage: "assignee" },
                { name: "status", w: 160, header_usage: "status" },
                { name: "updated_at", w: 160 },
                { name: "workspace_name", w: 160 },
                { name: "table_name", w: 160 },
                { name: "link_to_table", w: 160 },
            ];
            const headers = cols.map((c) => ({
                name: c.name,
                display_name: t.value(`projects.assigned_to_me.${c.name}`),
                width: c.w,
                header_usage: c.header_usage || "",
            }));

            tableHeaders.value = headers;
            columnWidths.value = headers.map((h) => h.width);
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error, "Failed to load data"));
        if (awaitingFilterLoad.value) {
            awaitingFilterLoad.value = false;
            tableLoaded.value = true;
        }
    } finally {
        loadingData = false;
        isLoading.value = false;
        setupReady.value = true;
        // If a filter token bump arrived while we were loading, apply it now
        if (pendingFilterRetry && reportStarted.value) {
            fetchPage(
                1,
                selectedWorkspacesForFilter.value?.length
                    ? selectedWorkspacesForFilter.value
                    : null,
            );
        }

        pendingFilterRetry = false;
    }
};

// Progressive rendering: show first batch immediately, add more each frame.
const pageSlice = computed(() => visibleData.value.slice(0, _dtrRenderedCount.value));

// Reload page 1 when filter is applied or cleared
watch(
    () => workspaceStore.getGridReloadToken,
    async () => {
        if (route.name === "detailed-task-report" && reportStarted.value) {
            // Use fetchPage instead of loadData. loadData would reset awaitingFilterLoad=true again.
            await fetchPage(
                1,
                selectedWorkspacesForFilter.value?.length
                    ? selectedWorkspacesForFilter.value
                    : null,
            );
        }
    },
);

// filterActive watcher removed. handleApplyFilters in TopNavTaskReport already
// loads data and clears saved filters when filters are removed. Calling loadData()
// here races with that and leaves awaitingFilterLoad=true (spinner stuck).

onMounted(async () => {
    resetState(true); // clear headers on fresh mount
    await loadData();
});
</script>

<style scoped src="../projectsTable.css"></style>

<style scoped>
.table-container {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    overflow-x: auto;
    background-color: rgb(255, 255, 255);
}

.header-row {
    position: sticky;
    top: 0;
    background-color: #f9fafb;
}

.table-content .grid-row:hover > .grid-item-text {
    background-color: #f9fafb;
}

.grid-item-text {
    background-color: white;
    font-size: 13px;
}

.sticky-row {
    height: 42px;
    position: sticky;
    bottom: 0;
    background-color: white;
    border-top: 1px solid #ddd;
    z-index: 1; /* Ensure it appears above the scrollable content */
    display: flex;
    justify-content: space-between;
    align-items: center;
}

/* Popover styles */
.popover {
    position: absolute;
}

.clickable-area {
    height: 100%;
    width: 100%;
    cursor: pointer;
    display: flex;
    align-items: center;
}
</style>
