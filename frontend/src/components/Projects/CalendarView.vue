<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="relative flex flex-col w-full h-[calc(100vh-140px)] bg-white">
        <div
            v-if="isInitialLoad"
            class="absolute inset-0 z-50 flex items-center justify-center bg-white"
        >
            <div class="flex flex-col items-center gap-y-3">
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>
        </div>
        <div
            class="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white flex-shrink-0"
        >
            <div class="flex items-center gap-2">
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronLeftIcon"
                    :aria-label="t('pagination.previous')"
                    @click="prev"
                />
                <span class="text-base font-semibold text-gray-900 min-w-[160px] text-center">{{
                    navLabel
                }}</span>
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    :prepend-icon="ChevronRightIcon"
                    :aria-label="t('pagination.next')"
                    @click="next"
                />
                <BaseButton
                    type="button"
                    size="small"
                    variant="secondary"
                    class="ml-2"
                    @click="goToToday"
                >
                    {{ t("projects.calendar_view.today") }}
                </BaseButton>
            </div>
            <div class="flex items-center gap-1">
                <BaseButton
                    v-for="v in ['day', 'week', 'month']"
                    :key="v"
                    type="button"
                    size="small"
                    :variant="currentView === v ? 'primary' : 'secondary'"
                    @click="setView(v)"
                >
                    {{ t("projects.calendar_view." + v) }}
                </BaseButton>
            </div>
        </div>

        <div v-if="currentView === 'day'" class="flex-1 flex flex-col overflow-hidden">
            <div
                class="flex items-baseline gap-2 px-5 pt-4 pb-2 border-b border-gray-200 flex-shrink-0"
            >
                <span
                    :class="
                        dayViewCell.isToday
                            ? 'flex items-center justify-center w-12 h-12 rounded-full bg-indigo-700 text-white text-[22px] font-bold'
                            : 'text-4xl font-bold text-gray-700 leading-none'
                    "
                    >{{ dayViewCell.day }}</span
                >
                <span class="text-sm text-gray-500 font-medium">{{ dayViewCell.weekday }}</span>
            </div>
            <div
                v-if="!dayViewCell.tasks.length"
                class="text-gray-400 text-[13px] text-center py-8"
            >
                {{ t("projects.calendar_view.no_tasks") }}
            </div>
            <div
                v-else
                ref="dayScrollEl"
                class="flex-1 overflow-y-auto px-4 py-2"
                @scroll="onDayScroll"
            >
                <div
                    :style="{
                        height: dayViewCell.tasks.length * DAY_ITEM_H + 'px',
                        minHeight: '100%',
                        position: 'relative',
                    }"
                >
                    <div
                        v-for="item in dayVisibleTasks"
                        :key="item.id"
                        :style="{
                            position: 'absolute',
                            top: item._top + 'px',
                            left: 0,
                            right: 0,
                            height: DAY_ITEM_H + 'px',
                            display: 'flex',
                            alignItems: 'center',
                        }"
                        :class="
                            item.id === selectedTaskId
                                ? 'bg-indigo-700 text-white border-indigo-700'
                                : 'bg-indigo-50 text-indigo-900 border-indigo-100 hover:bg-indigo-100'
                        "
                        class="p-3 mb-1 rounded-md text-[13px] cursor-pointer border"
                        @click.stop="openTask(item)"
                    >
                        {{ item.name }}
                    </div>
                </div>
            </div>
        </div>

        <template v-else-if="currentView === 'week'">
            <!-- DOW header: border-r on each cell so column lines start from the top -->
            <div class="grid grid-cols-7 border-b border-gray-200 border-l flex-shrink-0">
                <div
                    v-for="cell in weekSpans.cells"
                    :key="cell.dateStr"
                    class="py-1.5 px-2 text-center text-xs font-semibold text-gray-500 uppercase tracking-wide border-r border-gray-200"
                    :class="{ 'text-indigo-600 font-bold': cell.isToday }"
                >
                    {{ cell.weekday }} {{ cell.day }}
                </div>
            </div>
            <!-- Spanning bars band: gradient provides column lines behind the span bars -->
            <div
                v-if="weekSpans.spans.length"
                class="grid grid-cols-7 py-0.5 flex-shrink-0 border-l border-gray-200"
                :style="{
                    gridTemplateRows: `repeat(${weekSpans.bandRows}, 20px)`,
                    backgroundImage:
                        'repeating-linear-gradient(to right, transparent 0, transparent calc(100% / 7 - 1px), #e5e7eb calc(100% / 7 - 1px), #e5e7eb calc(100% / 7))',
                }"
            >
                <div
                    v-for="span in weekSpans.spans"
                    :key="span.id"
                    class="text-white text-[11px] leading-tight py-0.5 px-1.5 rounded whitespace-nowrap overflow-hidden text-ellipsis mx-0.5 my-px cursor-pointer self-center"
                    :class="[
                        span.id === selectedTaskId
                            ? 'bg-indigo-700'
                            : 'bg-indigo-500 hover:bg-indigo-600',
                        {
                            'rounded-l-none ml-0': span.continueLeft,
                            'rounded-r-none mr-0': span.continueRight,
                        },
                    ]"
                    :style="{
                        gridColumn: `${span.colStart + 1} / ${span.colEnd + 2}`,
                        gridRow: span.row + 1,
                    }"
                    :title="span.name"
                    @click.stop="openTask(span)"
                >
                    {{ span.name }}
                </div>
            </div>
            <!-- Week grid: gradient background draws column lines behind everything;
                 tasks paint on top as normal children. -->
            <div
                class="flex-1 min-h-0 border-l border-gray-200 overflow-hidden bg-white"
                :style="{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(7, 1fr)',
                    gridTemplateRows: '1fr',
                    backgroundImage:
                        'repeating-linear-gradient(to right, transparent 0, transparent calc(100% / 7 - 1px), #e5e7eb calc(100% / 7 - 1px), #e5e7eb calc(100% / 7))',
                }"
            >
                <div
                    v-for="cell in weekSpans.cells"
                    :key="cell.dateStr"
                    class="p-1 overflow-y-auto"
                    :class="cell.isToday ? 'bg-indigo-50/15' : 'bg-transparent'"
                    :style="{ height: '100%' }"
                    :ref="
                        (el) => {
                            if (el) weekCellHeights[cell.dateStr] = el.clientHeight;
                        }
                    "
                    @scroll="
                        (e) => {
                            weekScrollTops[cell.dateStr] = e.target.scrollTop;
                            weekCellHeights[cell.dateStr] = e.target.clientHeight;
                        }
                    "
                >
                    <div
                        :style="{
                            height: cell.tasks.length * WEEK_ITEM_H + 'px',
                            minHeight: '100%',
                            position: 'relative',
                        }"
                    >
                        <div
                            v-for="item in weekVisibleTasks(cell)"
                            :key="item.id"
                            :style="{
                                position: 'absolute',
                                top: item._top + 'px',
                                left: 0,
                                right: 0,
                                height: WEEK_ITEM_H + 'px',
                                display: 'flex',
                                alignItems: 'center',
                            }"
                            class="text-[11px] leading-[1.3] px-1 py-px rounded-sm whitespace-nowrap overflow-hidden text-ellipsis cursor-pointer"
                            :class="
                                item.id === selectedTaskId
                                    ? 'bg-indigo-600 text-white'
                                    : 'bg-indigo-50 text-indigo-800'
                            "
                            :title="item.name"
                            @click.stop="openTask(item)"
                        >
                            {{ item.name }}
                        </div>
                    </div>
                </div>
            </div>
        </template>

        <template v-else>
            <div class="grid grid-cols-7 border-b border-gray-200 flex-shrink-0">
                <div
                    v-for="day in dowLabels"
                    :key="day"
                    class="py-1.5 px-2 text-center text-xs font-semibold text-gray-500 uppercase tracking-wide"
                >
                    {{ day }}
                </div>
            </div>
            <div
                class="flex-1 grid grid-cols-7 auto-rows-fr overflow-y-auto border-l border-gray-200"
            >
                <div
                    v-for="cell in calendarCells"
                    :key="cell.key"
                    class="border-r border-b border-gray-200 p-1 min-h-0 overflow-hidden"
                    :class="{
                        'bg-gray-50': !cell.currentMonth,
                        'bg-[#fafafe]': cell.isToday && cell.currentMonth,
                        'bg-white': !cell.isToday && cell.currentMonth,
                    }"
                >
                    <div class="mb-0.5">
                        <span
                            class="inline-flex items-center justify-center w-[22px] h-[22px] text-xs font-medium leading-none"
                            :class="
                                cell.isToday
                                    ? 'rounded-full bg-indigo-600 text-white font-bold'
                                    : cell.currentMonth
                                      ? 'text-gray-700'
                                      : 'text-gray-300'
                            "
                        >
                            {{ cell.day }}
                        </span>
                    </div>
                    <div class="flex flex-col gap-0.5">
                        <div
                            v-for="task in cell.allTasks.slice(0, MAX_VISIBLE_TASKS)"
                            :key="task.id"
                            class="text-[11px] leading-[1.3] px-1 py-px rounded-sm whitespace-nowrap overflow-hidden text-ellipsis cursor-pointer"
                            :class="
                                task.id === selectedTaskId
                                    ? 'bg-indigo-700 text-white'
                                    : task.startYMD !== task.endYMD
                                      ? 'bg-indigo-500 text-white'
                                      : 'bg-indigo-50 text-indigo-800'
                            "
                            :title="task.name"
                            @click.stop="openTask(task)"
                        >
                            {{ task.name }}
                        </div>
                        <button
                            v-if="cell.allTasks.length > MAX_VISIBLE_TASKS"
                            class="text-[11px] text-gray-500 bg-transparent border-0 px-1 cursor-pointer text-left leading-[1.4] hover:text-indigo-600"
                            @click.stop="openMorePopover($event, { ...cell, tasks: cell.allTasks })"
                        >
                            {{
                                t("projects.calendar_view.n_more", {
                                    n: cell.allTasks.length - MAX_VISIBLE_TASKS,
                                })
                            }}
                        </button>
                    </div>
                </div>
            </div>
        </template>

        <Teleport to="body">
            <div
                v-if="popoverCell"
                ref="popoverFloating"
                class="z-[60] flex w-60 flex-col overflow-hidden rounded-md bg-white shadow-lg ring-1 ring-black/5"
                :style="popoverStyle"
                @click.stop
            >
                <div
                    class="flex items-center justify-between px-2.5 pt-2 pb-1.5 border-b border-gray-100 flex-shrink-0"
                >
                    <span class="text-xs font-semibold text-gray-700">{{ popoverDateLabel }}</span>
                    <button
                        type="button"
                        class="text-gray-400 bg-transparent border-0 cursor-pointer p-0.5 hover:text-gray-700"
                        @click="closePopover"
                    >
                        <span class="sr-only">{{ t("common.button.close") }}</span>
                        <XMarkIcon class="h-3.5 w-3.5" aria-hidden="true" />
                    </button>
                </div>
                <div class="px-2 pt-1.5 pb-1 border-b border-gray-100 flex-shrink-0">
                    <input
                        ref="popoverSearchInput"
                        v-model="popoverSearch"
                        type="text"
                        :placeholder="t('projects.calendar_view.search')"
                        class="block w-full rounded-md border-0 px-2 py-1 text-xs text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                    />
                </div>
                <div
                    ref="popoverScrollEl"
                    class="overflow-y-auto max-h-60 flex-1"
                    @scroll="onPopoverScroll"
                >
                    <div
                        :style="{
                            height: popoverFilteredTasks.length * POP_ITEM_H + 'px',
                            position: 'relative',
                            minHeight: '4px',
                        }"
                    >
                        <div
                            v-for="item in popoverVisibleTasks"
                            :key="item.id"
                            :style="{
                                position: 'absolute',
                                top: item._top + 'px',
                                left: 0,
                                right: 0,
                                height: POP_ITEM_H + 'px',
                            }"
                            class="flex items-center px-2.5 text-xs text-gray-900 cursor-pointer whitespace-nowrap overflow-hidden text-ellipsis box-border hover:bg-indigo-50 hover:text-indigo-800"
                            :title="item.name"
                            @click.stop="
                                openTask(item);
                                closePopover();
                            "
                        >
                            {{ item.name }}
                        </div>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<script setup>
