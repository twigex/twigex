<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ul role="list" class="space-y-6 py-6">
        <li
            v-for="(activityItem, activityItemIdx) in activity"
            :key="activityItem.id"
            class="relative flex gap-x-4"
        >
            <div
                :class="[
                    activityItemIdx === activity.length - 1 ? 'h-6' : '-bottom-6',
                    'absolute left-0 top-0 flex w-6 justify-center',
                ]"
            >
                <div class="w-px bg-gray-200" />
            </div>
            <template v-if="activityItem.type === 'commented'">
                <div class="relative mt-3 size-6 flex-none">
                    <UserAvatar :user-id="activityItem.userId" :name="activityItem.person.name" />
                </div>
                <div
                    class="min-w-0 flex-auto rounded-md bg-white p-3 ring-1 ring-inset ring-gray-200"
                >
                    <div class="flex justify-between gap-x-4">
                        <div class="py-0.5 text-xs/5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.person.name
                            }}</span>
                            {{ t("projects.task_comments.commented") }}
                        </div>
                        <time
                            :datetime="activityItem.dateTime"
                            class="flex-none py-0.5 text-xs/5 text-gray-500"
                        >
                            {{ activityItem.date }}
                        </time>
                    </div>
                    <p class="whitespace-pre-wrap break-words text-sm/6 text-gray-500">
                        <template
                            v-for="(token, i) in commentTokens(activityItem.comment)"
                            :key="i"
                        >
                            <MentionTag v-if="token.userId" :user-id="token.userId" />
                            <MentionTag v-else-if="token.username" :username="token.username" />
                            <EmojiIcon v-else-if="token.emoji" :short-name="token.emoji" />
                            <template v-else>{{ token.text }}</template>
                        </template>
                    </p>
                </div>
            </template>
            <template v-else>
                <div class="relative flex size-6 flex-none items-center justify-center bg-white">
                    <div class="size-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300" />
                </div>
                <p class="min-w-0 flex-auto break-words py-0.5 text-xs/5 text-gray-500">
                    <span class="font-medium text-gray-900">{{ activityItem.person.name }}</span>
                    {{ activityItem.comment }}
                </p>
                <time
                    :datetime="activityItem.dateTime"
                    class="flex-none py-0.5 text-xs/5 text-gray-500"
                >
                    {{ activityItem.date }}
                </time>
            </template>
        </li>
    </ul>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import MentionTag from "@/components/Mentions/MentionTag.vue";
import EmojiIcon from "@/components/EmojiIcon.vue";
import { splitMentionTokens } from "@/utils/mentions";
import { splitEmojiTokens } from "@/utils/emoji";

defineProps({
    activity: { type: Array, required: true },
});

// The list is rendered again whenever its parent is, so a comment is split
// into its mentions and emoji once, not on every render.
const tokensByComment = new Map();
const commentTokens = (comment) => {
    if (!tokensByComment.has(comment)) {
        const tokens = splitMentionTokens(comment).flatMap((token) =>
            token.text === undefined ? [token] : splitEmojiTokens(token.text),
        );

        tokensByComment.set(comment, tokens);
    }

    return tokensByComment.get(comment);
};
</script>
