<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col" v-if="loaded">
        <div
            @mouseenter="hovered = true"
            @mouseleave="!menuOpen && !emojiOpen ? (hovered = false) : (hovered = true)"
            class="relative flex space-x-3 cursor-pointer px-3 transition-colors h-full"
            :class="{
                'border-t border-gray-200 pt-2': !collapsed,
                'border-none': collapsed,
                'bg-indigo-100 hover:bg-indigo-50': highlight,
                ' hover:bg-gray-100': !highlight,
            }"
        >
            <div
                v-if="!collapsed"
                class="h-8 w-8 flex-none mt-2"
                :class="{ 'opacity-60': showSending }"
            >
                <UserAvatar :user="user" status />
            </div>
            <div v-if="collapsed" class="w-8 pt-1" :class="{ 'opacity-0': !hovered }">
                <p class="flex items-center text-xs whitespace-nowrap">
                    {{ useDateOperations().getTime(message.created) }}
                </p>
            </div>
            <div
                class="w-full h-full flex flex-col max-w-full min-w-full"
                :class="[contentPadding, { 'opacity-60': showSending }]"
            >
                <div v-if="!collapsed" class="flex items-center justify-between">
                    <div class="space-x-2">
                        <span class="text-sm font-semibold text-gray-800"
                            >{{ user.name ? user.name + " " + user.lastname : "" }}
                        </span>
                        <span class="text-xs text-gray-500">{{
                            useDateOperations().getDateAndTime(message.created)
                        }}</span>
                    </div>

                    <span
                        v-if="message.type == 'forward'"
                        class="inline-flex items-center rounded-md bg-indigo-100 px-2 py-1 text-xs font-medium text-indigo-700"
                        >{{ t("channels.message.forwarded") }}</span
                    >
                </div>

                <ReplyMessage
                    @click.stop="replyClicked"
                    v-if="message.reply"
                    :message="message.reply"
                    class="px-2 mt-1"
                />
                <div class="flex flex-col w-full h-full max-w-full" v-if="!edit">
                    <div class="flex items-center justify-between h-full w-full max-w-full">
                        <div
                            class="w-full max-w-full break-words pb-1 items-center text-sm text-gray-600 prose chat-markdown"
                        >
                            <MessageContent :text="message.message" />
                        </div>

                        <span
                            v-if="message.type == 'forward' && collapsed"
                            class="inline-flex items-center rounded-md bg-indigo-100 px-2 py-1 text-xs font-medium text-indigo-700"
                            >{{ t("channels.message.forwarded") }}</span
                        >
                    </div>

                    <ChatMessageLinkPreview
                        v-if="
                            message.metadata &&
                            message.metadata.links &&
                            message.metadata.links.length > 0
                        "
                        v-model:video-playing="videoPlaying"
                        :link="message.metadata.links[0]"
                    />

                    <!-- Display attachments if present -->
                    <ChatMessageAttachments
                        v-if="
                            message.metadata &&
                            message.metadata.files &&
                            message.metadata.files.length > 0
                        "
                        v-model:image-loaded="imageLoaded"
                        :message="message"
                        :hovered="hovered"
                    />
                </div>

                <div class="flex w-full" v-else>
                    <ChatInput
                        :message="message"
                        :busy="savingEdit"
                        @cancel="edit = false"
                        @update="handleMessageUpdate"
                    />
                </div>

                <!-- Reactions below the message -->
                <ChatMessageReactions
                    v-if="reactions.length > 0"
                    :reactions="reactions"
                    @react="addReaction"
                    @reaction-clicked="$emit('reaction-clicked', message)"
                />

                <ChatMessageSendState
                    :send-state="message.sendState"
                    :show-sending="showSending"
                    :collapsed="collapsed"
                    @retry="emits('retry')"
                    @discard="emits('discard')"
                />

                <!-- Hover Actions Inside the Gray Area -->
                <ChatMessageActions
                    v-if="!message.sendState"
                    :message="message"
                    :hovered="hovered"
                    :menu-open="menuOpen"
                    :emoji-open="emojiOpen"
                    :disabled="disabled"
                    :collapsed="collapsed"
                    @react="addReaction"
                    @reply="reply"
                    @forward="forward"
                    @toggle-edit="edit = !edit"
                    @delete="deleteMessage"
                    @update:menu-open="setMenuOpen"
                    @toggle-emoji="toggleEmoji"
                    @dismiss="clearEmojiHover"
                />
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, onMounted, computed } from "vue";
import { useUserStore } from "@/store/user";
import { useUser, useUsers, useMentionUsers } from "@/composables/useUser";
import chatService from "@/services/chatService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useDelayedSending } from "@/composables/chat/useDelayedSending";
import ChatInput from "@/components/Chat/Messages/ChatInput.vue";
import ChatMessageActions from "@/components/Chat/Messages/ChatMessageActions.vue";
import ChatMessageAttachments from "@/components/Chat/Messages/ChatMessageAttachments.vue";
import ChatMessageLinkPreview from "@/components/Chat/Messages/ChatMessageLinkPreview.vue";
import ChatMessageReactions from "@/components/Chat/Messages/ChatMessageReactions.vue";
import ChatMessageSendState from "@/components/Chat/Messages/ChatMessageSendState.vue";
import MessageContent from "@/components/Chat/Messages/MessageContent.vue";
import ReplyMessage from "@/components/Chat/Messages/ReplyMessage.vue";
import UserAvatar from "@/components/UserAvatar.vue";

