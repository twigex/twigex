// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { beforeEach, describe, expect, it } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { useWorkspaceRolesStore } from "../workspaceRoles";

describe("canRowAction", () => {
    let roles;

    beforeEach(() => {
        setActivePinia(createPinia());
        roles = useWorkspaceRolesStore();
    });

    it("allows everything before a role is loaded", () => {
        expect(roles.canRowAction("t1", "update_task")).toBe(true);
    });

    it("follows the workspace role outside per-table mode", () => {
        roles.setUser({ permissions: ["create_task"] });

        expect(roles.canRowAction("t1", "create_task")).toBe(true);
        expect(roles.canRowAction("t1", "update_task")).toBe(false);
    });

    it("follows each table's own permissions in per-table mode", () => {
        roles.setUser({
            permissions: ["update_task"],
            per_table_mode: true,
            table_permissions: [
                { table_id: "t1", action: "update_task" },
                { table_id: "t2", action: "create_task" },
            ],
        });

        expect(roles.canRowAction("t1", "update_task")).toBe(true);
        expect(roles.canRowAction("t2", "update_task")).toBe(false);
    });
});
