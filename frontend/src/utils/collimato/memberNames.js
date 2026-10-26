// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

export function typingMemberName(text) {
    const withoutAccents = text.normalize("NFD").replace(/\p{M}/gu, "");
    const withUnderscores = withoutAccents.replace(/[\s-]/g, "_");

    return withUnderscores.replace(/[^A-Za-z0-9_]/g, "");
}

export function finishedMemberName(text) {
    const typed = typingMemberName(text);
    const squashed = typed.replace(/_+/g, "_");
    const trimmed = squashed.replace(/^_|_$/g, "");

    if (/^[0-9]/.test(trimmed)) {
        return "_" + trimmed;
    }

    return trimmed;
}

export function duplicateMemberNames(names) {
    const seen = new Set();
    const duplicates = new Set();

    for (const name of names) {
        if (!name) continue;

        if (seen.has(name)) {
            duplicates.add(name);
        }

        seen.add(name);
    }

    return duplicates;
}
