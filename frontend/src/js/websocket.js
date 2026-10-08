// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRouter, useRoute } from "vue-router";
import { useChannelsStore } from "@/store/channels";
import { useUserStore } from "@/store/user";
import { useNotificationsStore } from "@/store/notifications";
import { useMeetingsStore } from "@/store/meetings";
import { useCallsStore } from "@/store/calls";
import { useAlertStore } from "@/store/alerts";
import useChatOperations from "@/composables/chat/useChatOperations";
import useNotificationOperations from "../composables/useNotifications";
import notificationService from "@/services/notificationService";
import chatService from "@/services/chatService";
import userService from "@/services/userService";
import constructNotificationMessage, { postNotificationText } from "@/utils/notifications.js";
import { toDisplayText } from "@/utils/mentions";
import { postFetchDirection, channelTypes } from "@/constants/channels";

var router = null;
var route = null;
var channelsStore = null;
var userStore = null;
var notificationsStore = null;
var reconnecting = false;
var closed = false;
var getCurrentChannel = null;
var addPostsToChannel = null;
var handleLinkPreviewUpdate = null;

var redirect = null;

var everOpened = false;
var retries = 0;
var retryTimer = null;
var initialized = false;

const MAX_BACKOFF = 30000;

export var connection = null;

export function initialize() {
    if (!initialized) {
        const {
            getCurrentChannel: getCurrentChannelFunc,
            addPostsToChannel: addPostsToChannelFunc,
            handleLinkPreviewUpdate: handleLinkPreviewUpdateFunc,
        } = useChatOperations();
        const { redirect: redirectFunc } = useNotificationOperations();

        getCurrentChannel = getCurrentChannelFunc;
        addPostsToChannel = addPostsToChannelFunc;
        handleLinkPreviewUpdate = handleLinkPreviewUpdateFunc;
        redirect = redirectFunc;

        router = useRouter();
        route = useRoute();
        channelsStore = useChannelsStore();
        notificationsStore = useNotificationsStore();
        userStore = useUserStore();

        initialized = true;
    }

    closed = false;
    everOpened = false;
    retries = 0;
    connect();
}

function connect() {
    if (!initialized) {
        return;
    }

    if (
        connection &&
        (connection.readyState === WebSocket.CONNECTING || connection.readyState === WebSocket.OPEN)
    ) {
        return;
    }

    clearTimeout(retryTimer);

    let protocol = "wss://";

    if (window.location.protocol == "http:") {
        protocol = "ws://";
    }

    let address = protocol + window.location.host + "/api/notifications/ws";

    connection = new WebSocket(address);

    connection.onmessage = (event) => {
        const messages = event.data.trim().split("\n");

        for (const message of messages) {
            if (!message.trim()) continue;
            const object = JSON.parse(message);

            handleEvent(object);
        }
    };

    connection.onopen = () => {
        everOpened = true;
        retries = 0;

        if (reconnecting) {
            loadUnreadPosts();

            notificationService.notifications().then((res) => {
                notificationsStore.setNotifications(res.data);
            });

            userService.statuses().then((response) => {
                userStore.setStatuses(response.data);
            });

            reconnectListeners.forEach((listener) => listener());
        }
    };

    connection.onclose = () => {
        if (closed) {
            return;
        }

        if (!everOpened) {
            return;
        }

        scheduleReconnect();
    };
}

function scheduleReconnect() {
    clearTimeout(retryTimer);

    const delay = Math.min(1000 * 2 ** retries, MAX_BACKOFF) + Math.random() * 250;

    retries++;
    reconnecting = true;

    retryTimer = setTimeout(connect, delay);
}

export function reconnectIfDead() {
    if (closed || !initialized) {
        return;
    }

    if (
        !connection ||
        connection.readyState === WebSocket.CLOSED ||
        connection.readyState === WebSocket.CLOSING
    ) {
        retries = 0;
        reconnecting = true;
        connect();
    }
}

export function close() {
    closed = true;
    clearTimeout(retryTimer);

    if (connection) {
        connection.close();
    }
}

export function sendPresenceEvent() {
    if (connection.readyState !== WebSocket.OPEN) {
        return;
    }

    connection.send(
        JSON.stringify({
            event: "presence",
        }),
    );
}

const projectChangeListeners = new Set();
const reconnectListeners = new Set();

export function onProjectChange(listener) {
    projectChangeListeners.add(listener);

    return () => projectChangeListeners.delete(listener);
}

// Changes made while the socket was down are not replayed, so views that
// show live data reload when it comes back.
export function onReconnect(listener) {
    reconnectListeners.add(listener);

    return () => reconnectListeners.delete(listener);
}

