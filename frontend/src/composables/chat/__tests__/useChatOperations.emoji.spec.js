// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { describe, it, expect, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useChatOperations from "@/composables/chat/useChatOperations";

let renderEmoji;

beforeEach(() => {
    setActivePinia(createPinia());
    ({ renderEmoji } = useChatOperations());
});

describe("renderEmoji", () => {
    it("keeps the sprite markup a known shortcode produces", () => {
        const html = renderEmoji(":smile:");

        expect(html).toContain("emoji-sizer");
        expect(html).toContain("/emoji-datasource/img/google/64/1f604.png");
        expect(html).toContain('data-codepoints="1f604"');
    });

    it("drops a payload replace_colons passes through untouched", () => {
        expect(renderEmoji("<img src=x onerror=alert(1)>")).toBe("");
        expect(renderEmoji("<script>alert(1)</script>")).toBe("");
    });

    it("strips an anchor while still converting the shortcode after it", () => {
        const html = renderEmoji("<a href=# onclick=steal()>hi</a>:+1:");

        expect(html).not.toContain("onclick");
        expect(html).not.toContain("<a");
        expect(html).toContain("/emoji-datasource/img/google/64/1f44d.png");
    });

    it("leaves plain text alone", () => {
        expect(renderEmoji("no emoji here")).toBe("no emoji here");
    });
});
