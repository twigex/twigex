// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// /src/services/userService.js
import axios from "./api";

const filesService = {
    getFile(id) {
        return axios.get(`/files/file/${id}`);
    },

    addMetadata(data) {
        return axios.post("files/metadata/entry/add", data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    updateMetadata(id, data) {
        return axios.put(`files/metadata/entry/${id}`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    removeMetadata(data) {
        return axios.post("files/metadata/entry/remove", data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    getDrive() {
        return axios.get("files/get-drive");
    },

    getFilesFromFolder(id) {
        return axios.get(`files/folder/${id}`);
    },

    recentFiles() {
        return axios.get("files/recents");
    },

    favoriteFiles() {
        return axios.get("files/favorites");
    },

    sharedFiles() {
        return axios.get("files/shared");
    },

    deletedFiles() {
        return axios.get("files/deleted");
    },

    createFile(data) {
        return axios.post("files/create-document", data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    renameFile(data) {
        return axios.post(
            `files/rename/${data.id}`,
            { name: data.name },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    trashFile(data) {
        return axios.post(`files/trash`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    deleteFile(data) {
        return axios.post(`files/delete`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    restoreFile(data) {
        return axios.post(`files/restore`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    moveFile(data) {
        return axios.post(`files/move`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    getParent(id) {
        return axios.get(`files/parent/${id}`);
    },

    getDetails(id) {
        return axios.get(`files/details/${id}`);
    },

    addToFavorites(id) {
        return axios.post(`files/favorites/${id}`);
    },

    createFolder(data) {
        return axios.post("files/create-folder", data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    startUpload(formData) {
        return axios.post("files/upload?action=start", formData, {
            headers: {
                Accept: "application/json",
                "Content-Type": "multipart/form-data",
            },
        });
    },

    cancelUpload(sessionId) {
        return axios.post(`files/upload?action=abort&session_id=${sessionId}`);
    },

    upload(formData, action, onProgress, signal) {
        return axios.post(`files/upload?action=${action}`, formData, {
            headers: {
                Accept: "application/json",
                "Content-Type": "multipart/form-data",
            },
            signal,
            onUploadProgress: (progressEvent) => {
                if (!onProgress) return;

                if (!progressEvent.total) return;

                const percentCompleted = Math.round(
                    (progressEvent.loaded * 100) / progressEvent.total,
                );

                onProgress(percentCompleted);
            },
        });
    },

    finishUpload(formData) {
        return axios.post("files/upload?action=finish", formData, {
            headers: {
                Accept: "application/json",
                "Content-Type": "multipart/form-data",
            },
        });
    },

    getRunningJobs() {
        return axios.get(`jobs/running`);
    },

    cancelJob(id) {
        return axios.post(`jobs/cancel/${id}`);
    },

    acknowledgeJob(id) {
        return axios.post(`jobs/acknowledge/${id}`);
    },
};

export default filesService;
