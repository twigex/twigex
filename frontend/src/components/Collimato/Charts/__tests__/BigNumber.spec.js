// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { mount } from "@vue/test-utils";
import BigNumber from "@/components/Collimato/Charts/BigNumber.vue";
import { number } from "@/utils/collimato/chartUtils.js";

const result = (measures, row) => ({
    query: { measures: measures, dimensions: [] },
    annotation: {
        measures: {
            "orders.count": { title: "Count" },
            "orders.revenue": { title: "Revenue" },
        },
    },
    data: [row],
});

describe("number", () => {
    it("totals each measure on its own", () => {
        const options = number({
            ...result(["orders.count", "orders.revenue"], {}),
            data: [
                { "orders.count": 2, "orders.revenue": 10 },
                { "orders.count": 3, "orders.revenue": 40 },
            ],
        });

        expect(options.values).toEqual([
            { label: "Count", formatted: "5", value: 5 },
            { label: "Revenue", formatted: "50", value: 50 },
        ]);
    });

    it("reads totals out of the strings the engine returns", () => {
        const options = number(result(["orders.count"], { "orders.count": "7" }));

        expect(options.values[0].value).toBe(7);
    });

    it("counts a missing value as nothing rather than NaN", () => {
        const options = number(result(["orders.count"], {}));

        expect(options.values[0].value).toBe(0);
    });
});

describe("BigNumber", () => {
    it("shows one value per measure", () => {
        const options = number({
            ...result(["orders.count", "orders.revenue"], {}),
            data: [{ "orders.count": 5, "orders.revenue": 50 }],
        });

        const w = mount(BigNumber, { props: { options } });

        expect(w.findAll("dd").map((d) => d.text())).toEqual(["5", "50"]);
        expect(w.findAll("dt").map((d) => d.text())).toEqual(["Count", "Revenue"]);
    });

    it("leaves the label off when there is only one measure", () => {
        const options = number(result(["orders.count"], { "orders.count": 5 }));

        const w = mount(BigNumber, { props: { options } });

        expect(w.find("dd").text()).toBe("5");
        expect(w.findAll("dt")).toHaveLength(0);
    });

    it("renders nothing before the data arrives", () => {
        const w = mount(BigNumber, {
            props: { options: { model: "orders", query: {} } },
        });

        expect(w.findAll("dd")).toHaveLength(0);
    });
});
