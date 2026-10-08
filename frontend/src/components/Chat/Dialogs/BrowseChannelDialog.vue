<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="$emit('update:modelValue', false)">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0"
                enter-to="opacity-100"
                leave="ease-in duration-200"
                leave-from="opacity-100"
                leave-to="opacity-0"
            >
                <div class="fixed inset-0 bg-gray-500/75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                <div
                    class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                >
                    <TransitionChild
                        as="template"
                        enter="ease-out duration-300"
                        enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                        enter-to="opacity-100 translate-y-0 sm:scale-100"
                        leave="ease-in duration-200"
                        leave-from="opacity-100 translate-y-0 sm:scale-100"
                        leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    >
                        <DialogPanel
                            class="relative transform overflow-hidden rounded-xl bg-white text-left shadow-xl transition-all w-full sm:my-8 sm:w-full sm:max-w-2xl"
                        >
                            <div class="px-6 pt-6 pb-4 border-b border-gray-200">
                                <div class="flex items-center justify-between">
                                    <DialogTitle
                                        as="h3"
                                        class="text-lg font-semibold text-gray-900"
                                        >{{ t("channels.browse_dialog.title") }}</DialogTitle
                                    >
                                    <button
                                        type="button"
                                        class="rounded-md text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                                        @click="$emit('update:modelValue', false)"
                                    >
                                        <span class="sr-only">Close</span>
                                        <XMarkIcon class="size-5" aria-hidden="true" />
                                    </button>
                                </div>
                                <div class="mt-3 relative">
                                    <div
                                        class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3"
                                    >
                                        <MagnifyingGlassIcon
                                            class="size-4 text-gray-400"
                                            aria-hidden="true"
                                        />
                                    </div>
                                    <input
                                        v-model="searchQuery"
                                        type="text"
                                        class="block w-full rounded-lg border-0 py-2 pl-9 pr-3 text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 text-sm"
                                        :placeholder="
                                            t('channels.browse_dialog.search_placeholder')
                                        "
                                    />
                                </div>
                            </div>

                            <div class="overflow-y-auto" style="max-height: 420px">
                                <div v-if="loading" class="flex items-center justify-center py-16">
                                    <svg
                                        class="animate-spin size-6 text-indigo-600"
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="none"
                                        viewBox="0 0 24 24"
                                    >
                                        <circle
                                            class="opacity-25"
                                            cx="12"
                                            cy="12"
                                            r="10"
                                            stroke="currentColor"
                                            stroke-width="4"
                                        />
                                        <path
                                            class="opacity-75"
                                            fill="currentColor"
                                            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                                        />
                                    </svg>
                                </div>

                                <ul
                                    v-else-if="filteredChannels.length > 0"
                                    role="list"
                                    class="divide-y divide-gray-100"
                                >
                                    <li
                                        v-for="channel in filteredChannels"
                                        :key="channel.id"
                                        class="flex items-center justify-between gap-x-4 px-6 py-3.5 hover:bg-gray-50"
                                    >
                                        <div class="min-w-0 flex-1">
                                            <div class="flex items-center gap-x-2">
                                                <GlobeAltIcon
                                                    v-if="channel.type === channelTypes.Public"
                                                    class="size-4 shrink-0 text-gray-400"
                                                    aria-hidden="true"
                                                />
                                                <LockClosedIcon
                                                    v-if="channel.type === channelTypes.Private"
                                                    class="size-4 shrink-0 text-gray-400"
                                                    aria-hidden="true"
                                                />
                                                <span
                                                    class="text-sm font-semibold text-gray-900 truncate"
                                                    >{{ channel.displayname }}</span
                                                >
                                                <span
                                                    v-if="isJoined(channel)"
                                                    class="inline-flex items-center rounded-full bg-green-50 px-2 py-0.5 text-xs font-medium text-green-700 ring-1 ring-inset ring-green-600/20"
                                                    >{{ t("channels.browse_dialog.joined") }}</span
                                                >
                                            </div>
                                            <div
                                                class="mt-1 flex items-center gap-x-2 text-xs text-gray-500"
                                            >
                                                <div class="flex items-center gap-x-1">
                                                    <UserGroupIcon
                                                        class="size-3.5 text-gray-400"
                                                        aria-hidden="true"
                                                    />
                                                    <span>{{
                                                        channel.channel_members.length
                                                    }}</span>
                                                </div>
                                                <template v-if="channel.description">
                                                    <svg
                                                        viewBox="0 0 2 2"
                                                        class="size-0.5 fill-current"
                                                    >
                                                        <circle cx="1" cy="1" r="1" />
                                                    </svg>
                                                    <span class="truncate">{{
                                                        channel.description
                                                    }}</span>
                                                </template>
                                            </div>
                                        </div>
                                        <button
                                            type="button"
                                            :disabled="joiningChannelId === channel.id"
                                            @click="handleChannelClick(channel)"
                                            class="shrink-0 rounded-md px-3 py-1.5 text-sm font-semibold shadow-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600 disabled:opacity-50 disabled:cursor-not-allowed"
                                            :class="
                                                isJoined(channel)
                                                    ? 'bg-white text-gray-900 ring-1 ring-inset ring-gray-300 hover:bg-gray-50'
                                                    : 'bg-indigo-600 text-white hover:bg-indigo-500'
                                            "
                                        >
                                            <span
                                                v-if="joiningChannelId === channel.id"
                                                class="flex items-center gap-x-1.5"
                                            >
                                                <svg
                                                    class="animate-spin size-3.5"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                    fill="none"
                                                    viewBox="0 0 24 24"
                                                >
                                                    <circle
                                                        class="opacity-25"
                                                        cx="12"
                                                        cy="12"
                                                        r="10"
                                                        stroke="currentColor"
                                                        stroke-width="4"
                                                    />
                                                    <path
                                                        class="opacity-75"
                                                        fill="currentColor"
                                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                                                    />
                                                </svg>
                                            </span>
                                            <span v-else>{{
                                                isJoined(channel)
                                                    ? t("channels.browse_dialog.button_view")
                                                    : t("channels.browse_dialog.button_join")
                                            }}</span>
                                        </button>
                                    </li>
                                </ul>

                                <div
                                    v-else
                                    class="flex flex-col items-center justify-center py-16 px-6"
                                >
                                    <ChatBubbleBottomCenterTextIcon
                                        class="size-10 text-gray-300"
                                        aria-hidden="true"
                                    />
                                    <p class="mt-3 text-sm font-semibold text-gray-900">
                                        {{
                                            searchQuery
                                                ? t("channels.browse_dialog.no_results")
                                                : t("channels.browse_dialog.nothing_to_browse")
                                        }}
                                    </p>
                                    <p v-if="!searchQuery" class="mt-1 text-sm text-gray-500">
                                        {{
                                            t(
                                                "channels.browse_dialog.nothing_to_browse_description",
                                            )
                                        }}
                                    </p>
                                </div>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, watch, computed } from "vue";
