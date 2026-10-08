// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const shareService = {
    shareFile(data) {
        return axios.post(`files/share`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    unshareFile(fileId, userId) {
        return axios.post(
            `files/${fileId}/unshare`,
            { user: userId },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    unshareGroup(fileId, groupId) {
        return axios.post(
            `files/${fileId}/unshare-group`,
            { group: groupId },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },

    updateShare(data) {
        return axios.post(`files/share/update`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    shareLink(data) {
        return axios.post(`link/share`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    updateLinkShare(data) {
        return axios.post(`link/update`, data, {
            headers: {
                Accept: "application/json",
                "Content-Type": "application/json;charset=UTF-8",
            },
        });
    },

    deleteLink(token) {
        return axios.post(
            `link/delete`,
            { token },
            {
                headers: {
                    Accept: "application/json",
                    "Content-Type": "application/json;charset=UTF-8",
                },
            },
        );
    },
};

export default shareService;
