// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useUserStore } from "@/store/user";
import { app } from "@/constants/system";
import { t } from "@/i18n/index";

const notificationMessages = {
    file_create: (user, file) => t.value("notifications.message.file_create", { file }),
    file_share: (user, file) => t.value("notifications.message.file_share", { file }),
    file_upload: (user, file) => t.value("notifications.message.file_upload", { file }),
    file_rename: (user, file, data) =>
        t.value("notifications.message.file_rename", {
            old_name: data.details.old_name,
            file,
        }),
    file_delete: (user, file) => t.value("notifications.message.file_delete", { file }),
    file_restore: (user, file) => t.value("notifications.message.file_restore", { file }),
    file_comment: (user, file) => t.value("notifications.message.file_comment", { file }),
    file_download: (user, file) => t.value("notifications.message.file_download", { file }),
    file_public_download: (user, file) =>
        t.value("notifications.message.file_public_download", { file }),
    task_created: (user, file, data) =>
        t.value("notifications.message.task_created", {
            task: data.details.taskName,
        }),
    project_deleted: (user, file, data) =>
        t.value("notifications.message.project_deleted", {
            workspace: data.details.workspaceName,
        }),
    project_invite: (user, file, data) =>
        t.value("notifications.message.project_invite", {
            workspace: data.details.workspaceName,
        }),
    task_deleted: (user, file, data) =>
        t.value("notifications.message.task_deleted", {
            task: data.details.taskName,
        }),
    task_assigned: (user, file, data) =>
        t.value("notifications.message.task_assigned", {
            task: data.details.taskName,
        }),
    task_status_changed: (user, file, data) =>
        t.value("notifications.message.task_status_changed", {
            task: data.details.taskName,
            status: data.details.statusName,
        }),
    task_comment: (user, file, data) =>
        t.value("notifications.message.task_comment", {
            task: data.details.taskName,
        }),
    meeting_scheduled: (user, file, data) =>
        t.value("notifications.message.meeting_scheduled", {
            channel: data.details.channelName,
        }),
    meeting_updated: (user, file, data) =>
        t.value("notifications.message.meeting_updated", {
            channel: data.details.channelName,
        }),
    meeting_cancelled: (user, file, data) =>
        t.value("notifications.message.meeting_cancelled", {
            channel: data.details.channelName,
        }),
    post_mention: (user, item, data) =>
        t.value("notifications.message.post_mention", {
            channel: data.details.channelName,
        }),
};

// A public download has no signed-in actor, so the stored sender is the owner
// and naming them would credit them with someone else's action.
export const actorlessNotifications = new Set(["file_public_download"]);

export function postNotificationText(handle, message) {
    return handle ? `@${handle}: ${message}` : message;
}

export default function constructNotificationMessage(data, prefix = false) {
    const userStore = useUserStore();
    let user = userStore.getUserById(data.sender);

    if (!user) {
        // Sender not cached yet: lazy-load so the name fills in on re-render,
        // and use a blank stand-in meanwhile so rendering never throws.
        if (data.sender) userStore.ensureUsers([data.sender]);
        user = { name: "", lastname: "" };
    }

    switch (data.app) {
        case app.Files:
            return constructFileNotificationMessage(user, data, prefix);
        case app.Projects:
            return constructProjectNotificationMessage(user, data, prefix);
        case app.Chat:
            return constructChatNotificationMessage(user, data, prefix);
        default:
            throw new Error(`Unknown app type: ${data.app}`);
    }
}

function constructFileNotificationMessage(user, data, prefix) {
    const { filename, notificationType } = data.details;

    const messageBuilder = notificationMessages[notificationType];

    if (!messageBuilder) {
        throw new Error(`Unknown notification type: ${notificationType}`);
    }

    let message = messageBuilder(user.name + " " + user.lastname, filename, data);

    if (prefix && !actorlessNotifications.has(notificationType)) {
        message = user.name + " " + user.lastname + message;
    }

    return message;
}

function constructProjectNotificationMessage(user, data, prefix) {
    const { taskName, notificationType } = data.details;

    const messageBuilder = notificationMessages[notificationType];

    if (!messageBuilder) {
        throw new Error(`Unknown notification type: ${notificationType}`);
    }

    let message = messageBuilder(user.name + " " + user.lastname, taskName, data);

    if (prefix) {
        message = user.name + " " + user.lastname + message;
    }

    return message;
}

function constructChatNotificationMessage(user, data, prefix) {
    const { notificationType, meetingTitle } = data.details;

    const messageBuilder = notificationMessages[notificationType];

    if (!messageBuilder) {
        throw new Error(`Unknown notification type: ${notificationType}`);
    }

    let message = messageBuilder(user.name + " " + user.lastname, meetingTitle, data);

    if (prefix) {
        message = user.name + " " + user.lastname + message;
    }

    return message;
}
