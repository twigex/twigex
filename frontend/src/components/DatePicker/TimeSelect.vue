<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex items-center justify-center gap-2">
        <ClockIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />

        <div
            class="flex items-center gap-1 rounded-md px-2 py-1 shadow-sm ring-1 ring-inset ring-gray-300 focus-within:ring-2 focus-within:ring-indigo-600"
        >
            <TimeUnit
                :model-value="hour"
                @update:model-value="onHour"
                @step="step('hour', $event)"
            />
            <span class="text-sm font-medium text-gray-500">:</span>
            <TimeUnit
                :model-value="minute"
                @update:model-value="onMinute"
                @step="step('minute', $event)"
            />
        </div>

        <div v-if="is12h" class="flex overflow-hidden rounded-md ring-1 ring-inset ring-gray-300">
            <button
                v-for="m in ['AM', 'PM']"
                :key="m"
                type="button"
                @click="setMeridiem(m)"
                :class="[
                    meridiem === m
                        ? 'bg-indigo-600 text-white'
                        : 'bg-white text-gray-700 hover:bg-gray-50',
                    'px-2.5 py-1.5 text-xs font-semibold',
                ]"
            >
                {{ m }}
            </button>
        </div>
    </div>
</template>

<script setup>
import { ref, computed, watch, h } from "vue";
import { ClockIcon, ChevronUpIcon, ChevronDownIcon } from "@heroicons/vue/20/solid";
import { useUserStore } from "@/store/user";

// Small inline render component: a centered number field with up/down steppers.
const TimeUnit = {
    props: { modelValue: { type: Number, default: 0 } },
    emits: ["update:modelValue", "step"],
    setup(props, { emit }) {
        const pad = (n) => String(n).padStart(2, "0");

        return () =>
            h("div", { class: "flex items-center gap-0.5" }, [
                h("input", {
                    type: "text",
                    inputmode: "numeric",
                    value: pad(props.modelValue),
                    onChange: (e) => emit("update:modelValue", e.target.value),
                    onFocus: (e) => e.target.select(),
                    class: "w-9 border-0 bg-transparent p-0 text-center text-sm text-gray-900 focus:outline-none focus:ring-0",
                }),
                h("div", { class: "flex flex-col" }, [
                    h(
                        "button",
                        {
                            type: "button",
                            onClick: () => emit("step", 1),
                            class: "text-gray-400 hover:text-gray-700",
                        },
                        [h(ChevronUpIcon, { class: "h-3.5 w-3.5" })],
                    ),
                    h(
                        "button",
                        {
                            type: "button",
                            onClick: () => emit("step", -1),
                            class: "text-gray-400 hover:text-gray-700",
                        },
                        [h(ChevronDownIcon, { class: "h-3.5 w-3.5" })],
                    ),
                ]),
            ]);
    },
};

const props = defineProps({
    // "HH:MM" in 24h
    modelValue: {
        type: String,
        default: "",
    },
});
const emit = defineEmits(["update:modelValue"]);

const userStore = useUserStore();
const is12h = computed(() => userStore.getClockDisplay === "12h");

const hour = ref(9);
const minute = ref(0);
const meridiem = ref("AM");

function pad(n) {
    return String(n).padStart(2, "0");
}

function clamp(n, min, max) {
    n = Number(n);
    if (!Number.isFinite(n)) return min;

    return Math.min(Math.max(Math.round(n), min), max);
}

const hourMax = computed(() => (is12h.value ? 12 : 23));
const hourMin = computed(() => (is12h.value ? 1 : 0));

function syncFromModel() {
    const [h24, m] = String(props.modelValue || "")
        .split(":")
        .map(Number);
    const hh = Number.isFinite(h24) ? h24 : 9;

    minute.value = clamp(Number.isFinite(m) ? m : 0, 0, 59);
    if (is12h.value) {
        meridiem.value = hh >= 12 ? "PM" : "AM";
        hour.value = hh % 12 || 12;
    } else {
        hour.value = clamp(hh, 0, 23);
    }
}

watch(() => props.modelValue, syncFromModel, { immediate: true });
watch(is12h, syncFromModel);

function to24() {
    let hr = clamp(hour.value, hourMin.value, hourMax.value);

    if (is12h.value) {
        hr = hr % 12;
        if (meridiem.value === "PM") hr += 12;
    }

    return `${pad(hr)}:${pad(clamp(minute.value, 0, 59))}`;
}

function commit() {
    hour.value = clamp(hour.value, hourMin.value, hourMax.value);
    minute.value = clamp(minute.value, 0, 59);
    emit("update:modelValue", to24());
}

function onHour(value) {
    hour.value = value;
    commit();
}

function onMinute(value) {
    minute.value = value;
    commit();
}

// Step a unit up/down with wrap-around.
function step(unit, dir) {
    if (unit === "hour") {
        let hr = clamp(hour.value, hourMin.value, hourMax.value) + dir;

        if (hr > hourMax.value) hr = hourMin.value;
        if (hr < hourMin.value) hr = hourMax.value;
        hour.value = hr;
    } else {
        let m = clamp(minute.value, 0, 59) + dir;

        if (m > 59) m = 0;
        if (m < 0) m = 59;
        minute.value = m;
    }

    commit();
}

function setMeridiem(m) {
    meridiem.value = m;
    commit();
}
</script>
