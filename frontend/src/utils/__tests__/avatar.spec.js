// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { colorClassFor, initialsFor } from "../avatar";

describe("initialsFor", () => {
    it("takes the first letter, uppercased", () => {
        expect(initialsFor("rihards")).toBe("R");
    });

    it("takes one letter from a full name, not two", () => {
        expect(initialsFor("Rihards Kalnins")).toBe("R");
        expect(initialsFor("Rihards Kalnins")).toBe(initialsFor("Rihards"));
    });

    it("takes a whole code point, not a byte", () => {
        expect(initialsFor("Āboliņš")).toBe("Ā");
        expect(initialsFor("😀 emoji")).toBe("😀");
    });

    it("falls back when there is no name", () => {
        expect(initialsFor("")).toBe("?");
        expect(initialsFor("   ")).toBe("?");
        expect(initialsFor(null)).toBe("?");
        expect(initialsFor(undefined)).toBe("?");
    });
});

describe("colorClassFor", () => {
    it("returns the same colour for the same id", () => {
        const id = "0dc1d5a2f4b64e0c9d7a1b2c3d4e5f60";

        expect(colorClassFor(id)).toBe(colorClassFor(id));
    });

    it("returns a full class name so Tailwind emits the CSS", () => {
        expect(colorClassFor("abc")).toMatch(/^bg-[a-z]+-\d{3}$/);
    });

    it("spreads ids across the palette", () => {
        const seen = new Set();

        for (let i = 0; i < 200; i++) {
            seen.add(colorClassFor(`user-${i}`));
        }

        expect(seen.size).toBeGreaterThan(1);
    });

    it("handles a missing id without throwing", () => {
        expect(colorClassFor(undefined)).toMatch(/^bg-/);
        expect(colorClassFor("")).toMatch(/^bg-/);
    });
});
