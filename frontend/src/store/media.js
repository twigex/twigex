// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useMediaStore = defineStore("nedia", {
    state: () => {
        return { files: [], index: 0, show: false };
    },

    getters: {},

    actions: {
        openViewer(files, index) {
            this.files = files;
            this.index = index;
            this.show = true;
        },

        close() {
            this.files = [];
            this.index = 0;
            this.show = false;
        },
    },
});
