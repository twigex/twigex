// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { transformData, applyConfiguration } from "@/utils/collimato/charts/chartDrawing.js";
import { cubeResult } from "@/utils/collimato/charts/__tests__/cubeResult.js";

describe("applyConfiguration", () => {
    const barConfig = {
        legend: true,
        bottomMargin: 5,
        topMargin: 30,
        leftMargin: 5,
        rightMargin: 5,
        containLabel: true,
        dataZoom: false,
        stack: false,
        orientation: "vertical",
        showValues: false,
        angle: 0,
    };

    it.each(["pie", "bar", "line", "time_bar"])(
        "returns an empty option untouched for %s",
        (chartType) => {
            expect(applyConfiguration(chartType, {}, barConfig)).toEqual({});
        },
    );

    it("returns a null option untouched", () => {
        expect(applyConfiguration("bar", null, barConfig)).toBeNull();
    });
});

describe("missing values setting", () => {
    function postsPerDay() {
        return cubeResult({
            measures: ["posts.count"],
            dimensions: [],
            timeDimensions: [
                {
                    dimension: "posts.created",
                    granularity: "day",
                    dateRange: ["2026-01-01T00:00:00.000", "2026-01-03T00:00:00.000"],
                },
            ],
            rows: [{ "posts.created.day": "2026-01-02T00:00:00.000", "posts.count": "7" }],
        });
    }

    it("connects the line across missing slots by default", () => {
        const option = transformData("line", postsPerDay(), {});

        expect(option.series[0].data).toEqual([null, 7, null]);
        expect(option.series[0].connectNulls).toBe(true);
    });

    it("leaves a gap when asked to", () => {
        const option = transformData("line", postsPerDay(), { missingValues: "gap" });

        expect(option.series[0].data).toEqual([null, 7, null]);
        expect(option.series[0].connectNulls).toBe(false);
    });

    it("draws missing slots as zero when asked to", () => {
        const option = transformData("line", postsPerDay(), { missingValues: "zero" });

        expect(option.series[0].data).toEqual([0, 7, 0]);
    });

    it("keeps a real zero as zero", () => {
        const data = postsPerDay();

        data.data.push({ "posts.created.day": "2026-01-03T00:00:00.000", "posts.count": "0" });

        const option = transformData("line", data, { missingValues: "gap" });

        expect(option.series[0].data).toEqual([null, 7, 0]);
    });
});
