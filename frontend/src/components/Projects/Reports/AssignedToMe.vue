<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="tableLoaded && totalTaskCount === 0" class="h-full bg-white">
        <div class="flex h-full w-full flex-col items-center justify-center text-center">
            <CubeIcon class="mx-auto h-12 w-12 text-gray-400" aria-hidden="true" />
            <h2 class="mt-4 text-lg font-semibold text-gray-900">
                {{ t("projects.assigned_to_me.title") }}
            </h2>
            <p class="mt-2 max-w-sm text-sm text-gray-500">
                {{ t("projects.assigned_to_me.description") }}
            </p>
        </div>
    </div>
    <div v-else-if="tableLoaded" class="flex h-full flex-col">
        <div class="min-h-0 flex-1 overflow-auto bg-white" :data-table-id="tableUniqueId">
            <div class="w-max min-w-full">
                <div
                    v-for="section in groupedSections"
                    :key="section.key"
                    class="border-t border-gray-200 first:border-t-0"
                >
                    <button
                        type="button"
                        class="sticky top-0 z-10 block w-full select-none border-b border-gray-200 bg-gray-50 text-left hover:bg-gray-100 focus:outline-none focus-visible:bg-gray-100"
                        :aria-expanded="!collapsedSections[section.key]"
                        @click="toggleSection(section.key)"
                    >
                        <span class="sticky left-0 inline-flex items-center gap-x-2 px-4 py-2.5">
                            <ChevronRightIcon
                                :class="[
                                    'h-4 w-4 text-gray-400 transition-transform duration-150',
                                    !collapsedSections[section.key] && 'rotate-90',
                                ]"
                                aria-hidden="true"
                            />
                            <span
                                :class="[
                                    'h-2 w-2 flex-none rounded-full',
                                    sectionDotClass(section.accent),
                                ]"
                                aria-hidden="true"
                            />
                            <span class="text-sm font-semibold text-gray-900">{{
                                section.label
                            }}</span>
                            <span
                                class="rounded-full bg-gray-200/70 px-2 py-0.5 text-xs font-medium text-gray-600"
                            >
                                {{ section.total }}
                            </span>
                        </span>
                    </button>

                    <!-- Section table -->
                    <div v-if="!collapsedSections[section.key]" class="grid-container">
                        <!-- Column headers -->
                        <ReportHeaderRow
                            class="col-header-row"
                            :headers="tableHeaders"
                            :column-widths="columnWidths"
                            @resize-start="startResize"
                        />

                        <!-- Data rows -->
                        <div
                            class="grid-row"
                            v-for="(item, rowIndex) in visibleBySection.get(section.key) || []"
                            :key="rowIndex"
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
                                :style="{ width: columnWidths[colIndex] + 'px' }"
                                :data-column-index="colIndex"
                            >
                                <div
                                    class="clickable-area"
                                    style="display: flex; align-items: center; height: 100%"
                                >
                                    <!-- Assignee -->
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

                                    <!-- Status -->
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

                                    <!-- Link to table -->
                                    <div
                                        v-else-if="tableHeaders[colIndex].name === 'link_to_table'"
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

                                    <!-- Default text/date -->
                                    <div
                                        v-else
                                        class="text-left alignLeft text-sm font-medium leading-6 text-gray-600"
                                        style="margin-right: 6px; font-size: 13px"
                                    >
                                        <div
                                            :style="{
                                                width: columnWidths[colIndex] - 6 + 'px',
                                                paddingRight: '6px',
                                                paddingLeft:
                                                    header.name === 'name'
                                                        ? hasChildren(item.id) ||
                                                          item._is_subtask ||
                                                          (item.parent_task_id &&
                                                              item.parent_task_id !== item.id)
                                                            ? (item._indent || 0) * 20 + 'px'
                                                            : '0px'
                                                        : '0px',
                                                overflow: 'hidden',
                                                whiteSpace: 'nowrap',
                                                textOverflow: 'ellipsis',
                                            }"
                                            class="flex items-center gap-2"
                                            :title="item[header.name]"
                                        >
                                            <span
                                                v-if="header.name === 'name'"
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
                                                    header.name === 'name' &&
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
                                                    header.name === 'name'
                                                        ? showFullView(
                                                              item,
                                                              item.link_to_table.table_id,
                                                              item.link_to_table.workspace_id,
                                                              item.link_to_table.view_id,
                                                          )
                                                        : null
                                                "
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

                                            <div
                                                v-if="
                                                    header.name === 'name' &&
                                                    getSubtaskCount(item.id) > 0
                                                "
                                                class="flex items-center gap-1 text-gray-400 text-xs flex-shrink-0"
                                                :title="`${getSubtaskCount(item.id)} ${t('projects.assigned_to_me.subtasks')}`"
                                            >
                                                <Square2StackIcon class="w-4 h-4" />
                                                <span>{{ getSubtaskCount(item.id) }}</span>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <button
                            v-if="section.next && section.tasks.length < section.total"
                            type="button"
                            class="block w-full border-b border-gray-200 text-left text-sm font-medium text-indigo-600 hover:bg-gray-50 hover:text-indigo-500 disabled:cursor-wait disabled:text-gray-400"
                            :disabled="loadingMore[section.key]"
                            @click="loadMore(section.key)"
                        >
                            <span class="sticky left-0 inline-block px-4 py-2.5">
                                {{ t("projects.kanban_view.load_more") }}
                                {{ Math.min(PAGE_SIZE, section.total - section.tasks.length) }}
                                <span class="font-normal text-gray-500">
                                    ({{ section.total - section.tasks.length }}
                                    {{ t("projects.kanban_view.remaining") }})
                                </span>
                            </span>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Bottom bar -->
        <div class="flex h-[42px] flex-none items-center border-t border-gray-200 bg-white px-4">
            <span class="text-sm text-gray-700 font-medium">
                {{ t("projects.assigned_to_me.total_tasks") }}
                {{ totalTaskCount }}
            </span>
        </div>
    </div>
    <div v-else-if="loadFailed" class="flex h-full items-center justify-center bg-white px-4">
        <div class="text-center">
            <ExclamationTriangleIcon class="mx-auto h-10 w-10 text-gray-400" aria-hidden="true" />
            <p class="mt-3 text-sm text-gray-600">
                {{ t("projects.assigned_to_me.failed_to_load_data") }}
            </p>
            <button
                type="button"
                class="mt-4 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                @click="loadData"
            >
                {{ t("common.button.try_again") }}
            </button>
        </div>
    </div>
    <div v-else class="flex h-full items-center justify-center bg-white">
        <div class="flex flex-col items-center gap-y-3">
            <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
            <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, reactive, onMounted, onBeforeUnmount, watch } from "vue";
