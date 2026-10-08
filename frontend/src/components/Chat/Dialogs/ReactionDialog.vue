<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="modelValue">
        <Dialog class="relative z-50" @close="emits('update:modelValue', false)">
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
                            class="relative transform overflow-hidden rounded-lg bg-white h-96 px-4 pb-4 pt-5 text-left w-full shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-2xl sm:p-6"
                        >
                            <div class="sm:flex sm:items-start h-full">
                                <div
                                    class="mt-3 text-center sm:ml-4 sm:mt-0 sm:text-left h-full w-full"
                                >
                                    <div class="flex flex-row justify-between items-center">
                                        <DialogTitle
                                            as="h3"
                                            class="text-base font-semibold text-gray-900 text-left"
                                            >Reactions</DialogTitle
                                        >

                                        <button
                                            type="button"
                                            class="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-2 focus:outline-offset-2 focus:outline-indigo-600"
                                            @click="emits('update:modelValue', false)"
                                        >
                                            <span class="sr-only">Close</span>
                                            <XMarkIcon class="size-6" aria-hidden="true" />
                                        </button>
                                    </div>

                                    <div
                                        v-if="message"
                                        class="mt-2 h-full overflow-y-auto flex flex-row"
                                    >
                                        <div
                                            class="mb-5 pr-5 flex flex-col items-left overflow-y-auto border-r"
                                        >
                                            <nav class="flex flex-1 flex-col" aria-label="Sidebar">
                                                <ul role="list" class="space-y-1">
                                                    <li
                                                        v-for="item in groupedReactions"
                                                        :key="item"
                                                        @click="selectedReaction = item"
                                                        :class="[
                                                            item === selectedReaction
                                                                ? 'bg-indigo-100'
                                                                : 'text-gray-700 bg-gray-50 hover:bg-gray-100 hover:text-indigo-600 ',
                                                            'flex rounded-md items-center justify-center gap-x-1 py-1 px-2 cursor-pointer',
                                                        ]"
                                                    >
                                                        <!-- eslint-disable vue/no-v-html -- renderEmoji sanitizes with DOMPurify -->
                                                        <div
                                                            class="text-sm"
                                                            v-html="renderEmoji(item)"
                                                        ></div>
                                                        <!-- eslint-enable vue/no-v-html -->

                                                        <div class="text-xs text-gray-500 pt-1">
                                                            {{ reactionCount(item) }}
                                                        </div>
                                                    </li>
                                                </ul>
                                            </nav>
                                        </div>

                                        <div
                                            class="0 flex w-full mb-5 flex-col overflow-y-auto divide-y"
                                        >
                                            <div v-for="user in userList" :key="user.id">
                                                <div class="mx-2 my-1 flex flex-col items-left">
                                                    <MemberItem :user="user" />
                                                </div>
                                            </div>
                                        </div>
                                    </div>
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
import { ref, computed, watch } from "vue";
import MemberItem from "../Sidebar/MemberItem.vue";
import { Dialog, DialogPanel, DialogTitle, TransitionRoot, TransitionChild } from "@headlessui/vue";
import { XMarkIcon } from "@heroicons/vue/24/outline";
import { useUserStore } from "@/store/user";
import { useUsers } from "@/composables/useUser";
import useChatOperations from "@/composables/chat/useChatOperations";

const props = defineProps({
    modelValue: {
        type: Boolean,
        required: true,
    },
    message: {
        type: Object,
        required: false,
        default: null,
    },
});

// Trigger a lazy load so reactor names resolve in the list.
useUsers(() => (props.message?.reactions ?? []).map((r) => r.user_id));

const emits = defineEmits(["update:modelValue", "forward"]);

const { renderEmoji } = useChatOperations();

const groupedReactions = computed(() => {
    const reactions = [];

    if (props.message && props.message.reactions) {
        props.message.reactions.forEach((reaction) => {
            if (!reactions.includes(reaction.reaction)) {
                reactions.push(reaction.reaction);
            }
        });
    }

    return reactions;
});

const userList = computed(() => {
    const users = [];

    if (props.message && props.message.reactions) {
        props.message.reactions.forEach((reaction) => {
            if (reaction.reaction === selectedReaction.value) {
                const user = userStore.getUserById(reaction.user_id);

                if (user && !users.includes(user)) {
                    users.push(user);
                }
            }
        });
    }

    return users;
});

function reactionCount(reaction) {
    if (props.message && props.message.reactions) {
        return props.message.reactions.filter((r) => r.reaction === reaction).length;
    }

    return 0;
}

const userStore = useUserStore();
const selectedReaction = ref(null);

watch(
    groupedReactions,
    (reactions) => {
        // Mark the first reaction as selected by default
        if (reactions.length > 0) {
            selectedReaction.value = reactions[0];
        }
    },
    { immediate: true },
);
</script>
