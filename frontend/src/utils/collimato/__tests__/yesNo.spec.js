// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import { yesNoFieldsOf, yesNoLabel } from "@/utils/collimato/yesNo.js";

describe("yesNoLabel", () => {
    it("names a row's true and false as Yes and No", () => {
        expect(yesNoLabel(true)).toBe("Yes");
        expect(yesNoLabel(false)).toBe("No");
    });

    it("names a filter's text, also as saved before the engine", () => {
        expect(yesNoLabel("true")).toBe("Yes");
        expect(yesNoLabel("1")).toBe("Yes");
        expect(yesNoLabel("false")).toBe("No");
        expect(yesNoLabel("0")).toBe("No");
    });

    it("gives anything else back as it is", () => {
        expect(yesNoLabel("maybe")).toBe("maybe");
    });
});

describe("yesNoFieldsOf", () => {
    it("lists the keys of the yes/no fields", () => {
        const dataModels = [
            {
                dimensions: [
                    { name: "orders.is_large", type: "boolean" },
                    { name: "orders.country", type: "string" },
                ],
            },
            { dimensions: [{ name: "stores.is_open", type: "boolean" }] },
        ];

        expect([...yesNoFieldsOf(dataModels)]).toEqual(["orders.is_large", "stores.is_open"]);
    });

    it("is empty before the data models are loaded", () => {
        expect(yesNoFieldsOf(undefined).size).toBe(0);
    });
});
