<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="w-full h-full flex">
        <!-- Sidebar component, swap this element with another sidebar if you like -->
        <div v-if="!mini" class="flex grow flex-col gap-y-5 overflow-y-auto bg-white pb-4">
            <nav class="flex flex-1 flex-col pt-2" aria-label="Sidebar">
                <ul role="list" class="flex flex-1 flex-col gap-y-7">
                    <li v-if="meetingsStore.myMeetings.length">
                        <ChatNavigationMeetings />
                    </li>
                    <li>
                        <ChatNavigationChannels
                            @rename="openRenameDialog"
                            @archive="archiveChannel"
                            @leave="leaveChannel"
                        />
                    </li>
                    <li>
                        <ChatNavigationDirectMessages />
                    </li>
                </ul>
            </nav>
        </div>

        <ChatNavigationMini v-else />

        <RenameChannelDialog
            v-model="channelRenameDialog"
            :channelName="channelToRename?.displayname"
            @rename="renameChannel"
        />
        <ArchiveChannelDialog
            v-model="archiveConfirmDialog"
            :channelName="channelToArchive?.displayname"
            @confirm="confirmArchive"
        />
        <LeaveChannelDialog
            v-model="leaveConfirmDialog"
            :channelName="channelToLeave?.displayname"
            @confirm="confirmLeave"
        />
    </div>
</template>

<script setup>
import { onMounted } from "vue";
import LeaveChannelDialog from "@/components/Chat/Dialogs/LeaveChannelDialog.vue";
import ArchiveChannelDialog from "@/components/Chat/Dialogs/ArchiveChannelDialog.vue";
import RenameChannelDialog from "@/components/Chat/Dialogs/RenameChannelDialog.vue";
import ChatNavigationMeetings from "@/components/Chat/Navigation/ChatNavigationMeetings.vue";
import ChatNavigationChannels from "@/components/Chat/Navigation/ChatNavigationChannels.vue";
import ChatNavigationDirectMessages from "@/components/Chat/Navigation/ChatNavigationDirectMessages.vue";
import ChatNavigationMini from "@/components/Chat/Navigation/ChatNavigationMini.vue";
import { useUsers } from "@/composables/useUser";
import { useChannelsStore } from "@/store/channels";
import { useMeetingsStore } from "@/store/meetings";
import { useChannelManagement } from "@/composables/chat/useChannelManagement";

defineProps({
    mini: {
        type: Boolean,
        default: false,
    },
});

const channelsStore = useChannelsStore();

// Trigger a lazy load so direct-message counterpart names/avatars resolve.
useUsers(() =>
    channelsStore.channels
        .filter((c) => c.type === "D")
        .flatMap((c) => (c.channel_members ?? []).map((m) => m.user_id)),
);
const meetingsStore = useMeetingsStore();
const {
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
} = useChannelManagement();

onMounted(() => meetingsStore.loadMyMeetings());
</script>
