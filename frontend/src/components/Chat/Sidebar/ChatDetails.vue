<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="channelStore.currentChannel != null"
        class="relative hidden md:block h-full flex-shrink-0 overflow-hidden border-l border-gray-200 bg-white transition-[width] duration-200 ease-in-out"
        :style="{ width: channelStore.details ? '24rem' : '0px' }"
    >
        <!-- Fixed width so content never squishes during the animation; pinned right so collapsing extends it inward, not past the viewport edge -->
        <div class="absolute inset-y-0 right-0 w-96">
            <ChatDetailsContent />
        </div>
    </div>
</template>

<script setup>
import ChatDetailsContent from "@/components/Chat/Sidebar/ChatDetailsContent.vue";
import { useChannelsStore } from "@/store/channels";

const channelStore = useChannelsStore();
</script>
