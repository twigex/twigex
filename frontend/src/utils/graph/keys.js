// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const KEY_LIMIT = 64;

export const KEY_PATTERN = /^[a-z][a-z0-9_]{0,63}$/;

// suggestKey turns a name into a key the server accepts: lowercase letters,
// digits and underscores, starting with a letter, at most 64 characters.
// "Purchase order" becomes "purchase_order", "Ēka 2" becomes "eka_2".
export function suggestKey(...parts) {
    const text = parts.join(" ");

    const plain = text.normalize("NFD").replace(/[̀-ͯ]/g, "");
    const lowered = plain.toLowerCase().replace(/ł/g, "l");
    const joined = lowered.replace(/[^a-z0-9]+/g, "_");
    const trimmed = joined.replace(/^[^a-z]+/, "").replace(/_+$/, "");

    return trimmed.slice(0, KEY_LIMIT).replace(/_+$/, "");
}

// fallbackKey is for a name with no Latin letters to make a key from, such as
// "Компания". The admin can still type a key of their own.
export function fallbackKey(prefix) {
    const suffix = Math.random().toString(36).slice(2, 8);

    return `${prefix}_${suffix}`;
}
