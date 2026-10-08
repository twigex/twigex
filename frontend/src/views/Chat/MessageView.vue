<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full min-h-0">
        <!-- Header -->
        <header class="shrink-0 h-14 border-b border-gray-200 bg-white">
            <RoomHeader :channel="getCurrentChannel(route.params.chatId)" />
        </header>

        <!-- Messages -->
        <div class="relative flex-1 min-h-0">
            <div
                v-if="loadingPosts"
                class="absolute inset-0 z-10 flex flex-col items-center justify-center gap-3 bg-white"
            >
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>

            <div
                class="h-full overflow-y-auto overflow-x-hidden bg-white"
                ref="messageListContainer"
                @scroll="handleScrollUpdate"
            >
                <div v-show="!loadingPosts" class="flex flex-col min-h-full justify-end">
                    <ChannelStart
                        v-if="
                            getCurrentChannel(route.params.chatId) &&
                            getCurrentChannel(route.params.chatId).type != channelTypes.Direct &&
                            showChannelStart
                        "
                        :channel="getCurrentChannel(route.params.chatId)"
                    />
                    <DirectMessageStart
                        v-else-if="
                            getCurrentChannel(route.params.chatId) &&
                            getCurrentChannel(route.params.chatId).type == channelTypes.Direct &&
                            showChannelStart
                        "
                        :channel="getCurrentChannel(route.params.chatId)"
                    />
                    <div v-for="(item, index) in posts" :key="item.id">
                        <!-- Date separator -->
                        <div
                            v-if="startsADay(index)"
                            class="flex items-center gap-x-3 px-4 py-3"
                            aria-hidden="true"
                        >
                            <div class="flex-1 border-t border-gray-200" />
                            <span
                                class="px-3 py-0.5 rounded-full bg-gray-100 text-xs font-medium text-gray-500 whitespace-nowrap select-none"
                            >
                                {{ dateLabel(item.created) }}
                            </span>
                            <div class="flex-1 border-t border-gray-200" />
                        </div>

                        <!-- Unread messages divider -->
                        <div
                            v-if="item.id && channelStore.unreadMessageId === item.id"
                            class="flex items-center gap-x-3 px-4 py-1"
                        >
                            <div class="flex-1 border-t border-indigo-400/50" />
                            <span
                                class="px-3 py-0.5 rounded-full bg-indigo-50 text-xs font-semibold text-indigo-600 whitespace-nowrap select-none"
                            >
                                {{ t("channels.message_view.new_messages") }}
                            </span>
                            <div class="flex-1 border-t border-indigo-400/50" />
                        </div>

                        <ChatMessage
                            :id="`message-${item.id}`"
                            :message="item"
                            :collapsed="collapsesIntoPrevious(index)"
                            :highlight="item.id == reply?.id"
                            :disabled="disabled"
                            @reply="handleReply(item)"
                            @delete="handleDelete(item)"
                            @forward="handleForward(item)"
                            @update="handleUpdate"
                            @reply-clicked="handleReplyClicked"
                            @reaction-clicked="handleReactionClicked(item)"
                            @retry="handleRetry(item)"
                            @discard="handleDiscard(item)"
                            @disable-hover="disabled = $event"
                        />
                    </div>
                </div>
            </div>

            <Transition
                enter-active-class="transition ease-out duration-200"
                enter-from-class="opacity-0 translate-y-2"
                enter-to-class="opacity-100 translate-y-0"
                leave-active-class="transition ease-in duration-150"
                leave-from-class="opacity-100 translate-y-0"
                leave-to-class="opacity-0 translate-y-2"
            >
                <button
                    v-show="showScrollButton"
                    type="button"
                    :aria-label="t('channels.message_view.scroll_to_bottom')"
                    :title="t('channels.message_view.scroll_to_bottom')"
                    class="absolute bottom-4 right-4 z-10 flex size-10 items-center justify-center rounded-full bg-white text-gray-600 shadow-lg ring-1 ring-gray-200 hover:bg-gray-50 hover:text-gray-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                    @click="scrollToLatest"
                >
                    <ChevronDownIcon class="size-5" aria-hidden="true" />
                </button>
            </Transition>
        </div>

        <!-- Input area -->
        <div
            v-if="isArchived"
            class="shrink-0 border-t border-gray-200 bg-gray-50 px-4 py-4 text-center text-sm text-gray-600"
        >
            {{ t("channels.message_view.archived") }}
        </div>
        <div v-else class="shrink-0 bg-white border-t border-gray-200">
            <div class="px-4 pt-2 pb-1">
                <ChatInput
                    :reply="reply"
                    @send-message="handleSendMessage"
                    @cancel-reply="reply = null"
                />
                <!-- Typing indicator -->
                <div class="h-5 flex items-center mt-0.5">
                    <div v-if="typingHere.length > 0" class="flex items-center gap-x-1.5">
                        <div class="flex gap-x-0.5">
                            <div
                                class="w-1.5 h-1.5 rounded-full bg-gray-400 animate-bounce"
                                style="animation-delay: 0ms"
                            />
                            <div
                                class="w-1.5 h-1.5 rounded-full bg-gray-400 animate-bounce"
                                style="animation-delay: 150ms"
                            />
                            <div
                                class="w-1.5 h-1.5 rounded-full bg-gray-400 animate-bounce"
                                style="animation-delay: 300ms"
                            />
                        </div>
                        <span class="text-xs text-gray-500">
                            <template v-if="typingHere.length == 1">
                                {{ userStore.getUserFullName(typingHere[0].user) }}
                                {{ t("channels.message_view.is_typing") }}
                            </template>
                            <template v-else>
                                {{ t("channels.message_view.multiple_typing") }}
                            </template>
                        </span>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <TransitionRoot as="template" :show="deleteDialog">
        <Dialog class="relative z-10" @close="deleteDialog = false">
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
                            class="relative transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-lg sm:p-6"
                        >
                            <div class="sm:flex sm:items-start">
                                <div
                                    class="mx-auto flex size-12 shrink-0 items-center justify-center rounded-full bg-red-100 sm:mx-0 sm:size-10"
                                >
                                    <ExclamationTriangleIcon
                                        class="size-6 text-red-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left">
                                    <DialogTitle
                                        as="h3"
                                        class="text-base font-semibold text-gray-900"
                                        >{{
                                            t("channels.message_view.dialog.delete_message.title")
                                        }}</DialogTitle
                                    >
                                    <div class="mt-2">
                                        <p class="text-sm text-gray-500">
                                            {{
                                                t(
                                                    "channels.message_view.dialog.delete_message.confirm",
                                                )
                                            }}
                                        </p>
                                    </div>
                                </div>
                            </div>
                            <div class="mt-5 sm:mt-4 sm:flex sm:flex-row-reverse">
                                <button
                                    type="button"
                                    class="inline-flex w-full justify-center rounded-md bg-red-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-red-500 sm:ml-3 sm:w-auto"
                                    @click="deleteMessage"
                                >
                                    {{ t("common.button.delete") }}
                                </button>
                                <button
                                    type="button"
                                    class="mt-3 inline-flex w-full justify-center rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 sm:mt-0 sm:w-auto"
                                    @click="deleteDialog = false"
                                    ref="cancelButtonRef"
                                >
                                    {{ t("common.button.cancel") }}
                                </button>
                            </div>
                        </DialogPanel>
                    </TransitionChild>
                </div>
            </div>
        </Dialog>
    </TransitionRoot>
    <ForwardDialog v-model="forwardDialog" @forward="forward" />
    <ReactionDialog v-model="reactionDialog" :message="reactionMessage" />
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import { useRoute } from "vue-router";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import { useUserStore } from "@/store/user";
import { useUsers } from "@/composables/useUser";
import { channelTypes, postFetchDirection, LAST_VIEWED_CHANNEL } from "@/constants/channels.js";
import RoomHeader from "@/components/Chat/Messages/RoomHeader.vue";
import ChatMessage from "@/components/Chat/Messages/ChatMessage.vue";
import ChatInput from "@/components/Chat/Messages/ChatInput.vue";
import ChannelStart from "@/components/Chat/Messages/ChannelStart.vue";
import DirectMessageStart from "@/components/Chat/Messages/DirectMessageStart.vue";
import BaseSpinner from "@/components/BaseSpinner.vue";
import ForwardDialog from "@/components/Chat/Dialogs/ForwardDialog.vue";
import ReactionDialog from "@/components/Chat/Dialogs/ReactionDialog.vue";
import useChatOperations from "@/composables/chat/useChatOperations";
import { useChannelScroll } from "@/composables/chat/useChannelScroll";
import { useMessageActions } from "@/composables/chat/useMessageActions";
import {
    isFirstMessageOfDay,
    shouldCollapse,
    dateSeparatorLabel,
} from "@/composables/chat/useMessageGrouping";