// Exported for tests.
export function handleEvent(msg) {
    switch (msg.event) {
        case "project_change":
            projectChangeListeners.forEach((listener) => listener(msg.data));
            break;
        case "file_create":
        case "file_share":
        case "file_upload":
        case "file_rename":
        case "file_delete":
        case "file_restore":
        case "file_download":
        case "file_public_download":
            handleFileNotifications(msg);
            break;
        case "task_created":
        case "project_deleted":
        case "project_invite":
        case "task_deleted":
        case "task_assigned":
        case "task_status_changed":
        case "task_comment":
            handleProjectNotifications(msg);
            break;

        case "meeting_scheduled":
        case "meeting_updated":
        case "meeting_cancelled":
            handleMeetingNotification(msg);
            break;
        case "meeting_started":
        case "meeting_ended":
        case "meetings_refresh":
            handleMeetingLiveUpdate();
            break;
        case "meeting_presence":
            useMeetingsStore().updateParticipants(msg.data.meeting_id, msg.data.participants);
            break;
        case "incoming_call":
            handleIncomingCall(msg);
            break;
        case "call_connected":
            handleCallConnected(msg);
            break;
        case "call_declined":
            handleCallDeclined(msg);
            break;
        case "call_cancelled":
            handleCallCancelled(msg);
            break;
        case "call_handled":
            handleCallHandled(msg);
            break;
        case "post":
            handlePost(msg);
            break;
        case "post_mention":
            handleMentionNotification(msg);
            break;
        case "post_update":
            handlePostUpdate(msg);
            break;
        case "post_delete":
            handlePostDelete(msg);
            break;
        case "typing":
            handleTyping(msg);
            break;
        case "presence":
            handlePresence(msg);
            break;
        case "channel_update":
            handleChannelUpdate(msg);
            break;
        case "post_reaction":
            handlePostReaction(msg);
            break;
        case "link_preview_update":
            handleLinkPreviewUpdate(msg);
            break;
        default:
            return;
    }
}

function handleProjectNotifications(msg) {
    addToNotifications(msg.data.message);

    sendDesktopNotification(msg.app, "", msg.data);
}

function handleMentionNotification(msg) {
    // Bell entry only; the sound/desktop alert for mentions is played by
    // handlePost (via the "post" event) so it is not fired twice.
    addToNotifications(msg.data.message);
}

// Transient live signal (no bell entry); just refresh the meetings list so live/ended state updates whether or not the lobby is open.
function handleMeetingLiveUpdate() {
    useMeetingsStore().loadMyMeetings();
}

function handleIncomingCall(msg) {
    const d = msg.data;
    const calls = useCallsStore();

    // Ignore a ring for a channel we're already calling on or in a call on:
    // that's our own glare, which the backend resolves into one room.
    const busyOnChannel =
        calls.outgoing?.channelId === d.channel_id ||
        (channelsStore.videoIsOpen && channelsStore.currentMeeting?.channel_id === d.channel_id);

    if (busyOnChannel) return;

    calls.setIncoming({
        channelId: d.channel_id,
        callerId: d.caller_id,
        callerName: d.caller_name,
    });

    // Attention ping that just focuses the app on click; Accept/Decline live in the in-app dialog.
    if ("Notification" in window) {
        Notification.requestPermission().then((permission) => {
            if (permission !== "granted") return;
            const n = new Notification("Twigex", {
                body: `${d.caller_name} is calling…`,
                icon: window.location.origin + "/icon.png",
            });

            n.onclick = () => window.focus();
        });
    }
}

function handleCallConnected(msg) {
    const d = msg.data;
    const calls = useCallsStore();

    // Delivered to every tab the caller has open; an idle one must not join.
    if (calls.outgoing?.channelId !== d.channel_id) return;

    calls.clearOutgoing();
    calls.clearIncoming();
    channelsStore.openVideoChat({
        id: d.meeting_id,
        channel_id: d.channel_id,
        title: d.title,
        status: "active",
        participants: 1,
    });
}

function handleCallDeclined(msg) {
    useCallsStore().clearOutgoing();
    useAlertStore().showError(`${msg.data.decliner_name} declined the call.`);
}

function handleCallCancelled() {
    const calls = useCallsStore();

    calls.clearIncoming();
    calls.clearOutgoing();
}

function handleCallHandled(msg) {
    const calls = useCallsStore();

    if (calls.incoming?.channelId !== msg.data.channel_id) return;
    calls.clearIncoming();
}

function handleMeetingNotification(msg) {
    addToNotifications(msg.data.message);

    useMeetingsStore().loadMyMeetings();

    sendDesktopNotification("chat", constructNotificationMessage(msg.data.message, true), {
        app: "chat",
        channel_id: msg.data.channel_id,
        id: msg.data.message.id,
    });
}

function handleFileNotifications(msg) {
    addToNotifications(msg.data.message);

    sendDesktopNotification(msg.app, "", msg.data);
}

