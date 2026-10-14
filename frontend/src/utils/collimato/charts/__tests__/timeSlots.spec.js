// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, afterEach } from "vitest";
import { currentLocale } from "@/i18n/index";
import { timeSlotLabels } from "@/utils/collimato/charts/timeSlots.js";
import { plainSpaces } from "@/utils/collimato/charts/__tests__/plainSpaces.js";

function labels(slots, granularity, dateRange, display) {
    const result = timeSlotLabels(slots, granularity, dateRange, display);

    return plainSpaces(result);
}

const SEPTEMBER_TO_OCTOBER = ["2026-09-15T00:00:00.000", "2026-10-14T23:59:59.999"];
const WHOLE_YEAR = ["2026-01-01T00:00:00.000", "2026-12-31T23:59:59.999"];

describe("time slot labels", () => {
    afterEach(() => {
        currentLocale.value = "en";
    });

    it("names a whole month by its month and year", () => {
        expect(labels(["2026-10-01"], "month", WHOLE_YEAR)).toEqual(["Oct 2026"]);
    });

    it("cuts the first and last month to the chosen range", () => {
        const result = labels(["2026-09-01", "2026-10-01"], "month", SEPTEMBER_TO_OCTOBER);

        expect(result).toEqual(["15 – 30 Sept 2026", "1 – 14 Oct 2026"]);
    });

    it("names a week by its days, across two months", () => {
        expect(labels(["2026-09-28"], "week", WHOLE_YEAR)).toEqual(["28 Sept – 4 Oct 2026"]);
    });

    it("names a week across the new year with both years", () => {
        const range = ["2026-01-01T00:00:00.000", "2027-12-31T23:59:59.999"];

        expect(labels(["2026-12-28"], "week", range)).toEqual(["28 Dec 2026 – 3 Jan 2027"]);
    });

    it("names a whole quarter and a whole year", () => {
        expect(labels(["2026-07-01"], "quarter", WHOLE_YEAR)).toEqual(["Q3 2026"]);
        expect(labels(["2026-01-01"], "year", WHOLE_YEAR)).toEqual(["2026"]);
    });

    it("names a day by its date", () => {
        expect(labels(["2026-09-15"], "day", SEPTEMBER_TO_OCTOBER)).toEqual(["15 Sept 2026"]);
    });

    it("follows the user's clock for hours", () => {
        const slot = ["2026-09-15T15:00:00.000"];

        expect(labels(slot, "hour", SEPTEMBER_TO_OCTOBER, { hour12: false })).toEqual([
            "15 Sept 2026, 15:00",
        ]);
        expect(labels(slot, "hour", SEPTEMBER_TO_OCTOBER, { hour12: true })).toEqual([
            "15 Sept 2026, 03:00 pm",
        ]);
    });

    it("writes the labels in the user's language", () => {
        currentLocale.value = "lv";

        const result = labels(["2026-09-01"], "month", SEPTEMBER_TO_OCTOBER);

        expect(result).toEqual(["2026. gada 15.–30. sept."]);
    });
});
