// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { defineStore } from "pinia";

export const useChannelsStore = defineStore("channels", {
    state: () => ({
        currentChannel: null,
        channels: [],
        posts: [],
        unreadMessageId: null,
        reply_posts: [],
        details: false,
        mobileDetails: false,
        videoIsOpen: false,
        currentMeeting: null,
        typing: [], // [{ channel: id, user: id }]
        typingTimers: {}, // Store timeout references for each user
        isMinimized: false,

        addUserDialog: {
            open: false,
            channel: null,
        },
        newChannelDialog: false,
        browseChannelDialog: false,
    }),

    getters: {
        getUsersTypingInChannel: (state) => (id) => {
            return state.typing.filter((user) => user.channel === id);
        },

        getTotalUnreadPosts: (state) => (userId) => {
            let totalUnread = 0;

            state.channels.forEach((channel) => {
                channel.channel_members.forEach((member) => {
                    if (member.user_id === userId) {
                        totalUnread += channel.msg_count - member.msg_count;
                    }
                });
            });

            return totalUnread;
        },

        getChannelUnreadPosts: (state) => (channelId, userId) => {
            const channel = state.channels.find((c) => c.id === channelId);

            if (!channel) return 0;
            const member = channel.channel_members.find((m) => m.user_id === userId);

            if (!member) return 0;

            return channel.msg_count - member.msg_count;
        },

        getChannelMentions: (state) => (channelId, userId) => {
            const channel = state.channels.find((c) => c.id === channelId);

            if (!channel) return 0;
            const member = channel.channel_members.find((m) => m.user_id === userId);

            return member?.mention_count ?? 0;
        },

        getUnreadChannelsCount: (state) => (userId) => {
            return state.channels.filter((channel) => {
                const member = channel.channel_members.find((m) => m.user_id === userId);

                return member && channel.msg_count > member.msg_count;
            }).length;
        },

        getChannelById: (state) => (id) => {
            return state.channels.find((channel) => channel.id === id);
        },
    },

    actions: {
        setCurrentChannel(channel) {
            this.currentChannel = channel;
        },

        setPosts(posts) {
            this.posts = posts;
        },

        setChannels(channels) {
            channels.forEach((newChannel) => {
                const index = this.channels.findIndex((c) => c.id === newChannel.id);

                if (index !== -1) {
                    // Update existing channel
                    this.channels[index] = {
                        ...this.channels[index],
                        ...newChannel,
                        posts: this.channels[index].posts ?? [],
                    };
                } else {
                    this.channels.push({
                        ...newChannel,
                        posts: newChannel.posts ?? [],
                    });
                }
            });
        },

        readChannelPosts(channelId, userId) {
            const channel = this.channels.find((c) => c.id === channelId);

            if (!channel) return;
            const member = channel.channel_members.find((m) => m.user_id === userId);

            if (!member) return;
            member.msg_count = channel.msg_count;
            member.mention_count = 0;

            member.last_viewed_at = Math.floor(Date.now());
        },

        incrementMention(channelId, userId) {
            const channel = this.channels.find((c) => c.id === channelId);

            if (!channel) return;
            const member = channel.channel_members.find((m) => m.user_id === userId);

            if (!member) return;
            member.mention_count = (member.mention_count ?? 0) + 1;
        },

        openAddUserDialog(channel) {
            this.addUserDialog.open = true;
            this.addUserDialog.channel = channel;
        },

        closeAddUserDialog() {
            this.addUserDialog.open = false;
            this.addUserDialog.channel = null;
        },

        setLastPost(channelId) {
            const channel = this.channels.find((c) => c.id === channelId);

            if (channel) {
                channel.last_post = Date.now();
            }
        },

        /**
         * Add a user to the typing list and reset the timeout if they continue typing.
         */
        addTypingUser(channel, user) {
            const key = `${channel}-${user}`;

            // Add user to typing list if not already present
            if (!this.typing.some((t) => t.channel === channel && t.user === user)) {
                this.typing.push({ channel, user });
            }

            // Clear previous timeout if exists
            clearTimeout(this.typingTimers[key]);

            // Set a new timeout (3.5 seconds)
            this.typingTimers[key] = setTimeout(() => {
                this.removeTypingUser(channel, user);
            }, 5000);
        },

        /**
         * Remove a user from the typing list
         */
        removeTypingUser(channel, user) {
            this.typing = this.typing.filter((t) => !(t.channel === channel && t.user === user));
        },

        /**
         * Open the video view for a specific meeting
         */
        openVideoChat(meeting) {
            this.currentMeeting = meeting;
            this.videoIsOpen = true;
        },

        /**
         * Close video view
         */
        closeVideoChat() {
            this.videoIsOpen = false;
            this.currentMeeting = null;
        },
    },
});
