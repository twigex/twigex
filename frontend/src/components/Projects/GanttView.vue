<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="gv-wrapper">
        <div v-if="isLoading" class="flex flex-1 items-center justify-center bg-white">
            <div class="flex flex-col items-center gap-y-3">
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>
        </div>
        <div v-else-if="noDateFields" class="gv-center">
            <p class="text-sm text-gray-500">
                {{ t("projects.gantt_view.no_date_fields") }}
            </p>
        </div>
        <div v-else-if="ganttTasks.length === 0" class="gv-center">
            <p class="text-sm text-gray-500">
                {{ t("projects.gantt_view.no_tasks") }}
            </p>
        </div>
        <template v-else>
            <div class="gv-toolbar">
                <BaseButton
                    v-for="mode in viewModes"
                    :key="mode"
                    type="button"
                    size="small"
                    :variant="currentMode === mode ? 'primary' : 'secondary'"
                    :is-disabled="isSwitchingMode"
                    @click="setMode(mode)"
                >
                    {{ mode }}
                </BaseButton>
                <div
                    v-if="isSwitchingMode"
                    class="h-4 w-4 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent ml-2 self-center"
                ></div>
                <div class="ml-auto flex items-center gap-2 text-xs text-gray-500">
                    <span>{{ ganttTasks.length }} {{ t("projects.gantt_view.tasks") }}</span>
                    <span
                        v-if="isCapped"
                        class="flex items-center gap-1 text-amber-600 font-medium"
                        :title="t('projects.gantt_view.cap_tooltip')"
                    >
                        <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M12 9v2m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"
                            />
                        </svg>
                        {{ t("projects.gantt_view.cap_warning") }}
                    </span>
                    <span
                        v-else-if="tasksOverflowVertical"
                        class="flex items-center gap-1 text-indigo-500 font-medium"
                    >
                        <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M19 9l-7 7-7-7"
                            />
                        </svg>
                        {{ t("projects.gantt_view.scroll_to_see_all") }}
                    </span>
                </div>
            </div>
            <div class="gv-scroll" ref="scrollEl">
                <div ref="ganttEl" class="gv-chart-el"></div>
            </div>
        </template>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import BaseSpinner from "@/components/BaseSpinner.vue";
import BaseButton from "@/components/BaseButton.vue";
import { ref, computed, onBeforeUnmount, watch, nextTick, toRaw } from "vue";
import { storeToRefs } from "pinia";
import Gantt from "frappe-gantt";
import "frappe-gantt/dist/frappe-gantt.css";
import { useRoute } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";
import workspaceService from "@/services/workspaceService";
import { onReconnect } from "@/js/websocket";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { appliedFlatFilters } from "@/utils/projects/filterCombine";
import useDateOperations from "@/composables/useDateOperations.js";
import moment from "moment-timezone";
import { useOpenTask } from "@/composables/projects/useOpenTask";
import { useLatestRequest } from "@/composables/useLatestRequest";

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const { getSavedFlatFilters, getSavedGroups } = storeToRefs(workspaceStore);

const ganttEl = ref(null);
const scrollEl = ref(null);
const isLoading = ref(true);
const viewModes = ["Day", "Week", "Month", "Year"];
const currentMode = ref("Month");
let ganttInstance = null;

const tasksOverflowVertical = computed(() => {
    if (!scrollEl.value) return false;

    return scrollEl.value.scrollHeight > scrollEl.value.clientHeight + 10;
});

// Tracks the timestamp window currently loaded so scroll-edge extension only fetches new data.
let loadedFrom = 0;
let loadedTo = 0;

const tableData = computed(() => workspaceStore.getTableData || []);
const tableHeaders = computed(() => workspaceStore.getTableHeaders || []);

const hasStartDate = computed(() => tableHeaders.value.some((h) => h.name === "start_date"));
const hasDueDate = computed(() => tableHeaders.value.some((h) => h.name === "due_date"));

const noDateFields = computed(() => !hasStartDate.value && !hasDueDate.value);

const { getDayOf: toYMD } = useDateOperations();

const GANTT_MAX_TASKS = 500;

