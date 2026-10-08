// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useRoleStore = defineStore("roles", {
    state: () => ({
        roles: [],
        permissions: [],
    }),

    getters: {
        getRoleById: (state) => (id) => state.roles.find((r) => r.id === id),
    },

    actions: {
        setRoles(roles) {
            this.roles = roles;
        },
        setPermissions(permissions) {
            this.permissions = permissions;
        },
        addRole(role) {
            this.roles.push(role);
        },
        updateRole(updated) {
            const idx = this.roles.findIndex((r) => r.id === updated.id);

            if (idx !== -1) this.roles[idx] = updated;
        },
        removeRole(id) {
            this.roles = this.roles.filter((r) => r.id !== id);
        },
    },
});
