<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="!loaded" class="w-full h-full flex items-center justify-center">
        <div
            class="h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
        ></div>
    </div>
    <div v-else class="w-full h-full flex flex-col">
        <ChartEditorToolbar
            v-model:name="name"
            :name-error="v$.name.$error"
            :show-sql-button="chartType !== 'map' && collimatoStore.hasPermissionToViewDataModels"
            :sql-preview="sqlPreview"
            :is-edit="route.name === 'edit-chart'"
            :show-save-button="
                route.name !== 'edit-chart' || collimatoStore.hasPermissionToEditCharts
            "
            @save="create"
        />
        <div class="flex flex-1 min-h-0">
            <!-- LEFT SIDEBAR -->
            <div class="w-1/3 border-r border-gray-200 bg-white flex flex-col">
                <div class="flex-1 min-h-0 overflow-y-auto">
                    <ChartEditorTypePicker
                        :charts="charts"
                        :chart-type="chartType"
                        @select="setChartType"
                    />

                    <ChartEditorModelPicker
                        v-if="chartType !== 'map'"
                        :data-models="dataModels"
                        :selected-model="selectedModel"
                        @select="selectModel"
                    />

                    <!-- Chart config -->
                    <div class="border-t border-gray-200">
                        <ChartConfig
                            v-if="chartType !== 'map'"
                            :chart-type="chartType"
                            :measures="measures"
                            :dimensions="dimensions"
                            :orders="orders"
                            :cubes="availableDataModels"
                            :filters="filters"
                            :time="time"
                            :configuration="chartOptions"
                            @add-measure="onAddMeasure"
                            @remove-measure="removeMeasure"
                            @add-dimension="onAddDimension"
                            @remove-dimension="removeDimension"
                            @set-order="setOrder"
                            @reorder="reorderOrders"
                            @add-filter="addFilter"
                            @update-filter="updateFilter"
                            @remove-filter="removeFilter"
                            @update:time="setTime"
                            @update:configuration="chartOptions = $event"
                        />

                        <MapOptions
                            v-else
                            :chartOptions="mapConfiguration"
                            :cubes="dataModels"
                            :statuses="layerStates"
                            @update="updateMapData"
                            @add-layer="addMapLayer"
                            @remove-layer="removeMapLayer"
                            @toggle-layer="toggleMapLayer"
                            @update-layer-name="updateLayerName"
                            @update-layer-model="updateLayerModel"
                            @update-layer-type="updateLayerType"
                            @update-layer-config="updateLayerConfig"
                            @update-layer-coordinates="updateLayerCoordinates"
                            @update-layer-fields="updateLayerFields"
                            @reload-layer="reloadMapLayer"
                            @add-layer-filter="addMapLayerFilter"
                            @update-layer-filter="updateMapLayerFilter"
                            @remove-layer-filter="removeMapLayerFilter"
                        />
                    </div>
                </div>

                <ChartEditorRunActions
                    :is-map="chartType === 'map'"
                    :loading="loading"
                    :layers-loading="layersLoading"
                    :has-layers="mapLayers.length > 0"
                    @cancel-load="cancelLoad"
                    @load="load"
                    @cancel-layers="cancelAllLayerLoads"
                    @reload-layers="reloadMapLayers"
                />
            </div>

            <div class="flex-1 flex flex-col min-h-0 bg-gray-50">
                <!-- MAP: full-height chart, no disclosure -->
                <div
                    v-if="chartType === 'map'"
                    class="relative flex flex-1 min-h-0 items-center justify-center bg-white border-b border-gray-200"
                >
                    <div
                        v-if="loading"
                        class="absolute inset-0 z-10 flex items-center justify-center bg-white/75"
                    >
                        <div
                            class="h-10 w-10 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent"
                        ></div>
                    </div>

                    <MapChart ref="map" :data="mapConfiguration" />
                </div>

                <!-- NON-MAP: collapsible chart + results panes -->
                <ChartEditorPreview
                    v-else
                    v-model:chart-open="chartOpen"
                    v-model:results-open="resultsOpen"
                    :loading="loading"
                    :options="options"
                    :chart-type="chartType"
                    :chart-options="chartOptions"
                    :chart-icon="selectedChartType.icon"
                    :table-data="tableData"
                />
            </div>
        </div>
        <ErrorDialog v-model="errorDialog" :message="error" />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted, onUnmounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import MapChart from "@/components/Collimato/Charts/Map/MapChart.vue";
import ErrorDialog from "@/components/Collimato/Dialogs/ErrorDialog.vue";
import ChartEditorToolbar from "@/components/Collimato/ChartEditorToolbar.vue";
import ChartEditorTypePicker from "@/components/Collimato/ChartEditorTypePicker.vue";
import ChartEditorModelPicker from "@/components/Collimato/ChartEditorModelPicker.vue";
import ChartEditorRunActions from "@/components/Collimato/ChartEditorRunActions.vue";
import ChartEditorPreview from "@/components/Collimato/ChartEditorPreview.vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useCollimatoStore } from "@/store/collimato";
import { CHART_TYPE_IDS } from "@/utils/collimato/chartTypes.js";
import { presentationFor } from "@/components/Collimato/Charts/chartTypePresentation.js";
import { useChartBuilder } from "@/composables/collimato/useChartBuilder.js";
import { useChartEditorQuery } from "@/composables/collimato/useChartEditorQuery.js";
import { useChartEditorMap } from "@/composables/collimato/useChartEditorMap.js";
import { useChartEditorSave } from "@/composables/collimato/useChartEditorSave.js";
import collimatoService from "@/services/collimatoService.js";
import ChartConfig from "@/components/Collimato/Charts/Configurations/ChartConfig.vue";
import MapOptions from "@/components/Collimato/Charts/Map/MapOptions.vue";

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();
const collimatoStore = useCollimatoStore();