const ganttTasks = computed(() => {
    const today = new Date().toISOString().split("T")[0];
    const nowTs = Date.now();

    const mapped = tableData.value
        .filter((row) => {
            const s = toYMD(row.start_date);
            const e = toYMD(row.due_date);

            if (!s && !e) return false;

            return true;
        })
        .map((row) => {
            const rawStart = toYMD(row.start_date) || toYMD(row.due_date) || today;
            const rawEnd = toYMD(row.due_date) || toYMD(row.start_date) || today;
            const start = rawStart <= rawEnd ? rawStart : rawEnd;
            const end = rawStart <= rawEnd ? rawEnd : rawStart;

            return {
                id: String(row.id),
                name: row.name || String(row.id),
                start,
                end,
                progress: 0,
                _dist: Math.abs(new Date(start).getTime() - nowTs),
            };
        });

    if (mapped.length > GANTT_MAX_TASKS) {
        mapped.sort((a, b) => a._dist - b._dist);
        mapped.length = GANTT_MAX_TASKS;
    }

    mapped.forEach((t) => delete t._dist);

    return mapped;
});

// How many tasks the loaded range holds on the server, which sends only the
// ones nearest today.
const rangeTotal = ref(0);
const isCapped = computed(() => rangeTotal.value > GANTT_MAX_TASKS);

// How far around today each zoom level loads first; scrolling to an edge
// loads more, and zooming into a range already loaded needs no new request.
const MODE_WINDOWS = {
    Day: [1, 2, "months"],
    Week: [3, 3, "months"],
    Month: [12, 12, "months"],
    Year: [3, 3, "years"],
};

function windowForMode(mode) {
    const tz = userStore.getTimezone;
    const now = moment().tz(tz);
    const [back, ahead, unit] = MODE_WINDOWS[mode] || MODE_WINDOWS.Month;

    return [
        now.clone().subtract(back, unit).startOf("month").unix(),
        now.clone().add(ahead, unit).endOf("month").unix(),
    ];
}

function rowsFromTable(table) {
    return (table.data_base || []).map((item) => {
        const row = {};

        for (const h of table.headers || []) {
            row[h.name] = item[h.name] ?? "";
        }

        row.id = item.id == null ? "" : String(item.id);
        row.name = item.name ?? "";
        row.start_date = item.start_date ?? "";
        row.due_date = item.due_date ?? "";

        return row;
    });
}

const tasksKey = (tasks) => tasks.map((t) => `${t.id}|${t.name}|${t.start}|${t.end}`).join("\n");

// What the chart on the page was last drawn from.
let drawnTasksKey = "";

// Guards to prevent duplicate/loop renders.
let isExtending = false;
let renderPending = false;
const isSwitchingMode = ref(false);
// Saved horizontal scroll position, restored after every re-render so resize doesn't jump.
let savedScrollLeft = null;

function scheduleRender() {
    if (renderPending) return;
    renderPending = true;
    requestAnimationFrame(() => {
        renderPending = false;
        renderGantt();
    });
}

function activeFilters() {
    const groups = toRaw(getSavedGroups.value || []);
    const flatFilters = appliedFlatFilters(toRaw(getSavedFlatFilters.value));

    if (!groups.length && !flatFilters.length) return null;

    return { groups, flatFilters };
}

// Cancels the previous loadData request when navigation fires a new one.
const rangeLoad = useLatestRequest();