import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { ExclamationTriangleIcon, ChevronDownIcon } from "@heroicons/vue/24/outline";

const { getCurrentChannel } = useChatOperations();

const scroll = useChannelScroll();
const messageListContainer = scroll.container;

const {
    reply,
    deleteDialog,
    reactionDialog,
    reactionMessage,
    forwardDialog,
    handleSendMessage,
    handleRetry,
    handleDiscard,
    handleReply,
    handleUpdate,
    handleDelete,
    deleteMessage,
    handleReactionClicked,
    handleForward,
    forward,
} = useMessageActions({
    channelId: () => route.params.chatId,
    onSent: () => scroll.scrollToNewest(),
});

const POST_COLLAPSE_TIMEOUT = 300000;
const userStore = useUserStore();
const channelStore = useChannelsStore();

const isArchived = computed(() => !!channelStore.currentChannel?.deleted);
const route = useRoute();

const typingHere = computed(() => channelStore.getUsersTypingInChannel(route.params.chatId));

// Trigger a lazy load so typing-indicator names resolve in the template.
useUsers(() => typingHere.value.map((entry) => entry.user));
const history = ref(false);
const loaded = ref(false);
const allowToLoadPosts = ref(true);
const loadingPosts = ref(true);
const showChannelStart = ref(false);
const disabled = ref(false);
const showScrollButton = ref(false);

