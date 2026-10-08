// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import { useRouter } from "vue-router";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";

export function useChannelManagement() {
    const router = useRouter();
    const channelsStore = useChannelsStore();

    const channelRenameDialog = ref(false);
    const channelToRename = ref(null);
    const leaveConfirmDialog = ref(false);
    const archiveConfirmDialog = ref(false);
    const channelToArchive = ref(null);
    const channelToLeave = ref(null);

    function openRenameDialog(channel) {
        channelToRename.value = channel;
        channelRenameDialog.value = true;
    }

    function renameChannel(name) {
        const channelId = channelToRename.value.id;

        chatService
            .updateChannel(channelId, {
                displayname: name,
            })
            .then((response) => {
                const index = channelsStore.channels.findIndex((c) => c.id === channelId);

                if (index !== -1) {
                    channelsStore.channels[index] = response.data;
                }

                if (channelsStore.currentChannel?.id === channelId) {
                    channelsStore.setCurrentChannel(response.data);
                }

                if (channelToRename.value?.id === channelId) {
                    channelToRename.value = null;
                }
            });

        channelRenameDialog.value = false;
    }

    function archiveChannel(channel) {
        channelToArchive.value = channel;
        archiveConfirmDialog.value = true;
    }

    function confirmArchive() {
        const channel = channelToArchive.value;

        channelToArchive.value = null;
        if (!channel) return;

        chatService.archiveChannel(channel.id).then(() => {
            channelsStore.channels = channelsStore.channels.filter((c) => c.id !== channel.id);
            if (channelsStore.currentChannel?.id === channel.id) {
                router.push({ name: "chat" });
            }
        });
    }

    function leaveChannel(channel) {
        channelToLeave.value = channel;
        leaveConfirmDialog.value = true;
    }

    function confirmLeave() {
        const channel = channelToLeave.value;

        if (!channel) return;

        channelToLeave.value = null;

        chatService.leaveFromChannel({ id: channel.id }).then(() => {
            channelsStore.channels = channelsStore.channels.filter((c) => c.id !== channel.id);

            if (channelsStore.channels.length === 0) {
                channelsStore.setCurrentChannel(null);

                router.push({ name: "chat" });
            } else if (channelsStore.currentChannel?.id === channel.id) {
                channelsStore.setCurrentChannel(channelsStore.channels[0]);

                router.push({
                    name: "chat",
                    params: { chatId: channelsStore.channels[0]?.id || "" },
                });
            }
        });
    }

    return {
        channelRenameDialog,
        channelToRename,
        leaveConfirmDialog,
        archiveConfirmDialog,
        channelToArchive,
        channelToLeave,
        openRenameDialog,
        renameChannel,
        archiveChannel,
        confirmArchive,
        leaveChannel,
        confirmLeave,
    };
}