// Replaces all loaded data with a fresh fetch for [from, to].
async function loadData(from, to) {
    if (!route.params.id || !route.params.tid || route.params.tid === "undefined") return;
    // Abort any previous in-flight request so it doesn't show a stale error toast.
    const { signal } = rangeLoad.start();

    isLoading.value = true;
    const f = activeFilters();

    try {
        // Fetch tasks and full headers concurrently so FilterBuilder sees all fields
        // (with linked_id for status names) before FilterRow's immediate watch fires.
        const [res, navHeaders] = await Promise.all([
            workspaceService.getTasksByDateRange(
                route.params.id,
                route.params.tid,
                from,
                to,
                f,
                signal,
            ),
            getFullHeaders(),
            // Load status options upfront so ALL FilterRow instances have them ready
            // the moment the filter panel opens, no per-row API calls needed.
            workspaceService
                .getTableStatusTypes(route.params.id, route.params.tid)
                .then((r) => {
                    const opts = (r?.data?.options || []).map((o) => ({ id: o.id, name: o.name }));

                    if (opts.length) workspaceStore.setStatusOptions(opts);
                })
                .catch(() => {}),
        ]);

        // Background fetch of the option tables for other linked fields (custom single-selects etc.).
        if (!workspaceStore.getWorkspaceTablesForLinked?.length) {
            workspaceService
                .getWorkspaceTables(route.params.id, "nav")
                .then((r) => workspaceStore.setWorkspaceTablesForLinked(r?.data || []))
                .catch(() => {});
        }

        const table = res?.data;

        if (!table) return;
        // Use full nav headers so FilterBuilder shows all fields with proper linked_id.
        workspaceStore.setTableHeaders(navHeaders.length ? navHeaders : table.headers || []);
        workspaceStore.setTableData(rowsFromTable(table));
        rangeTotal.value = table.total || 0;
        loadedFrom = from;
        loadedTo = to;
    } catch (err) {
        // Ignore cancellation, happens when the user navigates away before the request completes.
        if (
            err?.code === "ERR_CANCELED" ||
            err?.name === "AbortError" ||
            err?.name === "CanceledError"
        )
            return;
        useAlertStore().showError(extractErrorMessage(err));
    } finally {
        isLoading.value = false;
    }
}

// Fetches [from, to], merges new rows, refreshes gantt without resetting scroll.
async function extendRange(from, to) {
    if (isExtending) return;
    isExtending = true;
    try {
        const res = await workspaceService.getTasksByDateRange(
            route.params.id,
            route.params.tid,
            from,
            to,
            activeFilters(),
        );
        const table = res?.data;

        if (!table) return;
        rangeTotal.value += table.total || 0;
        const existingIds = new Set(tableData.value.map((r) => r.id));
        const newRows = rowsFromTable(table).filter((r) => !existingIds.has(r.id));

        if (newRows.length > 0) {
            workspaceStore.setTableData([...tableData.value, ...newRows]);
            // Wait for Vue to process the reactive update while isExtending=true,
            // so the ganttTasks watch skips re-render, then refresh in-place.
            await nextTick();
            if (ganttInstance) {
                ganttInstance.refresh(ganttTasks.value);
                drawnTasksKey = tasksKey(ganttTasks.value);
            }
        }

        loadedFrom = Math.min(loadedFrom, from);
        loadedTo = Math.max(loadedTo, to);
    } catch (err) {
        useAlertStore().showError(extractErrorMessage(err));
    } finally {
        isExtending = false;
    }
}

function applyHeight() {
    if (!ganttEl.value) return;
    const scrollEl = ganttEl.value.parentElement;
    const ROW_HEIGHT = 48;
    const HEADER_HEIGHT = 85;
    const available = scrollEl.clientHeight || 500;
    const minNeeded = HEADER_HEIGHT + ganttTasks.value.length * ROW_HEIGHT;
    const tasksOverflow = minNeeded > available;

    ganttEl.value.style.height = tasksOverflow ? minNeeded + "px" : "";
    scrollEl.style.overflowY = tasksOverflow ? "auto" : "";

    return tasksOverflow ? minNeeded : available;
}

// frappe-gantt adds a mouseup listener to document when a chart is built and
// never removes it, so every chart ever built stayed alive through it. One
// chart is built per table and updated in place; the listeners its
// constructor adds are recorded and removed with it.
let ganttListeners = [];

function destroyGantt() {
    for (const args of ganttListeners) document.removeEventListener(...args);
    ganttListeners = [];
    ganttInstance = null;
    if (ganttEl.value) ganttEl.value.innerHTML = "";
}

// Without infinite padding a chart spans its tasks and a little either side,
// which in Month and Year can be narrower than the page. Each side is
// widened to at least half the width shown. Per mode: the length of one
// column, its unit, and the library's own padding in that unit.
const COLUMN_SPANS = {
    Day: [1, "d", 7],
    Week: [7, "d", 30],
    Month: [1, "m", 2],
    Year: [1, "y", 2],
};

