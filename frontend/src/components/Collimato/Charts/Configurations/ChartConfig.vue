<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ChartConfigTabs v-if="customization.customization" v-model="tab" />

    <div v-if="!customization.customization || tab == 'data'">
        <ChartQuery
            :measures="measures"
            :dimensions="dimensions"
            :orders="orders"
            :cubes="cubes"
            @add-measure="$emit('add-measure', $event)"
            @remove-measure="$emit('remove-measure', $event)"
            @add-dimension="$emit('add-dimension', $event)"
            @remove-dimension="$emit('remove-dimension', $event)"
            @set-order="(name, direction) => $emit('set-order', name, direction)"
            @reorder="$emit('reorder', $event)"
        />

        <ChartFilters
            :filters="filters"
            :cubes="cubes"
            @add-filter="$emit('add-filter', $event)"
            @update-filter="(member, filter) => $emit('update-filter', member, filter)"
            @remove-filter="$emit('remove-filter', $event)"
        />

        <TimeDimensions
            v-if="showTimeDimensions"
            :time="time"
            :cubes="cubes"
            @update:time="$emit('update:time', $event)"
        />
    </div>

    <div v-if="customization.customization && tab == 'customize'">
        <component
            :is="customization.customization"
            v-bind="customization.customizationProps"
            :configuration="configuration"
            :dimensions="dimensions"
            @update:configuration="$emit('update:configuration', $event)"
        />
    </div>
</template>

<script setup>
import { ref, computed } from "vue";
import ChartQuery from "@/components/Collimato/Charts/Sidebar/ChartQuery.vue";
import ChartFilters from "@/components/Collimato/Charts/Sidebar/ChartFilters.vue";
import TimeDimensions from "@/components/Collimato/Charts/Sidebar/TimeDimensions.vue";
import ChartConfigTabs from "@/components/Collimato/Charts/Configurations/ChartConfigTabs.vue";
import { presentationFor } from "@/components/Collimato/Charts/chartTypePresentation.js";
import { hasTimeAxis } from "@/utils/collimato/chartTypes.js";

const props = defineProps({
    chartType: {
        type: String,
        required: true,
    },
    measures: {
        type: Array,
        default: () => [],
    },
    dimensions: {
        type: Array,
        default: () => [],
    },
    orders: {
        type: Array,
        default: () => [],
    },
    cubes: {
        type: Array,
        default: () => [],
    },
    filters: {
        type: Array,
        default: () => [],
    },
    time: {
        type: Object,
        default: () => ({}),
    },
    configuration: {
        type: Object,
        default: () => ({}),
    },
});

defineEmits([
    "add-measure",
    "remove-measure",
    "add-dimension",
    "remove-dimension",
    "set-order",
    "reorder",
    "add-filter",
    "update-filter",
    "remove-filter",
    "update:time",
    "update:configuration",
]);

const tab = ref("data");

const customization = computed(() => presentationFor(props.chartType));

const showTimeDimensions = computed(() => hasTimeAxis(props.chartType));
</script>
