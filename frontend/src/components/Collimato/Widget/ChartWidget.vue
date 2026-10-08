<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        ref="gridItem"
        class="grid-stack-item"
        :gs-w="chart.position.w"
        :gs-h="chart.position.h"
        :gs-x="chart.position.x"
        :gs-y="chart.position.y"
    >
        <div class="grid-stack-item-content">
            <div
                class="flex flex-col items-center justify-center rounded border bg-white h-full w-full p-5"
            >
                <div v-if="error != null" class="w-full rounded-md bg-red-50 p-3">
                    <div class="flex">
                        <div class="shrink-0">
                            <XCircleIcon class="h-5 w-5 text-red-400" aria-hidden="true" />
                        </div>
                        <div class="ml-3">
                            <p class="text-sm font-medium text-red-800">
                                {{ error }}
                            </p>
                        </div>
                        <div class="ml-auto pl-3">
                            <button
                                type="button"
                                class="-mx-1.5 -my-1.5 inline-flex rounded-md bg-red-50 p-1.5 text-red-500 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-600 focus:ring-offset-2 focus:ring-offset-red-50"
                                @click="error = null"
                            >
                                <span class="sr-only">{{ t("common.button.dismiss") }}</span>
                                <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                            </button>
                        </div>
                    </div>
                </div>
                <nav
                    v-if="drillQueries.length > 1"
                    class="flex"
                    :aria-label="t('collimato.dashboard.widget.breadcrumb')"
                >
                    <ol role="list" class="flex items-center space-x-0">
                        <li
                            v-for="(crumb, index) in breadcrumbs"
                            :key="index"
                            class="cursor-pointer"
                            @click="onBreadcrumbClick(crumb.query, index)"
                        >
                            <div class="flex items-center">
                                <a
                                    class="ml-1 text-xs font-normal text-gray-500 hover:text-gray-700"
                                    ><span v-if="crumb.value" class="font-medium text-gray-700"
                                        >{{ crumb.value }}:
                                    </span>
                                    {{ crumb.name }}</a
                                >
                                <ChevronRightIcon
                                    v-if="index < drillQueries.length - 1"
                                    class="h-5 w-5 flex-shrink-0 text-gray-400"
                                    aria-hidden="true"
                                />
                            </div>
                        </li>
                    </ol>
                </nav>

                <div v-show="loaded" class="flex flex-row w-full justify-between items-center">
                    <h1>{{ props.chart.name }}</h1>
                    <span
                        v-if="busy"
                        class="ml-2 h-4 w-4 shrink-0 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent"
                        :title="t('collimato.dashboard.widget.loading')"
                    ></span>
                    <Popover v-if="activeFilters.length > 0" class="relative">
                        <PopoverButton
                            class="inline-flex items-center gap-x-1 text-sm/6 font-semibold text-gray-900 hover:bg-gray-200 p-2 rounded-full focus-visible:outline focus-visible:outline-0 focus-visible:outline-offset-0"
                        >
                            <FunnelIcon class="h-5 w-5" />
                            <p>{{ activeFilters.length }}</p>
                        </PopoverButton>

                        <transition
                            enter-active-class="transition ease-out duration-200"
                            enter-from-class="opacity-0 translate-y-1"
                            enter-to-class="opacity-100 translate-y-0"
                            leave-active-class="transition ease-in duration-150"
                            leave-from-class="opacity-100 translate-y-0"
                            leave-to-class="opacity-0 translate-y-1"
                        >
                            <PopoverPanel
                                class="absolute -right-40 z-10 mt-1 flex w-screen max-w-min -translate-x-1/2 px-4"
                            >
                                <div
                                    class="w-56 shrink rounded-xl bg-white p-4 text-sm/6 font-semibold text-gray-900 shadow-lg ring-1 ring-gray-900/5"
                                >
                                    <h2 class="text-gray-500 px-2">
                                        {{ t("collimato.dashboard.widget.active_filters") }}
                                    </h2>
                                    <a
                                        v-for="item in activeFilters"
                                        :key="item.name"
                                        class="p-2 hover:text-indigo-600 justify-between flex flex-row items-center"
                                        >{{ item.name }}

                                        <CheckCircleIcon class="h-5 w-5 text-green-500" />
                                    </a>
                                </div>
                            </PopoverPanel>
                        </transition>
                    </Popover>
                </div>
                <div
                    v-show="loaded"
                    class="flex flex-col items-start justify-center h-full w-full overflow-x-auto"
                >
                    <BigNumber
                        v-if="props.chart.chart_type == 'big_number'"
                        :options="currentData"
                    />
                    <MapChart
                        ref="map"
                        v-else-if="props.chart.chart_type == 'map'"
                        :data="props.chart.configuration"
                    />
                    <div
                        v-else
                        ref="chartRef"
                        :style="{
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            width: '100%',
                            height: '100%',
                            padding: '10px', // Adjust padding if needed to prevent overflow
                            boxSizing: 'border-box', // Ensures padding doesn't affect width/height
                            minWidth: '400px', // Ensures overflow on the x-axis
                        }"
                    ></div>
                </div>
                <div v-if="error == null && !loaded">
                    <div class="loader">
                        <div class="dot"></div>
                        <div class="dot"></div>
                        <div class="dot"></div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { t } from "@/i18n/index.js";