function fitPadding(modeName) {
    const mode = Object.values(Gantt.VIEW_MODE).find((m) => m.name === modeName);
    const span = COLUMN_SPANS[modeName];
    const width = ganttEl.value?.clientWidth || 0;

    if (!mode || !span || !width) return;

    const [length, unit, least] = span;
    const columns = Math.ceil(width / (mode.column_width || 45) / 2) + 1;

    mode.padding = `${Math.max(columns * length, least)}${unit}`;
}

function renderGantt() {
    if (!ganttEl.value || ganttTasks.value.length === 0) return;

    fitPadding(currentMode.value);

    const oldContainer = ganttEl.value.querySelector(".gantt-container");

    if (oldContainer) savedScrollLeft = oldContainer.scrollLeft;

    const container_height = applyHeight();
    // Disable weekend highlights in Year mode. A padding of "2y" would create 400+ SVG rects.
    const holidays =
        currentMode.value === "Year" ? {} : { "var(--g-weekend-highlight-color)": "weekend" };

    if (ganttInstance && ganttEl.value.contains(ganttInstance.$container)) {
        ganttInstance.options.view_mode = currentMode.value;
        ganttInstance.options.container_height = container_height;
        ganttInstance.options.holidays = holidays;
        ganttInstance.refresh(ganttTasks.value);
        drawnTasksKey = tasksKey(ganttTasks.value);
        restoreScroll();

        return;
    }

    destroyGantt();
    const add = document.addEventListener;
    const added = [];

    document.addEventListener = function (...args) {
        added.push(args);

        return add.apply(this, args);
    };

    try {
        ganttInstance = new Gantt(ganttEl.value, ganttTasks.value, {
            view_mode: currentMode.value,
            date_format: "YYYY-MM-DD",
            readonly: true,
            popup: false,
            today_button: false,
            container_height,
            scroll_to: null,
            holidays,
            // On by default, it widens the chart and draws all of it again
            // inside every mouse wheel event. Scrolling to an edge loads more
            // here instead.
            infinite_padding: false,
        });
    } finally {
        document.addEventListener = add;
    }

    ganttListeners = added;
    drawnTasksKey = tasksKey(ganttTasks.value);

    // Suppress any internal scroll attempt during or after construction.
    ganttInstance.set_scroll_position = (date) => {
        const gc = ganttEl.value?.querySelector(".gantt-container");

        if (!gc) return;
        gc.style.scrollBehavior = "auto";
        if (savedScrollLeft !== null) {
            gc.scrollLeft = savedScrollLeft;
        }

        // Explicit today-button call, scroll to today marker.
        if (date === "today" || date instanceof Date) {
            const todayEl = ganttEl.value?.querySelector(".current-date-highlight");

            if (todayEl) gc.scrollLeft = Math.max(0, todayEl.offsetLeft - gc.clientWidth / 2);
            savedScrollLeft = gc.scrollLeft;
        }
    };

    // The chart keeps its container and svg across refreshes, so these are
    // attached once per chart.
    const gc = ganttEl.value?.querySelector(".gantt-container");

    gc?.addEventListener(
        "scroll",
        () => {
            savedScrollLeft = gc.scrollLeft;
        },
        { passive: true },
    );
    restoreScroll();
    requestAnimationFrame(setupClickHandler);
}

// Scrolls back to where the user was, or to today on a new chart, using the
// today marker element, the most reliable approach across versions.
function restoreScroll() {
    const gc = ganttEl.value?.querySelector(".gantt-container");

    if (gc) {
        gc.style.scrollBehavior = "auto";
        if (savedScrollLeft !== null) {
            gc.scrollLeft = savedScrollLeft;
        } else {
            const todayEl = ganttEl.value?.querySelector(".current-date-highlight");

            if (todayEl) gc.scrollLeft = Math.max(0, todayEl.offsetLeft - gc.clientWidth / 2);
            savedScrollLeft = gc.scrollLeft;
        }
    }

    requestAnimationFrame(setupScrollEdgeDetection);
}

// Full table headers + workspace tables for linked fields (status names etc).
// Loaded once per table, reused for task detail view and FilterRow status options.
let fullHeaders = null;

