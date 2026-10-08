// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, reactive } from "vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useCollimatoStore } from "@/store/collimato";
import { useAlertStore } from "@/store/alerts";
import collimatoService from "@/services/collimatoService.js";
import { extractErrorMessage } from "@/utils/errors";
import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";
import {
    filterApplies,
    filtersForChart,
    withFilterApplied,
    withFilterRemoved,
} from "@/utils/collimato/dashboardFilters.js";

export function useDashboardFilters({ charts, pollLoad, loadData }) {
    const route = useRoute();
    const collimatoStore = useCollimatoStore();
    const alertStore = useAlertStore();

    const crossFilterDialog = ref(false);
    const filters = ref([]);
    const filterToEdit = ref(null);
    const filterValues = ref({});

    const openFilters = reactive({});
    const localSelections = reactive({});
    const filterSearch = reactive({});

    function initLocalSelections() {
        filters.value.forEach((filter) => {
            localSelections[filter.id] = [...(filter.values ?? [])];
            filterSearch[filter.id] = "";
            openFilters[filter.id] = true;
        });
    }

    function toggleFilter(id) {
        openFilters[id] = !openFilters[id];
    }

    function toggleSelection(filterId, val) {
        const current = localSelections[filterId] ?? [];
        const idx = current.indexOf(val);

        if (idx >= 0) {
            current.splice(idx, 1);
        } else {
            current.push(val);
        }

        localSelections[filterId] = [...current];
    }

    function clearFilter(filterId) {
        localSelections[filterId] = [];
    }

    function applyLocalFilters() {
        filters.value.forEach((filter) => {
            const newValues = localSelections[filter.id] ?? [];

            filter.values = [...newValues];

            collimatoService
                .updateDashboardFilter(route.params.workspaceId, route.params.id, filter.id, filter)
                .catch((error) => alertStore.showError(extractErrorMessage(error)));

            charts.value.forEach((chart, i) => {
                if (chart.chart_type === "map") return;
                if (!filterApplies(filter, chart.id)) return;

                chart.data.query.filters = withFilterApplied(
                    chart.data.query.filters,
                    filter,
                    newValues,
                );

                loadData(chart, i);
            });
        });
    }

    function addFilter(filter) {
        collimatoService
            .createDashboardFilter(route.params.workspaceId, route.params.id, filter)
            .then((result) => {
                filters.value.push(result.data);
                localSelections[result.data.id] = [...(result.data.values ?? [])];
                filterSearch[result.data.id] = "";
                openFilters[result.data.id] = true;
                loadFilterData(result.data);

                charts.value.forEach((chart, i) => {
                    if (chart.chart_type === "map") return;
                    if (!filterApplies(filter, chart.id)) return;

                    chart.data.query.filters = withFilterApplied(
                        chart.data.query.filters,
                        filter,
                        filter.values,
                    );
                    loadData(chart, i);
                });
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });

        closeCrossFilter();
    }

    function updateFilter(data) {
        const { id, filter: newFilter } = data;

        filters.value = filters.value.map((f) => (f.id === id ? { ...f, ...newFilter } : f));

        collimatoService
            .updateDashboardFilter(route.params.workspaceId, route.params.id, id, newFilter)
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });

        charts.value.forEach((chart, i) => {
            if (chart.chart_type === "map") return;
            if (!filterApplies(newFilter, chart.id)) return;

            chart.data.query.filters = withFilterApplied(
                withFilterRemoved(chart.data.query.filters, filterToEdit.value),
                newFilter,
                newFilter.values,
            );

            loadData(chart, i);
        });

        closeCrossFilter();
    }

    function deleteFilter(filter) {
        filters.value = filters.value.filter((f) => f.id !== filter.id);
        delete localSelections[filter.id];
        delete filterSearch[filter.id];
        delete openFilters[filter.id];

        collimatoService
            .deleteDashboardFilter(route.params.workspaceId, route.params.id, filter.id)
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });

        charts.value.forEach((chart, i) => {
            if (chart.chart_type === "map") return;
            if (!filterApplies(filter, chart.id)) return;

            chart.data.query.filters = withFilterRemoved(chart.data.query.filters, filter);

            loadData(chart, i);
        });
    }

    function applyFilter(chart, index) {
        if (chart.chart_type === "map") {
            loadData(chart, index);

            return;
        }

        filters.value.forEach((filter) => {
            if (!filterApplies(filter, chart.id)) return;

            const alreadyApplied = chart.data.query.filters.some((f) => f.member === filter.column);

            if (!alreadyApplied && filter.values.length > 0) {
                chart.data.query.filters = withFilterApplied(
                    chart.data.query.filters,
                    filter,
                    filter.values,
                );
            }
        });

        loadData(charts.value[index], index);
    }

    function applyFilters() {
        charts.value.forEach((chart, i) => applyFilter(chart, i));
    }

    function activeFilters(chartID) {
        return filtersForChart(filters.value, chartID);
    }

    function closeCrossFilter() {
        crossFilterDialog.value = false;
        filterToEdit.value = null;
    }

    function editFilter(filter) {
        if (!collimatoStore.hasPermissionToEditDashboardFilters) {
            alertStore.showError(t.value("collimato.dashboard.error.no_permission_edit_filters"));

            return;
        }

        filterToEdit.value = filter;
        crossFilterDialog.value = true;
    }

    function loadFilterData(filter) {
        const filterQuery = {
            measures: [filter.table + ".count"],
            dimensions: [filter.column],
            filters: [],
            timeDimensions: [],
            order: [],
            limit: MAX_QUERY_LIMIT,
        };

        pollLoad(filterQuery, {
            onSuccess: (result) => {
                filterValues.value[filter.id] = result.data.data.map(
                    (value) => value[filter.column],
                );
            },
            onError: (message) => {
                // Clear the filter's stuck "loading" state (F8) and surface the error.
                filterValues.value[filter.id] = [];
                alertStore.showError(message);
            },
        });
    }

    function loadFilterValues() {
        filters.value.forEach((filter) => loadFilterData(filter));
    }

    return {
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
    };
}
