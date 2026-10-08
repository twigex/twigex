// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const notificationService = {
    notifications() {
        return axios.get("/notifications");
    },

    resolve(id) {
        return axios.get(`/notifications/${id}/resolve`);
    },

    read(data) {
        return axios.put("/notifications/read", data);
    },

    delete(id) {
        return axios.post(`/notifications/${id}/delete`);
    },
};

export default notificationService;
