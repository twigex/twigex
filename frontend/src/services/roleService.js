// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const roleService = {
    getRoles() {
        return axios.get("/settings/roles");
    },

    getPermissions() {
        return axios.get("/settings/permissions");
    },

    createRole(data) {
        return axios.post("/settings/roles", data);
    },

    updateRole(id, data) {
        return axios.put(`/settings/roles/${id}`, data);
    },

    deleteRole(id) {
        return axios.delete(`/settings/roles/${id}`);
    },
};

export default roleService;