import { useEcharts } from "@/composables/collimato/useEcharts.js";
import BigNumber from "@/components/Collimato/Charts/BigNumber.vue";
import MapChart from "@/components/Collimato/Charts/Map/MapChart.vue";
import { ChevronRightIcon, XMarkIcon } from "@heroicons/vue/20/solid";
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { isEchartsType, drillDownKeyFor } from "@/utils/collimato/chartTypes.js";
import { CheckCircleIcon, FunnelIcon, XCircleIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    chart: {
        type: Object,
        required: true,
    },
    activeFilters: {
        type: Array,
        default: () => [],
    },
});

const emits = defineEmits(["right-click", "drill-down"]);
const loaded = ref(false);
const map = ref(null);
const gridItem = ref(null);
const chartRef = ref(null);
const nameClicked = ref(null);
const drillQueries = ref([props.chart.data ? props.chart.data.query : null]);

const currentQuery = computed(() => drillQueries.value.at(-1));
const currentData = ref(props.chart.data);
const error = ref(null);
const busy = ref(false);

const { setData, resize, on } = useEcharts(
    chartRef,
    () => props.chart.chart_type,
    () => props.chart.configuration,
);

function onWidgetResize() {
    resize();
    updatePosition();
}

const updatePosition = () => {
    let x = gridItem.value.getAttribute("gs-x");
    let y = gridItem.value.getAttribute("gs-y");
    let w = gridItem.value.getAttribute("gs-w");
    let h = gridItem.value.getAttribute("gs-h");

    props.chart.position.x = parseInt(x);
    props.chart.position.y = parseInt(y);
    props.chart.position.w = parseInt(w);
    props.chart.position.h = parseInt(h);
};

function breadcrumbName(dimensions) {
    return dimensions.map((dimension) => dimension.split(".").at(-1)).join(", ");
}

function filtersAddedSince(previous, query) {
    const before = (previous?.filters ?? []).map((filter) => JSON.stringify(filter));

    return (query.filters ?? []).filter((filter) => !before.includes(JSON.stringify(filter)));
}

const breadcrumbs = computed(() =>
    drillQueries.value.map((query, index) => ({
        query: query,
        name: breadcrumbName(query.dimensions ?? []),
        value: filtersAddedSince(drillQueries.value[index - 1], query)
            .flatMap((filter) => filter.values ?? [])
            .join(", "),
    })),
);

function onBreadcrumbClick(query, index) {
    drillQueries.value = drillQueries.value.slice(0, index + 1);

    emits("drill-down", { chart: props.chart, query: query });
}

function loadEchart(data) {
    currentData.value = setData(data);
}

function drillDown(query, data) {
    busy.value = false;

    if (
        JSON.stringify(query) != JSON.stringify(drillQueries.value[drillQueries.value.length - 1])
    ) {
        drillQueries.value.push(query);
    }

    loadEchart(data);
}

function updateChart(query, data) {
    currentData.value = data;
    if (isEchartsType(props.chart.chart_type)) {
        loadEchart(data);
    } else if (props.chart.chart_type == "map") {
        map.value.updateLayers(data);
    }

    loaded.value ||= true;
}

function showError(status) {
    busy.value = false;
    error.value = status;
}

function setBusy(on) {
    busy.value = on;
}

on("contextmenu", (params) => {
    params.event.event.preventDefault();
    nameClicked.value = params[drillDownKeyFor(props.chart.chart_type)];

    if (!Array.isArray(nameClicked.value)) {
        nameClicked.value = [nameClicked.value];
    }

    emits("right-click", {
        x: params.event.event.x,
        y: params.event.event.y,
        chart: props.chart,
        name: nameClicked.value,
        query: currentQuery.value,
    });
});

onMounted(() => {
    window.addEventListener("widget_resize", onWidgetResize);
    window.addEventListener("resize", updatePosition);
});

onUnmounted(() => {
    window.removeEventListener("widget_resize", onWidgetResize);
    window.removeEventListener("resize", updatePosition);
});

defineExpose({
    drillDown,
    updateChart,
    showError,
    setBusy,
});
</script>

<style>
.grid-stack > .grid-stack-item > .grid-stack-item-content {
    overflow: visible !important; /* Allow content to overflow horizontally and vertically */
}

.loader {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 80px;
}

.dot {
    width: 16px;
    height: 16px;
    background-color: #4f46e5;
    border-radius: 50%;
    animation: bounce 1.5s infinite ease-in-out;
}

.dot:nth-child(1) {
    animation-delay: 0s;
}

.dot:nth-child(2) {
    animation-delay: 0.2s;
}

.dot:nth-child(3) {
    animation-delay: 0.4s;
}

@keyframes bounce {
    0%,
    80%,
    100% {
        transform: translateY(0);
        opacity: 0.5;
    }
    40% {
        transform: translateY(-15px);
        opacity: 1;
    }
}
</style>