import UserAvatarWithText from "@/components/UserAvatarWithText.vue";
import { useUserStore } from "@/store/user";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { isBlank } from "@/utils/projects/rows";
import { useColumnResize } from "@/composables/projects/useColumnResize";
import { flattenTaskTree } from "@/utils/projects/tree";
import { useRouter } from "vue-router";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import useDateOperations from "@/composables/useDateOperations.js";
import {
    ArrowTopRightOnSquareIcon,
    Square2StackIcon,
    ChevronRightIcon,
    ChevronDownIcon,
    CubeIcon,
} from "@heroicons/vue/20/solid";
import { ExclamationTriangleIcon, UserPlusIcon } from "@heroicons/vue/24/outline";
import BaseSpinner from "@/components/BaseSpinner.vue";
import { getStatusStyle } from "@/utils/projects/cells";
import { useSubtaskCounts } from "@/composables/projects/useSubtaskCounts";
import ReportHeaderRow from "@/components/Projects/Reports/ReportHeaderRow.vue";
import { useOpenTask } from "@/composables/projects/useOpenTask";

const router = useRouter();
const { getDate } = useDateOperations();

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

const workspaces = ref([]);
const tableData = computed({
    get() {
        return workspaceStore.getTableData;
    },
    set(value) {
        workspaceStore.setTableData(value);
    },
});
const { hasChildren, getSubtaskCount } = useSubtaskCounts(tableData);

const tableUniqueId = `assigned-${Math.random().toString(36).slice(2, 11)}`;
const { startResize } = useColumnResize({
    headers: tableHeaders,
    widths: columnWidths,
    tableId: tableUniqueId,
});

const { openTask } = useOpenTask();

