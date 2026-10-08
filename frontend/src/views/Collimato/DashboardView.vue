<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="flex flex-col h-full">
        <ContextMenu
            ref="contextMenu"
            :dimensions="contextMenuDimensions"
            @drill-by="handleDrillBy"
        />

        <DashboardHeader
            :title="dashboard.title"
            :edit="edit"
            :saving="saving"
            @cancel="cancelEdit"
            @save="saveLayout"
            @start-edit="startEdit"
        />

        <div class="flex flex-row w-full flex-1 min-h-0 overflow-hidden">
            <DashboardFilterSidebar
                v-if="!edit && collimatoStore.hasPermissionToViewDashboardFilters"
                v-model:dialog-open="crossFilterDialog"
                :open="sidebarOpen"
                :filters="filters"
                :filter-values="filterValues"
                :local-selections="localSelections"
                :open-filters="openFilters"
                :filter-search="filterSearch"
                :filter-to-edit="filterToEdit"
                :data-models="meta?.cubes ?? []"
                :charts="charts"
                @toggle-sidebar="toggleSidebar"
                @add-filter="addFilter"
                @update-filter="updateFilter"
                @close-dialog="closeCrossFilter"
                @toggle-filter="toggleFilter"
                @edit-filter="editFilter"
                @delete-filter="deleteFilter"
                @toggle-selection="toggleSelection"
                @clear-filter="clearFilter"
                @apply="applyLocalFilters"
            />

            <!-- Grid -->
            <div class="flex flex-col flex-1 min-w-0 overflow-y-auto">
                <div
                    v-show="charts.length > 0"
                    id="grid"
                    class="grid-stack w-full min-h-full bg-gray-50"
                >
                    <ChartWidget
                        v-for="chart in charts"
                        :key="chart.id"
                        :chart="chart"
                        :id="`widget-${chart.id}`"
                        :active-filters="activeFilters(chart.id)"
                        ref="widgets"
                        @right-click="showContextMenu"
                        @drill-down="handleDrillDown"
                    />
                </div>

                <div
                    v-show="charts.length === 0"
                    class="flex flex-col items-center justify-center h-full w-full"
                >
                    <ChartBarIcon class="h-16 w-16 text-gray-300" aria-hidden="true" />
                    <h3 class="mt-3 text-sm font-semibold text-gray-900">
                        {{ t("collimato.dashboard.empty.title") }}
                    </h3>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("collimato.dashboard.empty.description") }}
                    </p>
                </div>
            </div>

            <DashboardChartPanel
                v-if="edit"
                :available-charts="availableCharts"
                :charts="charts"
                @add="addWidget"
                @remove="removeWidget"
            />
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import ChartWidget from "@/components/Collimato/Widget/ChartWidget.vue";
import ContextMenu from "@/components/Collimato/Widget/ContextMenu.vue";
import DashboardHeader from "@/components/Collimato/DashboardHeader.vue";
import DashboardFilterSidebar from "@/components/Collimato/DashboardFilterSidebar.vue";
import DashboardChartPanel from "@/components/Collimato/DashboardChartPanel.vue";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import collimatoService from "@/services/collimatoService.js";
import { ChartBarIcon } from "@heroicons/vue/24/outline";
import { useDashboardQueries } from "@/composables/collimato/useDashboardQueries.js";
import { useDashboardGrid } from "@/composables/collimato/useDashboardGrid.js";
import { useDashboardFilters } from "@/composables/collimato/useDashboardFilters.js";
import { useDashboardDrillDown } from "@/composables/collimato/useDashboardDrillDown.js";

const collimatoStore = useCollimatoStore();
const alertStore = useAlertStore();
const route = useRoute();

const contextMenu = ref(null);
const widgets = ref([]);
const dashboard = ref({});
const charts = ref([]);
const availableCharts = ref([]);
const loaded = ref(false);
const meta = ref(null);
const sidebarOpen = ref(true);

const { pollLoad, loadMapLayer, loadChartData, loadData } = useDashboardQueries({ widgets });

const {
    edit,
    saving,
    loadGrid,
    resizeWidget,
    addWidget,
    removeWidget,
    startEdit,
    saveLayout,
    cancelEdit,
} = useDashboardGrid({
    dashboard,
    charts,
    widgets,
    loaded,
    loadMapLayer,
    loadChartData,
});

const {
    crossFilterDialog,
    filters,
    filterToEdit,
    filterValues,
    openFilters,
    localSelections,
    filterSearch,
    initLocalSelections,
    toggleFilter,
    toggleSelection,
    clearFilter,
    applyLocalFilters,
    addFilter,
    updateFilter,
    deleteFilter,
    applyFilters,
    activeFilters,
    closeCrossFilter,
    editFilter,
    loadFilterValues,
} = useDashboardFilters({ charts, pollLoad, loadData });

const { contextMenuDimensions, showContextMenu, handleDrillDown, handleDrillBy } =
    useDashboardDrillDown({ charts, widgets, meta, contextMenu, pollLoad });

function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value;
    setTimeout(() => resizeWidget(), 300);
}

onMounted(async () => {
    if (!collimatoStore.hasPermissionToViewDashboards) {
        alertStore.showError(t.value("collimato.dashboards.error.no_permission_view"));

        return;
    }

    try {
        await collimatoService.meta(route.params.workspaceId).then((response) => {
            meta.value = response.data;
        });

        if (collimatoStore.hasPermissionToViewCharts) {
            await collimatoService.getCharts(route.params.workspaceId).then((response) => {
                availableCharts.value = response.data;
            });
        }

        await collimatoService
            .getDashboardById(route.params.workspaceId, route.params.id)
            .then((response) => {
                dashboard.value = response.data;
                charts.value = response.data.charts;
                filters.value = response.data.filters;
                initLocalSelections();
                loadFilterValues();
                setTimeout(() => {
                    loadGrid();
                    applyFilters();
                }, 10);
            });
    } catch (error) {
        alertStore.showError(extractErrorMessage(error));
    } finally {
        loaded.value = true; // never leave the page stuck on the spinner
    }
});
</script>
