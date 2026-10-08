<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex space-x-2 mb-1">
        <span
            v-for="reaction in reactions"
            :key="reaction.reaction"
            @click="emit('react', reaction.reaction)"
        >
            <BaseTooltip position="top">
                <template v-slot:text>
                    <div @click="emit('reactionClicked')" class="select-none cursor-pointer">
                        {{ getReactionText(reaction) }}
                    </div>
                </template>

                <div
                    class="flex flex-row items-center justify-center text-sm cursor-pointer bg-gray-200 hover:bg-gray-300 p-1 rounded-md text-gray-500"
                >
                    <!-- eslint-disable-next-line vue/no-v-html -- renderEmoji sanitizes with DOMPurify -->
                    <span v-html="renderEmoji(reaction.reaction)" class="pr-1"> </span>
                    {{ reaction.count }}
                </div>
            </BaseTooltip>
        </span>
    </div>
</template>

<script setup>
import BaseTooltip from "@/components/BaseTooltip.vue";
import useChatOperations from "@/composables/chat/useChatOperations";
import { useUserStore } from "@/store/user";
import { t } from "@/i18n/index.js";

defineProps({
    reactions: {
        type: Array,
        required: true,
    },
});

const emit = defineEmits(["react", "reactionClicked"]);

const { renderEmoji } = useChatOperations();
const userStore = useUserStore();

function userName(id) {
    const user = userStore.getUserById(id);

    return user ? user.name + " " + user.lastname : t.value("common.label.unknown_user");
}

function getReactionText(reaction) {
    const me = userStore.user;
    const emoji = reaction.reaction;

    if (reaction.users.includes(me.id) && reaction.count === 1) {
        return t.value("channels.message.reaction_you", { emoji });
    } else if (reaction.users.includes(me.id) && reaction.count === 2) {
        let otherUserId = reaction.users.find((id) => id !== me.id);

        return t.value("channels.message.reaction_you_and_user", {
            name: userName(otherUserId),
            emoji,
        });
    } else if (reaction.users.includes(me.id) && reaction.count > 2) {
        return t.value("channels.message.reaction_you_and_others", {
            count: reaction.count - 1,
            emoji,
        });
    } else {
        let firstName = userName(reaction.users[0]);

        if (reaction.count === 1) {
            return t.value("channels.message.reaction_user", { name: firstName, emoji });
        } else if (reaction.count === 2) {
            return t.value("channels.message.reaction_two_users", {
                name: firstName,
                other: userName(reaction.users[1]),
                emoji,
            });
        } else {
            return t.value("channels.message.reaction_user_and_others", {
                name: firstName,
                count: reaction.count - 1,
                emoji,
            });
        }
    }
}
</script>
