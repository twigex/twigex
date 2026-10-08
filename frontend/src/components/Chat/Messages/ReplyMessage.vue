<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="flex flex-row bg-indigo-50 border-l-4 border-indigo-400 rounded-r-md p-2 cursor-pointer hover:bg-indigo-100/60 transition-colors"
    >
        <div class="flex flex-col w-full min-w-0">
            <div class="flex flex-row items-center justify-between mb-0.5">
                <span class="text-xs font-semibold text-indigo-700 truncate">
                    {{ userStore.getUserFullName(message.user_id) }}
                </span>
                <span
                    class="ml-2 shrink-0 inline-flex items-center rounded-md bg-indigo-100 px-1.5 py-0.5 text-xs font-medium text-indigo-600"
                    >{{ t("channels.reply_message.reply") }}</span
                >
            </div>
            <div
                v-if="message.metadata && message.metadata.files"
                class="flex items-center gap-x-1 text-xs text-gray-500"
            >
                <PhotoIcon class="size-3.5 text-gray-400 shrink-0" />
                <span>Attachment</span>
            </div>
            <div
                v-if="message.message != ''"
                class="text-xs text-gray-600 truncate line-clamp-2 leading-relaxed prose chat-markdown"
            >
                <MessageContent :text="message.message" />
            </div>
            <div v-else class="text-xs text-gray-500 italic">
                {{ t("channels.reply_message.click_to_see_attachments") }}
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useUserStore } from "@/store/user";
import { useUser, useMentionUsers } from "@/composables/useUser";
import MessageContent from "@/components/Chat/Messages/MessageContent.vue";
import { PhotoIcon } from "@heroicons/vue/24/outline";

const props = defineProps({
    message: {
        type: Object,
        required: true,
    },
});

const userStore = useUserStore();

// Trigger a lazy load so the reply author's name resolves in the template.
useUser(() => props.message?.user_id);

useMentionUsers(() => props.message);
</script>
