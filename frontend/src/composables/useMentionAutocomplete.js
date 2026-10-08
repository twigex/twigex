// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, nextTick, onBeforeUnmount, ref } from "vue";

// An @ that starts a mention: at the start of the text or after a space, so
// an email address does not open the list, followed by what a handle may hold.
const MENTION_AT_CARET = /(^|\s)@([\p{L}\p{M}0-9._-]*)$/u;

// useMentionAutocomplete is the list of people a text box offers after an @:
// who matches what follows the @ before the cursor, read page by page from
// search, plus the special mentions that match. Choosing one puts its handle
// in place of what was typed. Each caller draws the list its own way.
//
// text is the box's text, input returns its element, and search(query,
// offset) resolves to the next page of users.
export function useMentionAutocomplete({
    text,
    input,
    search,
    pageSize = 20,
    specials = [],
    specialsFirst = false,
}) {
    const visible = ref(false);
    const query = ref("");
    const results = ref([]);
    const selected = ref(0);
    const hasMore = ref(true);
    const loading = ref(false);

    let timer = null;
    let seq = 0;

    const matchingSpecials = computed(() =>
        specials.filter((name) => name.startsWith(query.value.toLowerCase())),
    );

    // entries is what arrow keys and Enter move through, in the order shown.
    const entries = computed(() => {
        const people = results.value.map((user) => ({ user }));
        const special = matchingSpecials.value.map((name) => ({ special: name }));

        return specialsFirst ? [...special, ...people] : [...people, ...special];
    });

    async function loadPage() {
        if (loading.value || !hasMore.value) return;
        const at = seq;

        loading.value = true;
        try {
            const batch = (await search(query.value, results.value.length)) || [];

            if (at !== seq) return;
            results.value = [...results.value, ...batch];
            hasMore.value = batch.length === pageSize;
        } catch {
            if (at === seq) hasMore.value = false;
        } finally {
            if (at === seq) loading.value = false;
        }
    }

    function reset() {
        seq++;
        results.value = [];
        hasMore.value = true;
        loading.value = false;

        return loadPage();
    }

    function close() {
        clearTimeout(timer);
        visible.value = false;
        query.value = "";
    }

    // onInput looks before the cursor for a mention being typed, and opens or
    // closes the list; the search waits for typing to pause.
    function onInput() {
        const el = input();
        const caret = el?.selectionStart ?? text.value.length;
        const m = text.value.slice(0, caret).match(MENTION_AT_CARET);

        if (!m) {
            close();

            return;
        }

        if (!visible.value || query.value !== m[2]) selected.value = 0;
        query.value = m[2];
        visible.value = true;
        clearTimeout(timer);
        timer = setTimeout(reset, 200);
    }

    // insert puts @handle and a space in place of the mention being typed,
    // and leaves the cursor after it.
    function insert(handle) {
        const el = input();
        const caret = el?.selectionStart ?? text.value.length;
        const before = text.value.slice(0, caret);
        const at = before.lastIndexOf("@");
        const start = at === -1 ? caret : at;

        text.value = `${text.value.slice(0, start)}@${handle} ${text.value.slice(caret)}`;
        close();

        nextTick(() => {
            const pos = start + handle.length + 2;

            el?.focus();
            el?.setSelectionRange(pos, pos);
        });
    }

    function choose(entry) {
        if (!entry) return;
        insert(entry.special || entry.user.username);
    }

    function move(direction) {
        const n = entries.value.length;

        if (n) selected.value = (selected.value + direction + n) % n;
    }

    // chooseSelected takes the highlighted entry when the list is open, and
    // reports whether it did, so Enter can fall back to its usual job.
    function chooseSelected() {
        if (!visible.value || entries.value.length === 0) return false;
        choose(entries.value[selected.value]);

        return true;
    }

    onBeforeUnmount(() => clearTimeout(timer));

    return {
        visible,
        query,
        results,
        selected,
        hasMore,
        loading,
        entries,
        matchingSpecials,
        onInput,
        reset,
        loadMore: loadPage,
        close,
        choose,
        insert,
        move,
        chooseSelected,
    };
}
