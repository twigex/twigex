// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { CHART_PALETTE, CHART_LABEL_COLOR } from "@/utils/collimato/chartPalette.js";
import { CHART_TYPE_IDS, isEchartsType, transformData } from "@/utils/collimato/chartTypes.js";

const result = {
    query: {
        measures: ["orders.count"],
        dimensions: ["orders.status"],
        timeDimensions: [],
    },
    annotation: { measures: { "orders.count": { title: "Count" } } },
    data: [
        { "orders.status": "new", "orders.count": 5 },
        { "orders.status": "done", "orders.count": 3 },
    ],
};

describe("chart palette", () => {
    it("holds distinct hex colours", () => {
        CHART_PALETTE.forEach((c) => expect(c).toMatch(/^#[0-9a-f]{6}$/));

        expect(new Set(CHART_PALETTE).size).toBe(CHART_PALETTE.length);
    });

    it.each(CHART_TYPE_IDS.filter(isEchartsType))("%s draws from the shared palette", (id) => {
        expect(transformData(id, result).color).toBe(CHART_PALETTE);
    });

    it("labels data points in the text colour, not pure black", () => {
        const option = transformData("bar", result);

        expect(option.series[0].label.color).toBe(CHART_LABEL_COLOR);
        expect(CHART_LABEL_COLOR).not.toBe("#000");
    });
});
