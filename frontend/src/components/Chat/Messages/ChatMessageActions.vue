<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="absolute z-10 right-2 flex space-x-1 bg-white border rounded-md p-1 items-center"
        :class="[
            (hovered && !menuOpen && !emojiOpen && !disabled) ||
            (hovered && (menuOpen || emojiOpen))
                ? 'opacity-100 visible pointer-events-auto'
                : 'opacity-0 invisible pointer-events-none',

            collapsed ? '-top-11' : '-top-5',
        ]"
    >
        <!-- eslint-disable vue/no-v-html -- renderEmoji sanitizes with DOMPurify -->
        <button
            v-for="reaction in recentlyUsedEmojis()"
            :key="reaction"
            v-html="renderEmoji(reaction)"
            class="hover:bg-gray-100 size-8 rounded-md flex items-center justify-center text-sm"
            @click="emit('react', reaction)"
        ></button>
        <!-- eslint-enable vue/no-v-html -->

        <div ref="emojiRef">
            <ExpressionPicker
                @click="emit('toggleEmoji')"
                :disable-gifs="true"
                @select-emoji="emit('react', $event)"
            />
        </div>

        <Menu v-slot="{ open }" ref="menuRef" as="div" class="relative inline-block text-left">
            <MenuButton ref="reference" v-sync-open="open" class="hover:bg-gray-100 p-1 rounded-md">
                <EllipsisHorizontalIcon class="size-6 text-gray-400" aria-hidden="true" />
            </MenuButton>

            <transition
                enter-active-class="transition-opacity ease-out duration-100"
                enter-from-class="opacity-0"
                enter-to-class="opacity-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
            >
                <MenuItems
                    ref="floating"
                    class="absolute right-0 z-50 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                    :style="floatingStyles"
                >
                    <div class="py-1">
                        <MenuItem v-slot="{ active }">
                            <a
                                @click="emit('reply')"
                                :class="[
                                    active
                                        ? 'bg-gray-100 text-gray-900 outline-none'
                                        : 'text-gray-700',
                                    'group flex items-center px-4 py-2 text-sm',
                                ]"
                            >
                                <ArrowUturnLeftIcon
                                    :class="[
                                        active ? 'text-gray-500' : '',
                                        'mr-3 size-5 text-gray-400',
                                    ]"
                                    aria-hidden="true"
                                />
                                {{ t("channels.message.reply") }}
                            </a>
                        </MenuItem>
                        <MenuItem v-slot="{ active }">
                            <a
                                @click.stop="emit('forward')"
                                :class="[
                                    active
                                        ? 'bg-gray-100 text-gray-900 outline-none'
                                        : 'text-gray-700',
                                    'group flex items-center px-4 py-2 text-sm',
                                ]"
                            >
                                <ArrowRightIcon
                                    :class="[
                                        active ? 'text-gray-500' : '',
                                        'mr-3 size-5 text-gray-400',
                                    ]"
                                    aria-hidden="true"
                                />
                                {{ t("channels.message.forward") }}
                            </a>
                        </MenuItem>
                    </div>
                    <div v-if="userStore.user.id == message.user_id" class="py-1">
                        <MenuItem v-slot="{ active }">
                            <a
                                @click="emit('toggleEdit')"
                                :class="[
                                    active
                                        ? 'bg-gray-100 text-gray-900 outline-none'
                                        : 'text-gray-700',
                                    'group flex items-center px-4 py-2 text-sm',
                                ]"
                            >
                                <PencilIcon
                                    :class="[
                                        active ? 'text-gray-500' : '',
                                        'mr-3 size-5 text-gray-400',
                                    ]"
                                    aria-hidden="true"
                                />
                                {{ t("channels.message.edit") }}
                            </a>
                        </MenuItem>
                    </div>
                    <div v-if="userStore.user.id == message.user_id" class="py-1">
                        <MenuItem v-slot="{ active }">
                            <a
                                @click="emit('delete')"
                                :class="[
                                    active
                                        ? 'bg-gray-100 text-red-600 outline-none'
                                        : 'text-red-500',
                                    'group flex items-center px-4 py-2 text-sm',
                                ]"
                            >
                                <TrashIcon
                                    :class="[
                                        active ? 'text-red-600' : '',
                                        'mr-3 size-5 text-red-500',
                                    ]"
                                    aria-hidden="true"
                                />
                                {{ t("channels.message.delete") }}
                            </a>
                        </MenuItem>
                    </div>
                </MenuItems>
            </transition>
        </Menu>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { useFloating, autoPlacement } from "@floating-ui/vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { EllipsisHorizontalIcon } from "@heroicons/vue/24/solid";
import { ArrowUturnLeftIcon, TrashIcon, ArrowRightIcon } from "@heroicons/vue/20/solid";
import { PencilIcon } from "@heroicons/vue/24/outline";
import ExpressionPicker from "@/components/ExpressionPicker/ExpressionPicker.vue";
import { useConditionalClickOutside } from "@/composables/useClickOutside";
import useChatOperations from "@/composables/chat/useChatOperations";
import { t } from "@/i18n/index.js";
import { useUserStore } from "@/store/user";
import { quickReactionEmojis } from "@/utils/emoji";

const props = defineProps({
    message: {
        type: Object,
        required: true,
    },
    hovered: {
        type: Boolean,
        default: false,
    },
    menuOpen: {
        type: Boolean,
        default: false,
    },
    emojiOpen: {
        type: Boolean,
        default: false,
    },
    disabled: {
        type: Boolean,
        default: false,
    },
    collapsed: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits([
    "react",
    "reply",
    "forward",
    "toggleEdit",
    "delete",
    "update:menuOpen",
    "toggleEmoji",
    "dismiss",
]);

const { renderEmoji } = useChatOperations();
const userStore = useUserStore();

const menuRef = ref(null);
const emojiRef = ref(null);
const reference = ref(null);
const floating = ref(null);

const { floatingStyles } = useFloating(reference, floating, {
    middleware: [
        autoPlacement({
            crossAxis: true,
            alignment: "start",
        }),
    ],
});

const vSyncOpen = {
    updated(el, binding) {
        if (binding.value !== binding.oldValue) {
            emit("update:menuOpen", binding.value);
        }
    },
};

function recentlyUsedEmojis() {
    const recents = JSON.parse(localStorage.getItem("recentEmojis") || "[]");

    return quickReactionEmojis(recents).map((name) => `:${name}:`);
}

useConditionalClickOutside(
    menuRef,
    () => props.menuOpen,
    () => emit("dismiss"),
);

useConditionalClickOutside(
    emojiRef,
    () => props.emojiOpen,
    () => emit("dismiss"),
);
</script>
