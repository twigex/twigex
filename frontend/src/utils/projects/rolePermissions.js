// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const SECTIONS = [
    { key: "rows", permissions: ["create_task", "update_task", "delete_task"] },
    {
        key: "tables",
        permissions: ["create_table", "update_workspace_table", "delete_workspace_table"],
    },
    { key: "fields", permissions: ["create_fields", "edit_fields"] },
    {
        key: "views",
        permissions: ["create_workspace_view", "update_workspace_view", "delete_table_view"],
    },
    { key: "folders", permissions: ["update_workspace_folder", "delete_workspace_folder"] },
    { key: "members", permissions: ["add_member_to_workspace", "delete_workspace_member"] },
    { key: "roles", permissions: ["create_roles", "update_roles", "delete_roles"] },
    { key: "workspace", permissions: ["update_workspace"] },
    { key: "other", permissions: ["show_assigned_tasks_only"] },
];

// rolePermissionGroups lists the workspace permissions a role can be given,
// by section. A role restricted to chosen tables gets its field permissions
// per table, so the workspace-wide ones are left out.
export function rolePermissionGroups(translate, perTable) {
    return SECTIONS.map(({ key, permissions }) => ({
        key,
        label: translate(`projects.create_role.section.${key}`),
        items:
            key === "fields" && perTable
                ? []
                : permissions.map((value) => ({
                      value,
                      label: translate(`projects.create_role.${value}.label`),
                      description: translate(`projects.create_role.${value}.description`),
                  })),
    }));
}