import { ref, computed, reactive, onBeforeUnmount, onMounted, watch, toRaw, nextTick } from "vue";
import { useAnchoredPopup } from "@/composables/useAnchoredPopup";
import { storeToRefs } from "pinia";
import { useRoute } from "vue-router";
import { ChevronLeftIcon, ChevronRightIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { useWorkspaceStore } from "@/store/workspaces";
import { useUserStore } from "@/store/user";
import workspaceService from "@/services/workspaceService";
import { onReconnect } from "@/js/websocket";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { appliedFlatFilters } from "@/utils/projects/filterCombine";
import useDateOperations from "@/composables/useDateOperations.js";
import { t } from "@/i18n/index.js";
import BaseSpinner from "@/components/BaseSpinner.vue";
import BaseButton from "@/components/BaseButton.vue";
import moment from "moment-timezone";
import { useOpenTask } from "@/composables/projects/useOpenTask";
import { useLatestRequest } from "@/composables/useLatestRequest";

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();
const alertStore = useAlertStore();
const { getSavedFlatFilters, getSavedGroups } = storeToRefs(workspaceStore);
const isLoading = ref(false);
const isInitialLoad = ref(true); // spinner only on first entry, not on view switches
const selectedTaskId = ref(null);

// Clear selection when the task detail panel closes
watch(
    () => workspaceStore.getFullTask,
    (open) => {
        if (!open) selectedTaskId.value = null;
    },
);

const today = new Date();

// Localised day-of-week labels. dowLabels = Mon-first (calendar header).
// dowShort = Sun-first (indexed by JS Date.getDay()).
// t is a computed ref; use t.value() inside JS, templates auto-unwrap.
const dowLabels = computed(() => [
    t.value("projects.calendar_view.dow_mon"),
    t.value("projects.calendar_view.dow_tue"),
    t.value("projects.calendar_view.dow_wed"),
    t.value("projects.calendar_view.dow_thu"),
    t.value("projects.calendar_view.dow_fri"),
    t.value("projects.calendar_view.dow_sat"),
    t.value("projects.calendar_view.dow_sun"),
]);
const dowShort = computed(() => [
    t.value("projects.calendar_view.dow_sun"),
    t.value("projects.calendar_view.dow_mon"),
    t.value("projects.calendar_view.dow_tue"),
    t.value("projects.calendar_view.dow_wed"),
    t.value("projects.calendar_view.dow_thu"),
    t.value("projects.calendar_view.dow_fri"),
    t.value("projects.calendar_view.dow_sat"),
]);

const currentView = ref("month");
const currentDate = ref(new Date(today));

const currentYear = computed(() => currentDate.value.getFullYear());
const currentMonth = computed(() => currentDate.value.getMonth());

const navLabel = computed(() => {
    const tz = userStore.getTimezone;
    const m = moment(currentDate.value).tz(tz);
    const is12h = userStore.getClockDisplay === "12h";

    if (currentView.value === "day")
        return m.format(is12h ? "dddd, MMMM D, YYYY" : "dddd, D MMMM YYYY");
    if (currentView.value === "week") {
        const mon = m.clone().startOf("isoWeek");
        const sun = mon.clone().add(6, "days");

        return is12h
            ? mon.format("MMM D") + " - " + sun.format("MMM D, YYYY")
            : mon.format("D MMM") + " - " + sun.format("D MMM YYYY");
    }

    return m.format("MMMM YYYY");
});

function prev() {
    const d = new Date(currentDate.value);

    if (currentView.value === "day") d.setDate(d.getDate() - 1);
    else if (currentView.value === "week") d.setDate(d.getDate() - 7);
    else {
        d.setDate(1);
        d.setMonth(d.getMonth() - 1);
    }

    currentDate.value = d;
}

function next() {
    const d = new Date(currentDate.value);

    if (currentView.value === "day") d.setDate(d.getDate() + 1);
    else if (currentView.value === "week") d.setDate(d.getDate() + 7);
    else {
        d.setDate(1);
        d.setMonth(d.getMonth() + 1);
    }

    currentDate.value = d;
}

function goToToday() {
    currentDate.value = new Date(today);
}

function setView(v) {
    currentView.value = v;
    loadTasks();
}

const { getDayOf: toYMD } = useDateOperations();

const tableData = computed(() => workspaceStore.getTableData || []);

function dateRange(startYMD, endYMD) {
    const dates = [];
    const cur = moment.tz(startYMD, userStore.getTimezone);
    const end = moment.tz(endYMD, userStore.getTimezone);

    while (cur <= end) {
        dates.push(cur.format("YYYY-MM-DD"));
        cur.add(1, "day");
    }

    return dates;
}

const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, "0")}-${String(today.getDate()).padStart(2, "0")}`;

// The days on screen, as YYYY-MM-DD, which compare correctly as text.
const visibleDays = computed(() => {
    const { from, to } = visibleWindow();
    const tz = userStore.getTimezone;

    return {
        from: moment.unix(from).tz(tz).format("YYYY-MM-DD"),
        to: moment.unix(to).tz(tz).format("YYYY-MM-DD"),
    };
});

// A task is placed on the days it covers that are on screen only: one that
// runs for years would otherwise add an entry for every day of them.
const tasksByDate = computed(() => {
    const map = {};
    const { from, to } = visibleDays.value;

    for (const row of tableData.value) {
        const due = toYMD(row.due_date);
        const start = toYMD(row.start_date);
        let dates;
        const startYMD = start || due;
        const endYMD = due || start;

        if (start && due) {
            const a = start <= due ? start : due;
            const b = start <= due ? due : start;

            dates = a > to || b < from ? [] : dateRange(a < from ? from : a, b > to ? to : b);
        } else {
            dates = [startYMD].filter(Boolean);
        }

        for (const d of dates) {
            if (!map[d]) map[d] = [];
            map[d].push({ id: row.id, name: row.name || String(row.id), startYMD, endYMD });
        }
    }

    return map;
});

// Greedy row assignment so non-overlapping spans share rows in the band.
function assignSpanRows(spans) {
    const rows = []; // last colEnd per row

    return spans.map((span) => {
        let row = rows.findIndex((lastEnd) => lastEnd < span.colStart);

        if (row === -1) {
            row = rows.length;
            rows.push(span.colEnd);
        } else rows[row] = span.colEnd;

        return { ...span, row };
    });
}

function computeSpans(cells) {
    const seen = new Map();

    cells.forEach((cell, col) => {
        for (const task of cell.allTasks || cell.tasks || []) {
            if (task.startYMD !== task.endYMD && task.startYMD && task.endYMD) {
                if (!seen.has(task.id)) {
                    seen.set(task.id, {
                        id: task.id,
                        name: task.name,
                        colStart: col,
                        colEnd: col,
                        continueLeft: task.startYMD < cells[0].dateStr,
                        continueRight: task.endYMD > cells[cells.length - 1].dateStr,
                    });
                } else {
                    seen.get(task.id).colEnd = col;
                }
            }
        }
    });
    const spans = [...seen.values()].filter((s) => s.colEnd > s.colStart);

    return assignSpanRows(spans);
}

function makeDateStr(date) {
    return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

const dayViewCell = computed(() => {
    const d = currentDate.value;
    const dateStr = makeDateStr(d);

    return {
        day: d.getDate(),
        weekday: dowShort.value[d.getDay()],
        dateStr,
        isToday: dateStr === todayStr,
        tasks: tasksByDate.value[dateStr] || [],
    };
});

const weekCells = computed(() => {
    const d = new Date(currentDate.value);
    const dayOfWeek = d.getDay(); // 0=Sun
    const diffToMon = (dayOfWeek + 6) % 7;

    d.setDate(d.getDate() - diffToMon);

    return Array.from({ length: 7 }, (_, i) => {
        const date = new Date(d);

        date.setDate(d.getDate() + i);
        const dateStr = makeDateStr(date);
        const allTasks = tasksByDate.value[dateStr] || [];

        return {
            dateStr,
            day: date.getDate(),
            weekday: dowShort.value[date.getDay()],
            isToday: dateStr === todayStr,
            allTasks,
            tasks: allTasks,
        };
    });
});

const weekSpans = computed(() => {
    const cells = weekCells.value;
    const spans = computeSpans(cells);
    const spanIds = new Set(spans.map((s) => s.id));

    return {
        spans,
        bandRows: spans.length ? Math.max(...spans.map((s) => s.row)) + 1 : 0,
        cells: cells.map((c) => ({
            ...c,
            tasks: c.allTasks.filter((t) => !spanIds.has(t.id)),
        })),
    };
});

const calendarCells = computed(() => {
    const year = currentYear.value;
    const month = currentMonth.value;

    const firstDay = new Date(year, month, 1);
    const startOffset = (firstDay.getDay() + 6) % 7; // Monday = 0
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const totalCells = Math.ceil((startOffset + daysInMonth) / 7) * 7;

    const cells = [];

    for (let i = 0; i < totalCells; i++) {
        const date = new Date(year, month, 1 - startOffset + i);
        const y = date.getFullYear();
        const m = date.getMonth();
        const d = date.getDate();
        const dateStr = `${y}-${String(m + 1).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
        const allTasks = tasksByDate.value[dateStr] || [];

        cells.push({
            key: dateStr,
            day: d,
            dateStr,
            currentMonth: m === month,
            isToday: dateStr === todayStr,
            allTasks,
            tasks: allTasks.filter((t) => !t.startYMD || !t.endYMD || t.startYMD === t.endYMD),
        });
    }

    return cells;
});

