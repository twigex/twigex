// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const usePermissionsStore = defineStore("permissions", {
    state: () => ({
        permissions: [],
        loaded: false,
    }),

    actions: {
        setPermissions(permissions) {
            this.permissions = permissions;
            this.loaded = true;
        },
    },
});
