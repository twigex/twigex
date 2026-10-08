// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// safeHref gives what a stored link may open: http, https and mailto as they
// are, a link without a scheme as https, and nothing for any other scheme,
// which javascript: would use to run in the browser of whoever clicks it.
// The server refuses such links, but some were stored before it did.
export function safeHref(link) {
    if (typeof link !== "string") return null;

    const trimmed = link.trim();
    const lower = trimmed.toLowerCase();

    if (!trimmed) return null;
    if (["http://", "https://", "mailto:"].some((scheme) => lower.startsWith(scheme)))
        return trimmed;

    const i = lower.search(/[:/?#]/);

    if (i !== -1 && lower[i] === ":") return null;

    return `https://${trimmed}`;
}