let _loadedHeadersKey = "";

async function loadHeaders() {
    const key = `${route.params.id}:${route.params.tid}`;

    if (_loadedHeadersKey === key) return;
    try {
        const res = await workspaceService.getWorkspaceTables(route.params.id, "nav");
        const tables = res?.data || [];
        const table = tables.find((tbl) => tbl.id === route.params.tid);

        if (!table) return;
        workspaceStore.setTableHeaders(table.headers || []);
        _loadedHeadersKey = key;
        const statusHeader = (table.headers || []).find((h) => h.header_usage === "status");

        if (statusHeader) {
            workspaceService
                .getTableStatusTypes(route.params.id, route.params.tid)
                .then((r) => {
                    const opts = (r?.data?.options || []).map((o) => ({ id: o.id, name: o.name }));

                    if (opts.length) workspaceStore.setStatusOptions(opts);
                })
                .catch((err) => alertStore.showError(extractErrorMessage(err)));
        }
    } catch (err) {
        alertStore.showError(extractErrorMessage(err));
    }
}

function visibleWindow() {
    const tz = userStore.getTimezone;
    const d = currentDate.value;

    if (currentView.value === "day") {
        const ymd = makeDateStr(d);

        return {
            from: Math.floor(moment.tz(ymd, tz).startOf("day").valueOf() / 1000),
            to: Math.floor(moment.tz(ymd, tz).endOf("day").valueOf() / 1000),
        };
    }

    if (currentView.value === "week") {
        const base = moment.tz(makeDateStr(d), tz);
        const mon = base.clone().startOf("isoWeek");
        const sun = mon.clone().endOf("isoWeek");

        return {
            from: Math.floor(mon.valueOf() / 1000),
            to: Math.floor(sun.valueOf() / 1000),
        };
    }

    // Month view: includes adjacent month days visible in the grid.
    const year = currentYear.value;
    const month = currentMonth.value;
    const firstDay = new Date(year, month, 1);
    const startOffset = (firstDay.getDay() + 6) % 7;
    const daysInMonth = new Date(year, month + 1, 0).getDate();
    const totalCells = Math.ceil((startOffset + daysInMonth) / 7) * 7;
    const firstCell = new Date(year, month, 1 - startOffset);
    const lastCell = new Date(year, month, 1 - startOffset + totalCells - 1);

    return {
        from: Math.floor(moment.tz(makeDateStr(firstCell), tz).startOf("day").valueOf() / 1000),
        to: Math.floor(moment.tz(makeDateStr(lastCell), tz).endOf("day").valueOf() / 1000),
    };
}

