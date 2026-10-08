// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const groupService = {
    page(q = "", limit = 20, offset = 0, sort = null) {
        return axios.get("/groups/page", {
            params: {
                q,
                limit,
                offset,
                sort: sort?.key || undefined,
                dir: sort?.desc ? "desc" : undefined,
            },
        });
    },

    search(q = "", limit = 20) {
        return axios.get("/groups/search", {
            params: { q, limit },
        });
    },

    get(id) {
        return axios.get(`/groups/${id}`);
    },

    create(data) {
        return axios.post("/groups", data);
    },

    update(id, data) {
        return axios.put(`/groups/${id}`, data);
    },

    delete(id) {
        return axios.delete(`/groups/${id}`);
    },

    members(id, q = "", limit = 30, offset = 0, sort = null) {
        return axios.get(`/groups/${id}/members`, {
            params: {
                q,
                limit,
                offset,
                sort: sort?.key || undefined,
                dir: sort?.desc ? "desc" : undefined,
            },
        });
    },

    addMembers(id, userIds) {
        return axios.post(`/groups/${id}/members`, { user_ids: userIds });
    },

    removeMember(id, userId) {
        return axios.delete(`/groups/${id}/members/${userId}`);
    },
};

export default groupService;