async function getFullHeaders() {
    if (fullHeaders) return fullHeaders;
    try {
        // "nav" mode returns headers only (no row data), fast.
        // FilterRow fetches linked table data on-demand when user opens the filter panel.
        const res = await workspaceService.getWorkspaceTables(route.params.id, "nav");
        const tables = res?.data || [];
        const table = tables.find((t) => t.id === route.params.tid);

        fullHeaders = table?.headers || [];
    } catch {
        fullHeaders = [];
    }

    return fullHeaders;
}

const { openTask: openInPanel } = useOpenTask();

function openTask(taskId) {
    return openInPanel({
        workspaceId: route.params.id,
        tableId: route.params.tid,
        taskId,
        headers: getFullHeaders(),
    });
}

// Attached once per chart: refreshes keep the svg, so another listener would
// open the task once more per refresh.
function setupClickHandler() {
    if (!ganttEl.value) return;
    // Listen on the SVG itself (same approach as frappe-gantt's own event system).
    const svg = ganttEl.value.querySelector("svg.gantt");

    if (!svg) return;
    svg.addEventListener("click", (e) => {
        const bar = e.target.closest(".bar-wrapper");

        if (!bar) return;
        const taskId = bar.getAttribute("data-id");

        if (taskId) openTask(taskId);
    });
}

// Attaches horizontal-scroll edge detection to frappe-gantt's own scroll container.
let scrollEdgeCleanup = null;

function setupScrollEdgeDetection() {
    if (scrollEdgeCleanup) scrollEdgeCleanup();

    const ganttContainer = ganttEl.value?.querySelector(".gantt-container");

    if (!ganttContainer) return;

    const EDGE_PX = 300;
    const EXTEND_MONTHS = 2;
    const tz = userStore.getTimezone;
    let debounceTimer = null;
    // Record initial scroll position so we don't trigger on the very first render.
    let lastScrollLeft = ganttContainer.scrollLeft;

    const handler = () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(async () => {
            if (isExtending) return;
            const { scrollLeft, scrollWidth, clientWidth } = ganttContainer;

            // Ignore tiny movements (< 50px) to avoid triggering on initial render.
            if (Math.abs(scrollLeft - lastScrollLeft) < 50 && scrollLeft !== 0) return;
            lastScrollLeft = scrollLeft;
            if (scrollLeft < EDGE_PX) {
                const newFrom = moment
                    .unix(loadedFrom)
                    .tz(tz)
                    .subtract(EXTEND_MONTHS, "months")
                    .startOf("month")
                    .unix();

                if (newFrom < loadedFrom) await extendRange(newFrom, loadedFrom - 1);
            }

            if (scrollLeft + clientWidth > scrollWidth - EDGE_PX) {
                const newTo = moment
                    .unix(loadedTo)
                    .tz(tz)
                    .add(EXTEND_MONTHS, "months")
                    .endOf("month")
                    .unix();

                if (newTo > loadedTo) await extendRange(loadedTo + 1, newTo);
            }
        }, 150);
    };

    ganttContainer.addEventListener("scroll", handler, { passive: true });
    scrollEdgeCleanup = () => {
        clearTimeout(debounceTimer);
        ganttContainer.removeEventListener("scroll", handler);
    };
}

async function setMode(mode) {
    isSwitchingMode.value = true;
    await new Promise((r) => setTimeout(r, 0));
    try {
        currentMode.value = mode;
        const [from, to] = windowForMode(mode);

        if (from < loadedFrom || to > loadedTo) {
            await loadData(from, to);
            // ganttTasks watch fires after loadData and schedules renderGantt().
        } else {
            if (ganttInstance) {
                // Reset saved scroll so set_scroll_position finds today in the new zoom level
                // instead of restoring a stale pixel offset from the previous mode.
                savedScrollLeft = null;
                // Disable weekend highlights for Year mode to skip generating 400+ SVG rects.
                ganttInstance.options.holidays =
                    mode === "Year" ? {} : { "var(--g-weekend-highlight-color)": "weekend" };
                fitPadding(mode);
                ganttInstance.change_view_mode(mode, false);
                // Manually scroll to today after re-render.
                const gc = ganttEl.value?.querySelector(".gantt-container");

                if (gc) {
                    gc.style.scrollBehavior = "auto";
                    const todayEl = ganttEl.value?.querySelector(".current-date-highlight");

                    if (todayEl) {
                        gc.scrollLeft = Math.max(0, todayEl.offsetLeft - gc.clientWidth / 2);
                        savedScrollLeft = gc.scrollLeft;
                    }
                }
            }
        }
    } finally {
        isSwitchingMode.value = false;
    }
}