function activeFilters() {
    const groups = toRaw(getSavedGroups.value || []);
    const flatFilters = appliedFlatFilters(toRaw(getSavedFlatFilters.value));

    if (!groups.length && !flatFilters.length) return { timezone: userStore.getTimezone };

    return { groups, flatFilters, timezone: userStore.getTimezone };
}

// The latest range asked for wins: moving quickly between months, or a
// filter changing during a load, must not leave an older answer on screen.
const rangeLoad = useLatestRequest();

const stopReconnect = onReconnect(() => {
    _loadedHeadersKey = "";
    loadHeaders();
    loadTasks();
});

onBeforeUnmount(stopReconnect);

async function loadTasks() {
    const req = rangeLoad.start();
    const { signal } = req;

    isLoading.value = true;
    try {
        const { from, to } = visibleWindow();
        const res = await workspaceService.getTasksByCalendarRange(
            route.params.id,
            route.params.tid,
            from,
            to,
            activeFilters(),
            signal,
        );

        if (!req.isCurrent()) return;
        const raw = res?.data;
        let items = [];

        if (Array.isArray(raw)) {
            items = raw;
        } else if (Array.isArray(raw?.data_base)) {
            items = raw.data_base;
        } else if (raw && typeof raw === "object") {
            items = Object.values(raw);
        }

        items = items.filter((item) => item != null && item.id != null);
        workspaceStore.setTableData(
            items.map((item) => ({
                id: item.id == null ? "" : String(item.id),
                name: item.name ?? "",
                start_date: item.start_date ?? "",
                due_date: item.due_date ?? "",
            })),
        );
    } catch (err) {
        if (req.isCurrent()) alertStore.showError(extractErrorMessage(err));
    } finally {
        if (req.isCurrent()) {
            isLoading.value = false;
            isInitialLoad.value = false;
        }
    }
}

