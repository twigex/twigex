// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Mirrors mentionHandleRe in app/notifications.go. The ID branch comes first so
// a malformed <@abc> falls through to the name branch instead of being dropped.
const MENTION_PATTERN = /<@([0-9a-f]{32})>|@([a-zA-Z0-9](?:[a-zA-Z0-9._-]*[a-zA-Z0-9])?)/g;

// Mirrors the strip patterns in app/notifications.go. The renderer excludes
// these through its DOM walk, so the composer has to exclude them here.
const PROTECTED_PATTERN = /```[\s\S]*?```|`[^`]*`|\[[^\]]*\]\([^)]*\)/g;

// Not translated, and ASCII: it stands in for a handle, and a localised one
// would not survive the round trip back to an id. Shared with the renderer so
// a reader and an editor see the same word.
export const UNKNOWN_USER = "unknown-user";

function mapOutsideProtected(text, fn) {
    PROTECTED_PATTERN.lastIndex = 0;

    let out = "";
    let last = 0;
    let match;

    while ((match = PROTECTED_PATTERN.exec(text)) !== null) {
        if (match.index > last) out += fn(text.slice(last, match.index));
        out += match[0];
        last = PROTECTED_PATTERN.lastIndex;
    }

    if (last < text.length) out += fn(text.slice(last));

    return out;
}

// Returns {text}, {userId} and {username} tokens. Free of Vue so this and the
// Go parser can be tested on one corpus.
export function splitMentionTokens(text) {
    MENTION_PATTERN.lastIndex = 0;

    const out = [];
    let last = 0;
    let match;

    while ((match = MENTION_PATTERN.exec(text)) !== null) {
        if (match.index > last) {
            out.push({ text: text.slice(last, match.index) });
        }

        out.push(match[1] ? { userId: match[1] } : { username: match[2] });
        last = MENTION_PATTERN.lastIndex;
    }

    if (last < text.length) out.push({ text: text.slice(last) });

    return out;
}

// Returns what the edit textarea shows, the restore list that puts it back,
// and the name-form handles the body already had. resolve(id) gives a username
// or null. Every id is restorable, not only the nameless ones: otherwise an
// edit could degrade an id to a name.
export function toComposerText(text, resolve) {
    const unresolved = [];
    const existing = [];

    mapOutsideProtected(text, (chunk) => {
        for (const token of splitMentionTokens(chunk)) {
            if (token.username !== undefined) {
                existing.push(token.username.toLowerCase());
                continue;
            }

            if (token.userId === undefined || resolve(token.userId)) continue;
            if (!unresolved.includes(token.userId)) {
                unresolved.push(token.userId);
            }
        }

        return chunk;
    });

    // Two deleted users would otherwise share one marker, and reordering them
    // in the textarea would swap who the message refers to.
    const numbered = unresolved.length > 1;
    const restore = [];

    const composed = mapOutsideProtected(text, (chunk) =>
        splitMentionTokens(chunk)
            .map((token) => {
                if (token.text !== undefined) return token.text;
                if (token.username !== undefined) return `@${token.username}`;

                const username = resolve(token.userId);
                const marker =
                    username ||
                    (numbered
                        ? `${UNKNOWN_USER}-${unresolved.indexOf(token.userId) + 1}`
                        : UNKNOWN_USER);

                restore.push({ marker, id: token.userId });

                return `@${marker}`;
            })
            .join(""),
    );

    return { text: composed, restore, existing };
}

// A stored body as readable text, for somewhere that cannot render a
// component: a desktop notification, a preview.
export function toDisplayText(text, resolve) {
    return toComposerText(text, resolve).text;
}

// Produces the body to store. `restore` is consulted before `resolve` so a
// mention already in the message keeps its id, and each entry is spent once so
// a handle the writer typed cannot take one. `skip` holds handles the body
// already had: one written before ids existed may already name the wrong
// person, and editing a typo should not freeze that guess in.
export function fromComposerText(text, restore, resolve, skip) {
    const pending = restore ? [...restore] : [];
    const untouched = new Set(skip ?? []);

    return mapOutsideProtected(text, (chunk) =>
        splitMentionTokens(chunk)
            .map((token) => {
                if (token.text !== undefined) return token.text;
                if (token.userId !== undefined) return `<@${token.userId}>`;

                const i = pending.findIndex((r) => r.marker === token.username);

                if (i !== -1) return `<@${pending.splice(i, 1)[0].id}>`;

                if (untouched.has(token.username.toLowerCase())) {
                    return `@${token.username}`;
                }

                const id = resolve?.(token.username);

                return id ? `<@${id}>` : `@${token.username}`;
            })
            .join(""),
    );
}
