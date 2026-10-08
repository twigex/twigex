// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, watch, nextTick, onBeforeUnmount } from "vue";

export function useCallSidebarScroll(sharedView, participants) {
    const sidebarScroll = ref(null);
    const canScrollUp = ref(false);
    const canScrollDown = ref(false);
    let ro;

    function updateScrollButtons() {
        const el = sidebarScroll.value;

        if (!el) return;

        const eps = 1;

        canScrollUp.value = el.scrollTop > eps;
        canScrollDown.value = el.scrollTop + el.clientHeight < el.scrollHeight - eps;
    }

    function scrollSidebar(direction) {
        const el = sidebarScroll.value;

        if (!el) return;

        const amount = 2 * 128;

        el.scrollBy({ top: direction * amount, behavior: "smooth" });

        requestAnimationFrame(updateScrollButtons);
    }

    function attachSidebarObservers() {
        const el = sidebarScroll.value;

        if (!el) return;

        updateScrollButtons();

        el.addEventListener("scroll", updateScrollButtons, { passive: true });

        ro = new ResizeObserver(updateScrollButtons);
        ro.observe(el);
    }

    function detachSidebarObservers() {
        const el = sidebarScroll.value;

        if (el) el.removeEventListener("scroll", updateScrollButtons);
        if (ro) {
            ro.disconnect();
            ro = null;
        }
    }

    onBeforeUnmount(() => {
        const el = sidebarScroll.value;

        if (el) el.removeEventListener("scroll", updateScrollButtons);
        if (ro) ro.disconnect();
    });

    // If participants changes height, re-check
    watch(
        () => participants.value.length,
        async () => {
            await nextTick();
            updateScrollButtons();
        },
    );

    watch(
        sharedView,
        async (isOn) => {
            await nextTick();

            detachSidebarObservers();

            if (!isOn) {
                canScrollUp.value = false;
                canScrollDown.value = false;

                return;
            }

            attachSidebarObservers();
        },
        { immediate: true },
    );

    return {
        sidebarScroll,
        canScrollUp,
        canScrollDown,
        scrollSidebar,
    };
}
