// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed } from "vue";
import { buildQuery, joinableDataModels, ORDER_NONE } from "@/utils/collimato/chartQuery.js";
import { layerQuery } from "@/utils/collimato/mapQuery.js";
import { maxMeasuresFor } from "@/utils/collimato/chartTypes.js";
import { pollQuery } from "@/utils/collimato/queryPolling.js";
import {
    defaultLayerConfig,
    layerTypeValue,
    layerTypeIsReady,
    layerQueryFields,
} from "@/utils/collimato/mapLayerTypes.js";

export const DEFAULT_MAP_STYLE =
    "https://basemaps.cartocdn.com/gl/dark-matter-nolabels-gl-style/style.json";
export const DEFAULT_MAP_CENTER = { latitude: "", longitude: "" };
export const DEFAULT_MAP_ZOOM = 1;

function withoutModelScopedFields(config) {
    const next = {
        ...config,
        coordinates: { lat: null, long: null },
        data: null,
    };

    if ("geojson_field" in next) {
        next.geojson_field = null;
    }

    if ("source" in next) {
        next.source = { lat: null, long: null };
    }

    if ("target" in next) {
        next.target = { lat: null, long: null };
    }

    if ("source_color_field" in next) {
        next.source_color_field = null;
    }

    if ("target_color_field" in next) {
        next.target_color_field = null;
    }

    if ("radius_field" in next) {
        next.radius_field = null;
    }

    if ("fill_color_field" in next) {
        next.fill_color_field = null;
    }

    if ("weight" in next) {
        next.weight = null;
    }

    if (next.popover) {
        next.popover = { fields: [] };
    }

    return next;
}

const MAX_LAYER_RETRIES = 30;
const LAYER_RETRY_MS = 2000;

