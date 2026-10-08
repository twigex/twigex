<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="md:grid md:grid-cols-1 md:divide-x md:divide-gray-200">
        <div>
            <div class="flex items-center justify-center space-x-2">
                <button
                    @click="previousYear"
                    type="button"
                    class="flex flex-none items-center justify-center p-1 text-gray-400 hover:text-gray-500"
                >
                    <span class="sr-only">Previous year</span>
                    <ChevronDoubleLeftIcon class="h-4 w-4" aria-hidden="true" />
                </button>

                <button
                    @click="previousMonth"
                    type="button"
                    class="flex flex-none items-center justify-center p-1 text-gray-400 hover:text-gray-500"
                >
                    <span class="sr-only">Previous month</span>
                    <ChevronLeftIcon class="h-4 w-4" aria-hidden="true" />
                </button>

                <div class="flex items-center space-x-2">
                    <Menu as="div" class="relative inline-block text-left">
                        <div>
                            <MenuButton
                                class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                            >
                                {{ monthsList[state.cur.m - 1] }}
                            </MenuButton>
                        </div>

                        <transition
                            enter-active-class="transition ease-out duration-100"
                            enter-from-class="transform opacity-0 scale-95"
                            enter-to-class="transform opacity-100 scale-100"
                            leave-active-class="transition ease-in duration-75"
                            leave-from-class="transform opacity-100 scale-100"
                            leave-to-class="transform opacity-0 scale-95"
                        >
                            <MenuItems
                                class="absolute max-h-48 overflow-y-auto right-0 z-10 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                            >
                                <div class="py-1">
                                    <MenuItem v-for="(m, i) in monthsList" :key="m" as="template">
                                        <button
                                            class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                                            @click="selectMonth(i)"
                                        >
                                            {{ m }}
                                        </button>
                                    </MenuItem>
                                </div>
                            </MenuItems>
                        </transition>
                    </Menu>

                    <Menu as="div" class="relative inline-block text-left">
                        <div>
                            <MenuButton
                                class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                            >
                                {{ state.currentYear }}
                            </MenuButton>
                        </div>

                        <transition
                            enter-active-class="transition ease-out duration-100"
                            enter-from-class="transform opacity-0 scale-95"
                            enter-to-class="transform opacity-100 scale-100"
                            leave-active-class="transition ease-in duration-75"
                            leave-from-class="transform opacity-100 scale-100"
                            leave-to-class="transform opacity-0 scale-95"
                        >
                            <MenuItems
                                class="absolute max-h-48 right-0 z-10 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                            >
                                <div class="py-1">
                                    <RecycleScroller
                                        class="max-h-48 overflow-y-auto"
                                        :items="availableYears"
                                        :buffer="10"
                                        :item-size="32"
                                        v-slot="{ item }"
                                    >
                                        <MenuItem as="template">
                                            <button
                                                class="block w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100"
                                                @click="selectYear(item)"
                                            >
                                                {{ item }}
                                            </button>
                                        </MenuItem>
                                    </RecycleScroller>
                                </div>
                            </MenuItems>
                        </transition>
                    </Menu>
                </div>

                <button
                    @click="nextMonth"
                    type="button"
                    class="flex flex-none items-center justify-center p-1 text-gray-400 hover:text-gray-500"
                >
                    <span class="sr-only">Next month</span>
                    <ChevronRightIcon class="h-4 w-4" aria-hidden="true" />
                </button>

                <button
                    @click="nextYear"
                    type="button"
                    class="flex flex-none items-center justify-center p-1 text-gray-400 hover:text-gray-500"
                >
                    <span class="sr-only">Next year</span>
                    <ChevronDoubleRightIcon class="h-4 w-4" aria-hidden="true" />
                </button>
            </div>

            <div class="mt-4 grid grid-cols-7 text-center text-xs leading-6 text-gray-500">
                <div>M</div>
                <div>T</div>
                <div>W</div>
                <div>T</div>
                <div>F</div>
                <div>S</div>
                <div>S</div>
            </div>

            <div class="mt-1 grid grid-cols-7 text-sm">
                <div
                    v-for="(day, dayIdx) in state.days"
                    :key="day.key"
                    :class="[dayIdx > 6 && 'border-t border-gray-200', 'py-1']"
                >
                    <button
                        @click="updateActive(day)"
                        type="button"
                        :class="[
                            day.isSelected && 'text-white',
                            !day.isSelected && day.isToday && 'text-indigo-600',
                            !day.isSelected &&
                                !day.isToday &&
                                day.isCurrentMonth &&
                                'text-indigo-900',
                            !day.isSelected &&
                                !day.isToday &&
                                !day.isCurrentMonth &&
                                'text-gray-400',
                            day.isSelected && 'bg-indigo-600',
                            !day.isSelected && 'hover:bg-gray-200',
                            (day.isSelected || day.isToday) && 'font-semibold',
                            'mx-auto flex h-7 w-7 items-center justify-center rounded-full',
                        ]"
                    >
                        <time>{{ day.day }}</time>
                    </button>
                </div>
            </div>

            <div
                v-if="withTime"
                class="mt-3 flex items-center justify-center border-t border-gray-100 pt-3"
            >
                <TimeSelect :model-value="timeStr" @update:model-value="onTimePick" />
            </div>
        </div>
    </div>
