// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach } from "vitest";
import { mount } from "@vue/test-utils";
import { computed, ref, nextTick } from "vue";
import { Map as MaplibreMap } from "maplibre-gl";
import MapChart from "@/components/Collimato/Charts/Map/MapChart.vue";

let overlayProps = null;
const overlay = {
    setProps: vi.fn((props) => (overlayProps = props)),
    finalize: vi.fn(),
};

const instance = {
    on: vi.fn(),
    remove: vi.fn(),
    setStyle: vi.fn(),
    jumpTo: vi.fn(),
    zoomTo: vi.fn(),
    getZoom: vi.fn(() => 8),
    addControl: vi.fn(),
    fitBounds: vi.fn(),
};

vi.mock("maplibre-gl", () => ({
    Map: vi.fn(function () {
        return instance;
    }),
    NavigationControl: vi.fn(),
    setWorkerUrl: vi.fn(),
}));

vi.mock("maplibre-gl/dist/maplibre-gl-worker.mjs?worker&url", () => ({
    default: "maplibre-gl-worker.mjs",
}));

vi.mock("@deck.gl/mapbox", () => ({
    MapboxOverlay: vi.fn(function () {
        return overlay;
    }),
}));

beforeEach(() => {
    Object.values(instance).forEach((fn) => fn.mockClear?.());
    instance.getZoom.mockReturnValue(8);
    MaplibreMap.mockImplementation(function () {
        return instance;
    });
});

function setup() {
    const layers = ref([]);
    const zoom = ref(1);
    const data = computed(() => ({
        map_style: "style-a",
        latitude: "",
        longitude: "",
        zoom: zoom.value,
        layers: layers.value,
    }));

    const w = mount(MapChart, { props: { data: data.value } });

    return { w, layers, zoom, data };
}

describe("Map view state", () => {
    it("leaves the zoom alone when only the layers changed", async () => {
        const { w, layers, data } = setup();

        layers.value = [{ id: 0, show: true, type: "point", config: {} }];
        await w.setProps({ data: data.value });

        expect(instance.zoomTo).not.toHaveBeenCalled();
        expect(instance.jumpTo).not.toHaveBeenCalled();
    });

    it("leaves the basemap alone when only the layers changed", async () => {
        const { w, layers, data } = setup();

        layers.value = [{ id: 0, show: true, type: "point", config: {} }];
        await w.setProps({ data: data.value });

        expect(instance.setStyle).not.toHaveBeenCalled();
    });

    it("still follows a real zoom change", async () => {
        const { w, zoom, data } = setup();

        zoom.value = 5;
        await w.setProps({ data: data.value });
        await nextTick();

        expect(instance.zoomTo).toHaveBeenCalledWith(5, { duration: 0 });
    });

    it("still follows a real style change", async () => {
        const { w, data } = setup();

        await w.setProps({ data: { ...data.value, map_style: "style-b" } });
        await nextTick();

        expect(instance.setStyle).toHaveBeenCalledWith("style-b");
    });
});

describe("layers arriving before the map is ready", () => {
    function mountMap() {
        const data = {
            map_style: "s",
            latitude: "",
            longitude: "",
            zoom: 1,
            layers: [],
        };

        return mount(MapChart, { props: { data } });
    }

    const layer = {
        name: "points",
        type: "point",
        show: true,
        config: {
            coordinates: { lat: "t.lat", long: "t.long" },
            fill_color: "#4f46e5",
            radius: 10,
            min_radius: 1,
            max_radius: 10,
            data: [{ "t.lat": "40.7", "t.long": "-74.0" }],
        },
    };

    it("does not throw when data lands before the basemap loads", () => {
        const w = mountMap();

        expect(() => w.vm.updateLayers([layer])).not.toThrow();
    });

    it("draws them once the basemap has loaded", () => {
        const w = mountMap();

        w.vm.updateLayers([layer]);

        const onLoad = instance.on.mock.calls.find(([event]) => event === "load")[1];

        onLoad();

        expect(overlayProps.layers).toHaveLength(1);
    });

    it("keeps drawing after the basemap is ready", () => {
        const w = mountMap();

        const onLoad = instance.on.mock.calls.find(([event]) => event === "load")[1];

        onLoad();

        w.vm.updateLayers([layer]);

        expect(overlayProps.layers).toHaveLength(1);
    });

    it("draws nothing for a hidden layer", () => {
        const w = mountMap();

        const onLoad = instance.on.mock.calls.find(([event]) => event === "load")[1];

        onLoad();

        w.vm.updateLayers([{ ...layer, show: false }]);

        expect(overlayProps.layers).toEqual([]);
    });
});
