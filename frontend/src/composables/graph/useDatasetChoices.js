// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import graphService from "@/services/graphService";
import { extractErrorMessage } from "@/utils/errors";

// useDatasetChoices reads the datasets a graph type can be made from, with
// their fields, for the type forms. A failed read is shown in errorMessage.
export function useDatasetChoices(errorMessage) {
    const choices = ref([]);
    const loading = ref(false);

    async function load() {
        loading.value = true;

        try {
            const response = await graphService.getDatasetChoices();

            choices.value = response.data;
        } catch (e) {
            errorMessage.value = extractErrorMessage(e);
        } finally {
            loading.value = false;
        }
    }

    // The fields of one dataset as options, named by their title where they
    // have one.
    function fieldOptions(datasetID) {
        const dataset = choices.value.find((choice) => choice.id === datasetID);

        if (!dataset) {
            return [];
        }

        return dataset.fields.map((field) => {
            let label = field.name;

            if (field.title) {
                label = `${field.title} (${field.name})`;
            }

            return { value: field.id, label: label, name: field.name };
        });
    }

    return { choices, loading, load, fieldOptions };
}
