// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { onMounted, onBeforeUnmount, toValue } from "vue";
import * as echarts from "echarts";
import { applyConfiguration } from "@/utils/collimato/chartTypes.js";

export function useEcharts(container, chartType, configuration) {
    let instance = null;
    let observer = null;
    const pending = [];

    function resize() {
        instance?.resize();
    }

    function setData(data) {
        if (!instance) {
            return null;
        }

        const option = applyConfiguration(toValue(chartType), data, toValue(configuration) ?? {});

        instance.setOption(option, true);

        return option;
    }

    function on(event, handler) {
        if (instance) {
            instance.on(event, handler);

            return;
        }

        pending.push([event, handler]);
    }

    onMounted(() => {
        if (!container.value) {
            return;
        }

        instance = echarts.init(container.value);
        pending.splice(0).forEach(([event, handler]) => instance.on(event, handler));

        observer = new ResizeObserver(resize);
        observer.observe(container.value);
        window.addEventListener("resize", resize);
    });

    onBeforeUnmount(() => {
        observer?.disconnect();
        observer = null;

        window.removeEventListener("resize", resize);

        instance?.dispose();
        instance = null;
    });

    return { setData, resize, on };
}
