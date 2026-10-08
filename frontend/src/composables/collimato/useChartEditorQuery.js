// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed, onUnmounted } from "vue";
import { useRoute } from "vue-router";
import collimatoService from "@/services/collimatoService.js";
import { useCollimatoStore } from "@/store/collimato";
import { extractErrorMessage } from "@/utils/errors";
import { toTableData } from "@/utils/collimato/chartUtils.js";
import { transformData } from "@/utils/collimato/chartTypes.js";
import { pollQuery } from "@/utils/collimato/queryPolling.js";

export function useChartEditorQuery({ query, chartType, chartOptions, onError }) {
    const route = useRoute();
    const collimatoStore = useCollimatoStore();

    const loading = ref(false);
    const queryResult = ref(null);
    const sqlPreview = ref("");
    const tableData = ref({
        headers: [],
        data: [],
    });

    const options = computed(() =>
        queryResult.value
            ? transformData(chartType.value, queryResult.value, chartOptions.value)
            : null,
    );

    let loadAbort = null;

    // The DB query keeps running server-side until CUBEJS_DB_QUERY_TIMEOUT.
    function cancelLoad() {
        if (loadAbort) loadAbort.abort();
        loading.value = false;
    }

    async function load() {
        loading.value = true;
        loadAbort = new AbortController();

        try {
            const result = await pollQuery(
                (q, signal) => collimatoService.loadData(route.params.workspaceId, q, signal),
                query.value,
                { signal: loadAbort.signal },
            );

            queryResult.value = result.data;
            tableData.value = toTableData(result.data);

            if (!collimatoStore.hasPermissionToViewDataModels) return;

            collimatoService
                .previewSql(route.params.workspaceId, query.value)
                .then((response) => {
                    sqlPreview.value = response.data.sql.sql[0];
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
