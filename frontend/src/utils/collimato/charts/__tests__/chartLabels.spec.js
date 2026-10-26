// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { shortNumber } from "@/utils/collimato/charts/chartLabels.js";

describe("shortNumber", () => {
    it("leaves numbers below a thousand as they are", () => {
        expect(shortNumber(0)).toBe("0");
        expect(shortNumber(250)).toBe("250");
        expect(shortNumber(12.5)).toBe("12.5");
    });

    it("shortens large numbers", () => {
        expect(shortNumber(1500)).toBe("1.5K");
        expect(shortNumber(1200000)).toBe("1.2M");
        expect(shortNumber(3400000000)).toBe("3.4B");
        expect(shortNumber(7100000000000)).toBe("7.1T");
    });

    it("leaves out a trailing .0", () => {
        expect(shortNumber(1000)).toBe("1K");
        expect(shortNumber(2000000)).toBe("2M");
    });

    it("shortens negative numbers too", () => {
        expect(shortNumber(-5000)).toBe("-5K");
        expect(shortNumber(-1500000)).toBe("-1.5M");
        expect(shortNumber(-250)).toBe("-250");
    });
});
