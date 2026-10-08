// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import moment from "moment-timezone";
import useDateOperations from "@/composables/useDateOperations.js";
import { useUserStore } from "@/store/user";

function dayIn(timezone, timestamp) {
    return moment.unix(timestamp).tz(timezone).format("YYYY-MM-DD");
}

describe("getDateInputValue", () => {
    beforeEach(() => setActivePinia(createPinia()));

    it("gives a date field's day, and a timestamp's day where the user is", () => {
        useUserStore().user = {
            ...useUserStore().user,
            timezone: { useAutomaticTimezone: "false", manualTimezone: "America/New_York" },
        };
        const { getDateInputValue } = useDateOperations();

        expect(getDateInputValue("2026-09-30T00:00:00Z", true)).toBe("2026-09-30");
        // 03:00 UTC on 30 September is still the 29th in New York.
        expect(getDateInputValue(1790737200, false)).toBe("2026-09-29");
        expect(getDateInputValue("", false)).toBe("");
        expect(getDateInputValue(0, false)).toBe("");
    });
});

describe("getTimestampFromDateString", () => {
    beforeEach(() => setActivePinia(createPinia()));

    it("keeps a date field's day west and east of UTC", () => {
        for (const timezone of ["America/New_York", "Europe/Riga", "Asia/Tokyo"]) {
            useUserStore().user = {
                ...useUserStore().user,
                timezone: { useAutomaticTimezone: "false", manualTimezone: timezone },
            };
            const { getTimestampFromDateString } = useDateOperations();

            expect(dayIn(timezone, getTimestampFromDateString("2026-09-30T00:00:00Z"))).toBe(
                "2026-09-30",
            );
            expect(dayIn(timezone, getTimestampFromDateString("2026-09-30"))).toBe("2026-09-30");
        }
    });
});

describe("getDayOf", () => {
    beforeEach(() => setActivePinia(createPinia()));

    it("gives the day of a day, or of a timestamp where the user is", () => {
        useUserStore().user = {
            ...useUserStore().user,
            timezone: { useAutomaticTimezone: "false", manualTimezone: "America/New_York" },
        };
        const { getDayOf } = useDateOperations();

        expect(getDayOf("2026-09-30")).toBe("2026-09-30");
        expect(getDayOf("2026-09-30T00:00:00Z")).toBe("2026-09-30");
        // 03:00 UTC on 30 September is still the 29th in New York, in seconds
        // or milliseconds.
        expect(getDayOf(1790737200)).toBe("2026-09-29");
        expect(getDayOf(1790737200000)).toBe("2026-09-29");
        expect(getDayOf("")).toBeNull();
        expect(getDayOf("soon")).toBeNull();
    });
});
