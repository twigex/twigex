// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

// incoming: a call ringing on this client (callee). outgoing: a call this client placed, which drives the ringback until answered or declined.
export const useCallsStore = defineStore("calls", {
    state: () => ({
        incoming: null,
        outgoing: null,
    }),
    actions: {
        setIncoming(call) {
            this.incoming = call;
        },
        clearIncoming() {
            this.incoming = null;
        },
        setOutgoing(call) {
            this.outgoing = call;
        },
        clearOutgoing() {
            this.outgoing = null;
        },
    },
});