function handlePost(msg) {
    const data = {
        app: "chat",
        event: msg.event,
        channel_id: msg.data.post.channel_id,
        post_id: msg.data.post.id,
        notification_id: msg.data.id,
    };

    channelsStore.removeTypingUser(msg.data.post.channel_id, msg.data.post.user_id);

    let message = postNotificationText(senderHandle(msg), postBody(msg.data.post));

    // The backend resolves who is mentioned (skipping code/links, matching the
    // renderer) and puts it on the event, so the alert sound agrees with the
    // authoritative bell/badge instead of re-parsing the text here.
    const mentions = msg.data.mentions || [];
    let contains = msg.data.mention_all === true || mentions.includes(userStore.user.id);

    let channel = channelsStore.getChannelById(msg.data.post.channel_id);

    if (
        route.name == "messages" &&
        getCurrentChannel(route.params.chatId).id == msg.data.post.channel_id
    ) {
        addPostsToChannel(msg.data.post.channel_id, msg.data.post);
        if (userStore.user?.id == msg.data.post.user_id) {
            channel.channel_members.forEach((member) => {
                if (member.user_id == userStore.user.id) {
                    member.last_viewed_at = msg.data.post.created;
                }
            });
        }

        if (msg.data.reply_post) {
            channelsStore.reply_posts.push(msg.data.reply_post);
        }

        if (!document.hasFocus()) {
            if (contains && userStore.user?.id != msg.data.post.user_id) {
                channelsStore.incrementMention(msg.data.post.channel_id, userStore.user.id);
            }

            for (let i = 0; i < channelsStore.channels.length; i++) {
                if (channelsStore.channels[i].id == msg.data.post.channel_id) {
                    if (userStore.user?.id != msg.data.post.user_id) {
                        channelsStore.channels[i].msg_count += 1;

                        if (!channelsStore.unreadMessageId) {
                            channelsStore.unreadMessageId = msg.data.post.id;
                        }
                    }

                    break;
                }
            }

            if (
                (channel != null && channel.type == channelTypes.Direct) ||
                (channel != null && contains)
            ) {
                if (userStore.user?.id != msg.data.post.user_id) {
                    sendDesktopNotification(msg.data.app, message, data);
                }
            }
        }

        return;
    }

    for (let i = 0; i < channelsStore.channels.length; i++) {
        if (channelsStore.channels[i].id == msg.data.post.channel_id) {
            if (userStore.user?.id != msg.data.post.user_id) {
                channelsStore.channels[i].msg_count += 1;
            }

            break;
        }
    }

    if (contains && userStore.user?.id != msg.data.post.user_id) {
        channelsStore.incrementMention(msg.data.post.channel_id, userStore.user.id);
    }

    if ((channel != null && channel.type == channelTypes.Direct) || (channel != null && contains)) {
        if (userStore.user?.id != msg.data.post.user_id) {
            sendDesktopNotification(msg.data.app, message, data);
        } else {
            channelsStore.setLastPost(msg.data.post.channel_id);
        }
    }
}

function handlePostUpdate(msg) {
    // The event carries only the newly added mentions, so an edit that adds
    // your @mention alerts like a new one while existing mentions never re-fire.
    const mentions = msg.data.mentions || [];
    const mentionsMe = msg.data.mention_all === true || mentions.includes(userStore.user.id);

    if (mentionsMe && userStore.user?.id != msg.data.post.user_id) {
        const viewing =
            route.name == "messages" &&
            getCurrentChannel(route.params.chatId)?.id == msg.data.post.channel_id;

        if (!viewing || !document.hasFocus()) {
            channelsStore.incrementMention(msg.data.post.channel_id, userStore.user.id);
            const message = postNotificationText(senderHandle(msg), postBody(msg.data.post));

            sendDesktopNotification("chat", message, {
                app: "chat",
                event: msg.event,
                channel_id: msg.data.post.channel_id,
                post_id: msg.data.post.id,
            });
        }
    }

    const current = getCurrentChannel(route.params.chatId);

    if (!current || current.id != msg.data.post.channel_id) {
        return;
    }

    const index = current.posts.findIndex((p) => p.id == msg.data.post.id);

    if (index !== -1) {
        current.posts[index] = msg.data.post;
    }
}

function handlePostDelete(msg) {
    const current = getCurrentChannel(route.params.chatId);

    if (!current || current.id != msg.data.post.channel_id) {
        return;
    }

    const index = current.posts.findIndex((p) => p.id == msg.data.post.id);

    if (index !== -1) {
        current.posts.splice(index, 1);
    }
}

function handleTyping(msg) {
    channelsStore.addTypingUser(msg.data.channel_id, msg.data.user_id);
}