</template>

<script setup>
import {
    ChevronLeftIcon,
    ChevronRightIcon,
    ChevronDoubleLeftIcon,
    ChevronDoubleRightIcon,
} from "@heroicons/vue/20/solid";
import { Menu, MenuButton, MenuItems, MenuItem } from "@headlessui/vue";
import TimeSelect from "@/components/DatePicker/TimeSelect.vue";
import { RecycleScroller } from "vue-virtual-scroller";
import "vue-virtual-scroller/dist/vue-virtual-scroller.css";
import { reactive, ref, onMounted, watch, computed } from "vue";
import { useUserStore } from "@/store/user";
import { currentLocale } from "@/i18n/index";

const userStore = useUserStore();
const tz = computed(() => {
    let timezone = userStore.user?.timezone?.manualTimezone;

    if (userStore.user?.timezone?.useAutomaticTimezone === "true") {
        timezone = userStore.user?.timezone?.automaticTimezone || timezone;
    }

    return timezone || "UTC";
});

const props = defineProps({
    modelValue: String,
    // When true, shows a time field and emits "YYYY-MM-DDTHH:MM" instead of
    // "YYYY-MM-DD". Off by default so existing date-only consumers are unaffected.
    withTime: {
        type: Boolean,
        default: false,
    },
});
const emit = defineEmits(["update:modelValue"]);

const now = new Date();
const timeStr = ref(
    `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}`,
);

function fmtParts(date, timeZone, opts) {
    const dtf = new Intl.DateTimeFormat("en-GB", { timeZone, ...opts });

    return dtf.formatToParts(date).reduce((acc, p) => {
        acc[p.type] = p.value;

        return acc;
    }, {});
}

function todayParts(timeZone) {
    const now = new Date();
    const p = fmtParts(now, timeZone, {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
    });

    return { y: Number(p.year), m: Number(p.month), d: Number(p.day) };
}

function ymdString({ y, m, d }) {
    return `${y}-${String(m).padStart(2, "0")}-${String(d).padStart(2, "0")}`;
}

function parseYMD(str, timeZone) {
    const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(String(str || ""));

    if (!m) return todayParts(timeZone);

    return { y: Number(m[1]), m: Number(m[2]), d: Number(m[3]) };
}

function daysInMonth(y, m1) {
    return new Date(y, m1, 0).getDate();
}

const WD_MAP = { Mon: 0, Tue: 1, Wed: 2, Thu: 3, Fri: 4, Sat: 5, Sun: 6 };

function firstOfMonthWeekday(y, m1, timeZone) {
    const instant = new Date(Date.UTC(y, m1 - 1, 1, 12));
    const p = fmtParts(instant, timeZone, { weekday: "short" });

    return WD_MAP[p.weekday];
}

function clampDay(y, m1, d) {
    return Math.min(d, daysInMonth(y, m1));
}

// Localized month names for the current app locale (index 0 = January).
const monthsList = computed(() => {
    const fmt = new Intl.DateTimeFormat(currentLocale.value, { month: "long" });

    return Array.from({ length: 12 }, (_, i) => fmt.format(new Date(2000, i, 1)));
});

const startYear = 1900;
const currentYearNum = new Date().getFullYear();
const availableYears = Array.from(
    { length: currentYearNum - startYear + 1 },
    (_, i) => startYear + i,
).reverse();

// The month shown can be moved away from the chosen day, which stays chosen
// until another is clicked.
const selectedYMD = ref("");

const state = reactive({
    days: [],
    cur: todayParts(tz.value),
    currentYear: String(new Date().getFullYear()),
    currentDay: String(new Date().getDate()),
});

function refreshLabels() {
    state.currentYear = String(state.cur.y);
    state.currentDay = String(state.cur.d);
}

function emitModel() {
    const day = selectedYMD.value || ymdString(state.cur);

    if (props.withTime) {
        emit("update:modelValue", `${day}T${timeStr.value || "00:00"}`);
    } else {
        emit("update:modelValue", day);
    }
}

function onTimePick(value) {
    timeStr.value = value;
    emitModel();
}