const props = defineProps({
    message: {
        type: Object,
        required: true,
    },
    collapsed: {
        type: Boolean,
        default: false,
    },
    highlight: {
        type: Boolean,
        default: false,
    },
    disabled: {
        type: Boolean,
        default: false,
    },
});

const emits = defineEmits([
    "reply",
    "reply-clicked",
    "reaction-clicked",
    "delete",
    "update",
    "forward",
    "disable-hover",
    "retry",
    "discard",
]);

const reactions = computed(() => {
    let sortedReactions = [];

    props.message.reactions.forEach((reaction) => {
        let reactionIndex = sortedReactions.findIndex((r) => r.reaction === reaction.reaction);

        if (reactionIndex === -1) {
            sortedReactions.push({
                reaction: reaction.reaction,
                count: 1,
                users: [reaction.user_id],
            });
        } else {
            sortedReactions[reactionIndex].count++;
            sortedReactions[reactionIndex].users.push(reaction.user_id);
        }
    });

    return sortedReactions;
});

const menuOpen = ref(false);
const emojiOpen = ref(false);
const loaded = ref(false);
const userStore = useUserStore();
const fetchedUser = useUser(() => props.message.user_id);
const user = computed(() => fetchedUser.value || { id: props.message.user_id });

// Trigger a lazy load so reactor names resolve in reaction tooltips.
useUsers(() => (props.message.reactions ?? []).map((r) => r.user_id));

useMentionUsers(() => props.message);
const edit = ref(false);
const imageLoaded = ref(false);
const hovered = ref(false);
const videoPlaying = ref(false);

function addReaction(reaction) {
    let user = userStore.user;

    let reactIndex = reactions.value.findIndex((r) => r.reaction === reaction);

    if (reactIndex > -1) {
        let userIndex = reactions.value[reactIndex].users.findIndex((u) => u === user.id);

        if (userIndex > -1) {
            chatService
                .deletePostReaction(props.message.channel_id, props.message.id, reaction)
                .then(() => {
                    //Search if there is reaction in message.reactions with user from user.value.id and if yes then delete it
                    props.message.reactions = props.message.reactions.filter(
                        (r) => !(r.reaction === reaction && r.user_id === user.id),
                    );
                });

            clearEmojiHover();

            return;
        }
    }

    chatService
        .createPostReaction(props.message.channel_id, props.message.id, reaction)
        .then((response) => {
            props.message.reactions.push(response.data);
        });

    clearEmojiHover();
}