function handlePresence(msg) {
    userStore.setUserStatus(msg.data.status);
}

function handleChannelUpdate(msg) {
    for (let i = 0; i < channelsStore.channels.length; i++) {
        if (channelsStore.channels[i].id == msg.data.channel.id) {
            channelsStore.channels[i] = msg.data.channel;

            const current = getCurrentChannel(route.params.chatId);

            if (current && current.id == msg.data.channel.id) {
                channelsStore.setCurrentChannel(msg.data.channel);
            }

            break;
        }
    }
}

function handlePostReaction(msg) {
    if (!msg.data.post.channel_id) {
        return;
    }

    const channel = getCurrentChannel(msg.data.post.channel_id);

    if (!channel) {
        return;
    }

    const post = channel.posts.find((p) => p.id == msg.data.post.id);

    if (post) {
        post.reactions = msg.data.post.reactions;
    }
}

//move to separate file later
function sendDesktopNotification(title, body, data) {
    let message = "";

    switch (title) {
        case "chat":
            message = body;
            break;
        case "files":
            message = constructNotificationMessage(data.message, true);
            break;
        case "projects":
            message = constructNotificationMessage(data.message, true);
            break;
        default:
            message = body || "You have a new notification";
            break;
    }

    if (title === "chat") {
        new Audio("/notification.ogg").play().catch(() => {});
    }

    Notification.requestPermission().then((permission) => {
        if (permission === "granted") {
            const notification = new Notification(`Twigex ${title}`, {
                icon: window.location.origin + "/icon.png",
                body: message,
                data: data,
            });

            notification.onclick = (event) => {
                notificationService.read({
                    read: true,
                    id: [data.id],
                });

                window.focus();

                if (event.target.data.app == "projects") {
                    handleProjectRedirect(event);
                } else if (event.target.data.app == "chat") {
                    router.push({
                        name: "chat",
                        params: { chatId: event.target.data.channel_id },
                    });
                } else if (event.target.data.message.app == "files") {
                    handleFileRedirect(event.target.data.message);
                }
            };
        }
    });
}

function handleFileRedirect(notification) {
    redirect(notification);
}

function handleProjectRedirect(event) {
    const notificationData = {
        id: event.target.data.message.id,
        app: event.target.data.app,
        type: event.target.data.message?.type,
        details: event.target.data.message?.details || event.target.data.message,
    };

    redirect(notificationData);
}

function addToNotifications(message) {
    notificationsStore.addNotification(message);
}

function loadUnreadPosts() {
    chatService.getChannels().then((response) => {
        channelsStore.channels.forEach((channel) => {
            let chan = response.data.find((c) => c.id == channel.id);

            if (!chan) {
                return;
            }

            let me = channel.channel_members.find((m) => m.user_id === userStore.user.id);

            if (me == null) {
                return;
            }

            if (chan.msg_count > me.msg_count) {
                const fromId = channel.posts?.length
                    ? channel.posts[channel.posts.length - 1].id
                    : "";

                const direction = fromId ? postFetchDirection.Next : "";

                chatService.getChannelPosts(channel.id, fromId, direction).then((response) => {
                    const newPosts = response.data.posts;

                    const unread = newPosts.filter((post) => post.created > me.last_viewed_at);

                    // REST posts carry no handle, and the store is cold on
                    // a fresh tab.
                    if (unread.length) {
                        userStore
                            .ensureUsers([...new Set(unread.map((p) => p.user_id))])
                            .then(() => unread.forEach((post) => searchForMentions(post)));
                    }

                    if (fromId) {
                        addPostsToChannel(channel.id, newPosts);
                    } else {
                        channel.posts = newPosts;
                    }

                    if (route.params.chatId == channel.id && newPosts.length > 0) {
                        channelsStore.unreadMessageId = newPosts[0].id;
                    }

                    //Update the channel here ?
                    channelsStore.setChannels([chan]);
                });
            }
        });
    });
}

function senderHandle(msg) {
    return msg.data.sender_username || userStore.getUserById(msg.data.post.user_id)?.username;
}

// A notification is plain text, so a stored <@id> has to be spelled out here
// rather than by the renderer.
function postBody(post) {
    return toDisplayText(
        post.message,
        (id) => post.mentions?.[id]?.username || userStore.getUserById(id)?.username || null,
    );
}

function searchForMentions(post) {
    if (
        post.message.includes(`<@${userStore.user.id}>`) ||
        post.message.includes(`@${userStore.user.username}`) ||
        post.message.includes("@all")
    ) {
        sendDesktopNotification(
            "chat",
            postNotificationText(userStore.getUserById(post.user_id)?.username, postBody(post)),
            {
                app: "chat",
                event: "post",
                channel_id: post.channel_id,
                post_id: post.id,
                notification_id: null,
            },
        );
    }
}
