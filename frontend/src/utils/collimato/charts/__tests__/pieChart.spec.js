// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { transformData, applyConfiguration } from "@/utils/collimato/charts/chartDrawing.js";
import { cubeResult } from "@/utils/collimato/charts/__tests__/cubeResult.js";

describe("pie configuration", () => {
    let warn;

    beforeEach(() => {
        warn = vi.spyOn(console, "warn").mockImplementation(() => {});
    });

    afterEach(() => {
        warn.mockRestore();
    });

    function pieOption() {
        return transformData(
            "pie",
            cubeResult({
                measures: ["orders.count"],
                dimensions: ["orders.status"],
                rows: [
                    { "orders.status": "new", "orders.count": 5 },
                    { "orders.status": "done", "orders.count": 5 },
                ],
            }),
        );
    }

    it("falls back to default radii when the configuration omits them", () => {
        const option = applyConfiguration("pie", pieOption(), {
            legend: true,
            threshold: 5,
            labelType: "category",
            orientation: "Top",
        });

        expect(option.series[0].radius).toEqual(["0%", "55%"]);
    });

    it("honours configured radii", () => {
        const option = applyConfiguration("pie", pieOption(), {
            legend: true,
            threshold: 5,
            labelType: "category",
            orientation: "Top",
            innerRadius: 20,
            outerRadius: 70,
        });

        expect(option.series[0].radius).toEqual(["20%", "70%"]);
    });

    it("yields an empty option when there is nothing to plot", () => {
        const option = transformData("pie", cubeResult({ measures: [], dimensions: [], rows: [] }));

        expect(option).toEqual({});
    });
});
