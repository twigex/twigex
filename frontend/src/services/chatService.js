// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import axios from "./api";

const chatService = {
    getChannels(urlParam = "") {
        let url = "/channels";

        if (urlParam) {
            url += `?${urlParam}`;
        }

        return axios.get(url);
    },

    getAllChannels() {
        return axios.get("/channels/all");
    },

    createChannel(data) {
        return axios.post("/channels", data);
    },

    updateChannel(id, data) {
        return axios.post(`/channels/${id}/patch`, data);
    },

    addUsersToChannel(data) {
        return axios.post(`/channels/${data.id}/users`, {
            users: data.users,
        });
    },

    addGroupsToChannel(channelId, groupIds) {
        return axios.post(`/channels/${channelId}/groups`, {
            group_ids: groupIds,
        });
    },

    getChannelGroups(channelId) {
        return axios.get(`/channels/${channelId}/groups`);
    },

    removeGroupFromChannel(channelId, groupId) {
        return axios.delete(`/channels/${channelId}/groups/${groupId}`);
    },

    archiveChannel(id) {
        return axios.post(`/channels/${id}/archive`);
    },

    unarchiveChannel(id) {
        return axios.post(`/channels/${id}/unarchive`);
    },

    deleteChannel(id) {
        return axios.delete(`/channels/${id}`);
    },

    joinChannel(id) {
        return axios.get(`/channels/${id}/join`);
    },

    removeUserFromChannel(data) {
        return axios.delete(`/channels/${data.id}/users/${data.user}`);
    },

    leaveFromChannel(data) {
        return axios.delete(`/channels/${data.id}/users/leave`);
    },

    changeUserRole(data) {
        return axios.post(`/channels/${data.id}/users/${data.user}/roles`, {
            role: data.role,
        });
    },

    getChannelPosts(id, postId, param) {
        let url = `/channels/${id}/posts`;

        if (postId !== "" && param !== "") {
            url += `?param=${param}&id=${postId}`;
        }

        return axios.get(url);
    },

    createChannelPost(id, message, attachments, reply, gif, pendingPostId = "") {
        return axios.post(
            `/channels/${id}/posts`,
            {
                message: message,
                reply: reply,
                attachments: attachments,
                gif: gif,
                pending_post_id: pendingPostId,
            },
            {
                headers: {
                    "Content-Type": "application/json",
                },
            },
        );
    },

    updateChannelPost(channelId, postId, message, remove, attachments) {
        return axios.post(`/channels/${channelId}/posts/${postId}`, {
            message: message,
            remove: remove,
            attachments: attachments,
        });
    },

    forwardChannelPost(channelId, postId, channels, users) {
        return axios.post(`/channels/${channelId}/posts/${postId}/forward`, {
            channels: channels,
            users: users,
        });
    },

    deleteChannelPost(channelId, postId) {
        return axios.delete(`/channels/${channelId}/posts/${postId}`);
    },

    createPostReaction(channelId, postId, reaction) {
        return axios.post(`/channels/${channelId}/posts/${postId}/reactions`, {
            reaction: reaction,
        });
    },

    deletePostReaction(channelId, postId, reaction) {
        return axios.delete(`/channels/${channelId}/posts/${postId}/reactions/${reaction}`);
    },

    getChannelById(id) {
        return axios.get(`/channels/${id}`);
    },

    readChannelPosts(data) {
        return axios.post(`/channels/${data.id}/posts/read`);
    },

    getChannelMembers(channelId) {
        return axios.get(`/channels/${channelId}/members`);
    },

    searchChannelMembers(channelId, q, limit = 25) {
        return axios.get(`/channels/${channelId}/members/search`, {
            params: { q, limit },
        });
    },

    getMeetings(channelId) {
        return axios.get(`/channels/${channelId}/meetings`);
    },

    getMyMeetings() {
        return axios.get(`/users/me/meetings`);
    },

    getMeeting(meetingId) {
        return axios.get(`/users/me/meetings/${meetingId}`);
    },

    getScheduledMeetings(channelId) {
        return axios.get(`/channels/${channelId}/meetings/scheduled`);
    },

    createMeeting(
        channelId,
        {
            scheduledAt = 0,
            title = "",
            durationMinutes = 0,
            invitees = [],
            groups = [],
            guestPassword = "",
            timezone = Intl.DateTimeFormat().resolvedOptions().timeZone,
        } = {},
    ) {
        return axios.post(`/channels/${channelId}/meetings`, {
            scheduled_at: scheduledAt,
            title,
            duration_minutes: durationMinutes,
            invitees,
            groups,
            guest_password: guestPassword,
            timezone,
        });
    },

    updateMeeting(
        channelId,
        meetingId,
        {
            scheduledAt = 0,
            title = "",
            durationMinutes = 0,
            invitees = [],
            groups = [],
            guestPassword = "",
            timezone = Intl.DateTimeFormat().resolvedOptions().timeZone,
        } = {},
    ) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}`, {
            scheduled_at: scheduledAt,
            title,
            duration_minutes: durationMinutes,
            invitees,
            groups,
            guest_password: guestPassword,
            timezone,
        });
    },

    getMeetingGroups(channelId, meetingId) {
        return axios.get(`/channels/${channelId}/meetings/${meetingId}/groups`);
    },

    addMeetingGroups(channelId, meetingId, groupIds) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/groups`, {
            group_ids: groupIds,
        });
    },

    startMeeting(channelId, meetingId) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/start`);
    },

    endMeeting(channelId, meetingId) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/end`);
    },

    cancelMeeting(channelId, meetingId) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/cancel`);
    },

    startCall(channelId) {
        return axios.post(`/channels/${channelId}/call`);
    },

    answerCall(channelId) {
        return axios.post(`/channels/${channelId}/call/answer`);
    },

    declineCall(channelId) {
        return axios.post(`/channels/${channelId}/call/decline`);
    },

    cancelCall(channelId) {
        return axios.post(`/channels/${channelId}/call/cancel`);
    },

    endCall(channelId) {
        return axios.post(`/channels/${channelId}/call/end`);
    },

    setMeetingHost(channelId, meetingId, userId) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/host`, {
            user_id: userId,
        });
    },

    getMeetingVideoToken(channelId, meetingId) {
        return axios.get(`/channels/${channelId}/meetings/${meetingId}/video/token`);
    },

    announceMeetingPresence(channelId, meetingId) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/presence`);
    },

    inviteGuests(channelId, meetingId, emails, expiresInHours) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/invite`, {
            email: emails,
            expires_in_hours: expiresInHours,
        });
    },

    inviteMembers(channelId, meetingId, invitees) {
        return axios.post(`/channels/${channelId}/meetings/${meetingId}/members`, { invitees });
    },

    getGuestInvites(channelId, meetingId) {
        return axios.get(`/channels/${channelId}/meetings/${meetingId}/invites`);
    },

    revokeGuestInvite(channelId, meetingId, linkId) {
        return axios.delete(`/channels/${channelId}/meetings/${meetingId}/invite/${linkId}`);
    },

    searchKlipyGif(searchTerm, limit, next) {
        return axios.get(
            `/channels/klipy/search?search=${encodeURIComponent(searchTerm)}` +
                `&limit=${limit}` +
                (next ? `&next=${encodeURIComponent(next)}` : ""),
        );
    },

    getKlipyTrending(limit, next) {
        return axios.get(
            `/channels/klipy/trending` +
                `?limit=${limit}` +
                (next ? `&next=${encodeURIComponent(next)}` : ""),
        );
    },

    getKlipyCategories() {
        return axios.get(`/channels/klipy/categories`);
    },

    uploadFile(file, channelId, onProgress) {
        const formData = new FormData();

        formData.append("file", file);

        return axios.post(`/channels/${channelId}/upload`, formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
            onUploadProgress: (progressEvent) => {
                if (onProgress && progressEvent.total) {
                    const percentCompleted = Math.round(
                        (progressEvent.loaded * 100) / progressEvent.total,
                    );

                    onProgress(percentCompleted);
                }
            },
        });
    },
};

export default chatService;
