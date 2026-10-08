// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    DATABASE_TYPES,
    DEFAULT_DATABASE_TYPE,
    databaseLabel,
} from "@/utils/collimato/databaseTypes";

describe("the stored value is Cube's driver key", () => {
    it("uses the tokens Cube resolves a driver from", () => {
        expect(DATABASE_TYPES.map((db) => db.value)).toEqual(["mysql", "postgres"]);
    });

    it("defaults to one of the listed types", () => {
        expect(DATABASE_TYPES.some((db) => db.value === DEFAULT_DATABASE_TYPE)).toBe(true);
    });

    it("gives every type a label", () => {
        for (const db of DATABASE_TYPES) {
            expect(db.label).toBeTruthy();
            expect(db.label).not.toBe(db.value);
        }
    });
});

describe("databaseLabel", () => {
    it("names mysql for MariaDB users too", () => {
        expect(databaseLabel("mysql")).toBe("MySQL / MariaDB");
    });

    it("spells postgres out in full", () => {
        expect(databaseLabel("postgres")).toBe("PostgreSQL");
    });

    it("falls back to the stored value for a type it does not know", () => {
        expect(databaseLabel("clickhouse")).toBe("clickhouse");
    });

    it("survives a connection saved with no type", () => {
        expect(databaseLabel(undefined)).toBeUndefined();
    });
});
