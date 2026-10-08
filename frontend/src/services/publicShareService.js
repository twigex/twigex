// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

function uploadPath(token, action, childId) {
    const params = new URLSearchParams();

    if (action) params.set("action", action);
    if (childId) params.set("folder", childId);
    const query = params.toString();

    return `/public/${token}/upload${query ? `?${query}` : ""}`;
}

const publicShareService = {
    getShare(token, childId = "") {
        return childId
            ? axios.get(`/public/${token}/folder/${childId}`)
            : axios.get(`/public/${token}`);
    },

    thumbnailUrl(token, childId) {
        return `/api/public/${token}/thumbnail/${childId}`;
    },

    viewUrl(token, childId = "") {
        return childId ? `/api/public/${token}/view/${childId}` : `/api/public/${token}/view`;
    },

    authenticate(token, password) {
        return axios.post(`/public/${token}/auth`, { password });
    },

    getOffice(token, childId = "") {
        return childId
            ? axios.get(`/public/${token}/office/${childId}`)
            : axios.get(`/public/${token}/office`);
    },

    upload(token, file, childId = "") {
        const form = new FormData();

        form.append("file", file);
        form.append("name", file.name);
        form.append("type", file.type);

        return axios.post(uploadPath(token, "", childId), form);
    },

    startUpload(token, file, childId = "") {
        const form = new FormData();

        form.append("name", file.name);
        form.append("type", file.type);
        form.append("full_size", String(file.size));

        return axios.post(uploadPath(token, "start", childId), form);
    },

    uploadChunk(token, sessionId, chunk, partNumber, childId = "") {
        const form = new FormData();

        form.append("upload_session_id", sessionId);
        form.append("file", chunk);
        form.append("part_number", String(partNumber));

        return axios.post(uploadPath(token, "chunk", childId), form);
    },

    finishUpload(token, sessionId, childId = "") {
        const form = new FormData();

        form.append("upload_session_id", sessionId);

        return axios.post(uploadPath(token, "finish", childId), form);
    },

    downloadUrl(token, childId = "") {
        return childId
            ? `/api/public/${token}/download/${childId}`
            : `/api/public/${token}/download`;
    },
};

export default publicShareService;
