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

    getTypes() {
        return axios.get("/graph/types");
    },

    searchEntities(type, search, signal) {
        const path = `/graph/types/${encodeURIComponent(type)}/entities`;

        return axios.get(path, { params: { search: search }, signal: signal });
    },

    getRelations(type, id) {
        const path = `/graph/entities/${encodeURIComponent(type)}/relations`;

        return axios.get(path, { params: { id: id } });
    },

    getRelationPage(type, id, relationType, direction, after) {
        const path = `/graph/entities/${encodeURIComponent(type)}/relations/${encodeURIComponent(relationType)}`;

        return axios.get(path, { params: { id: id, direction: direction, after: after } });
    },
};

export default graphService;
