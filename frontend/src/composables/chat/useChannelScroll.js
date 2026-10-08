// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, computed, nextTick } from "vue";

// Answers where the reader is; deciding what to do about it is the view's
// business, because only the view knows what has already been fetched.

// Sub-pixel rounding means scrollTop rarely lands exactly on the maximum even
// when the view is at the bottom.
const BOTTOM_SLACK = 10;

const HIGHLIGHT_MS = 2000;
const SCROLL_START_MS = 200;
const SETTLE_MS = 1000;

export function useChannelScroll() {
    const container = ref(null);

    const scrollTop = ref(0);
    const maxScrollTop = ref(0);
    const viewportHeight = ref(0);
    const movedUp = ref(false);

    const distanceFromTop = computed(() => scrollTop.value);
    const distanceFromBottom = computed(() => maxScrollTop.value - scrollTop.value);
    const atBottom = computed(() => distanceFromBottom.value <= BOTTOM_SLACK);

    // Reads the direction before overwriting the position, since the direction
    // is the difference between the two.
    function record(event) {
        const el = event.target;
        const previousTop = scrollTop.value;

        scrollTop.value = el.scrollTop;
        movedUp.value = el.scrollTop < previousTop;
        maxScrollTop.value = el.scrollHeight - el.clientHeight;
        viewportHeight.value = el.clientHeight;
    }

    function scrollToNewest() {
        nextTick(() => {
            if (container.value) {
                container.value.scrollTop = container.value.scrollHeight;
            }
        });
    }

    function followNewest() {
        if (atBottom.value) scrollToNewest();
    }

    // Older messages go on above, pushing everything down by their height.
    // Measure before, restore after, or the reader is thrown forward by
    // however much arrived.
    function preservePositionWhile(mutate) {
        const el = container.value;

        if (!el) {
            mutate();

            return;
        }

        const heightBefore = el.scrollHeight;
        const topBefore = el.scrollTop;

        mutate();

        nextTick(() => {
            el.scrollTop = topBefore + (el.scrollHeight - heightBefore);
        });
    }

    // `onSettled` fires even when the message is not on the page, or whatever
    // the caller turned off for the smooth scroll would stay off for good.
    function scrollToMessage(id, { onSettled } = {}) {
        const el = document.getElementById(`message-${id}`);

        if (!el) {
            onSettled?.();

            return;
        }

        el.classList.add("bg-indigo-100");
        setTimeout(() => el.classList.remove("bg-indigo-100"), HIGHLIGHT_MS);

        setTimeout(() => {
            el.scrollIntoView({ behavior: "smooth", block: "center" });
            setTimeout(() => onSettled?.(), SETTLE_MS);
        }, SCROLL_START_MS);
    }

    return {
        container,
        movedUp,
        viewportHeight,
        distanceFromTop,
        distanceFromBottom,
        atBottom,
        record,
        scrollToNewest,
        preservePositionWhile,
        followNewest,
        scrollToMessage,
    };
}
