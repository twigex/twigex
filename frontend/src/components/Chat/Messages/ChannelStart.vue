<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col items-center justify-center py-12 px-6 text-center">
        <div class="rounded-full bg-indigo-100 p-4 mb-4">
            <ChatBubbleBottomCenterTextIcon class="size-10 text-indigo-600" aria-hidden="true" />
        </div>
        <h2 class="text-lg font-semibold text-gray-900 mb-1">
            {{
                t("channels.channel_start.title", {
                    channel: channel.displayname,
                })
            }}
        </h2>
        <p class="text-sm text-gray-500 max-w-sm mb-6">
            {{
                t("channels.channel_start.description", {
                    channel: channel.displayname,
                })
            }}
        </p>
        <button
            @click="channelsStore.openAddUserDialog(channel)"
            type="button"
            class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
        >
            <UsersIcon class="-ml-0.5 size-5" aria-hidden="true" />
            {{ t("channels.channel_start.add_members") }}
        </button>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useChannelsStore } from "@/store/channels";
import { ChatBubbleBottomCenterTextIcon, UsersIcon } from "@heroicons/vue/24/outline";

defineProps({
    channel: {
        type: Object,
        required: true,
    },
});

const channelsStore = useChannelsStore();
</script>
