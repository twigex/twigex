// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    isFirstMessageOfDay,
    shouldCollapse,
    dateSeparatorLabel,
} from "@/composables/chat/useMessageGrouping";

const SECOND = 1000;
const MINUTE = 60 * SECOND;

// 2026-03-10T12:00:00Z
const noon = Date.UTC(2026, 2, 10, 12, 0, 0);

const post = (created, userID = "u1") => ({ created, user_id: userID });

describe("isFirstMessageOfDay", () => {
    it("opens the day for the first message shown", () => {
        expect(isFirstMessageOfDay(post(noon), undefined)).toBe(true);
    });

    it("does not separate two messages on the same day", () => {
        expect(isFirstMessageOfDay(post(noon + MINUTE), post(noon))).toBe(false);
    });

    it("separates across midnight UTC", () => {
        const justBefore = Date.UTC(2026, 2, 10, 23, 59, 0);
        const justAfter = Date.UTC(2026, 2, 11, 0, 1, 0);

        expect(isFirstMessageOfDay(post(justAfter), post(justBefore))).toBe(true);
    });

    it("reads a seconds timestamp as the same instant as milliseconds", () => {
        const inSeconds = Math.floor(noon / 1000);

        expect(isFirstMessageOfDay(post(noon + MINUTE), post(inSeconds))).toBe(false);
    });
});

describe("shouldCollapse", () => {
    const window = 5 * MINUTE;

    it("never collapses the first message", () => {
        expect(shouldCollapse(post(noon), undefined, window)).toBe(false);
    });

    it("collapses a quick follow-up from the same person", () => {
        expect(shouldCollapse(post(noon + MINUTE), post(noon), window)).toBe(true);
    });

    it("does not collapse a different person", () => {
        expect(shouldCollapse(post(noon + MINUTE, "u2"), post(noon, "u1"), window)).toBe(false);
    });

    it("does not collapse once the window has passed", () => {
        expect(shouldCollapse(post(noon + 6 * MINUTE), post(noon), window)).toBe(false);
    });

    it("compares seconds and milliseconds on the same scale", () => {
        // Same instants as the collapsing case above, one side in seconds.
        expect(shouldCollapse(post(noon + MINUTE), post(Math.floor(noon / 1000)), window)).toBe(
            true,
        );
    });
});

describe("dateSeparatorLabel", () => {
    const t = (id) => id;
    const now = new Date(2026, 2, 10, 15, 0, 0); // local time, as the code uses

    it("names today", () => {
        const today = new Date(2026, 2, 10, 9, 0, 0).getTime();

        expect(dateSeparatorLabel(today, { locale: "en", translate: t, now })).toBe("dates.today");
    });

    it("names yesterday", () => {
        const yesterday = new Date(2026, 2, 9, 9, 0, 0).getTime();

        expect(dateSeparatorLabel(yesterday, { locale: "en", translate: t, now })).toBe(
            "dates.yesterday",
        );
    });

    it("uses the weekday within the last week", () => {
        const threeDaysAgo = new Date(2026, 2, 7, 9, 0, 0).getTime();

        expect(dateSeparatorLabel(threeDaysAgo, { locale: "en", translate: t, now })).toBe(
            "Saturday",
        );
    });

    it("uses a full date beyond a week", () => {
        const longAgo = new Date(2026, 0, 2, 9, 0, 0).getTime();

        expect(dateSeparatorLabel(longAgo, { locale: "en", translate: t, now })).toBe(
            "January 2, 2026",
        );
    });
});
