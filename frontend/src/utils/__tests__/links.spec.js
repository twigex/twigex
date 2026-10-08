// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { safeHref } from "../links";

describe("safeHref", () => {
    it("opens web and mail links as they are, and one without a scheme as https", () => {
        expect(safeHref("https://example.com/a")).toBe("https://example.com/a");
        expect(safeHref("HTTP://example.com")).toBe("HTTP://example.com");
        expect(safeHref("mailto:a@example.com")).toBe("mailto:a@example.com");
        expect(safeHref(" example.com/page ")).toBe("https://example.com/page");
    });

    it("opens nothing for a link that could run code", () => {
        for (const link of [
            "javascript:alert(1)",
            " JavaScript:alert(1)",
            "java\tscript:alert(1)",
            "data:text/html,x",
            "",
            null,
        ]) {
            expect(safeHref(link)).toBeNull();
        }
    });
});