function addMonths(delta) {
    let y = state.cur.y;
    let m1 = state.cur.m + delta;
    let d = state.cur.d;

    while (m1 <= 0) {
        m1 += 12;
        y -= 1;
    }

    while (m1 > 12) {
        m1 -= 12;
        y += 1;
    }

    d = clampDay(y, m1, d);
    state.cur = { y, m: m1, d };
    refreshLabels();
    state.days = getDays();
}

function addYears(delta) {
    const y = state.cur.y + delta;
    const m1 = state.cur.m;
    const d = clampDay(y, m1, state.cur.d);

    state.cur = { y, m: m1, d };
    refreshLabels();
    state.days = getDays();
}

onMounted(() => {
    const [datePart, timePart] = String(props.modelValue || "").split("T");

    state.cur = datePart ? parseYMD(datePart, tz.value) : todayParts(tz.value);
    selectedYMD.value = datePart ? ymdString(state.cur) : "";
    if (props.withTime && timePart) {
        timeStr.value = timePart.slice(0, 5);
    }

    refreshLabels();
    state.days = getDays();

    // Seed the bound value so the shown default (today + time) is usable
    // without forcing the user to re-click. Only in time mode, only if unset.
    if (props.withTime && !props.modelValue) {
        emitModel();
    }
});
watch(
    () => props.modelValue,
    (newValue) => {
        if (newValue) {
            const [datePart, timePart] = String(newValue).split("T");

            state.cur = parseYMD(datePart, tz.value);
            selectedYMD.value = ymdString(state.cur);
            if (props.withTime && timePart) {
                timeStr.value = timePart.slice(0, 5);
            }

            refreshLabels();
            state.days = getDays();
        } else {
            selectedYMD.value = "";
            state.days = getDays();
        }
    },
);

function nextMonth() {
    addMonths(1);
}

function previousMonth() {
    addMonths(-1);
}

function nextYear() {
    addYears(1);
}

function previousYear() {
    addYears(-1);
}

function updateStateFromDateParts(p) {
    state.cur = { ...p };
    refreshLabels();
    state.days = getDays();
}

function updateActive(day) {
    selectedYMD.value = day.date;
    updateStateFromDateParts(parseYMD(day.date, tz.value));
    emitModel();
}

function selectMonth(index) {
    const targetM1 = index + 1;
    const d = clampDay(state.cur.y, targetM1, state.cur.d);

    updateStateFromDateParts({ y: state.cur.y, m: targetM1, d });
}

function selectYear(year) {
    const d = clampDay(year, state.cur.m, state.cur.d);

    updateStateFromDateParts({ y: year, m: state.cur.m, d });
}

function getDays() {
    const viewYear = Number(state.cur.y);
    const viewMonthIndex = state.cur.m - 1;
    const viewMonth1 = viewMonthIndex + 1;
    const padStart = firstOfMonthWeekday(viewYear, viewMonth1, tz.value);
    const dimCurrent = daysInMonth(viewYear, viewMonth1);
    const prevMonthIndex = (viewMonthIndex + 11) % 12;
    const prevYear = viewMonthIndex === 0 ? viewYear - 1 : viewYear;
    const prevMonth1 = prevMonthIndex + 1;
    const dimPrev = daysInMonth(prevYear, prevMonth1);
    const selected = selectedYMD.value;
    const t = todayParts(tz.value);
    const todayYMD = ymdString(t);
    const cells = [];
    const startDayPrev = dimPrev - padStart + 1;

    for (let i = 0; i < padStart; i++) {
        const dayNum = startDayPrev + i;
        const ymd = ymdString({ y: prevYear, m: prevMonth1, d: dayNum });

        cells.push({
            key: `prev-${ymd}`,
            day: dayNum,
            date: ymd,
            isCurrentMonth: false,
            isToday: ymd === todayYMD,
            isSelected: ymd === selected,
        });
    }

    for (let i = 1; i <= dimCurrent; i++) {
        const ymd = ymdString({ y: viewYear, m: viewMonth1, d: i });

        cells.push({
            key: `cur-${ymd}`,
            day: i,
            date: ymd,
            isCurrentMonth: true,
            isToday: ymd === todayYMD,
            isSelected: ymd === selected,
        });
    }

    const totalNeeded = 42;
    const trailing = totalNeeded - cells.length;
    const nextMonthIndex = (viewMonthIndex + 1) % 12;
    const nextYear = viewMonthIndex === 11 ? viewYear + 1 : viewYear;
    const nextMonth1 = nextMonthIndex + 1;

    for (let i = 1; i <= trailing; i++) {
        const ymd = ymdString({ y: nextYear, m: nextMonth1, d: i });

        cells.push({
            key: `next-${ymd}`,
            day: i,
            date: ymd,
            isCurrentMonth: false,
            isToday: ymd === todayYMD,
            isSelected: ymd === selected,
        });
    }

    return cells;
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: all 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
    transform: scale(0.95);
}
</style>
