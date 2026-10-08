// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import en from "@/i18n/en.json";
import lv from "@/i18n/lv.json";
import kk from "@/i18n/kk.json";
import pl from "@/i18n/pl.json";
import { rolePermissionGroups } from "../rolePermissions";

const keysOf = (perTable) => {
    const keys = [];

    rolePermissionGroups((key) => (keys.push(key), key), perTable);

    return keys;
};

describe("rolePermissionGroups", () => {
    it("asks only for keys every locale has", () => {
        for (const [locale, messages] of Object.entries({ en, lv, kk, pl })) {
            const missing = keysOf(false).filter((key) => !(key in messages));

            expect(missing, locale).toEqual([]);
        }
    });

    it("leaves out the workspace-wide field permissions for a role set per table", () => {
        const fields = (perTable) =>
            rolePermissionGroups((key) => key, perTable).find((g) => g.key === "fields").items;

        expect(fields(false).map((p) => p.value)).toEqual(["create_fields", "edit_fields"]);
        expect(fields(true)).toEqual([]);
    });
});
