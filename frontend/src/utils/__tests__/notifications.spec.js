// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { postNotificationText } from "@/utils/notifications.js";

describe("postNotificationText", () => {
    it("prefixes the sender's handle", () => {
        expect(postNotificationText("ada.abele", "ping")).toBe("@ada.abele: ping");
    });

    it("omits the prefix rather than naming an unresolved sender", () => {
        for (const handle of [undefined, null, ""]) {
            expect(postNotificationText(handle, "ping")).toBe("ping");
        }
    });

    it("never invents a stand-in handle", () => {
        expect(postNotificationText(undefined, "ping")).not.toContain("@");
    });
});
