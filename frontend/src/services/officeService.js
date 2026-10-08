// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const officeService = {
    openFile(fileId, app) {
        return axios.post(`/office/${app}/file/open/${fileId}`);
    },

    supportFileType(fileType) {
        return axios.post(`/office/file/support`, { officeType: fileType });
    },
};

export default officeService;
