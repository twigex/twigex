<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <NewChannelDialog v-model="channelsStore.newChannelDialog" @create="createChannel" />
    <BrowseChannelDialog v-model="channelsStore.browseChannelDialog" />
</template>

<script setup>
import { useRouter } from "vue-router";
import { useChannelsStore } from "@/store/channels";
import chatService from "@/services/chatService";
import NewChannelDialog from "@/components/Chat/Dialogs/NewChannelDialog.vue";
import BrowseChannelDialog from "@/components/Chat/Dialogs/BrowseChannelDialog.vue";

const router = useRouter();
const channelsStore = useChannelsStore();

function createChannel(data) {
    chatService.createChannel(data).then((response) => {
        const channel = response.data;

        channelsStore.setChannels([channel]);
        channelsStore.setCurrentChannel(channel);

        router.push({ name: "chat", params: { chatId: channel.id } });
    });

    channelsStore.newChannelDialog = false;
}
</script>
