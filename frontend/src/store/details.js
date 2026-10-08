// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import fileService from "@/services/fileService";
import activityService from "@/services/activityService";

export const useDetailsStore = defineStore("details", {
    state: () => {
        return {
            details: null,
            open: true,
            mobileOpen: false,
            file: null,
            activities: [],
        };
    },

    getters: {
        getDetails() {
            return this.details;
        },

        getFile() {
            return this.file;
        },

        isOpen() {
            return this.open;
        },
    },

    actions: {
        loadDetails(file) {
            if (this.file != null && this.file.id === file.id) {
                return;
            }

            this.file = file;

            // Drop a stale response so a slower earlier fetch can't overwrite a newer selection.
            const requestedId = file.id;

            fileService.getDetails(file.id).then((res) => {
                if (this.file?.id === requestedId) {
                    this.details = res.data;
                }
            });

            activityService.fileActivity(file.id).then((res) => {
                if (this.file?.id === requestedId) {
                    this.activities = res.data;
                }
            });
        },

        setFile(file) {
            this.file = file;
        },
    },
});
