// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    typingMemberName,
    finishedMemberName,
    duplicateMemberNames,
} from "@/utils/collimato/memberNames.js";

describe("typingMemberName", () => {
    it("turns spaces and dashes into underscores", () => {
        expect(typingMemberName("Full Name")).toBe("Full_Name");
        expect(typingMemberName("order-id")).toBe("order_id");
    });

    it("keeps capitals", () => {
        expect(typingMemberName("fullName")).toBe("fullName");
    });

    it("removes accents and drops other characters", () => {
        expect(typingMemberName("Vārds")).toBe("Vards");
        expect(typingMemberName("Total (EUR)")).toBe("Total_EUR");
        expect(typingMemberName("50%")).toBe("50");
    });

    it("keeps a trailing underscore so the next word can follow", () => {
        expect(typingMemberName("full ")).toBe("full_");
    });
});

describe("finishedMemberName", () => {
    it("squashes and trims underscores", () => {
        expect(finishedMemberName(" full  name ")).toBe("full_name");
        expect(finishedMemberName("Total (EUR)")).toBe("Total_EUR");
    });

    it("prefixes a name that starts with a digit", () => {
        expect(finishedMemberName("2024 sales")).toBe("_2024_sales");
    });

    it("leaves a valid name unchanged", () => {
        expect(finishedMemberName("full_name")).toBe("full_name");
    });
});

describe("duplicateMemberNames", () => {
    it("returns names used more than once", () => {
        const duplicates = duplicateMemberNames(["count", "id", "count", "name", "id"]);

        expect([...duplicates].sort()).toEqual(["count", "id"]);
    });

    it("ignores empty names", () => {
        expect(duplicateMemberNames(["", "", "id"]).size).toBe(0);
    });
});
