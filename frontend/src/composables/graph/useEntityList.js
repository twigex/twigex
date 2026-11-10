// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";

// useEntityList reads the entities of one type 50 at a time, in key order.
// A new type or search starts again from the first page and cancels a read
// still running, so an older answer can't replace a newer one.
export function useEntityList() {
    const items = ref([]);
    const next = ref("");
    const loading = ref(false);
    const loadingMore = ref(false);
    const errorMessage = ref("");

    let typeKey = "";
    let search = "";
    let running = null;

    function cancelRunning() {
        if (running) {
            running.abort();
            running = null;
        }
    }

    async function readPage(after) {
        cancelRunning();

        const controller = new AbortController();

        running = controller;

        try {
            const response = await graphService.searchEntities(
                typeKey,
                search,
                controller.signal,
                after,
            );

            if (controller.signal.aborted) {
                return;
            }

            if (after) {
                items.value = [...items.value, ...response.data.items];
            } else {
                items.value = response.data.items;
            }

            next.value = response.data.next ?? "";
            errorMessage.value = "";
        } catch (e) {
            if (controller.signal.aborted) {
                return;
            }

            errorMessage.value = extractErrorMessage(e);
        } finally {
            if (running === controller) {
                running = null;
                loading.value = false;
                loadingMore.value = false;
            }
        }
    }

    function start(newTypeKey, newSearch) {
        typeKey = newTypeKey;
        search = newSearch;

        items.value = [];
        next.value = "";
        loading.value = true;
        loadingMore.value = false;

        return readPage("");
    }

    function cancel() {
        cancelRunning();

        loading.value = false;
        loadingMore.value = false;
    }

    function loadMore() {
        if (!next.value || loading.value || loadingMore.value) {
            return;
        }

        loadingMore.value = true;

        return readPage(next.value);
    }

    return {
        items,
        next,
        loading,
        loadingMore,
        errorMessage,
        start,
        loadMore,
        cancel: cancel,
    };
}
