// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { t } from "@/i18n/index.js";
import { useAlertStore } from "@/store/alerts";

export function useChartEditorMap(map, builder) {
    const {
        mapLayers,
        layerStates,
        hasMapCenter,
        setMapConfiguration,
        addLayer,
        removeLayer,
        setLayerVisible,
        setLayerName,
        setLayerType,
        setLayerModel,
        setLayerConfig,
        setLayerCoordinates,
        setLayerFields,
        addLayerFilter,
        updateLayerFilter,
        removeLayerFilter,
        loadLayer,
        loadAllLayers,
    } = builder;

    const alertStore = useAlertStore();

    const layersLoading = computed(() =>
        layerStates.value.some((layer) => layer.status === "loading"),
    );

    function reloadMapLayer(id) {
        redrawWhenLoaded(loadLayer(id));
    }

    function reloadMapLayers() {
        loadAllLayers().then(reportLayerFailures);
    }

    function reportLayerFailures(results) {
        const failed = results.filter((result) => result.status === "rejected");

        if (failed.length > 0) {
            alertStore.showError(t.value("collimato.charts.new_chart.map.error.load_data_failed"));
        }

        redrawMap();
        fitMapToData();
    }

    function redrawMap() {
        if (map.value) {
            map.value.updateLayers(mapLayers.value);
        }
    }

    function fitMapToData() {
        if (map.value && !hasMapCenter.value) {
            map.value.fitToLayers();
        }
    }

    function addMapLayer(defaults) {
        addLayer(defaults);
        redrawMap();
    }

    function removeMapLayer(id) {
        removeLayer(id);
        redrawMap();
    }

    function toggleMapLayer(id, show) {
        setLayerVisible(id, show);
        redrawMap();
    }

    function updateLayerName(id, name) {
        setLayerName(id, name);
    }

    function updateLayerModel(id, model) {
        setLayerModel(id, model);
        redrawMap();
    }

    function updateLayerType(id, type) {
        setLayerType(id, type);
        redrawMap();
    }

    function redrawWhenLoaded(pending) {
        pending
            .catch(() => {
                alertStore.showError(
                    t.value("collimato.charts.new_chart.map.error.load_data_failed"),
                );
            })
            .finally(() => redrawMap());
    }

    function addMapLayerFilter(id, filter) {
        redrawWhenLoaded(addLayerFilter(id, filter));
    }

    function updateMapLayerFilter(id, member, filter) {
        redrawWhenLoaded(updateLayerFilter(id, member, filter));
    }

    function removeMapLayerFilter(id, filter) {
        redrawWhenLoaded(removeLayerFilter(id, filter));
    }

    function updateLayerFields(id, patch) {
        setLayerFields(id, patch)
            .catch(() => {
                alertStore.showError(
                    t.value("collimato.charts.new_chart.map.error.load_data_failed"),
                );
            })
            .finally(() => {
                redrawMap();
                fitMapToData();
            });
    }

    function updateLayerConfig(id, patch) {
        redrawWhenLoaded(setLayerConfig(id, patch));
    }

    function updateLayerCoordinates(id, patch) {
        setLayerCoordinates(id, patch)
            .catch(() => {
                alertStore.showError(
                    t.value("collimato.charts.new_chart.map.error.load_data_failed"),
                );
            })
            .finally(() => {
                redrawMap();
                fitMapToData();
            });
    }

    function updateMapData(event) {
        setMapConfiguration(event);

        if (map.value) {
            map.value.updateLayers(mapLayers.value);
        }
    }

    return {
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
    };
}
