// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref, watch, onUnmounted } from "vue";

const SENDING_DELAY_MS = 1000;

export function useDelayedSending(sendState) {
    const showSending = ref(false);
    let sendingTimer = null;

    watch(
        sendState,
        (state) => {
            clearTimeout(sendingTimer);
            showSending.value = false;

            if (state === "sending") {
                sendingTimer = setTimeout(() => (showSending.value = true), SENDING_DELAY_MS);
            }
        },
        { immediate: true },
    );

    onUnmounted(() => clearTimeout(sendingTimer));

    return showSending;
}
