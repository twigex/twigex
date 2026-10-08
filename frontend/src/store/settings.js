// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useSettingsStore = defineStore("settings", {
    state: () => {
        return {
            metadata: [],
            collimatoStatus: "false",
            chatSettings: {},
            config: null,
            license: null,
        };
    },

    getters: {
        getVersion(state) {
            return state.config.Version;
        },
        getBuildDate(state) {
            return state.config.BuildDate;
        },
        getBuildHash(state) {
            return state.config.BuildHash?.slice(0, 7);
        },
        isEnterpriseBuild(state) {
            return state.config.Edition === "enterprise";
        },
        getLicenseFeature: (state) => (feature) => {
            if (state.license && state.license[feature] !== undefined) {
                return state.license[feature];
            }

            return false;
        },
    },

    actions: {
        setMetadata(metadata) {
            this.metadata = metadata;
        },

        setCollimatoStatus(status) {
            this.collimatoStatus = status;
        },

        setChatSettings(settings) {
            this.chatSettings = settings;
        },

        setConfig(config) {
            this.config = config;
        },
        setLicense(license) {
            this.license = license;
        },
    },
});