function reply() {
    emits("reply", props.message);
}

function replyClicked() {
    emits("reply-clicked", props.message.reply);
}

function deleteMessage() {
    emits("delete");
}

function forward() {
    emits("forward");
}

const savingEdit = ref(false);

function handleMessageUpdate(message) {
    if (savingEdit.value) return;

    savingEdit.value = true;

    emits("update", message, (saved) => {
        savingEdit.value = false;

        if (saved) edit.value = false;
    });
}

function setMenuOpen(open) {
    menuOpen.value = open;

    emits("disable-hover", open);
}

function toggleEmoji() {
    emojiOpen.value = !emojiOpen.value;

    emits("disable-hover", emojiOpen.value);
}

function clearEmojiHover() {
    emojiOpen.value = false;
    hovered.value = false;

    emits("disable-hover", false);
}

const contentPadding = computed(() => {
    if (props.message.sendState === "failed") return "pr-56";
    if (props.message.sendState === "sending") return "pr-20";

    return "pr-10";
});

const showSending = useDelayedSending(() => props.message.sendState);

onMounted(() => {
    loaded.value = true;
});
</script>

<style>
.emoji {
    display: inline-block;
    width: 1.4em;
    height: 1.4em;
    background-size: contain;
    background-repeat: no-repeat;
    vertical-align: -0.25em;
    overflow: hidden;
    color: transparent;
    /* Hides text visually */
    user-select: text;
    /* Ensures text is selectable and copyable */
}

.emoji::before {
    content: attr(title);
    /* Provide accessible fallback */
    visibility: hidden;
    /* Ensure it doesn't render visually */
}

.prose code::before,
.prose code::after {
    content: none !important;
}

/* Compact chat-tuned markdown. Overrides @tailwindcss/typography's zero-specificity
   prose rules so messages read at chat scale, not document scale. */
.chat-markdown > :first-child {
    margin-top: 0;
}

.chat-markdown > :last-child {
    margin-bottom: 0;
}

.chat-markdown h1,
.chat-markdown h2,
.chat-markdown h3,
.chat-markdown h4,
.chat-markdown h5,
.chat-markdown h6 {
    margin: 0.75rem 0 0.25rem;
    font-weight: 600;
    line-height: 1.3;
    color: #374151; /* gray-700 */
}

.chat-markdown h1 {
    font-size: 1.25rem;
}

.chat-markdown h2 {
    font-size: 1.15rem;
}

.chat-markdown h3 {
    font-size: 1.05rem;
}

.chat-markdown h4,
.chat-markdown h5,
.chat-markdown h6 {
    font-size: 1rem;
}

.chat-markdown ul,
.chat-markdown ol {
    margin: 0.25rem 0;
    padding-left: 1.5rem;
}

.chat-markdown li {
    margin: 0.125rem 0;
}

.chat-markdown blockquote {
    margin: 0.375rem 0;
    padding: 0.125rem 0 0.125rem 0.75rem;
    border-left: 3px solid #d1d5db; /* gray-300 */
    color: #6b7280; /* gray-500 */
    font-style: normal;
}

/* Prose wraps blockquotes in curly quotes; the accent bar is enough here. */
.chat-markdown blockquote p:first-of-type::before,
.chat-markdown blockquote p:last-of-type::after {
    content: none;
}

.chat-markdown hr {
    margin: 0.75rem 0;
    border: 0;
    border-top: 1px solid #e5e7eb; /* gray-200 */
}

/* Prose only draws row separators; give tables a full Obsidian-style cell grid. */
.chat-markdown table {
    margin: 0.5rem 0;
    border-collapse: collapse;
    border: 1px solid #d1d5db; /* gray-300 */
    font-size: 0.8125rem;
}

.chat-markdown th,
.chat-markdown td {
    border: 1px solid #d1d5db; /* gray-300 */
    padding: 0.375rem 0.75rem;
}

.chat-markdown thead th {
    background-color: #f9fafb; /* gray-50 */
    font-weight: 600;
}
</style>
