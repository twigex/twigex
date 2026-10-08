// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";
import chatService from "@/services/chatService";

// Personal, notification-independent list of the user's meetings, so it survives a deleted notification.
export const useMeetingsStore = defineStore("meetings", {
    state: () => ({
        myMeetings: [],
    }),
    getters: {
        liveMeetings: (state) => state.myMeetings.filter((m) => m.status === "active"),
        scheduledMeetings: (state) => state.myMeetings.filter((m) => m.status === "scheduled"),
    },
    actions: {
        async loadMyMeetings() {
            try {
                const { data } = await chatService.getMyMeetings();

                this.myMeetings = data || [];
            } catch {
                this.myMeetings = [];
            }
        },
        updateParticipants(meetingId, count) {
            const meeting = this.myMeetings.find((m) => m.id === meetingId);

            if (meeting) meeting.participants = count;
        },
    },
});
