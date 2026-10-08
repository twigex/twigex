// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const guestVideoService = {
    validate(token) {
        return axios.get(`/guest/video/${token}`);
    },

    getToken(token, name, password = "") {
        return axios.post(`/guest/video/${token}/token`, { name, password });
    },
};

export default guestVideoService;
