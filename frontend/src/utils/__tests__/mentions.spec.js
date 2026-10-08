// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect } from "vitest";
import {
    splitMentionTokens,
    toComposerText,
    toDisplayText,
    fromComposerText,
} from "@/utils/mentions";

// Mirrors app/notifications_test.go, which runs the same corpus through
// extractMentions. Keep the two in step.
const ID = "0f8fad5bd9cb469fa16570867728950e";

// The Go parser returns only mentions, so compare on that projection.
function mentionsOf(text) {
    return splitMentionTokens(text).filter((t) => t.text === undefined);
}

function handlesOf(text) {
    return mentionsOf(text).map((t) => t.username);
}

describe("splitMentionTokens", () => {
    it("extracts plain handles", () => {
        expect(handlesOf("hey @bob and @all")).toEqual(["bob", "all"]);
    });

    it("extracts directory style names", () => {
        expect(handlesOf("ping @jane.doe and @jane-doe and @jane_doe")).toEqual([
            "jane.doe",
            "jane-doe",
            "jane_doe",
        ]);
    });

    it("drops trailing punctuation", () => {
        expect(handlesOf("ask @jane.doe. then @bob- or @carol_")).toEqual([
            "jane.doe",
            "bob",
            "carol",
        ]);
    });

    it("extracts the id form", () => {
        expect(mentionsOf(`hey <@${ID}> there`)).toEqual([{ userId: ID }]);
    });

    it("extracts mixed forms", () => {
        expect(mentionsOf(`<@${ID}> and @bob and @all`)).toEqual([
            { userId: ID },
            { username: "bob" },
            { username: "all" },
        ]);
    });

    it("falls back to the name form for a malformed id", () => {
        expect(mentionsOf(`<@abc123> and <@${ID}x>`)).toEqual([
            { username: "abc123" },
            { username: `${ID}x` },
        ]);
    });

    it("keeps the surrounding text", () => {
        expect(splitMentionTokens(`hey <@${ID}> there`)).toEqual([
            { text: "hey " },
            { userId: ID },
            { text: " there" },
        ]);
    });

    it("returns one text token when there is no mention", () => {
        expect(splitMentionTokens("nothing here")).toEqual([{ text: "nothing here" }]);
    });

    it("returns nothing for empty text", () => {
        expect(splitMentionTokens("")).toEqual([]);
    });
});

describe("edit round trip", () => {
    const JANE = "0f8fad5bd9cb469fa16570867728950e";
    const GHOST = "11111111111111111111111111111111";
    const WRAITH = "22222222222222222222222222222222";

    const names = { [JANE]: "jane" };
    const resolve = (id) => names[id] ?? null;

    function roundTrip(stored, resolver = resolve) {
        const { text, restore } = toComposerText(stored, resolver);

        return { shown: text, saved: fromComposerText(text, restore) };
    }

    it("shows the current name and saves the same id", () => {
        const { shown, saved } = roundTrip(`hey <@${JANE}> there`);

        expect(shown).toBe("hey @jane there");
        expect(saved).toBe(`hey <@${JANE}> there`);
    });

    it("shows the new name after a rename and still saves the id", () => {
        const { shown, saved } = roundTrip(`hey <@${JANE}>`, () => "jsmith");

        expect(shown).toBe("hey @jsmith");
        expect(saved).toBe(`hey <@${JANE}>`);
    });

    it("never puts a raw id in the buffer", () => {
        const { text } = toComposerText(`a <@${JANE}> b <@${GHOST}>`, resolve);

        expect(text).not.toContain(JANE);
        expect(text).not.toContain(GHOST);
    });

    it("keeps a deleted user's mention through an edit", () => {
        const { shown, saved } = roundTrip(`hey <@${GHOST}> there`);

        expect(shown).toBe("hey @unknown-user there");
        expect(saved).toBe(`hey <@${GHOST}> there`);
    });

    it("keeps two deleted users distinct", () => {
        const stored = `<@${GHOST}> and <@${WRAITH}>`;
        const { shown, saved } = roundTrip(stored);

        expect(shown).toBe("@unknown-user-1 and @unknown-user-2");
        expect(saved).toBe(stored);
    });

    it("keeps two deleted users distinct when reordered", () => {
        const { restore } = toComposerText(`<@${GHOST}> and <@${WRAITH}>`, resolve);
        const swapped = "@unknown-user-2 and @unknown-user-1";

        expect(fromComposerText(swapped, restore)).toBe(`<@${WRAITH}> and <@${GHOST}>`);
    });

    // A real user could hold this username.
    it("does not hand an id to a freshly typed marker", () => {
        const { restore } = toComposerText(`hey <@${GHOST}>`, resolve);
        const edited = "hey @unknown-user and @unknown-user";

        expect(fromComposerText(edited, restore)).toBe(`hey <@${GHOST}> and @unknown-user`);
    });

    it("leaves a mention the user edited for the server to resolve", () => {
        const { restore } = toComposerText(`hey <@${JANE}>`, resolve);

        expect(fromComposerText("hey @janet", restore)).toBe("hey @janet");
    });

    it("drops the mention the user deleted without disturbing the rest", () => {
        const { restore } = toComposerText(`<@${JANE}> and <@${GHOST}>`, resolve);

        expect(fromComposerText("@unknown-user alone", restore)).toBe(`<@${GHOST}> alone`);
    });

    it("keeps a legacy name mention as a name", () => {
        const { shown, saved } = roundTrip(`<@${JANE}> and @bob and @all`);

        expect(shown).toBe("@jane and @bob and @all");
        expect(saved).toBe(`<@${JANE}> and @bob and @all`);
    });

    it("leaves ids inside code and links untouched", () => {
        const stored = `\`<@${JANE}>\` and [<@${JANE}>](https://e.com)`;
        const { text, restore } = toComposerText(stored, resolve);

        expect(text).toBe(stored);
        expect(restore).toHaveLength(0);
    });

    it("leaves ids inside a fenced block untouched", () => {
        const stored = "```\n<@" + JANE + ">\n```";

        expect(toComposerText(stored, resolve).text).toBe(stored);
    });

    it("passes a message with no mentions straight through", () => {
        const { shown, saved } = roundTrip("nothing to see");

        expect(shown).toBe("nothing to see");
        expect(saved).toBe("nothing to see");
    });
});

