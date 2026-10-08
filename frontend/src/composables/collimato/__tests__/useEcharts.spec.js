// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { mount } from "@vue/test-utils";
import { defineComponent, h, ref } from "vue";
import { init } from "echarts";
import { useEcharts } from "@/composables/collimato/useEcharts.js";
import { transformData } from "@/utils/collimato/chartTypes.js";

const instance = {
    setOption: vi.fn(),
    resize: vi.fn(),
    dispose: vi.fn(),
    on: vi.fn(),
};

vi.mock("echarts", () => ({ init: vi.fn() }));

let observed = [];
let disconnected = 0;
let observerCallback = null;
let wrappers = [];

function mountHost(...args) {
    const w = mount(host(...args));

    wrappers.push(w);

    return w;
}

beforeEach(() => {
    Object.values(instance).forEach((fn) => fn.mockClear());
    init.mockClear();
    init.mockReturnValue(instance);
    observed = [];
    disconnected = 0;
    observerCallback = null;
    wrappers = [];

    vi.stubGlobal(
        "ResizeObserver",
        class {
            constructor(cb) {
                observerCallback = cb;
            }
            observe(el) {
                observed.push(el);
            }
            disconnect() {
                disconnected += 1;
            }
        },
    );
});

afterEach(() => {
    wrappers.forEach((w) => w.unmount());
    vi.unstubAllGlobals();
});

function host(render, setup) {
    return defineComponent({
        setup() {
            const container = ref(null);
            const api = useEcharts(
                container,
                () => "bar",
                () => ({}),
            );

            setup?.(api);

            return () => h("div", { ref: container });
        },
    });
}

const chartOption = transformData("bar", {
    query: {
        measures: ["orders.count"],
        dimensions: ["orders.status"],
        timeDimensions: [],
    },
    annotation: { measures: { "orders.count": { title: "Count" } } },
    data: [{ "orders.status": "new", "orders.count": 5 }],
});

describe("useEcharts", () => {
    it("watches the container for size changes, not just the window", () => {
        const w = mountHost();

        expect(observed).toEqual([w.element]);
    });

    it("resizes the chart when the container changes size", () => {
        mountHost();

        observerCallback();

        expect(instance.resize).toHaveBeenCalled();
    });

    it("resizes the chart when the window changes size", () => {
        mountHost();

        window.dispatchEvent(new Event("resize"));

        expect(instance.resize).toHaveBeenCalled();
    });

    it("lets go of the observer, the listener and the chart on unmount", () => {
        const w = mountHost();

        w.unmount();
        wrappers.length = 0;
        instance.resize.mockClear();
        window.dispatchEvent(new Event("resize"));

        expect(disconnected).toBe(1);
        expect(instance.dispose).toHaveBeenCalled();
        expect(instance.resize).not.toHaveBeenCalled();
    });

    it("returns the option it drew so the caller can keep it", () => {
        let api = null;

        mountHost(null, (a) => (api = a));

        const option = api.setData(chartOption);

        expect(option.series).toBeTruthy();
        expect(instance.setOption).toHaveBeenCalledWith(option, true);
    });

    it("attaches handlers registered before the chart exists", () => {
        const handler = vi.fn();

        mountHost(null, (api) => api.on("contextmenu", handler));

        expect(instance.on).toHaveBeenCalledWith("contextmenu", handler);
    });

    it("stays quiet when there is no container to draw into", () => {
        const Blank = defineComponent({
            setup() {
                const container = ref(null);
                const api = useEcharts(
                    container,
                    () => "bar",
                    () => ({}),
                );

                return () => h("p", String(api.setData(chartOption)));
            },
        });

        expect(() => mount(Blank)).not.toThrow();
        expect(init).not.toHaveBeenCalled();
    });
});