watch(
    () => [route.params.id, route.params.tid, route.params.fid],
    async () => {
        destroyGantt();
        loadedFrom = 0;
        loadedTo = 0;
        savedScrollLeft = null;
        fullHeaders = null;
        workspaceStore.setTableData([]);
        // Kick off workspace tables fetch immediately so FilterRow has linked
        // options (status names) ready by the time the user opens the filter panel.
        const [from, to] = windowForMode(currentMode.value);

        await loadData(from, to);
    },
    { immediate: true },
);

watch(
    () => isLoading.value,
    async (loading) => {
        if (!loading && ganttTasks.value.length > 0) {
            await nextTick();
            setupResizeObserver();
        }
    },
);

watch(ganttTasks, async (tasks) => {
    // A change to the table that moves no bar, an edit to another field or
    // the same rows read again, leaves a chart on the page as it is.
    const shown = ganttInstance && ganttEl.value?.contains(ganttInstance.$container);

    if (shown && tasksKey(tasks) === drawnTasksKey) return;

    // Skip during extendRange. It calls ganttInstance.refresh() directly.
    if (!isLoading.value && !isExtending) {
        await nextTick();
        scheduleRender();
    }
});

let resizeObserver = null;
let resizeTimer = null;

function onResize() {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
        // Only update height, don't recreate the gantt (would reset scroll position).
        applyHeight();
        if (ganttInstance) {
            const gc = ganttEl.value?.querySelector(".gantt-container");

            if (gc && savedScrollLeft !== null) gc.scrollLeft = savedScrollLeft;
        }
    }, 150);
}

function setupResizeObserver() {
    resizeObserver?.disconnect();
    const wrapperEl = ganttEl.value?.closest(".gv-wrapper");

    if (!wrapperEl) return;
    // ResizeObserver fires immediately on first observe, skip that initial callback.
    let ready = false;

    resizeObserver = new ResizeObserver(() => {
        if (!ready) {
            ready = true;

            return;
        }

        onResize();
    });
    resizeObserver.observe(wrapperEl);
}

// Re-fetch when filters change. Debounced so batched store updates don't cause double fetches.
let filterDebounce = null;

watch(
    [getSavedFlatFilters, getSavedGroups],
    () => {
        clearTimeout(filterDebounce);
        filterDebounce = setTimeout(async () => {
            if (!loadedFrom || !loadedTo) return;
            await loadData(loadedFrom, loadedTo);
        }, 50);
    },
    { deep: true },
);

const stopReconnect = onReconnect(() => {
    if (loadedFrom && loadedTo) loadData(loadedFrom, loadedTo);
});

onBeforeUnmount(() => {
    stopReconnect();
    destroyGantt();
    resizeObserver?.disconnect();
    clearTimeout(resizeTimer);
    if (scrollEdgeCleanup) scrollEdgeCleanup();
});
</script>

<style scoped>
.gv-wrapper {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: calc(100vh - 140px);
}

.gv-toolbar {
    display: flex;
    gap: 4px;
    padding: 8px 16px;
    border-bottom: 1px solid #e5e7eb;
    flex-shrink: 0;
}

.gv-scroll {
    flex: 1;
    overflow-x: hidden;
    overflow-y: hidden; /* switched to auto via JS only when tasks overflow */
    display: flex;
    flex-direction: column;
}

.gv-chart-el {
    flex: 1;
    min-height: 0;
}

.gv-center {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
}

.gv-chart-el :deep(.gantt-container) {
    width: 100%;
    background-color: #fff;
    overflow-y: hidden !important;
    scroll-behavior: auto !important;
}

.gv-chart-el :deep(svg.gantt) {
    display: block;
}
</style>