function handleReplyClicked(data) {
    allowToLoadPosts.value = false;
    const chatId = route.params.chatId;
    const exists = getCurrentChannel(chatId).posts.some((post) => post.id == data.id);

    if (exists) {
        scroll.scrollToMessage(data.id, {
            onSettled: () => (allowToLoadPosts.value = true),
        });
    } else {
        history.value = true;

        chatService
            .getChannelPosts(chatId, data.id, postFetchDirection.History)
            .then((response) => {
                if (route.params.chatId !== chatId) return;

                if (response.data.posts.length < 61) {
                    showChannelStart.value = true;
                }

                getCurrentChannel(chatId).posts = response.data.posts.reverse();

                nextTick(() => {
                    scroll.scrollToMessage(data.id, {
                        onSettled: () => (allowToLoadPosts.value = true),
                    });
                });
            });
    }
}

async function getChannelPosts(id, postId = "", param = "") {
    try {
        const response = await chatService.getChannelPosts(id, postId, param);

        return Promise.resolve(response);
    } catch (error) {
        return Promise.reject(error);
    }
}

function addPosts(posts, direction = "", channelId = route.params.chatId) {
    const channel = getCurrentChannel(channelId);

    if (!channel) return;

    const isActive = route.params.chatId === channelId;

    if (direction == "") {
        channel.posts = posts.reverse();

        if (isActive) {
            channelStore.unreadMessageId = getFirstUnreadMessageId();
            readChannelPosts();
            isChannelStart(posts);
        }
    } else if (direction == postFetchDirection.History) {
        channelStore.setPosts(posts.reverse());
        if (isActive) isChannelStart(posts);
    } else if (direction == postFetchDirection.Previous) {
        if (isActive) {
            scroll.preservePositionWhile(() => {
                channel.posts.unshift(...posts.reverse());
                isChannelStart(posts);
            });
        } else {
            channel.posts.unshift(...posts.reverse());
        }
    } else {
        channel.posts.push(...posts);

        if (isActive) {
            channelStore.unreadMessageId = getFirstUnreadMessageId();
        }
    }
}

function getFirstUnreadMessageId() {
    const channel = getCurrentChannel(route.params.chatId);
    const member = channel?.channel_members.find((m) => m.user_id == userStore.user.id);

    if (!member) return null;
    if (member.msg_count >= channel.msg_count) return null;

    const message = channel.posts.find((p) => p.created > member.last_viewed_at);

    return message?.id ?? null;
}

function isChannelStart(posts) {
    if (posts.length < 30) {
        showChannelStart.value = true;
    }
}

function readChannelPosts() {
    let channel = getCurrentChannel(route.params.chatId);

    if (!channel) return;

    let me = channel.channel_members.find((m) => m.user_id == userStore.user.id);

    if (!me) return;

    if (me.msg_count >= channel.msg_count) {
        me.last_viewed_at = Math.floor(Date.now());

        return;
    }

    const channelId = route.params.chatId;

    setTimeout(() => {
        chatService.readChannelPosts({ id: channelId }).then(() => {
            channelStore.readChannelPosts(channelId, userStore.user.id);
        });
    }, 1000);
}

// A short page back means the end of what the channel holds.
const PAGE_SIZE = 30;
const NEAR_TOP = 300;
const NEAR_BOTTOM = 200;

function loadOlderPosts(chatId) {
    const channel = getCurrentChannel(chatId);
    const oldest = channel?.posts.length ? channel.posts[0].id : "";

    allowToLoadPosts.value = false;

    getChannelPosts(chatId, oldest, postFetchDirection.Previous).then((response) => {
        addPosts(response.data.posts, postFetchDirection.Previous, chatId);

        if (route.params.chatId !== chatId) return;

        allowToLoadPosts.value = response.data.posts.length >= PAGE_SIZE;
    });
}

