// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const userService = {
    me() {
        return axios.get("/users/me");
    },

    getMyPermissions() {
        return axios.get("/users/me/permissions");
    },

    createUser(data) {
        return axios.post("/users/create", data);
    },

    updateUser(data) {
        return axios.post("/users/update", data);
    },

    deactivate(id) {
        return axios.post(`/users/${id}/deactivate`);
    },

    resetPassword(id) {
        return axios.post(`/users/${id}/password/reset`);
    },

    revokeSessions(id) {
        return axios.post(`/users/${id}/sessions/revoke`);
    },

    enableMfa(data) {
        return axios.post("/users/mfa", data);
    },

    generateMfa() {
        return axios.post("/users/mfa/generate");
    },

    preferences() {
        return axios.get("/users/me/preferences");
    },

    updatePreferences(data) {
        return axios.post("/users/me/preferences", data);
    },

    userSessions() {
        return axios.get("/users/me/sessions");
    },

    logoutUserSession(sessionId) {
        return axios.post(`/users/sessions/${sessionId}/logout`);
    },

    users({ limit = 50, offset = 0, includeDeactivated = false, query = "", sort = null } = {}) {
        return axios.get("/users", {
            params: {
                limit,
                offset,
                include_deactivated: includeDeactivated,
                q: query,
                sort: sort?.key || undefined,
                dir: sort?.desc ? "desc" : undefined,
            },
        });
    },

    byIds(ids) {
        if (!ids || ids.length === 0) return Promise.resolve({ data: [] });

        return axios.get("/users/by-ids", { params: { ids: ids.join(",") } });
    },

    byUsernames(usernames) {
        if (!usernames || usernames.length === 0) return Promise.resolve({ data: [] });

        return axios.get("/users/by-usernames", {
            params: { usernames: usernames.join(",") },
        });
    },

    search(q, limit = 20) {
        return axios.get("/users/search", { params: { q, limit } });
    },

    updateProfile(data) {
        return axios.post("/users/profile/update", data);
    },

    updatePassword(data) {
        return axios.post("/users/me/password/update", data);
    },

    uploadPhoto(data) {
        return axios.post("/users/me/photo", data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "multipart/form-data",
            },
        });
    },

    deletePhoto() {
        return axios.delete("/users/me/photo");
    },

    statuses() {
        return axios.get("/users/status");
    },
};

export default userService;
