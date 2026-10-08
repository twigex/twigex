// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index";

// Built-in roles carry a translation key as their display name, or only their
// technical name where an older edit replaced it. Custom roles keep whatever
// their creator typed.
const KEY_PREFIX = "authentication.roles.";

const BUILT_IN = {
    system: {
        system_admin: "authentication.roles.system_admin",
        system_user: "authentication.roles.system_user",
    },
    collimato: {
        workspace_admin: "authentication.roles.collimato_admin",
        workspace_user: "authentication.roles.collimato_user",
    },
    projects: {
        admin: "authentication.roles.admin",
        user: "authentication.roles.user",
    },
};

function builtInKey(name, scope) {
    return BUILT_IN[scope]?.[name] ?? null;
}

function translated(text) {
    return text?.startsWith(KEY_PREFIX) ? t.value(text) : text;
}

export function isBuiltInRole(name, scope) {
    return builtInKey(name, scope) !== null;
}

export function roleLabel(role, scope) {
    const key = builtInKey(role?.name, scope);

    if (key) return t.value(`${key}.name`);

    return translated(role?.display_name) || role?.name || "";
}

export function roleDescription(role, scope) {
    const key = builtInKey(role?.name, scope);

    if (key) return t.value(`${key}.description`);

    return translated(role?.description) || "";
}

export function roleNameLabel(name, scope) {
    const key = builtInKey(name, scope);

    return key ? t.value(`${key}.name`) : translated(name);
}
