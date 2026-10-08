// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// /src/services/emailService.js
import axios from "./api";

const emailService = {
    emailServer() {
        return axios.get("/settings/email-server");
    },

    update(data) {
        return axios.post("/settings/email-server", data);
    },

    testConnection(data) {
        return axios.post("/settings/email-server/connection", data);
    },
};

export default emailService;
