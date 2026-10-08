// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, expect, it } from "vitest";
import { emojiByShortName, emojiCharacter, quickReactionEmojis, splitEmojiTokens } from "../emoji";

describe("splitEmojiTokens", () => {
    it("turns known short names into emoji and keeps the text around them", () => {
        expect(splitEmojiTokens("nice :+1: work :smile:")).toEqual([
            { text: "nice " },
            { emoji: "+1" },
            { text: " work " },
            { emoji: "smile" },
        ]);
    });

    it("leaves an unknown name as text", () => {
        expect(splitEmojiTokens("at 10:30:00 and :nope:")).toEqual([
            { text: "at 10:30:00 and :nope:" },
        ]);
    });

    it("finds an emoji whose first colon closed an unknown name", () => {
        expect(splitEmojiTokens("ratio 1:smile:")).toEqual([
            { text: "ratio 1" },
            { emoji: "smile" },
        ]);
        expect(splitEmojiTokens("a:nope:smile:")).toEqual([{ text: "a:nope" }, { emoji: "smile" }]);
    });
});

describe("emojiCharacter", () => {
    it("builds the character from its code points", () => {
        expect(emojiCharacter(emojiByShortName("smile"))).toBe("😄");
        expect(emojiCharacter(emojiByShortName("hash"))).toBe("#️⃣");
    });
});

describe("quickReactionEmojis", () => {
    it("uses the first three defaults when nothing was used recently", () => {
        expect(quickReactionEmojis([])).toEqual(["+1", "heart", "joy"]);
    });

    it("pads with defaults that are not already listed", () => {
        expect(quickReactionEmojis(["heart"])).toEqual(["heart", "+1", "joy"]);
        expect(quickReactionEmojis(["+1", "heart"])).toEqual(["+1", "heart", "joy"]);
        expect(quickReactionEmojis(["fire", "joy"])).toEqual(["fire", "joy", "+1"]);
    });

    it("keeps only the three most recent", () => {
        expect(quickReactionEmojis(["a", "b", "c", "d"])).toEqual(["a", "b", "c"]);
    });
});
