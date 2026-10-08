// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { onUnmounted } from "vue";
import { useRoute } from "vue-router";
import collimatoService from "@/services/collimatoService.js";
import { extractErrorMessage } from "@/utils/errors";
import { transformData } from "@/utils/collimato/chartTypes.js";
import { pollQuery } from "@/utils/collimato/queryPolling.js";
import { layerQuery } from "@/utils/collimato/mapQuery.js";

export function useDashboardQueries({ widgets }) {
    const route = useRoute();

    const pollAborts = new Set();

    onUnmounted(() => {
        pollAborts.forEach((controller) => controller.abort());
        pollAborts.clear();
    });

    // pollLoad runs a Cube query and handles Cube's "Continue wait" long-poll with a
    // bounded, delayed retry (matching DataView/NewChartView), routing the real
    // result or a real error message to the caller's handlers.
    function pollLoad(query, { onSuccess, onError }) {
        const controller = new AbortController();

        pollAborts.add(controller);

        pollQuery(
            (q, signal) => collimatoService.loadData(route.params.workspaceId, q, signal),
            query,
            { signal: controller.signal },
        )
            .then(onSuccess)
            .catch((error) => {
                if (controller.signal.aborted) return;
                onError(extractErrorMessage(error));
            })
            .finally(() => pollAborts.delete(controller));
    }

    function loadMapLayer(layer, lastWidgetIndex, item) {
        const mapQuery = layerQuery(layer);

        pollLoad(mapQuery, {
            onSuccess: (result) => {
                layer.config.data = result.data.data;
                widgets.value[lastWidgetIndex]?.updateChart(null, item.configuration.layers);
            },
            onError: (message) => widgets.value[lastWidgetIndex]?.showError(message),
        });
    }

    function loadChartData(item, lastWidgetIndex) {
        pollLoad(item.data.query, {
            onSuccess: (result) => {
                const data = transformData(item.chart_type, result.data, item.configuration);

                widgets.value[lastWidgetIndex]?.updateChart(item.data.query, data);
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
        pollLoad,
        loadMapLayer,
        loadChartData,
        loadData,
    };
}
