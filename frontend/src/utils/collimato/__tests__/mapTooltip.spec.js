// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { formatTooltip } from "@/utils/collimato/mapTooltip.js";

describe("formatTooltip", () => {
    it("renders one line per field", () => {
        const text = formatTooltip(["earthquakes.magnitude", "earthquakes.source"], {
            "earthquakes.magnitude": "3.49",
            "earthquakes.source": "NCSN",
        });

        expect(text).toBe("earthquakes.magnitude: 3.49\nearthquakes.source: NCSN");
    });

    it("never produces markup, whatever the row contains", () => {
        const text = formatTooltip(["source"], {
            source: "<img src=x onerror=alert(1)>",
        });

        expect(text).toBe("source: <img src=x onerror=alert(1)>");
        expect(text).not.toContain("<div");
        expect(text).not.toContain("<strong");
    });

    it("returns nothing when there is no row or no fields", () => {
        expect(formatTooltip(["a"], null)).toBeNull();
        expect(formatTooltip([], { a: 1 })).toBeNull();
        expect(formatTooltip(undefined, { a: 1 })).toBeNull();
    });

    it("shows missing fields as undefined rather than throwing", () => {
        expect(formatTooltip(["missing"], { a: 1 })).toBe("missing: undefined");
    });
});
