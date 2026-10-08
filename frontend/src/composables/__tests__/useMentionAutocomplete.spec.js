// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { effectScope, nextTick, ref } from "vue";
import { useMentionAutocomplete } from "../useMentionAutocomplete";

const people = ["anna", "andris", "bob"].map((username, i) => ({ id: `u${i}`, username }));

function setup(options = {}) {
    const text = ref("");
    const el = { selectionStart: 0, focus: vi.fn(), setSelectionRange: vi.fn() };
    const search = vi.fn(async (query, offset) =>
        people.filter((p) => p.username.startsWith(query)).slice(offset, offset + 2),
    );
    const scope = effectScope();
    const m = scope.run(() =>
        useMentionAutocomplete({
            text,
            input: () => el,
            search,
            pageSize: 2,
            specials: ["all", "here"],
            ...options,
        }),
    );
    const type = async (value) => {
        text.value = value;
        el.selectionStart = value.length;
        m.onInput();
        await vi.runAllTimersAsync();
    };

    return { m, text, el, search, type };
}

describe("useMentionAutocomplete", () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it("opens on an @ after a space and searches what follows it", async () => {
        const { m, search, type } = setup();

        await type("hi @an");

        expect(m.visible.value).toBe(true);
        expect(m.query.value).toBe("an");
        expect(search).toHaveBeenCalledWith("an", 0);
        expect(m.results.value.map((u) => u.username)).toEqual(["anna", "andris"]);
    });

    it("stays closed inside an email address", async () => {
        const { m, type } = setup();

        await type("mail anna@example");

        expect(m.visible.value).toBe(false);
    });

    it("reads the next page only while the last one was full", async () => {
        const { m, search, type } = setup();

        await type("@");
        expect(m.hasMore.value).toBe(true);

        await m.loadMore();
        expect(search).toHaveBeenLastCalledWith("", 2);
        expect(m.results.value.map((u) => u.username)).toEqual(["anna", "andris", "bob"]);
        expect(m.hasMore.value).toBe(false);
    });

    it("offers the special mentions that match, after or before the people", async () => {
        const after = setup();

        await after.type("@a");
        expect(after.m.entries.value).toEqual([
            { user: people[0] },
            { user: people[1] },
            { special: "all" },
        ]);

        const before = setup({ specialsFirst: true });

        await before.type("@h");
        expect(before.m.entries.value).toEqual([{ special: "here" }]);
    });

    it("puts the chosen handle in place of what was typed", async () => {
        const { m, text, el, type } = setup();

        await type("ask @an please");
        el.selectionStart = "ask @an".length;
        m.onInput();
        await vi.runAllTimersAsync();

        m.move(1);
        expect(m.chooseSelected()).toBe(true);
        await nextTick();

        expect(text.value).toBe("ask @andris  please");
        expect(m.visible.value).toBe(false);
        expect(el.setSelectionRange).toHaveBeenCalledWith(12, 12);
    });

    it("leaves Enter to the box when the list is closed", async () => {
        const { m, type } = setup();

        await type("no mention");

        expect(m.chooseSelected()).toBe(false);
    });

    it("drops results that arrive after a newer search began", async () => {
        const { m, search, type } = setup();
        let release;

        search.mockImplementationOnce(() => new Promise((resolve) => (release = resolve)));
        await type("@a");
        await type("@b");
        release([people[0]]);
        await vi.runAllTimersAsync();

        expect(m.results.value.map((u) => u.username)).toEqual(["bob"]);
    });
});
