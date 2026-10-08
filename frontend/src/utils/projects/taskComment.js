// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

const UNKNOWN_USER = { name: "Unknown", lastname: "User" };

// The history load and the websocket handler both go through here, so a
// comment that arrives live and the same comment after a reload cannot
// disagree.
export function toCommentRow(entry, { getUser }) {
    const user = getUser(entry.user_id) || UNKNOWN_USER;
    const createdAt = new Date(entry.created_at * 1000);

    return {
        id: entry.id,
        userId: entry.user_id,
        type: "commented",
        person: {
            name: `${user.name} ${user.lastname}`,
        },
        comment: entry.content,
        date: createdAt.toLocaleDateString(),
        dateTime: createdAt.toISOString(),
    };
}

const PLAIN_ACTIVITY = {
    task_created: "projects.task_comments.created_the_task",
    task_deleted: "projects.task_comments.deleted_the_task",
    project_deleted: "projects.task_comments.deleted_the_project",
};

// toActivityRow is one entry of a task's history as the panel lists it: a
// comment, or a line saying what someone did. An invitation is told from
// the invited person's side, so it is shown as theirs.
export function toActivityRow(entry, { getUser, translate }) {
    if (entry.activity_type === "Comment") return toCommentRow(entry, { getUser });

    const nameOf = (id) => {
        const user = getUser(id) || UNKNOWN_USER;

        return `${user.name} ${user.lastname}`;
    };
    let userId = entry.user_id;
    let comment;

    switch (entry.activity_type) {
        case "task_created":
        case "task_deleted":
        case "project_deleted":
            comment = translate(PLAIN_ACTIVITY[entry.activity_type]);
            break;
        case "task_assigned":
            comment = `${translate("projects.task_comments.assigned_the_task_to")} ${nameOf(entry.affected_user)}.`;
            break;
        case "task_status_changed":
            comment = `${translate("projects.task_comments.changed_the_status_of_assigned_task")} ${nameOf(entry.affected_user)}.`;
            break;
        case "project_invite":
            userId = entry.affected_user;
            comment = translate("projects.task_comments.was_invited_to_the_project");
            break;
        default:
            comment = entry.content;
    }

    const createdAt = new Date(entry.created_at * 1000);

    return {
        id: entry.id,
        userId,
        type: entry.activity_type,
        person: { name: nameOf(userId) },
        comment,
        date: createdAt.toLocaleDateString(),
        dateTime: createdAt.toISOString(),
    };
}
