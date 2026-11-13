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

    getDatasetChoices() {
        return axios.get("/graph/datasets");
    },

    createEntityType(input) {
        return axios.post("/graph/entity-types", input);
    },

    updateEntityType(id, input) {
        return axios.put(`/graph/entity-types/${encodeURIComponent(id)}`, input);
    },

    deleteEntityType(id) {
        return axios.delete(`/graph/entity-types/${encodeURIComponent(id)}`);
    },

    createRelationType(input) {
        return axios.post("/graph/relation-types", input);
    },

    updateRelationType(id, input) {
        return axios.put(`/graph/relation-types/${encodeURIComponent(id)}`, input);
    },

    deleteRelationType(id) {
        return axios.delete(`/graph/relation-types/${encodeURIComponent(id)}`);
    },

    searchEntities(type, search, signal, after) {
        const path = `/graph/types/${encodeURIComponent(type)}/entities`;

        return axios.get(path, { params: { search: search, after: after }, signal: signal });
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