async function showFullView(item, table_id, workspace_id) {
    if (!(await openTask({ workspaceId: workspace_id, tableId: table_id, taskId: item.id })))
        return;

    router.push({
        name: "grid-view",
        params: {
            id: workspace_id,
            tid: table_id,
            fid: item.link_to_table.view_id,
        },
        query: { task: item.id },
    });
}

function goToTable(workspace_id, table_id, item) {
    showFullView(item, table_id, workspace_id);
    resetState();
}

const resetState = () => {
    workspaceStore.clearTableData();
    workspaceStore.setTableHeaders([]);
    workspaceStore.setColumnsWidths([]);
    workspaces.value = [];
    tableHeaders.value = [];
    columnWidths.value = [];
    tableData.value = [];
    selectedView.value = null;
};

const isLoading = ref(false);
const tableLoaded = ref(false);
const loadFailed = ref(false);

// Subtask expand/collapse
const expandedTasks = ref(new Set());

const toggleExpand = (itemId) => {
    if (expandedTasks.value.has(itemId)) {
        expandedTasks.value.delete(itemId);
    } else {
        expandedTasks.value.add(itemId);
    }
};

// Built when the tasks or what is expanded change, not on every render:
// resizing a column renders the page on each mouse move.
const visibleBySection = computed(
    () =>
        new Map(
            groupedSections.value.map((s) => [
                s.key,
                flattenTaskTree(s.tasks, expandedTasks.value),
            ]),
        ),
);

// Section collapse state
const collapsedSections = reactive({});

function toggleSection(key) {
    collapsedSections[key] = !collapsedSections[key];
}

function sectionDotClass(accent) {
    const map = {
        red: "bg-red-500",
        orange: "bg-orange-400",
        yellow: "bg-yellow-400",
        blue: "bg-blue-500",
        gray: "bg-gray-400",
    };

    return map[accent] || "bg-gray-400";
}

const SECTIONS = [
    { key: "overdue", label: "projects.assigned_to_me.section.overdue", accent: "red" },
    { key: "today", label: "projects.assigned_to_me.section.today", accent: "orange" },
    { key: "thisWeek", label: "projects.assigned_to_me.section.this_week", accent: "yellow" },
    { key: "upcoming", label: "projects.assigned_to_me.section.upcoming", accent: "blue" },
    { key: "noDueDate", label: "projects.assigned_to_me.section.no_due_date", accent: "gray" },
];
const PAGE_SIZE = 50;
const sectionData = reactive({});
const loadingMore = reactive({});

const groupedSections = computed(() =>
    SECTIONS.map((s) => ({
        key: s.key,
        label: t.value(s.label),
        accent: s.accent,
        tasks: sectionData[s.key]?.tasks || [],
        total: sectionData[s.key]?.total || 0,
        next: sectionData[s.key]?.next || "",
    })).filter((s) => s.total > 0),
);

const totalTaskCount = computed(() =>
    SECTIONS.reduce((sum, s) => sum + (sectionData[s.key]?.total || 0), 0),
);

// The rows land in the store the grid and board share, so an answer that
// arrives after leaving this page would overwrite theirs.
const requests = new AbortController();

onBeforeUnmount(() => requests.abort());

const loadData = async () => {
    if (isLoading.value) return;
    isLoading.value = true;
    resetState();

    try {
        const res = await workspaceService.getAssignedToMe(
            {
                tz: userStore.getTimezone,
                limit: PAGE_SIZE,
            },
            requests.signal,
        );

        if (requests.signal.aborted) return;

        const defaultHeaders = [
            { name: "id", display_name: t.value("projects.assigned_to_me.id") },
            { name: "name", display_name: t.value("projects.assigned_to_me.name") },
            { name: "start_date", display_name: t.value("projects.assigned_to_me.start_date") },
            { name: "due_date", display_name: t.value("projects.assigned_to_me.due_date") },
            { name: "assignee", display_name: t.value("projects.assigned_to_me.assignee") },
            { name: "status", display_name: t.value("projects.assigned_to_me.status") },
            { name: "updated_at", display_name: t.value("projects.assigned_to_me.updated_at") },
            {
                name: "workspace_name",
                display_name: t.value("projects.assigned_to_me.workspace_name"),
            },
            { name: "table_name", display_name: t.value("projects.assigned_to_me.table_name") },
            {
                name: "link_to_table",
                display_name: t.value("projects.assigned_to_me.link_to_table"),
            },
        ];

        tableHeaders.value = defaultHeaders.map((header) => ({
            name: header.name,
            display_name: header.display_name,
            width: header.name === "id" ? 0 : header.name === "name" ? 320 : 160,
        }));
        columnWidths.value = tableHeaders.value.map((h) => h.width);

        const page = res.data || {};
        const rows = [];

        for (const key of Object.keys(sectionData)) delete sectionData[key];
        for (const s of page.sections || []) {
            const tasks = (s.tasks || []).map((item) => toRow(item, page.tables || {}));

            sectionData[s.key] = { total: s.total, tasks, next: s.next || "" };
            rows.push(...tasks);
        }

        // The tasks show at once; their assignees' names fill in as they load.
        ensureAssignees(rows);
        syncTableData();
        loadFailed.value = false;
        tableLoaded.value = true;
    } catch (error) {
        if (requests.signal.aborted) return;
        loadFailed.value = true;
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.assigned_to_me.failed_to_load_data")),
        );
    } finally {
        isLoading.value = false;
    }
};

