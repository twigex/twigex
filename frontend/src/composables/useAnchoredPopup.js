// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, isRef, ref, shallowRef, watch } from "vue";
import { autoUpdate, flip, offset, shift, size, useFloating } from "@floating-ui/vue";

// An anchor at a point on screen, for a menu opened by a right click.
export function pointAnchor(x, y) {
    return {
        getBoundingClientRect: () => ({
            x,
            y,
            left: x,
            top: y,
            right: x,
            bottom: y,
            width: 0,
            height: 0,
        }),
    };
}

// What a popup opened by event is anchored to: the pointer for a right
// click, otherwise what was clicked. Kept, since the event's currentTarget
// is gone once the handler returns.
export function anchorFromEvent(event) {
    if (event?.type === "contextmenu") return pointAnchor(event.clientX, event.clientY);

    return event?.currentTarget || event?.target || null;
}

// Places a popup next to its anchor, on screen whatever the popup sits in:
// render it teleported to body with floatingStyles. fitHeight caps its
// height to the room there is, for a popup that scrolls inside; matchWidth
// makes it as wide as its anchor, for a dropdown. anchor may be a ref the
// caller keeps, such as a prop.
export function useAnchoredPopup({
    placement = "bottom-start",
    fitHeight = false,
    matchWidth = false,
    floating = ref(null),
    anchor,
} = {}) {
    const reference = shallowRef(null);

    if (isRef(anchor)) {
        watch(anchor, (value) => (reference.value = value || null), { immediate: true });
    }

    const middleware = [offset(4), flip({ padding: 8 }), shift({ padding: 8, crossAxis: true })];

    if (fitHeight || matchWidth) {
        middleware.push(
            size({
                padding: 8,
                apply({ availableHeight, rects, elements }) {
                    if (fitHeight)
                        elements.floating.style.maxHeight = `${Math.max(160, availableHeight)}px`;
                    if (matchWidth) elements.floating.style.width = `${rects.reference.width}px`;
                },
            }),
        );
    }

    // Placed with top and left rather than a transform, so an open animation
    // that scales the popup does not move it, and kept hidden until it has
    // been placed once, so it never shows at the corner of the screen.
    const {
        floatingStyles: placed,
        isPositioned,
        update,
    } = useFloating(reference, floating, {
        strategy: "fixed",
        placement,
        middleware,
        transform: false,
        open: computed(() => floating.value != null),
        whileElementsMounted: autoUpdate,
    });
    const floatingStyles = computed(() => ({
        ...placed.value,
        visibility: isPositioned.value ? "visible" : "hidden",
    }));

    function anchorTo(target) {
        reference.value = target || null;
    }

    return { floating, floatingStyles, anchorTo, update };
}
