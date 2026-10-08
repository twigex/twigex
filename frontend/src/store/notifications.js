// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useNotificationsStore = defineStore("notifications", {
    state: () => {
        return { notifications: [] };
    },

    getters: {},

    actions: {
        setNotifications(notifications) {
            this.notifications = notifications;
        },
        addNotification(notification) {
            this.notifications.unshift(notification);
        },
    },
});
