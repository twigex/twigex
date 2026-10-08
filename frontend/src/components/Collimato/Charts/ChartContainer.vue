<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div ref="chartRef" :style="{ width: '90%', height: '100%' }"></div>
</template>

<script setup>
import { ref, onMounted, watch } from "vue";
import { useEcharts } from "@/composables/collimato/useEcharts.js";

const props = defineProps({
    option: {
        type: Object,
        required: true,
    },
    chartType: {
        type: String,
        required: true,
    },
    chartOptions: {
        type: Object,
        default: () => ({}),
    },
});

const chartRef = ref(null);

const { setData } = useEcharts(
    chartRef,
    () => props.chartType,
    () => props.chartOptions,
);

onMounted(() => setData(props.option));

watch(() => props.option, setData, { deep: true });

watch(
    () => props.chartOptions,
    () => setData(props.option),
    { deep: true },
);
</script>