export function useChartBuilder(initialChartType, io = {}) {
    const { loadData } = io;
    const chartType = ref(initialChartType);
    const dataModels = ref([]);
    const selectedModel = ref(null);
    const measures = ref([]);
    const dimensions = ref([]);
    const orders = ref([]);
    const filters = ref([]);
    const chartOptions = ref({});

    const mapStyle = ref(DEFAULT_MAP_STYLE);
    const mapCenter = ref({ ...DEFAULT_MAP_CENTER });
    const mapZoom = ref(DEFAULT_MAP_ZOOM);
    const mapLayers = ref([]);
    const layerStatus = ref({});
    const layerTruncated = ref({});
    const mapExtras = ref({});

    const time = ref({
        dimension: null,
        dateRange: "",
        granularity: "",
        startTime: "",
        endTime: "",
    });

    const availableDataModels = computed(() =>
        joinableDataModels(dataModels.value, selectedModel.value),
    );

    const timeDimensionFields = computed(() =>
        dataModels.value.flatMap((model) => model.dimensions.filter((d) => d.type == "time")),
    );

    const hasMapCenter = computed(() => {
        const { latitude, longitude } = mapCenter.value;

        return (
            latitude !== "" &&
            longitude !== "" &&
            Number.isFinite(Number(latitude)) &&
            Number.isFinite(Number(longitude))
        );
    });

    const mapConfiguration = computed(() => ({
        ...mapExtras.value,
        map_style: mapStyle.value,
        longitude: mapCenter.value.longitude,
        latitude: mapCenter.value.latitude,
        zoom: mapZoom.value,
        layers: mapLayers.value,
    }));

    const layerStates = computed(() =>
        mapLayers.value.map((layer) => ({
            id: layer.id,
            name: layer.name,
            status: layerStatus.value[layer.id] ?? "",
            truncated: Boolean(layerTruncated.value[layer.id]),
        })),
    );

    const query = computed(() =>
        buildQuery({
            measures: measures.value,
            dimensions: dimensions.value,
            orders: orders.value,
            filters: filters.value,
            chartType: chartType.value,
            time: time.value,
        }),
    );

    function trackOrder(name) {
        if (!orders.value.some((entry) => entry.name == name)) {
            orders.value.push({ name: name, direction: ORDER_NONE });
        }
    }

    function untrackOrder(name) {
        orders.value = orders.value.filter((entry) => entry.name != name);
    }

    const measureLimit = computed(() => maxMeasuresFor(chartType.value));

    function addMeasure(field) {
        if (measures.value.some((m) => m.name == field.name)) {
            return false;
        }

        if (measureLimit.value != null && measures.value.length >= measureLimit.value) {
            return false;
        }

        measures.value.push(field);
        trackOrder(field.name);

        return true;
    }

    function removeMeasure(field) {
        measures.value = measures.value.filter((m) => m.name != field.name);
        untrackOrder(field.name);
    }

    function addDimension(field) {
        if (dimensions.value.some((d) => d.name == field.name)) {
            return false;
        }

        dimensions.value.push(field);
        trackOrder(field.name);

        return true;
    }

    function removeDimension(field) {
        dimensions.value = dimensions.value.filter((d) => d.name != field.name);
        untrackOrder(field.name);
    }

    function setOrder(name, direction) {
        const entry = orders.value.find((o) => o.name == name);

        if (entry) {
            entry.direction = direction;
        }
    }

    function reorderOrders(entries) {
        orders.value = entries;
    }

    function addFilter(filter) {
        filters.value.push(filter);
    }

    function updateFilter(member, filter) {
        const index = filters.value.findIndex((f) => f.member == member);

        if (index != -1) {
            filters.value[index] = filter;
        }
    }

    function removeFilter(filter) {
        filters.value = filters.value.filter((f) => f !== filter);
    }

    function setMapConfiguration(configuration) {
        const {
            map_style,
            longitude,
            latitude,
            zoom,
            layers: incomingLayers,
            ...rest
        } = configuration ?? {};

        mapStyle.value = map_style ?? DEFAULT_MAP_STYLE;
        mapCenter.value = {
            latitude: latitude ?? DEFAULT_MAP_CENTER.latitude,
            longitude: longitude ?? DEFAULT_MAP_CENTER.longitude,
        };
        mapZoom.value = zoom ?? DEFAULT_MAP_ZOOM;
        mapLayers.value = (incomingLayers ?? []).map((layer) => ({
            ...layer,
            type: layerTypeValue(layer),
            filters: layer.filters ?? [],
        }));
        mapExtras.value = rest;
    }

    const layerAborts = new Map();

    function setLayerStatus(id, status) {
        layerStatus.value[id] = status;
    }

    function findLayer(id) {
        return mapLayers.value.find((layer) => layer.id === id) ?? null;
    }

    function cancelLayerLoad(id) {
        layerAborts.get(id)?.abort();
        layerAborts.delete(id);
    }

    function cancelAllLayerLoads() {
        [...layerAborts.keys()].forEach(cancelLayerLoad);
    }

    function layerIsReady(layer) {
        return Boolean(layer?.model && layerTypeIsReady(layerTypeValue(layer), layer.config));
    }

    function nextLayerId() {
        return mapLayers.value.reduce((max, layer) => Math.max(max, layer.id + 1), 0);
    }

    function addLayer({ name, type }) {
        const layer = {
            id: nextLayerId(),
            type: type,
            show: true,
            name: name,
            model: null,
            filters: [],
            config: defaultLayerConfig(type),
        };

        mapLayers.value.push(layer);

        return layer;
    }

    function removeLayer(id) {
        cancelLayerLoad(id);
        mapLayers.value = mapLayers.value.filter((layer) => layer.id !== id);
        delete layerStatus.value[id];
        delete layerTruncated.value[id];
    }

    function setLayerVisible(id, show) {
        const layer = findLayer(id);

        if (layer) {
            layer.show = show;
        }
    }

    function setLayerName(id, name) {
        const layer = findLayer(id);

        if (layer) {
            layer.name = name;
        }
    }

    function setLayerType(id, type) {
        const layer = findLayer(id);

        if (!layer) {
            return;
        }

        cancelLayerLoad(id);
        layer.type = type;
        layer.config = defaultLayerConfig(type);
    }

    function setLayerModel(id, model) {
        const layer = findLayer(id);

        if (!layer) {
            return;
        }

        cancelLayerLoad(id);
        layer.model = model;
        layer.filters = [];
        layer.config = withoutModelScopedFields(layer.config);
    }

    function queryShapeOf(layer) {
        return [...layerQueryFields(layerTypeValue(layer), layer.config), layer.config.limit].join(
            "|",
        );
    }

    function setLayerConfig(id, patch) {
        const layer = findLayer(id);

        if (!layer) {
            return Promise.resolve(null);
        }

        const before = queryShapeOf(layer);

        layer.config = { ...layer.config, ...patch };

        return before === queryShapeOf(layer) ? Promise.resolve(null) : loadLayer(id);
    }

    function addLayerFilter(id, filter) {
        const layer = findLayer(id);

        if (!layer) {
            return Promise.resolve(null);
        }

        layer.filters = [...layer.filters, filter];

        return loadLayer(id);
    }

    function updateLayerFilter(id, member, filter) {
        const layer = findLayer(id);

        if (!layer) {
            return Promise.resolve(null);
        }

        const index = layer.filters.findIndex((f) => f.member == member);

        if (index == -1) {
            return Promise.resolve(null);
        }

        layer.filters = layer.filters.map((existing, i) => (i == index ? filter : existing));

        return loadLayer(id);
    }

    function removeLayerFilter(id, filter) {
        const layer = findLayer(id);

        if (!layer) {
            return Promise.resolve(null);
        }

        layer.filters = layer.filters.filter((f) => f !== filter);

        return loadLayer(id);
    }

    function setLayerCoordinates(id, patch) {
        return setLayerFields(id, { coordinates: patch });
    }

    function setLayerFields(id, patch) {
        const layer = findLayer(id);

        if (!layer) {
            return Promise.resolve(null);
        }

        const next = { ...layer.config };

        Object.entries(patch).forEach(([key, value]) => {
            next[key] =
                value && typeof value === "object" && !Array.isArray(value)
                    ? { ...next[key], ...value }
                    : value;
        });

        layer.config = next;

        return loadLayer(id);
    }

    async function loadLayer(id) {
        const layer = findLayer(id);

        if (!layerIsReady(layer)) {
            return null;
        }

        const model = dataModels.value.find((candidate) => candidate.table == layer.model);

        if (!model) {
            return null;
        }

        cancelLayerLoad(id);

        const controller = new AbortController();

        layerAborts.set(id, controller);
        setLayerStatus(id, "loading");

        const query = layerQuery(layer);

        try {
            const result = await pollQuery(loadData, query, {
                signal: controller.signal,
                retryLimit: MAX_LAYER_RETRIES,
                retryMs: LAYER_RETRY_MS,
            });

            const rows = result.data.data;

            layer.config.data = rows;
            layerTruncated.value[id] = rows.length >= query.limit;

            return rows;
        } catch (error) {
            if (controller.signal.aborted) {
                return null;
            }

            throw error;
        } finally {
            if (layerAborts.get(id) === controller) {
                layerAborts.delete(id);
                setLayerStatus(id, "loaded");
            } else if (!layerAborts.has(id)) {
                setLayerStatus(id, "");
            }
        }
    }

    function loadAllLayers() {
        return Promise.allSettled(mapLayers.value.map((layer) => loadLayer(layer.id)));
    }

    function setTime(patch) {
        time.value = { ...time.value, ...patch };
    }

    function clearQuery() {
        measures.value = [];
        dimensions.value = [];
        orders.value = [];
        filters.value = [];
        time.value = {
            dimension: null,
            dateRange: "",
            granularity: "",
            startTime: "",
            endTime: "",
        };
    }

    function selectModel(model) {
        selectedModel.value = model;
        clearQuery();
    }

    function setChartType(value) {
        chartType.value = value;
        chartOptions.value = {};

        const limit = maxMeasuresFor(value);

        if (limit != null && measures.value.length > limit) {
            measures.value.slice(limit).forEach((field) => untrackOrder(field.name));
            measures.value = measures.value.slice(0, limit);
        }
    }

    function fieldByName(kind, name) {
        for (const model of dataModels.value) {
            const found = model[kind].find((field) => field.name == name);

            if (found) {
                return found;
            }
        }

        return null;
    }

    function loadQuery(savedQuery, modelName) {
        selectedModel.value = dataModels.value.find((model) => model.table == modelName) ?? null;

        orders.value = savedQuery.order.map(([name, direction]) => ({
            name: name,
            direction: direction,
        }));

        savedQuery.measures.forEach((name) => {
            const field = fieldByName("measures", name);

            if (field) {
                measures.value.push(field);
                trackOrder(name);
            }
        });

        savedQuery.dimensions.forEach((name) => {
            const field = fieldByName("dimensions", name);

            if (field) {
                dimensions.value.push(field);
                trackOrder(name);
            }
        });

        filters.value = savedQuery.filters.map((filter) => ({
            member: filter.member,
            operator: filter.operator,
            value: filter.values,
        }));

        savedQuery.timeDimensions.forEach((timeDimension) => {
            const isCustom = timeDimension.dateRange.length == 2;

            setTime({
                dimension: timeDimensionFields.value.find((d) => d.name == timeDimension.dimension),
                dateRange: isCustom ? "custom" : timeDimension.dateRange,
                granularity: timeDimension.granularity,
                startTime: isCustom ? timeDimension.dateRange[0] : "",
                endTime: isCustom ? timeDimension.dateRange[1] : "",
            });
        });
    }

    return {
        chartType,
        dataModels,
        selectedModel,
        measures,
        dimensions,
        orders,
        filters,
        time,
        chartOptions,
        mapStyle,
        mapCenter,
        mapZoom,
        mapLayers,
        layerStates,
        mapConfiguration,
        hasMapCenter,
        availableDataModels,
        timeDimensionFields,
        query,
        measureLimit,
        addMeasure,
        removeMeasure,
        addDimension,
        removeDimension,
        setOrder,
        reorderOrders,
        addFilter,
        updateFilter,
        removeFilter,
        setTime,
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
        cancelLayerLoad,
        cancelAllLayerLoads,
        clearQuery,
        selectModel,
        setChartType,
        loadQuery,
    };
}
