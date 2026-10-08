// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, shallowRef } from "vue";
import { anchorFromEvent } from "@/composables/useAnchoredPopup";

// The menu a row, card or field opens at the click: what it is for, which
// kind of thing that is, and where it is anchored. Toggling it on the thing
// it is already open for closes it.
export function useItemMenu({ canOpen, onOpen } = {}) {
    const visible = ref(false);
    const anchor = shallowRef(null);
    const item = ref(null);
    const context = ref("item");

    function toggle(event, forItem, forContext) {
        if (canOpen && !canOpen()) return;

        context.value = forContext;

        if (visible.value && item.value === forItem) {
            visible.value = false;
            item.value = null;
        } else {
            visible.value = true;
            item.value = forItem;
            onOpen?.(forItem);
            anchor.value = anchorFromEvent(event);
        }
    }

    return { visible, anchor, item, context, toggle };
}
