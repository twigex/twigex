// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { watch, onBeforeUnmount } from "vue";

export function useConditionalClickOutside(targetRef, isOpen, callback) {
    let listener = null;

    const add = () => {
        if (!listener) {
            listener = (event) => {
                const el = targetRef.value?.$el ?? targetRef.value;

                if (el instanceof Element) {
                    if (!el.contains(event.target)) {
                        callback(event);
                    }
                }
            };

            document.addEventListener("click", listener);
        }
    };

    const remove = () => {
        if (listener) {
            document.removeEventListener("click", listener);
            listener = null;
        }
    };

    // Add/remove listener depending on open state
    watch(isOpen, (value) => {
        if (value) add();
        else remove();
    });

    onBeforeUnmount(remove);
}