watch(
    () => [route.params.id, route.params.tid, route.params.fid],
    () => {
        isInitialLoad.value = true;
        Promise.all([loadHeaders(), loadTasks()]);
    },
    { immediate: true },
);

watch(currentDate, () => loadTasks());

let filterDebounce = null;

watch(
    [getSavedFlatFilters, getSavedGroups],
    () => {
        clearTimeout(filterDebounce);
        filterDebounce = setTimeout(() => loadTasks(), 50);
    },
    { deep: true },
);

const MAX_VISIBLE_TASKS = 3;
const POP_ITEM_H = 28;
const POP_MAX_H = 240;

const DAY_ITEM_H = 44;
const dayScrollEl = ref(null);
const dayScrollTop = ref(0);

function onDayScroll(e) {
    dayScrollTop.value = e.target.scrollTop;
}

watch(currentDate, () => {
    dayScrollTop.value = 0;
});

const dayVisibleTasks = computed(() => {
    const items = dayViewCell.value.tasks;

    if (!items.length) return [];
    const h = dayScrollEl.value?.clientHeight || 600;
    const st = dayScrollTop.value;
    const start = Math.max(0, Math.floor(st / DAY_ITEM_H) - 3);
    const end = Math.min(items.length, Math.ceil((st + h) / DAY_ITEM_H) + 3);

    return items.slice(start, end).map((t, i) => ({ ...t, _top: (start + i) * DAY_ITEM_H }));
});

