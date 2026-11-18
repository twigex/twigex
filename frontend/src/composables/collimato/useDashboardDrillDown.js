// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { transformData } from "@/utils/collimato/charts/chartDrawing.js";
import { drillDownFilters, filterMembers } from "@/utils/collimato/charts/drillDown.js";
import { useChartDisplay } from "@/composables/collimato/useChartDisplay.js";

export function useDashboardDrillDown({ charts, widgets, dataModels, contextMenu, loadQuery }) {
    const { chartDisplay } = useChartDisplay();
    const drillDownQuery = ref(null);
    const drillDownData = ref(null);
    const valuesClicked = ref(null);
    const currentChart = ref(null);
    const contextMenuDimensions = ref([]);

    function showContextMenu(event) {
        currentChart.value = event.chart;
        drillDownQuery.value = event.query ?? event.chart.data.query;
        valuesClicked.value = event.values;
        contextMenuDimensions.value = createContextMenuDimensions();
        contextMenu.value.showContextMenu(event);
    }

    function createContextMenuDimensions() {
        if (drillDownQuery.value === null) return [];

        const used = [];
        const dimensions = [];
        const names = [];

        drillDownQuery.value.measures.forEach((measure) => {
            const name = measure.split(".")[0];

            if (!names.includes(name)) names.push(name);
        });

        drillDownQuery.value.dimensions.forEach((dimension) => {
            const name = dimension.split(".")[0];

            if (!names.includes(name)) names.push(name);
            if (!used.includes(dimension)) used.push(dimension);
        });

        drillDownQuery.value.filters.forEach((filter) => {
            filterMembers(filter).forEach((member) => {
                if (!used.includes(member)) used.push(member);
            });
        });

        const models = dataModels.value ?? [];
        const root = models.find((model) => model.name === names[0]);

        if (!root) {
            return [];
        }

        const related = root.join > 0 ? models.filter((model) => model.join === root.join) : [root];

        related.forEach((model) => {
            (model.dimensions ?? []).forEach((dimension) => {
                if (!used.includes(dimension.name)) {
                    dimensions.push(dimension);
                }
            });
        });

        return dimensions;
    }

    function handleDrillDown({ chart, query }) {
        const chartIndex = charts.value.findIndex((c) => c.id === chart.id);

        widgets.value[chartIndex]?.setBusy(true);

        loadQuery(query, {
            onSuccess: (data) => {
                const display = chartDisplay();

                drillDownData.value = transformData(
                    chart.chart_type,
                    data,
                    chart.configuration,
                    display,
                );
                widgets.value[chartIndex].drillDown(query, drillDownData.value);
            },
            onError: (message) => widgets.value[chartIndex].showError(message),
        });
    }

    function handleDrillBy(item) {
        const chartIndex = charts.value.findIndex((c) => c.id === currentChart.value.id);
        const oldQuery = JSON.parse(JSON.stringify(drillDownQuery.value));
        const clickedFilters = drillDownFilters(valuesClicked.value);

        drillDownQuery.value = {
            dimensions: [item],
            measures: [...oldQuery.measures],
            filters: [...oldQuery.filters, ...clickedFilters],
            order: [...oldQuery.order],
            timeDimensions: [...oldQuery.timeDimensions],
            limit: oldQuery.limit,
        };

        const query = drillDownQuery.value;

        widgets.value[chartIndex]?.setBusy(true);

        loadQuery(query, {
            onSuccess: (data) => {
                const display = chartDisplay();

                drillDownData.value = transformData(
                    currentChart.value.chart_type,
                    data,
                    currentChart.value.configuration,
                    display,
                );
                widgets.value[chartIndex].drillDown(query, drillDownData.value);
            },
            onError: (message) => widgets.value[chartIndex].showError(message),
        });
    }

    return {
        contextMenuDimensions,
        showContextMenu,
        handleDrillDown,
        handleDrillBy,
    };
}
