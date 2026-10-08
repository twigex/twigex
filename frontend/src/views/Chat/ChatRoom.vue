<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <SidebarLayout :title="t('main.sections.chat')">
        <template #sidebar="{ mini }">
            <ChatNavigation :mini="mini" />
        </template>

        <div v-if="loaded" class="flex flex-row w-full h-full">
            <div
                v-if="!can('view_channels')"
                class="flex flex-col items-center justify-center w-full h-full text-center px-6"
            >
                <LockClosedIcon class="h-12 w-12 text-gray-300" aria-hidden="true" />
                <h3 class="mt-4 text-base font-semibold text-gray-900">Access restricted</h3>
                <p class="mt-1 text-sm text-gray-500 max-w-sm">
                    You don't have permission to access chat. Contact your administrator to request
                    access.
                </p>
            </div>

            <template v-else>
                <div class="bg-white w-full flex min-w-0">
                    <div
                        v-if="
                            channelsStore.channels.length > 0 &&
                            route.params.chatId != '' &&
                            route.params.chatId != undefined
                        "
                        class="h-full w-full flex flex-col bg-white text-gray-800 min-h-0"
                    >
                        <router-view />
                    </div>
                    <div
                        v-else-if="channelsStore.channels.length === 0"
                        class="flex w-full h-full flex-col items-center justify-center px-6 text-center"
                    >
                        <ChatBubbleBottomCenterTextIcon
                            class="mx-auto h-16 w-16 text-indigo-400"
                            aria-hidden="true"
                        />
                        <h2 class="mt-4 text-lg font-semibold text-gray-900">
                            {{ t("channels.empty_state.title") }}
                        </h2>
                        <p class="mt-2 max-w-sm text-sm text-gray-500">
                            {{ t("channels.empty_state.description") }}
                        </p>
                        <div
                            v-if="can('create_channel')"
                            class="mt-6 flex flex-wrap items-center justify-center gap-3"
                        >
                            <button
                                type="button"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500"
                                @click="channelsStore.newChannelDialog = true"
                            >
                                <PlusIcon class="-ml-0.5 h-4 w-4" aria-hidden="true" />
                                {{ t("channels.navigation.menu.create_channel") }}
                            </button>
                            <button
                                type="button"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                                @click="channelsStore.browseChannelDialog = true"
                            >
                                <GlobeAltIcon
                                    class="-ml-0.5 h-4 w-4 text-gray-400"
                                    aria-hidden="true"
                                />
                                {{ t("channels.navigation.menu.browse_channels") }}
                            </button>
                        </div>
                    </div>
                </div>

                <ChatDetails />
                <MobileDetails />
                <AddUserDialog
                    v-model="channelsStore.addUserDialog.open"
                    @close="channelsStore.addUserDialog.open = false"
                    @add="addUserToChannel"
                    @add-groups="addGroupsToChannel"
                />
                <ChannelDialogs />
            </template>
        </div>

        <div v-else class="flex h-full w-full flex-col items-center justify-center gap-3">
            <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
            <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
        </div>
    </SidebarLayout>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import SidebarLayout from "@/components/Navigation/SidebarLayout.vue";
import ChatNavigation from "@/components/Chat/Navigation/ChatNavigation.vue";

import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import ChatDetails from "@/components/Chat/Sidebar/ChatDetails.vue";
import MobileDetails from "@/components/Chat/Sidebar/MobileDetails.vue";
import AddUserDialog from "@/components/Chat/Dialogs/AddUserDialog.vue";
import ChannelDialogs from "@/components/Chat/Dialogs/ChannelDialogs.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import {
    ChatBubbleBottomCenterTextIcon,
    LockClosedIcon,
    GlobeAltIcon,
    PlusIcon,
} from "@heroicons/vue/24/outline";
import { usePermissions } from "@/composables/usePermissions";
import chatService from "@/services/chatService";
import useChatOperations from "@/composables/chat/useChatOperations";
import { useChannelsStore } from "@/store/channels";
import { LAST_VIEWED_CHANNEL } from "@/constants/channels.js";
const { addUser, addGroups } = useChatOperations();

const route = useRoute();
const router = useRouter();
const { can } = usePermissions();
const channelsStore = useChannelsStore();
const loaded = ref(false);

function addUserToChannel(user) {
    addUser(user, channelsStore.addUserDialog.channel);

    channelsStore.closeAddUserDialog();
}

function addGroupsToChannel(groups) {
    addGroups(groups, channelsStore.addUserDialog.channel);

    channelsStore.closeAddUserDialog();
}

onMounted(() => {
    chatService.getChannels().then((response) => {
        // Use the channels this request returned rather than waiting to see
        // whether another one has landed. setChannels merges, so arriving
        // after Main's identical fetch is harmless.
        channelsStore.setChannels(response.data);

        if (route.params.chatId == undefined || route.params.chatId == "") {
            if (channelsStore.channels.length > 0) {
                const lastViewed = localStorage.getItem(LAST_VIEWED_CHANNEL);

                if (lastViewed) {
                    const lastChannel = channelsStore.channels.find((c) => c.id === lastViewed);

                    if (lastChannel) {
                        channelsStore.setCurrentChannel(lastChannel);
                        router.push({
                            name: "chat",
                            params: { chatId: lastChannel.id },
                        });
                    } else {
                        channelsStore.setCurrentChannel(channelsStore.channels[0]);
                        router.push({
                            name: "chat",
                            params: { chatId: channelsStore.channels[0].id },
                        });
                    }
                } else {
                    channelsStore.setCurrentChannel(channelsStore.channels[0]);
                    router.push({
                        name: "chat",
                        params: { chatId: channelsStore.channels[0].id },
                    });
                }
            }
        } else {
            const channel = channelsStore.channels.find((c) => c.id === route.params.chatId);

            if (channel) {
                channelsStore.setCurrentChannel(channel);
            }
        }

        loaded.value = true;
    });
});
</script>