const WEEK_ITEM_H = 26;
const weekScrollTops = reactive({});
const weekCellHeights = reactive({});

function weekVisibleTasks(cell) {
    const items = cell.tasks;

    if (!items.length) return [];
    const scrollTop = weekScrollTops[cell.dateStr] || 0;
    const h = weekCellHeights[cell.dateStr] || 600;
    const start = Math.max(0, Math.floor(scrollTop / WEEK_ITEM_H) - 3);
    const end = Math.min(items.length, Math.ceil((scrollTop + h) / WEEK_ITEM_H) + 3);

    return items.slice(start, end).map((t, i) => ({ ...t, _top: (start + i) * WEEK_ITEM_H }));
}

const popoverCell = ref(null);
const {
    floating: popoverFloating,
    floatingStyles: popoverStyle,
    anchorTo: anchorPopover,
} = useAnchoredPopup({ placement: "right-start" });
const popoverSearch = ref("");
const popoverScrollEl = ref(null);
const popoverSearchInput = ref(null);
const popoverScrollTop = ref(0);

const popoverDateLabel = computed(() => {
    if (!popoverCell.value) return "";
    const is12h = userStore.getClockDisplay === "12h";

    return moment
        .tz(popoverCell.value.dateStr, userStore.getTimezone)
        .format(is12h ? "MMMM D, YYYY" : "D MMMM YYYY");
});

