// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { isBuiltInRole, roleDescription, roleLabel, roleNameLabel } from "../roleLabels";

describe("roleLabel", () => {
    it("translates built-in roles whether they carry a key or only their name", () => {
        expect(
            roleLabel(
                {
                    name: "workspace_admin",
                    display_name: "authentication.roles.collimato_admin.name",
                },
                "collimato",
            ),
        ).toBe("Workspace admin");
        expect(roleLabel({ name: "admin", display_name: "admin" }, "projects")).toBe("Admin");
        expect(roleLabel({ name: "system_admin", display_name: "Administrator" }, "system")).toBe(
            "Administrator",
        );
    });

    it("keeps what a custom role's creator typed", () => {
        expect(roleLabel({ name: "analysts", display_name: "Data analysts" }, "collimato")).toBe(
            "Data analysts",
        );
        expect(roleLabel({ name: "analysts" }, "collimato")).toBe("analysts");
    });

    it("only treats a name as built in within its own section", () => {
        expect(roleLabel({ name: "admin", display_name: "Team admin" }, "collimato")).toBe(
            "Team admin",
        );
        expect(isBuiltInRole("admin", "projects")).toBe(true);
        expect(isBuiltInRole("admin", "collimato")).toBe(false);
    });
});

describe("roleDescription", () => {
    it("translates built-in descriptions and keeps custom ones", () => {
        expect(roleDescription({ name: "user" }, "projects")).toBe("Can create and update tasks.");
        expect(
            roleDescription({ name: "analysts", description: "Reads reports" }, "projects"),
        ).toBe("Reads reports");
    });
});

describe("roleNameLabel", () => {
    it("labels a bare role name or a stored key", () => {
        expect(roleNameLabel("workspace_user", "collimato")).toBe("Workspace user");
        expect(roleNameLabel("authentication.roles.admin.name", "projects")).toBe("Admin");
        expect(roleNameLabel("analysts", "collimato")).toBe("analysts");
    });
});
