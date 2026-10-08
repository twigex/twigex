// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { getCurrentInstance, onBeforeUnmount } from "vue";

// Lets only the latest of a kind of request count: starting one cancels the
// one before, and a request whose answer arrives after a newer start, or
// after the component has gone, is told it is no longer current.
export function useLatestRequest() {
    let latest = 0;
    let controller = null;

    function start() {
        const mine = ++latest;

        controller?.abort();
        controller = new AbortController();

        return { signal: controller.signal, isCurrent: () => mine === latest };
    }

    function cancel() {
        latest++;
        controller?.abort();
        controller = null;
    }

    if (getCurrentInstance()) onBeforeUnmount(cancel);

    return { start, cancel };
}