const popoverFilteredTasks = computed(() => {
    if (!popoverCell.value) return [];
    const q = popoverSearch.value.trim().toLowerCase();
    const tasks = popoverCell.value.tasks;

    return q ? tasks.filter((t) => t.name?.toLowerCase().includes(q)) : tasks;
});

const popoverVisibleTasks = computed(() => {
    const items = popoverFilteredTasks.value;

    if (!items.length) return [];
    const start = Math.max(0, Math.floor(popoverScrollTop.value / POP_ITEM_H) - 2);
    const end = Math.min(
        items.length,
        Math.ceil((popoverScrollTop.value + POP_MAX_H) / POP_ITEM_H) + 2,
    );

    return items.slice(start, end).map((t, i) => ({ ...t, _top: (start + i) * POP_ITEM_H }));
});

function onPopoverScroll(e) {
    popoverScrollTop.value = e.target.scrollTop;
}

function openMorePopover(e, cell) {
    anchorPopover(e.currentTarget);
    popoverSearch.value = "";
    popoverScrollTop.value = 0;
    popoverCell.value = cell;
    nextTick(() => popoverSearchInput.value?.focus());
}

function closePopover() {
    popoverCell.value = null;
}

function handleDocClick() {
    closePopover();
}

onMounted(() => document.addEventListener("click", handleDocClick));
onBeforeUnmount(() => document.removeEventListener("click", handleDocClick));

const { openTask: openInPanel } = useOpenTask();

async function openTask(task) {
    selectedTaskId.value = task.id;
    const opened = await openInPanel({
        workspaceId: route.params.id,
        tableId: route.params.tid,
        taskId: task.id,
    });

    if (!opened) selectedTaskId.value = null;
}
</script>
