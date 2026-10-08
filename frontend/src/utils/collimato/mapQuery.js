// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { MAX_QUERY_LIMIT } from "@/utils/collimato/chartUtils.js";
import { mergeFilters } from "@/utils/collimato/chartQuery.js";
import { layerQueryFields, layerTypeValue } from "@/utils/collimato/mapLayerTypes.js";

export function buildLayerQuery(modelName, dimensionNames = [], filters = [], limit) {
    const measure = modelName + ".count";

    return {
        measures: [measure],
        dimensions: dimensionNames.filter((name) => name !== measure),
        order: [[measure, "desc"]],
        filters: mergeFilters(filters),
        timeDimensions: [],
        limit: limitWithin(limit),
    };
}

function limitWithin(limit) {
    const requested = Number(limit);

    if (!Number.isFinite(requested) || requested <= 0) {
        return MAX_QUERY_LIMIT;
    }

    return Math.min(requested, MAX_QUERY_LIMIT);
}

export function layerQuery(layer) {
    return buildLayerQuery(
        layer.model,
        layerQueryFields(layerTypeValue(layer), layer.config),
        layer.filters,
        layer.config.limit,
    );
}
