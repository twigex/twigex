// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { onUnmounted } from "vue";
import { useRoute } from "vue-router";
import collimatoService from "@/services/collimatoService.js";
import { extractErrorMessage } from "@/utils/errors";
import { transformData } from "@/utils/collimato/charts/chartDrawing.js";
import { useChartDisplay } from "@/composables/collimato/useChartDisplay.js";
import { fetchChartData } from "@/utils/collimato/chartData.js";
import { layerQuery } from "@/utils/collimato/mapQuery.js";

export function useDashboardQueries({ widgets }) {
    const route = useRoute();
    const { chartDisplay } = useChartDisplay();

    const loadAborts = new Set();

    onUnmounted(() => {
        loadAborts.forEach((controller) => controller.abort());
        loadAborts.clear();
    });

    // loadQuery loads a query's data and gives it, or the reason it failed, to
    // the caller's handlers. Leaving the dashboard stops every load.
    function loadQuery(query, { onSuccess, onError }) {
        const controller = new AbortController();

        loadAborts.add(controller);

        fetchChartData(
            (q, signal) => collimatoService.loadData(route.params.workspaceId, q, signal),
            query,
            { signal: controller.signal },
        )
            .then(onSuccess)
            .catch((error) => {
                if (controller.signal.aborted) return;
                onError(extractErrorMessage(error));
            })
            .finally(() => loadAborts.delete(controller));
    }

    function loadMapLayer(layer, lastWidgetIndex, item) {
        const mapQuery = layerQuery(layer);

        loadQuery(mapQuery, {
            onSuccess: (data) => {
                layer.config.data = data.rows;
                widgets.value[lastWidgetIndex]?.updateChart(null, item.configuration.layers);
            },
            onError: (message) => widgets.value[lastWidgetIndex]?.showError(message),
        });
    }

    function loadChartData(item, lastWidgetIndex) {
        loadQuery(item.data.query, {
            onSuccess: (data) => {
                const display = chartDisplay();
                const option = transformData(item.chart_type, data, item.configuration, display);

                widgets.value[lastWidgetIndex]?.updateChart(item.data.query, option);
            },
            onError: (message) => widgets.value[lastWidgetIndex]?.showError(message),
        });
    }

    function loadData(chart, i) {
        if (chart.chart_type === "map") {
            chart.configuration.layers.forEach((layer) => loadMapLayer(layer, i, chart));
        } else {
            loadChartData(chart, i);
        }
    }

    return {
        loadQuery,
        loadMapLayer,
        loadChartData,
        loadData,
    };
}
