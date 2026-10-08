// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import emojiData from "emoji-datasource-google/emoji.json";

const EMOJI_SHEET_URL = "/sheet_google_64_clean.png";

const SHEET_COLUMNS = 62;

// Percentages land on cell boundaries at any size and any device pixel ratio,
// where fixed pixel offsets bleed the neighbouring cell at fractional scaling.
function buildStyle(emoji) {
    return {
        width: "32px",
        height: "32px",
        backgroundImage: `url(${EMOJI_SHEET_URL})`,
        backgroundPosition: `${(emoji.sheet_x / (SHEET_COLUMNS - 1)) * 100}% ${(emoji.sheet_y / (SHEET_COLUMNS - 1)) * 100}%`,
        backgroundSize: `${SHEET_COLUMNS * 100}% ${SHEET_COLUMNS * 100}%`,
        display: "inline-block",
    };
}

// Identities have to stay stable: rebuilt per render, Vue re-diffs all ~1900
// picker cells on every scroll.
const byCategory = new Map();
const byShortName = new Map();
const styles = new Map();

export const searchableEmojis = [];

for (const emoji of emojiData) {
    if (!emoji.has_img_google) {
        continue;
    }

    if (!byCategory.has(emoji.category)) {
        byCategory.set(emoji.category, []);
    }

    byCategory.get(emoji.category).push(emoji);
    byShortName.set(emoji.short_name, emoji);
    styles.set(emoji.short_name, buildStyle(emoji));
    searchableEmojis.push(emoji);
}

for (const list of byCategory.values()) {
    list.sort((a, b) => a.sort_order - b.sort_order);
}

export function emojisInCategory(category) {
    return byCategory.get(category) ?? [];
}

export function emojiByShortName(shortName) {
    return byShortName.get(shortName);
}

export function emojiStyle(emoji) {
    return styles.get(emoji.short_name);
}

// One image per emoji, the files chat's renderer uses, so a comment with a few
// emoji does not pull the whole 13 MB sheet.
export function emojiImageUrl(emoji) {
    return `/emoji-datasource/img/google/64/${emoji.image}`;
}

export function emojiCharacter(emoji) {
    return String.fromCodePoint(...emoji.unified.split("-").map((hex) => parseInt(hex, 16)));
}

const SHORT_NAME_PATTERN = /:([a-z0-9_+-]+):/g;

// Splits text into {text} and {emoji} tokens; a :name: that is not an emoji
// stays text.
export function splitEmojiTokens(text) {
    SHORT_NAME_PATTERN.lastIndex = 0;

    const out = [];
    let last = 0;
    let match;

    while ((match = SHORT_NAME_PATTERN.exec(text)) !== null) {
        const emoji = byShortName.get(match[1]);

        if (!emoji) {
            SHORT_NAME_PATTERN.lastIndex = match.index + 1;
            continue;
        }

        if (match.index > last) {
            out.push({ text: text.slice(last, match.index) });
        }

        out.push({ emoji: emoji.short_name });
        last = SHORT_NAME_PATTERN.lastIndex;
    }

    if (last < text.length) out.push({ text: text.slice(last) });

    return out;
}

const QUICK_REACTION_COUNT = 3;

const DEFAULT_QUICK_REACTIONS = ["+1", "heart", "joy", "smile", "tada", "-1", "clap"];

export function quickReactionEmojis(recents) {
    const picked = recents.slice(0, QUICK_REACTION_COUNT);

    for (const name of DEFAULT_QUICK_REACTIONS) {
        if (picked.length >= QUICK_REACTION_COUNT) {
            break;
        }

        if (!picked.includes(name)) {
            picked.push(name);
        }
    }

    return picked;
}

let warmed = false;

// The sheet is 13.2 MB and decodes to ~64 MB, so leaving it to the click that
// opens the picker is what makes the first open slow.
export function warmEmojiSheet() {
    if (warmed || typeof window === "undefined") {
        return;
    }

    warmed = true;

    const load = () => {
        new Image().src = EMOJI_SHEET_URL;
    };

    if (window.requestIdleCallback) {
        window.requestIdleCallback(load, { timeout: 3000 });

        return;
    }

    window.setTimeout(load, 1000);
}
