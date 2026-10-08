// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import fileTypes from "@/constants/fileTypes";

const officeTypeHints = [
    "word",
    "excel",
    "spreadsheet",
    "presentation",
    "powerpoint",
    "opendocument",
    "officedocument",
    "msword",
    "rtf",
    "csv",
];

export function typeMeta(child) {
    if (child.IsFolder) return fileTypes["inode/directory"];

    return fileTypes[child.Type] ?? fileTypes["not-found"];
}

export function isOfficeType(type) {
    if (!type) return false;
    const lower = type.toLowerCase();

    return officeTypeHints.some((hint) => lower.includes(hint));
}

export function isImage(type) {
    return typeof type === "string" && type.startsWith("image/");
}

export function previewable(type) {
    return (
        typeof type === "string" &&
        (type.includes("image") || type.includes("video") || type.includes("pdf"))
    );
}
