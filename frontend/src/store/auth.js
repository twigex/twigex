// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useAuthStore = defineStore("auth", {
    state: () => {
        return { authenticated: false, csrf: "" };
    },

    actions: {
        setAuthenticated(authenticated) {
            this.authenticated = authenticated;
        },

        setCSRF(csrf) {
            this.csrf = csrf;
        },
    },
});