const TASK_FIELDS = ["id", "name", "start_date", "due_date", "assignee", "status", "updated_at"];

function toRow(item, tables) {
    const table = tables[item.table_id] || {};
    const row = {};

    for (const name of TASK_FIELDS) row[name] = item[name] || "";
    row.workspace_name = table.workspace_name || "";
    row.table_name = table.table_name || "";
    row.link_to_table = {
        workspace_id: table.workspace_id,
        table_id: item.table_id,
        view_id: table.view_id || "",
    };
    row.parent_task_id =
        typeof item.parent_task_id === "string" && item.parent_task_id.trim() !== ""
            ? item.parent_task_id.trim()
            : "";

    return row;
}

async function ensureAssignees(rows) {
    const missing = [...new Set(rows.map((r) => r.assignee).filter(Boolean))].filter(
        (id) => !userStore.usersMap[id],
    );

    if (missing.length) await userStore.ensureUsers(missing);
}

function syncTableData() {
    tableData.value = SECTIONS.flatMap((s) => sectionData[s.key]?.tasks || []);
}

// The task panel saves into the shared rows; the sections hold their own.
watch(
    () => workspaceStore.savedTaskField,
    (saved) => {
        if (!saved) return;

        let found = false;

        for (const section of Object.values(sectionData)) {
            const row = section.tasks?.find((task) => String(task.id) === String(saved.id));

            if (row) {
                row[saved.field] = saved.value;
                found = true;
            }
        }

        if (found) syncTableData();
    },
);

async function loadMore(key) {
    const section = sectionData[key];

    if (!section || loadingMore[key]) return;
    loadingMore[key] = true;
    try {
        const res = await workspaceService.getAssignedToMe(
            {
                tz: userStore.getTimezone,
                section: key,
                after: section.next,
                limit: PAGE_SIZE,
            },
            requests.signal,
        );

        if (requests.signal.aborted) return;
        const s = res.data?.sections?.[0];

        if (s) {
            const tasks = (s.tasks || []).map((item) => toRow(item, res.data.tables || {}));

            await ensureAssignees(tasks);
            if (requests.signal.aborted) return;
            section.total = s.total;
            section.next = s.next || "";
            section.tasks = [...section.tasks, ...tasks];
            syncTableData();
        }
    } catch (error) {
        if (requests.signal.aborted) return;
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.assigned_to_me.failed_to_load_data")),
        );
    } finally {
        loadingMore[key] = false;
    }
}

onMounted(async () => {
    resetState();
    await loadData();
});
</script>

<style scoped src="../projectsTable.css"></style>

<style scoped>
.grid-container {
    display: block;
    width: max-content;
}

.col-header-row {
    background-color: #f9fafb;
    border-bottom: 1px solid #e5e7eb;
}

.grid-row:not(.col-header-row):hover > .grid-item-text {
    background-color: #f9fafb;
}

.grid-item-text {
    background-color: white;
    font-size: 13px;
}

.sticky-row {
    height: 42px;
    position: sticky;
    padding-left: 10px;
    bottom: 0;
    background-color: white;
    border-top: 1px solid #ddd;
    z-index: 1;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.clickable-area {
    height: 100%;
    width: 100%;
    cursor: pointer;
    display: flex;
    align-items: center;
}
</style>
