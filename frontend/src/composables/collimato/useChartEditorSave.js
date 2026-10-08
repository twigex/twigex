// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRoute, useRouter } from "vue-router";
import { t } from "@/i18n/index.js";
import { useAlertStore } from "@/store/alerts";
import collimatoService from "@/services/collimatoService.js";
import { extractErrorMessage } from "@/utils/errors";

export function useChartEditorSave({ name, v$, builder }) {
    const {
        chartType,
        selectedModel,
        measures,
        dimensions,
        query,
        chartOptions,
        mapConfiguration,
    } = builder;

    const route = useRoute();
    const router = useRouter();
    const alertStore = useAlertStore();

    function create() {
        v$.value.name.$touch();

        if (v$.value.name.$error) {
            return;
        }

        if (chartType.value == "map") {
            route.name == "edit-chart" ? updateMap() : createMap();

            return;
        }

        if (!selectedModel.value) {
            alertStore.showError(t.value("collimato.charts.new_chart.error.select_model"));

            return;
        }

        if (measures.value.length == 0 && dimensions.value.length == 0) {
            alertStore.showError(t.value("collimato.charts.new_chart.error.no_query"));

            return;
        }

        route.name == "edit-chart" ? updateChart() : createChart();
    }

    function createMap() {
        let data = JSON.parse(JSON.stringify(mapConfiguration.value));

        data.layers.forEach((layer) => {
            layer.config.data = null;
        });

        collimatoService
            .createChart(route.params.workspaceId, {
                name: name.value,
                query: null,
                chart_type: chartType.value,
                configuration: data,
                model: "map",
            })
            .then(() => {
                alertStore.showSuccess(t.value("collimato.charts.new_chart.success.chart_created"));
                router.push({ name: "charts" });
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });
    }

    function updateMap() {
        let data = JSON.parse(JSON.stringify(mapConfiguration.value));

        if (chartType.value == "map") {
            data.layers.forEach((layer) => {
                layer.config.data = null;
            });
        }

        collimatoService
            .updateChart(route.params.workspaceId, route.params.id, {
                name: name.value,
                query: null,
                chart_type: chartType.value,
                configuration: data,
                model: "map",
            })
            .then(() => {
                alertStore.showSuccess(t.value("collimato.charts.new_chart.success.chart_updated"));
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });
    }

    function createChart() {
        collimatoService
            .createChart(route.params.workspaceId, {
                name: name.value,
                query: query.value,
                chart_type: chartType.value,
                configuration: chartOptions.value,
                model: selectedModel.value.table,
            })
            .then(() => {
                alertStore.showSuccess(t.value("collimato.charts.new_chart.success.chart_created"));
                router.push({ name: "charts" });
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });
    }

    function updateChart() {
        collimatoService
            .updateChart(route.params.workspaceId, route.params.id, {
                name: name.value,
                query: query.value,
                chart_type: chartType.value,
                configuration: chartOptions.value,
                model: selectedModel.value.table,
            })
            .then(() => {
                alertStore.showSuccess(t.value("collimato.charts.new_chart.success.chart_updated"));
            })
            .catch((error) => {
                alertStore.showError(extractErrorMessage(error));
            });
    }

    return {
        create,
    };
}