describe("resolving on send", () => {
    const JANE = "0f8fad5bd9cb469fa16570867728950e";
    const GHOST = "11111111111111111111111111111111";

    const members = { jane: JANE };
    const resolve = (handle) => members[handle] ?? null;

    function send(text, { restore = [], skip = [] } = {}) {
        return fromComposerText(text, restore, resolve, skip);
    }

    it("stores a resolvable handle as its id", () => {
        expect(send("hey @jane can you look")).toBe(`hey <@${JANE}> can you look`);
    });

    it("stores an unresolvable handle as the name", () => {
        expect(send("hey @nobody")).toBe("hey @nobody");
    });

    it("never resolves the broadcast keywords", () => {
        expect(send("@all and @here")).toBe("@all and @here");
    });

    it("prefers a restore entry over the resolver", () => {
        const restore = [{ marker: "jane", id: GHOST }];

        expect(send("hey @jane", { restore })).toBe(`hey <@${GHOST}>`);
    });

    it("resolves a second mention the restore entry did not cover", () => {
        const restore = [{ marker: "unknown-user", id: GHOST }];

        expect(send("@unknown-user and @jane", { restore })).toBe(`<@${GHOST}> and <@${JANE}>`);
    });

    it("leaves a handle the message already had", () => {
        expect(send("hey @jane, fixed a typo", { skip: ["jane"] })).toBe("hey @jane, fixed a typo");
    });

    it("matches an already-present handle regardless of case", () => {
        expect(send("hey @Jane", { skip: ["jane"] })).toBe("hey @Jane");
    });

    it("still resolves a handle the message did not already have", () => {
        expect(send("@jane and @jane", { skip: ["bob"] })).toBe(`<@${JANE}> and <@${JANE}>`);
    });

    it("leaves handles inside code and links alone", () => {
        const text = "`@jane` and [@jane](https://e.com)";

        expect(send(text)).toBe(text);
    });

    it("leaves an id already in the text untouched", () => {
        expect(send(`hey <@${GHOST}>`)).toBe(`hey <@${GHOST}>`);
    });

    it("sends the text unchanged with no resolver", () => {
        expect(fromComposerText("hey @jane", [], undefined, [])).toBe("hey @jane");
    });
});

describe("length after resolving", () => {
    const JANE = "0f8fad5bd9cb469fa16570867728950e";

    it("grows a short mention into a full id", () => {
        const shown = "@jane";
        const stored = fromComposerText(shown, [], () => JANE, []);

        expect([...shown].length).toBe(5);
        expect([...stored].length).toBe(35);
    });
});

describe("edit as the composer runs it", () => {
    const JANE = "0f8fad5bd9cb469fa16570867728950e";
    const GHOST = "11111111111111111111111111111111";
    const BOB = "22222222222222222222222222222222";

    const nameOf = { [JANE]: "jane" };
    const idOf = { jane: JANE, bob: BOB };

    function editAndSend(stored, edit) {
        const { text, restore, existing } = toComposerText(stored, (id) => nameOf[id] ?? null);

        return fromComposerText(edit(text), restore, (handle) => idOf[handle] ?? null, existing);
    }

    it("keeps every mention when nothing is touched", () => {
        const stored = `<@${JANE}> and <@${GHOST}> and @bob`;

        expect(editAndSend(stored, (t) => t)).toBe(stored);
    });

    it("resolves a mention added during the edit", () => {
        const stored = `hey <@${JANE}>`;

        expect(editAndSend(stored, (t) => `${t} and @bob`)).toBe(`hey <@${JANE}> and <@${BOB}>`);
    });

    it("leaves a mention that was already there as a name", () => {
        const stored = "hey @bob, typo here";

        expect(editAndSend(stored, (t) => t.replace("typo", "fixed"))).toBe("hey @bob, fixed here");
    });

    it("never exposes an id to the edit step", () => {
        const stored = `<@${JANE}> and <@${GHOST}>`;

        editAndSend(stored, (t) => {
            expect(t).not.toContain(JANE);
            expect(t).not.toContain(GHOST);

            return t;
        });
    });
});

describe("toDisplayText", () => {
    const JANE = "0f8fad5bd9cb469fa16570867728950e";
    const GHOST = "11111111111111111111111111111111";

    const resolve = (id) => (id === JANE ? "jane" : null);

    it("spells out an id for somewhere that cannot render a component", () => {
        expect(toDisplayText(`hey <@${JANE}>`, resolve)).toBe("hey @jane");
    });

    it("never leaks an id", () => {
        const out = toDisplayText(`<@${JANE}> and <@${GHOST}>`, resolve);

        expect(out).not.toContain(JANE);
        expect(out).not.toContain(GHOST);
    });

    it("leaves a legacy name mention alone", () => {
        expect(toDisplayText("hey @bob", resolve)).toBe("hey @bob");
    });
});
