// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export const DATABASE_TYPES = [
    { value: "mysql", label: "MySQL / MariaDB" },
    { value: "postgres", label: "PostgreSQL" },
];

export const DEFAULT_DATABASE_TYPE = DATABASE_TYPES[0].value;

export function databaseLabel(type) {
    return DATABASE_TYPES.find((db) => db.value === type)?.label ?? type;
}