const errorDialog = ref(false);
const loaded = ref(false);
const chartOpen = ref(true);
const resultsOpen = ref(true);
const charts = computed(() =>
    CHART_TYPE_IDS.map((id) => ({
        value: id,
        icon: presentationFor(id).icon,
        name: t.value(presentationFor(id).labelKey),
        shortName: t.value(presentationFor(id).shortLabelKey),
    })),
);

const map = ref(null);

const name = ref("");

const builder = useChartBuilder(charts.value[0].value, {
    loadData: (query, signal) => collimatoService.loadData(route.params.workspaceId, query, signal),
});

const {
    chartType,
    dataModels,
    selectedModel,
    measures,
    measureLimit,
    dimensions,
    orders,
    filters,
    time,
    chartOptions,
    availableDataModels,
    query,
    addMeasure,
    removeMeasure,
    addDimension,
    removeDimension,
    setOrder,
    reorderOrders,
    addFilter,
    updateFilter,
    removeFilter,
    setTime,
    mapConfiguration,
    mapLayers,
    layerStates,
    setMapConfiguration,
    loadAllLayers,
    cancelAllLayerLoads,
    selectModel,
    setChartType,
    loadQuery,
} = builder;

const selectedChartType = computed(
    () => charts.value.find((c) => c.value == chartType.value) ?? charts.value[0],
);

const error = ref("");

const { loading, sqlPreview, tableData, options, load, cancelLoad } = useChartEditorQuery({
    query,
    chartType,
    chartOptions,
    onError: (message) => {
        error.value = message;
        errorDialog.value = true;
    },
});

const {
    layersLoading,
    reloadMapLayer,
    reloadMapLayers,
    reportLayerFailures,
    addMapLayer,
    removeMapLayer,
    toggleMapLayer,
    updateLayerName,
    updateLayerModel,
    updateLayerType,
    addMapLayerFilter,
    updateMapLayerFilter,
    removeMapLayerFilter,
    updateLayerFields,
    updateLayerConfig,
    updateLayerCoordinates,
    updateMapData,
} = useChartEditorMap(map, builder);

const rules = {
    name: { required },
};
const v$ = useVuelidate(rules, { name });

const { create } = useChartEditorSave({ name, v$, builder });

onMounted(async () => {
    if (
        route.name == "new-chart" &&
        (!collimatoStore.hasPermissionToViewCharts || !collimatoStore.hasPermissionToCreateCharts)
    ) {
        alertStore.showError(t.value("collimato.charts.errors.no_permission_create_chart"));
        router.push({ name: "charts" });

        return;
    }

    if (route.name == "edit-chart" && !collimatoStore.hasPermissionToViewCharts) {
        alertStore.showError(t.value("collimato.charts.errors.no_permission_view_chart"));
        router.push({ name: "charts" });

        return;
    }

    try {
        if (!(await loadDataModels())) {
            return;
        }

        if (route.name != "edit-chart") {
            loaded.value = true;

            return;
        }

        const response = await collimatoService.getChartById(
            route.params.workspaceId,
            route.params.id,
        );

        name.value = response.data.name;

        if (response.data.chart_type != "map") {
            loadChart(response);
        } else {
            loadMap(response);
        }
    } catch (err) {
        alertStore.showError(extractErrorMessage(err));
        router.push({ name: "charts" });
    }
});

async function loadDataModels() {
    const response = await collimatoService.meta(route.params.workspaceId);

    if (response.data.error) {
        error.value = response.data.error;
        errorDialog.value = true;
        loaded.value = true;

        return false;
    }

    dataModels.value = response.data.cubes.map((model) => ({
        table: model.name,
        join: model.connectedComponent,
        dimensions: model.dimensions.map((dimension) => ({
            name: dimension.name,
            title: dimension.shortTitle,
            type: dimension.type,
        })),
        measures: model.measures.map((measure) => ({
            name: measure.name,
            title: measure.shortTitle,
            type: measure.type,
        })),
    }));

    return true;
}

onUnmounted(() => {
    cancelAllLayerLoads();
});

function onAddMeasure(field) {
    if (measureLimit.value != null && measures.value.length >= measureLimit.value) {
        alertStore.showError(
            t.value("collimato.charts.errors.measure_limit", {
                n: measureLimit.value,
            }),
        );

        return;
    }

    if (!addMeasure(field)) {
        alertStore.showError(t.value("collimato.charts.errors.measure_already_added"));
    }
}

function onAddDimension(field) {
    if (!addDimension(field)) {
        alertStore.showError(t.value("collimato.charts.errors.dimension_already_added"));
    }
}

function loadMap(response) {
    setChartType(response.data.chart_type);
    setMapConfiguration(response.data.configuration);

    loaded.value = true;

    loadAllLayers().then(reportLayerFailures);
}

function loadChart(response) {
    loadQuery(response.data.data.query, response.data.data.model);

    setChartType(response.data.chart_type);
    chartOptions.value = response.data.configuration;

    loaded.value = true;

    load();
}
</script>
