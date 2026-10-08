// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import fileTypes from "@/constants/fileTypes";
import { fileTypeOf, isViewable } from "../files";

describe("fileTypeOf", () => {
    it("takes the type Files lists", () => {
        expect(fileTypeOf({ type: "application/pdf" })).toBe(fileTypes["application/pdf"]);
    });

    it("takes the family's type for one Files does not list", () => {
        expect(fileTypeOf({ type: "image/webp" })).toBe(fileTypes["image/png"]);
        expect(fileTypeOf({ type: "video/quicktime" })).toBe(fileTypes["video/mp4"]);
    });

    it("falls back to a plain document", () => {
        expect(fileTypeOf({ type: "application/x-unknown" })).toBe(fileTypes["not-found"]);
        expect(fileTypeOf({})).toBe(fileTypes["not-found"]);
        expect(fileTypeOf(null)).toBe(fileTypes["not-found"]);
    });
});

describe("isViewable", () => {
    it("is true for images, videos and PDFs", () => {
        expect(isViewable({ type: "image/jpeg" })).toBe(true);
        expect(isViewable({ type: "video/mp4" })).toBe(true);
        expect(isViewable({ type: "application/pdf" })).toBe(true);
    });

    it("is false for office documents, archives and files without a type", () => {
        expect(
            isViewable({
                type: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            }),
        ).toBe(false);
        expect(isViewable({ type: "application/zip" })).toBe(false);
        expect(isViewable({ name: "notes.txt" })).toBe(false);
    });
});
