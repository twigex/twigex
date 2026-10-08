<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }" default-open>
        <DisclosureButton
            as="div"
            class="sticky top-0 z-10 flex w-full cursor-pointer items-center gap-x-1 rounded-md bg-white px-2 py-1 text-xs font-semibold leading-6 text-gray-400 hover:bg-gray-50 hover:text-gray-600"
        >
            <ChevronDownIcon
                :class="[
                    'h-3 w-3 shrink-0 transition-transform duration-200',
                    open ? '' : '-rotate-90',
                ]"
                aria-hidden="true"
            />
            <span>{{ t("channels.navigation.direct_messages") }}</span>
            <span class="flex-1" />
            <span
                v-if="!open && dmsUnreadCount > 0"
                class="mr-1 flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                aria-hidden="true"
            >
                {{ dmsUnreadCount > 99 ? "99+" : dmsUnreadCount }}
            </span>
            <button
                v-if="open"
                @click.stop="messageDialog = true"
                type="button"
                class="rounded p-0.5 text-gray-400 hover:bg-gray-200 hover:text-indigo-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                <PlusIcon class="h-4 w-4" aria-hidden="true" />
            </button>
        </DisclosureButton>
        <NewMessageDialog v-model="messageDialog" @create="createDM" />
        <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
        >
            <DisclosurePanel>
                <ul role="list" class="mt-2 space-y-1">
                    <li v-if="!directMessages.length" class="px-2 py-1.5 text-xs text-gray-400">
                        {{ t("channels.navigation.no_direct_messages") }}
                    </li>
                    <li
                        v-for="item in directMessages"
                        :key="item.id"
                        role="button"
                        tabindex="0"
                        class="cursor-pointer rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-indigo-500"
                        @click="openChannel(item)"
                        @keydown.enter="openChannel(item)"
                        @keydown.space.prevent="openChannel(item)"
                    >
                        <a
                            :href="item.href"
                            :class="[
                                item.id == route.params.chatId
                                    ? 'bg-gray-50 text-indigo-600'
                                    : 'text-gray-500 hover:text-indigo-600 hover:bg-gray-50',
                                'group flex gap-x-3 rounded-md px-2 py-1.5 text-sm leading-6 font-medium',
                            ]"
                        >
                            <div class="flex flex-row justify-between items-center w-full truncate">
                                <div class="flex flex-row gap-x-3 truncate py-0.5">
                                    <div class="min-h-6 min-w-6 h-6 w-6">
                                        <UserAvatar :user="getDirectChannelUser(item)" status />
                                    </div>

                                    <span
                                        :class="[
                                            'truncate',
                                            unreadByChannel[item.id] > 0
                                                ? 'font-semibold text-gray-900'
                                                : '',
                                        ]"
                                        >{{ getChannelName(item) }}</span
                                    >
                                </div>

                                <span
                                    v-if="unreadByChannel[item.id] > 0"
                                    class="flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                                    aria-hidden="true"
                                >
                                    {{
                                        unreadByChannel[item.id] > 99
                                            ? "99+"
                                            : unreadByChannel[item.id]
                                    }}
                                </span>
                            </div>
                        </a>
                    </li>
                </ul>
            </DisclosurePanel>
        </transition>
    </Disclosure>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref } from "vue";
import NewMessageDialog from "@/components/Chat/Dialogs/NewMessageDialog.vue";
import { Disclosure, DisclosureButton, DisclosurePanel } from "@headlessui/vue";
import UserAvatar from "@/components/UserAvatar.vue";
import { PlusIcon } from "@heroicons/vue/24/solid";
import { ChevronDownIcon } from "@heroicons/vue/20/solid";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import { useChatNavigation } from "@/composables/chat/useChatNavigation";

const channelsStore = useChannelsStore();
const {
    route,
    dmsUnreadCount,
    directMessages,
    unreadByChannel,
    openChannel,
    getDirectChannelUser,
    getChannelName,
} = useChatNavigation();

const messageDialog = ref(false);

function createDM(data) {
    chatService.createChannel(data).then((response) => {
        const channel = response.data;

        channelsStore.setChannels([channel]);
        openChannel(channel);
    });

    messageDialog.value = false;
}
</script>
