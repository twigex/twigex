// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, shallowRef } from "vue";
import { t } from "@/i18n/index.js";
import { useAlertStore } from "@/store/alerts";

// The date picker a task's date field opens: which task and field it is for,
// and the element it is anchored to. Toggling it on the field it is already
// open for closes it.
export function useDatePickerPopover({ canOpen, onOpen } = {}) {
    const visible = ref(false);
    const anchor = shallowRef(null);
    const item = ref(null);
    const colIndex = ref(null);

    function toggle(event, forItem, forColIndex) {
        if (canOpen && !canOpen()) {
            useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

            return;
        }

        if (visible.value && item.value === forItem && colIndex.value === forColIndex) {
            visible.value = false;

            return;
        }

        anchor.value = event.currentTarget;
        item.value = forItem;
        colIndex.value = forColIndex;
        visible.value = true;
        onOpen?.(forItem);
    }

    return { visible, anchor, item, colIndex, toggle };
}