import { useRouter } from "vue-router";
import chatService from "@/services/chatService";
import { channelTypes } from "@/constants/channels";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useChannelsStore } from "@/store/channels";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import {
    XMarkIcon,
    UserGroupIcon,
    GlobeAltIcon,
    LockClosedIcon,
    ChatBubbleBottomCenterTextIcon,
    MagnifyingGlassIcon,
} from "@heroicons/vue/24/outline";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
});

const emits = defineEmits(["update:modelValue"]);

const router = useRouter();
const userStore = useUserStore();
const channelsStore = useChannelsStore();
const channels = ref([]);
const searchQuery = ref("");
const loading = ref(false);
const joiningChannelId = ref(null);

const filteredChannels = computed(() => {
    const q = searchQuery.value.toLowerCase().trim();

    return channels.value
        .filter((channel) => [channelTypes.Public, channelTypes.Private].includes(channel.type))
        .filter((channel) => {
            if (!q) return true;

            return (
                channel.displayname.toLowerCase().includes(q) ||
                channel.description?.toLowerCase().includes(q)
            );
        });
});

function isJoined(channel) {
    return channel.channel_members.some((member) => member.user_id === userStore.user.id);
}

function handleChannelClick(channel) {
    if (isJoined(channel)) {
        router.push({ name: "chat", params: { chatId: channel.id } });
        emits("update:modelValue", false);

        return;
    }

    joiningChannelId.value = channel.id;

    chatService
        .joinChannel(channel.id)
        .then((response) => {
            channel.channel_members.push(...response.data);
            useAlertStore().showSuccess(`You have joined the channel ${channel.displayname}`);
            channelsStore.setChannels([channel]);
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        })
        .finally(() => {
            joiningChannelId.value = null;
        });
}

watch(
    () => props.modelValue,
    (newValue) => {
        if (newValue) {
            searchQuery.value = "";
            loading.value = true;
            chatService.getChannels("all=true").then((response) => {
                channels.value = response.data;
                loading.value = false;
            });
        }
    },
);
</script>
