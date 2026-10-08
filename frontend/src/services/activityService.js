// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const activityService = {
    fileActivity(id) {
        return axios.get(`/activity/file/${id}`);
    },
};

export default activityService;
