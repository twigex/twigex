// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import collimatoService from "@/services/collimatoService.js";
import { useCollimatoStore } from "@/store/collimato";
import { extractErrorMessage } from "@/utils/errors";
import { toTableData } from "@/utils/collimato/tableData.js";
import { transformData } from "@/utils/collimato/charts/chartDrawing.js";
import { fetchChartData, withoutDateAxisInFields } from "@/utils/collimato/chartData.js";
import { useChartDisplay } from "@/composables/collimato/useChartDisplay.js";

export function useChartEditorQuery({ query, chartType, chartOptions, onError }) {
    const route = useRoute();
    const collimatoStore = useCollimatoStore();
    const { chartDisplay } = useChartDisplay();

    const loading = ref(false);
    const queryResult = ref(null);
    const sqlPreview = ref("");
    const tableData = ref({
        headers: [],
        data: [],
    });

    const options = computed(() => {
        if (!queryResult.value) {
            return null;
        }

        const display = chartDisplay();

        return transformData(chartType.value, queryResult.value, chartOptions.value, display);
    });

    let loadAbort = null;

    // Cancelling the request also stops the query in the database.
    function cancelLoad() {
        if (loadAbort) loadAbort.abort();
        loading.value = false;
    }

    async function load() {
        loading.value = true;
        loadAbort = new AbortController();

        try {
            const data = await fetchChartData(
                (q, signal) => collimatoService.loadData(route.params.workspaceId, q, signal),
                query.value,
                { signal: loadAbort.signal },
            );

            queryResult.value = data;
            tableData.value = toTableData(data);

            if (!collimatoStore.hasPermissionToViewDataModels) return;

            collimatoService
                .previewSql(route.params.workspaceId, withoutDateAxisInFields(query.value))
                .then((response) => {
                    sqlPreview.value = response.data.sql;
                })
                .catch((error) => {
                    console.error(error?.response?.data ?? error);
                });
        } catch (err) {
            if (loadAbort?.signal.aborted || err?.code === "ERR_CANCELED") {
                return;
            }

            onError(extractErrorMessage(err));
        } finally {
            loading.value = false;
        }
    }

    onUnmounted(() => {
        if (loadAbort) loadAbort.abort();
    });

    return {
        loading,
        sqlPreview,
        tableData,
        options,
        load,
        cancelLoad,
    };
}
