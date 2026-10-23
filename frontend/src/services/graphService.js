// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const graphService = {
    getRebuildStatus() {
        return axios.get("/graph/rebuild");
    },

    rebuild() {
        return axios.post("/graph/rebuild");
    },
};

export default graphService;