function loadNewerPosts(chatId) {
    const channel = getCurrentChannel(chatId);

    if (!channel?.posts.length) return;

    allowToLoadPosts.value = false;

    const newest = channel.posts[channel.posts.length - 1].id;

    getChannelPosts(chatId, newest, postFetchDirection.Next).then((response) => {
        addPosts(response.data.posts, postFetchDirection.Next, chatId);

        if (route.params.chatId !== chatId) return;

        const full = response.data.posts.length >= PAGE_SIZE;

        allowToLoadPosts.value = full;
        if (!full) history.value = false;
    });
}

function handleScrollUpdate(event) {
    const chatId = route.params.chatId;

    scroll.record(event);

    if (allowToLoadPosts.value) {
        if (scroll.movedUp.value && scroll.distanceFromTop.value <= NEAR_TOP) {
            loadOlderPosts(chatId);
        } else if (
            !scroll.movedUp.value &&
            scroll.distanceFromBottom.value <= NEAR_BOTTOM &&
            history.value
        ) {
            loadNewerPosts(chatId);
        }
    }

    showScrollButton.value =
        history.value || scroll.distanceFromBottom.value > scroll.viewportHeight.value * 2;
}

function scrollToLatest() {
    const chatId = route.params.chatId;

    if (!history.value) {
        scroll.scrollToNewest();
        showScrollButton.value = false;

        return;
    }

    allowToLoadPosts.value = false;

    getChannelPosts(chatId).then((response) => {
        if (route.params.chatId !== chatId) return;

        addPosts(response.data.posts, "", chatId);
        history.value = false;
        showScrollButton.value = false;

        nextTick(() => {
            scroll.scrollToNewest();
            allowToLoadPosts.value = true;
        });
    });
}

const posts = computed(() => getCurrentChannel(route.params.chatId)?.posts ?? []);

const startsADay = (index) => isFirstMessageOfDay(posts.value[index], posts.value[index - 1]);

const collapsesIntoPrevious = (index) =>
    shouldCollapse(posts.value[index], posts.value[index - 1], POST_COLLAPSE_TIMEOUT);

const dateLabel = (timestamp) =>
    dateSeparatorLabel(timestamp, {
        locale: userStore.getLanguage || undefined,
        translate: t.value,
    });

onMounted(() => {
    const chatId = route.params.chatId;

    localStorage.setItem(LAST_VIEWED_CHANNEL, chatId);

    getChannelPosts(chatId).then((response) => {
        addPosts(response.data.posts, "", chatId);

        if (route.params.chatId !== chatId) return;

        nextTick(() => {
            scroll.scrollToNewest();
        });

        loadingPosts.value = false;
    });

    window.addEventListener("click", readChannelPosts);

    loaded.value = true;
});

onUnmounted(() => {
    window.removeEventListener("click", readChannelPosts);
});

//watch route.params.id for changes
watch(
    () => route.params.chatId,
    (id) => {
        localStorage.setItem(LAST_VIEWED_CHANNEL, id);
        showChannelStart.value = false;
        history.value = false;
        allowToLoadPosts.value = false;
        disabled.value = false;

        const channel = channelStore.channels.find((c) => c.id === id);

        if (channel.posts.length == 0) {
            loadingPosts.value = true;
        }

        if (channel.posts.length > 0) {
            getChannelPosts(
                id,
                channel.posts[channel.posts.length - 1].id,
                postFetchDirection.Next,
            ).then((reponse) => {
                addPosts(reponse.data.posts, postFetchDirection.Next, id);

                if (route.params.chatId !== id) return;

                isChannelStart(getCurrentChannel(id)?.posts ?? []);

                readChannelPosts();

                nextTick(() => {
                    scroll.scrollToNewest();

                    allowToLoadPosts.value = true;
                });
            });

            return;
        }

        getChannelPosts(id).then((reponse) => {
            addPosts(reponse.data.posts, "", id);

            if (route.params.chatId !== id) return;

            nextTick(() => {
                scroll.scrollToNewest();

                allowToLoadPosts.value = true;
            });

            loadingPosts.value = false;
        });
    },
);

watch(history, (isHistory) => {
    if (isHistory) showScrollButton.value = true;
});

// watch getCurrentChannel(route.params.chatId).posts for changes deep
watch(
    () => getCurrentChannel(route.params.chatId)?.posts,
    () => {
        nextTick(() => {
            scroll.followNewest();
        });
    },
    { deep: true },
);
</script>

<style scoped>
.scroller {
    height: 100%;
}
</style>
