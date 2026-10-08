// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { ref } from "vue";
import chatService from "@/services/chatService";
import { useChannelsStore } from "@/store/channels";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { t } from "@/i18n/index.js";
import useChatOperations from "@/composables/chat/useChatOperations";

// `channelId` is a getter rather than a value because the view outlives the
// route param.
export function useMessageActions({ channelId, onSent }) {
    const channelStore = useChannelsStore();
    const userStore = useUserStore();
    const alertStore = useAlertStore();
    const { getCurrentChannel, addPostsToChannel } = useChatOperations();

    const reply = ref(null);
    const deleteDialog = ref(false);
    const messageToDelete = ref(null);
    const reactionDialog = ref(false);
    const reactionMessage = ref(null);
    const forwardDialog = ref(false);
    const forwardMessage = ref(null);

    function handleSendMessage(message) {
        const channel = channelId();
        const replyTo = reply.value;

        const pending = {
            id: `pending-${Date.now()}-${Math.random().toString(36).slice(2)}`,
            channel_id: channel,
            user_id: userStore.user.id,
            message: message.message,
            created: Date.now(),
            reply: replyTo,
            reactions: [],
            metadata: {},
            sendState: "sending",
            sendPayload: {
                message: message.message,
                attachments: message.attachments,
                reply: replyTo ? replyTo.id : "",
                gif: message.gif,
            },
        };

        addPostsToChannel(channel, pending);

        reply.value = null;
        onSent?.();

        submitPending(channel, pending.id);
    }

    function findPending(channel, pendingId) {
        const posts = channelStore.getChannelById(channel)?.posts;

        if (!posts) return { posts: null, index: -1 };

        return { posts, index: posts.findIndex((post) => post.id === pendingId) };
    }

    function submitPending(channel, pendingId) {
        const { posts, index } = findPending(channel, pendingId);

        if (index === -1) return;

        const pending = posts[index];

        pending.sendState = "sending";

        const { message, attachments, reply: replyId, gif } = pending.sendPayload;

        chatService
            .createChannelPost(channel, message, attachments, replyId, gif, pendingId)
            .then((response) => {
                const current = findPending(channel, pendingId);

                if (current.index === -1) return;

                const created = response.data?.posts?.[0];

                if (!created || current.posts.some((post) => post.id === created.id)) {
                    current.posts.splice(current.index, 1);
                } else {
                    current.posts.splice(current.index, 1, created);
                }
            })
            .catch(() => {
                const current = findPending(channel, pendingId);

                if (current.index === -1) return;

                current.posts[current.index].sendState = "failed";
            });
    }

    function handleRetry(message) {
        submitPending(message.channel_id, message.id);
    }

    function handleDiscard(message) {
        const { posts, index } = findPending(message.channel_id, message.id);

        if (index === -1) return;

        posts.splice(index, 1);
    }

    function handleReply(message) {
        reply.value = message;
    }

    function handleUpdate(message, done) {
        chatService
            .updateChannelPost(
                channelId(),
                message.id,
                message.message,
                message.removedFiles,
                message.attachments,
            )
            .then((response) => {
                done?.(true);

                const posts = getCurrentChannel(channelId())?.posts;

                if (!posts) return;

                const index = posts.findIndex((post) => post.id === message.id);

                if (index !== -1) posts[index] = response.data;
            })
            .catch(() => {
                alertStore.showError(t.value("channels.message.edit_failed"));
                done?.(false);
            });
    }

    function handleDelete(message) {
        messageToDelete.value = message;
        deleteDialog.value = true;
    }

    function deleteMessage() {
        const id = messageToDelete.value?.id;

        if (!id) return;

        chatService.deleteChannelPost(channelId(), id).then(() => {
            const channel = getCurrentChannel(channelId());

            if (channel) {
                channel.posts = channel.posts.filter((post) => post.id !== id);
            }

            messageToDelete.value = null;
            deleteDialog.value = false;
        });
    }

    function handleReactionClicked(message) {
        reactionMessage.value = message;
        reactionDialog.value = true;
    }

    // The refresh runs behind the dialog so it does not wait on a request.
    async function handleForward(message) {
        forwardMessage.value = message;
        forwardDialog.value = true;

        const { data } = await chatService.getChannels();

        channelStore.setChannels(data);
    }

    function forward(data) {
        chatService.forwardChannelPost(
            channelId(),
            forwardMessage.value.id,
            data.channels,
            data.users,
        );

        forwardDialog.value = false;
    }

    return {
        reply,
        deleteDialog,
        messageToDelete,
        reactionDialog,
        reactionMessage,
        forwardDialog,
        forwardMessage,
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
    };
}
