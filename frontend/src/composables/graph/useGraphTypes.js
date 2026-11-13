// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, provide, inject } from "vue";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";

const GRAPH_TYPES = Symbol("graphTypes");

// The Graph section reads the types each time it opens and shares them with
// its sidebar and pages. They are not kept between visits: an admin's list
// names datasets and fields that another user signing in later must not see.
export function provideGraphTypes() {
    const types = ref({ entity_types: [], relation_types: [] });
    const errorMessage = ref("");

    async function read() {
        try {
            const response = await graphService.getTypes();

            types.value = response.data;
            errorMessage.value = "";
        } catch (e) {
            errorMessage.value = extractErrorMessage(e);
        }
    }

    // After a type is saved or deleted, the list is read again so every page
    // and the sidebar show the change.
    const shared = { types: types, errorMessage: errorMessage, ready: read(), reload: read };

    provide(GRAPH_TYPES, shared);

    return shared;
}

export function useGraphTypes() {
    return inject(GRAPH_TYPES);
}
