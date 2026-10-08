// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useAlertStore = defineStore("alert", {
    state: () => {
        return { show: false, title: "", message: "", type: "" };
    },

    getters: {},

    actions: {
        showSuccess(message) {
            this.show = true;
            this.title = "Success";
            this.message = message;
            this.type = "success";
        },

        showError(message) {
            this.show = true;
            this.title = "Error";
            this.message =
                typeof message === "string"
                    ? message
                    : message?.error || message?.message || "An error occurred";
            this.type = "error";
        },
    },
});
