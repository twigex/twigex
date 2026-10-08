<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <nav class="flex w-full flex-col items-center gap-y-1 pt-2" aria-label="Sidebar">
        <a
            v-for="m in meetingsStore.liveMeetings"
            :key="m.id"
            :title="m.title"
            :class="[
                rowActive(m)
                    ? 'bg-indigo-50 text-indigo-600'
                    : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                'relative flex size-10 cursor-pointer items-center justify-center rounded-md',
            ]"
            @click="openMeetingPage(m)"
        >
            <VideoCameraIcon class="size-5" aria-hidden="true" />
            <span
                class="absolute right-1.5 top-1.5 size-2 rounded-full bg-green-500"
                aria-hidden="true"
            />
            <span class="sr-only">{{ m.title }}</span>
        </a>
        <div v-if="meetingsStore.liveMeetings.length" class="my-1 h-px w-6 bg-gray-200" />

        <a
            v-for="item in regularChannels"
            :key="item.id"
            :title="item.displayname"
            :class="[
                item.id == route.params.chatId ? 'bg-gray-100' : 'hover:bg-gray-50',
                'relative flex size-10 cursor-pointer items-center justify-center rounded-md',
            ]"
            @click="openChannel(item)"
        >
            <span class="relative size-7">
                <LetterAvatar
                    :id="item.id"
                    :name="item.displayname"
                    class="size-7 rounded-md text-xs"
                />
                <span
                    v-if="item.type === 'P'"
                    class="absolute -bottom-1 -left-1 flex size-3.5 items-center justify-center rounded-full bg-white"
                    aria-hidden="true"
                >
                    <LockClosedIcon class="size-2.5 text-gray-500" />
                </span>
            </span>
            <span
                v-if="unreadByChannel[item.id] > 0"
                class="absolute bottom-1 right-1 size-2.5 rounded-full bg-red-500 ring-2 ring-white"
                aria-hidden="true"
            />
            <span class="sr-only">{{ item.displayname }}</span>
        </a>

        <template v-if="can('create_channel')">
            <button
                type="button"
                :title="t('channels.navigation.menu.create_channel')"
                class="flex size-10 items-center justify-center rounded-md text-gray-400 hover:bg-gray-50 hover:text-indigo-600"
                @click="channelsStore.newChannelDialog = true"
            >
                <PlusOutlineIcon class="size-5" aria-hidden="true" />
                <span class="sr-only">{{ t("channels.navigation.menu.create_channel") }}</span>
            </button>
            <button
                type="button"
                :title="t('channels.navigation.menu.browse_channels')"
                class="flex size-10 items-center justify-center rounded-md text-gray-400 hover:bg-gray-50 hover:text-indigo-600"
                @click="channelsStore.browseChannelDialog = true"
            >
                <GlobeAltIcon class="size-5" aria-hidden="true" />
                <span class="sr-only">{{ t("channels.navigation.menu.browse_channels") }}</span>
            </button>
        </template>

        <div v-if="directMessages.length" class="my-1 h-px w-6 bg-gray-200" />

        <a
            v-for="item in directMessages"
            :key="item.id"
            :title="getChannelName(item)"
            :class="[
                item.id == route.params.chatId ? 'bg-gray-100' : 'hover:bg-gray-50',
                'relative flex size-10 cursor-pointer items-center justify-center rounded-md',
            ]"
            @click="openChannel(item)"
        >
            <span class="size-7">
                <UserAvatar :user="getDirectChannelUser(item)" status />
            </span>
            <span
                v-if="unreadByChannel[item.id] > 0"
                class="absolute bottom-1 right-1 size-2.5 rounded-full bg-red-500 ring-2 ring-white"
                aria-hidden="true"
            />
            <span class="sr-only">{{ getChannelName(item) }}</span>
        </a>
    </nav>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import UserAvatar from "@/components/UserAvatar.vue";
import {
    LockClosedIcon,
    GlobeAltIcon,
    VideoCameraIcon,
    PlusIcon as PlusOutlineIcon,
} from "@heroicons/vue/24/outline";
import LetterAvatar from "@/components/LetterAvatar.vue";
import { useChannelsStore } from "@/store/channels";
import { useMeetingsStore } from "@/store/meetings";
import { usePermissions } from "@/composables/usePermissions";
import { useChatNavigation } from "@/composables/chat/useChatNavigation";

const channelsStore = useChannelsStore();
const meetingsStore = useMeetingsStore();
const { can } = usePermissions();
const {
    route,
    regularChannels,
    directMessages,
    unreadByChannel,
    rowActive,
    openMeetingPage,
    openChannel,
    getDirectChannelUser,
    getChannelName,
} = useChatNavigation();
</script>
