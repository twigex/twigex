// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useNavigationStore = defineStore("navigation", {
    state: () => {
        return { open: true, mobileOpen: false, sidebars: 0 };
    },

    getters: {
        hasSidebar: (state) => state.sidebars > 0,
    },

    actions: {},
});
