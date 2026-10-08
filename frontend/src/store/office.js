// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useOfficeStore = defineStore("office", {
    state: () => ({
        open: false,
        type: "",
        url: "",
        host: "",
        config: null,
    }),

    getters: {},

    actions: {
        openOffice(result) {
            this.type = result.type;
            this.url = result.url || "";
            this.host = result.host || "";
            this.config = result.config || null;
            this.open = true;
        },

        closeOffice() {
            this.open = false;
            this.type = "";
            this.url = "";
            this.host = "";
            this.config = null;
        },
    },
});
