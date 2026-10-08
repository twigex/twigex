// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const configService = {
    config() {
        return axios.get(`/config/client`);
    },
};

export default configService;
