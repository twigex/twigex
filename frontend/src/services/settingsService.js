// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// /src/services/settingsService.js
import axios from "./api";

const settingsService = {
    passwordPolicy() {
        return axios.get("/settings/password-policy");
    },

    updatePasswordPolicy(data) {
        return axios.post("/settings/password-policy", data);
    },

    metadata() {
        return axios.get("/settings/metadata");
    },

    createMetadata(data) {
        return axios.post("/settings/metadata", data);
    },

    deleteMetadata(id) {
        return axios.delete(`/settings/metadata/${id}`);
    },

    updateMetadata(id, data) {
        return axios.put(`/settings/metadata/${id}`, data);
    },

    chatSettings() {
        return axios.get("/settings/chat");
    },

    uploadLicense(file, onProgress) {
        const formData = new FormData();

        formData.append("file", file);

        return axios.post(`/license/add`, formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
            onUploadProgress: (progressEvent) => {
                if (onProgress && progressEvent.total) {
                    const percentCompleted = Math.round(
                        (progressEvent.loaded * 100) / progressEvent.total,
                    );

                    onProgress(percentCompleted);
                }
            },
        });
    },

    getActiveLicense() {
        return axios.get("/settings/license");
    },

    removeActiveLicense() {
        return axios.delete("/settings/license");
    },

    getLicense() {
        return axios.get("/license");
    },

    getChatSettings() {
        return axios.get("/settings/chat/configuration");
    },

    updateChatSettings(data) {
        return axios.post("/settings/chat/configuration", data);
    },

    getOfficeSettings() {
        return axios.get("/settings/office");
    },
    updateOfficeSettings(data) {
        return axios.post("/settings/office", data);
    },

    getLanguageSettings() {
        return axios.get("/settings/language");
    },
    updateLanguageSettings(data) {
        return axios.post("/settings/language", data);
    },

    getAuthSettings() {
        return axios.get("/settings/auth");
    },
    updateAuthSettings(data) {
        return axios.post("/settings/auth", data);
    },
    runLDAPSync() {
        return axios.post("/settings/auth/ldap/sync");
    },

    testLDAPConnection(data) {
        return axios.post("/settings/auth/ldap/test", data);
    },

    getOIDCProviders() {
        return axios.get("/settings/auth/oidc/providers");
    },
    prepareOIDCProvider() {
        return axios.post("/settings/auth/oidc/providers/prepare");
    },
    createOIDCProvider(data) {
        return axios.post("/settings/auth/oidc/providers", data);
    },
    updateOIDCProvider(id, data) {
        return axios.put(`/settings/auth/oidc/providers/${id}`, data);
    },
    deleteOIDCProvider(id) {
        return axios.delete(`/settings/auth/oidc/providers/${id}`);
    },
    publicAuthSettings() {
        return axios.get("/settings/oauth");
    },

    getStorages() {
        return axios.get("/settings/storage");
    },
    createStorage(data) {
        return axios.post("/settings/storage", data);
    },
    updateStorage(id, data) {
        return axios.put(`/settings/storage/${id}`, data);
    },
    deleteStorage(id) {
        return axios.delete(`/settings/storage/${id}`);
    },
    setPrimaryStorage(id) {
        return axios.put(`/settings/storage/${id}/primary`);
    },
};

export default settingsService;
