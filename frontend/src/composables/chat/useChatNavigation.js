// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import { useUserStore } from "@/store/user";
import { useChannelsStore } from "@/store/channels";
import useChatOperations from "@/composables/chat/useChatOperations";

export function useChatNavigation() {
    const router = useRouter();
    const route = useRoute();
    const userStore = useUserStore();
    const channelsStore = useChannelsStore();
    const { getDirectChannelName } = useChatOperations();

    const highlightId = computed(() => route.query.meeting || "");

    const channelsUnreadCount = computed(() =>
        channelsStore.channels
            .filter((c) => c.type !== "D")
            .reduce(
                (sum, c) => sum + channelsStore.getChannelUnreadPosts(c.id, userStore.user.id),
                0,
            ),
    );

    const dmsUnreadCount = computed(() =>
        channelsStore.channels
            .filter((c) => c.type === "D")
            .reduce(
                (sum, c) => sum + channelsStore.getChannelUnreadPosts(c.id, userStore.user.id),
                0,
            ),
    );

    // Channels alphabetically; DMs by most recent activity.
    const regularChannels = computed(() =>
        channelsStore.channels
            .filter((c) => c.type !== "D")
            .sort((a, b) => (a.displayname || "").localeCompare(b.displayname || "")),
    );

    const directMessages = computed(() =>
        channelsStore.channels
            .filter((c) => c.type === "D")
            .sort((a, b) => (b.last_post || 0) - (a.last_post || 0)),
    );

    const unreadByChannel = computed(() => {
        const map = {};

        for (const c of channelsStore.channels) {
            map[c.id] = channelsStore.getChannelUnreadPosts(c.id, userStore.user.id);
        }

        return map;
    });

    const mentionsByChannel = computed(() => {
        const map = {};

        for (const c of channelsStore.channels) {
            map[c.id] = channelsStore.getChannelMentions(c.id, userStore.user.id);
        }

        return map;
    });

    // Highlighted when it's the deep-link target or the meeting open in the call overlay.
    function rowActive(meeting) {
        return meeting.id === highlightId.value || meeting.id === channelsStore.currentMeeting?.id;
    }

    function openMeetingPage(meeting) {
        router.push({ name: "meeting", params: { id: meeting.id } });
    }

    function openChannel(channel) {
        channelsStore.setCurrentChannel(channel);

        router.push({ name: "chat", params: { chatId: channel.id } });
    }

    function getDirectChannelUser(channel) {
        let me = userStore.user;

        for (let i = 0; i < channel.channel_members.length; i++) {
            const memberId = channel.channel_members[i].user_id;

            if (me.id != memberId) {
                // Fall back to an id-only object so UserAvatar (which only needs
                // the id) renders while the user record is still loading.
                return userStore.getUserById(memberId) || { id: memberId };
            }
        }
    }

    function getChannelName(channel) {
        return getDirectChannelName(userStore, channel);
    }

    return {
        route,
        channelsUnreadCount,
        dmsUnreadCount,
        regularChannels,
        directMessages,
        unreadByChannel,
        mentionsByChannel,
        rowActive,
        openMeetingPage,
        openChannel,
        getDirectChannelUser,
        getChannelName,
    };
}
