<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <Disclosure v-slot="{ open }" default-open>
        <DisclosureButton
            as="div"
            class="sticky top-0 z-10 flex w-full cursor-pointer items-center gap-x-1 rounded-md bg-white px-2 py-1 text-xs font-semibold leading-6 text-gray-400 hover:bg-gray-50 hover:text-gray-600"
        >
            <ChevronDownIcon
                :class="[
                    'h-3 w-3 shrink-0 transition-transform duration-200',
                    open ? '' : '-rotate-90',
                ]"
                aria-hidden="true"
            />
            <span>{{ t("channels.navigation.channels") }}</span>
            <span class="flex-1" />
            <span
                v-if="!open && channelsUnreadCount > 0"
                class="mr-1 flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                aria-hidden="true"
            >
                {{ channelsUnreadCount > 99 ? "99+" : channelsUnreadCount }}
            </span>
            <button
                v-if="open && can('create_channel')"
                @click.stop="channelsStore.newChannelDialog = true"
                type="button"
                class="rounded p-0.5 text-gray-400 hover:bg-gray-200 hover:text-indigo-600 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            >
                <PlusIcon class="h-4 w-4" aria-hidden="true" />
            </button>
        </DisclosureButton>
        <transition
            enter-active-class="transition ease-out duration-100"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition ease-in duration-75"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
        >
            <DisclosurePanel>
                <ul role="list" class="space-y-1">
                    <li
                        v-for="item in regularChannels"
                        :key="item.id"
                        role="button"
                        tabindex="0"
                        class="cursor-pointer rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-indigo-500"
                        @click="openChannel(item)"
                        @keydown.enter="openChannel(item)"
                        @keydown.space.prevent="openChannel(item)"
                    >
                        <a
                            :class="[
                                item.id == route.params.chatId
                                    ? 'bg-gray-50 text-indigo-600'
                                    : 'text-gray-500 hover:text-indigo-600 hover:bg-gray-50',
                                'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-medium items-center justify-between',
                            ]"
                        >
                            <div class="flex flex-row justify-between items-center w-full truncate">
                                <div class="flex flex-row gap-x-3 truncate">
                                    <span class="relative size-6 shrink-0">
                                        <LetterAvatar
                                            :id="item.id"
                                            :name="item.displayname"
                                            class="size-6 rounded-md text-[0.625rem]"
                                        />
                                        <span
                                            v-if="item.type === 'P'"
                                            class="absolute -bottom-1 -right-1 flex size-3.5 items-center justify-center rounded-full bg-white"
                                            aria-hidden="true"
                                        >
                                            <LockClosedIcon class="size-2.5 text-gray-500" />
                                        </span>
                                    </span>
                                    <div
                                        :class="[
                                            'truncate',
                                            unreadByChannel[item.id] > 0
                                                ? 'font-semibold text-gray-900'
                                                : '',
                                        ]"
                                    >
                                        {{ item.displayname }}
                                    </div>
                                </div>

                                <div class="flex shrink-0 items-center gap-x-1 pl-2">
                                    <span
                                        v-if="mentionsByChannel[item.id] > 0"
                                        class="group-hover:hidden flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                                        aria-hidden="true"
                                    >
                                        @
                                    </span>
                                    <span
                                        v-if="unreadByChannel[item.id] > 0"
                                        class="group-hover:hidden flex items-center justify-center whitespace-nowrap rounded-full bg-red-500 px-1.5 text-center text-xs font-medium leading-5 text-white"
                                        aria-hidden="true"
                                    >
                                        {{
                                            unreadByChannel[item.id] > 99
                                                ? "99+"
                                                : unreadByChannel[item.id]
                                        }}
                                    </span>
                                </div>
                            </div>

                            <Menu as="div" class="relative text-left hidden group-hover:block">
                                <div>
                                    <MenuButton
                                        @click.stop
                                        class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100"
                                    >
                                        <EllipsisVerticalIcon class="size-5" aria-hidden="true" />
                                    </MenuButton>
                                </div>

                                <transition
                                    enter-active-class="transition ease-out duration-100"
                                    enter-from-class="transform opacity-0 scale-95"
                                    enter-to-class="transform opacity-100 scale-100"
                                    leave-active-class="transition ease-in duration-75"
                                    leave-from-class="transform opacity-100 scale-100"
                                    leave-to-class="transform opacity-0 scale-95"
                                >
                                    <MenuItems
                                        class="absolute right-0 z-20 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                                    >
                                        <div class="py-1">
                                            <MenuItem
                                                v-if="isChannelAdmin(item)"
                                                v-slot="{ active }"
                                            >
                                                <a
                                                    @click.stop="emit('rename', item)"
                                                    :class="[
                                                        active
                                                            ? 'bg-gray-100 text-gray-600 outline-none'
                                                            : 'text-gray-700',
                                                        'block px-4 py-2 text-sm font-normal',
                                                    ]"
                                                    >{{ t("channels.navigation.rename_channel") }}
                                                </a>
                                            </MenuItem>
                                            <MenuItem
                                                v-if="isChannelAdmin(item)"
                                                @click.stop="channelsStore.openAddUserDialog(item)"
                                                v-slot="{ active }"
                                            >
                                                <a
                                                    :class="[
                                                        active
                                                            ? 'bg-gray-100 text-gray-600 outline-none'
                                                            : 'text-gray-700',
                                                        'block px-4 py-2 text-sm font-normal',
                                                    ]"
                                                    >{{ t("channels.navigation.add_members") }}</a
                                                >
                                            </MenuItem>
                                            <MenuItem
                                                v-if="isChannelAdmin(item)"
                                                v-slot="{ active }"
                                                @click.stop="emit('archive', item)"
                                            >
                                                <a
                                                    :class="[
                                                        active
                                                            ? 'bg-gray-100 text-gray-600 outline-none'
                                                            : 'text-gray-700',
                                                        'block px-4 py-2 text-sm font-normal',
                                                    ]"
                                                    >{{
                                                        t("channels.navigation.archive_channel")
                                                    }}</a
                                                >
                                            </MenuItem>
                                            <MenuItem
                                                v-slot="{ active }"
                                                @click.stop="emit('leave', item)"
                                            >
                                                <a
                                                    :class="[
                                                        active
                                                            ? 'bg-gray-100 text-gray-600 outline-none'
                                                            : 'text-gray-700',
                                                        'block px-4 py-2 text-sm font-normal',
                                                    ]"
                                                    >{{ t("channels.navigation.leave_channel") }}</a
                                                >
                                            </MenuItem>
                                        </div>
                                    </MenuItems>
                                </transition>
                            </Menu>
                        </a>
                    </li>

                    <li v-if="can('create_channel')">
                        <Menu as="div" class="relative inline-block text-left w-full">
                            <div>
                                <MenuButton
                                    ref="reference"
                                    class="text-gray-700 hover:text-indigo-600 hover:bg-indigo-100 bg-indigo-50 group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold cursor-pointer w-full"
                                >
                                    <PlusIcon
                                        class="text-gray-400 group-hover:text-indigo-600 h-6 w-6 shrink-0"
                                        aria-hidden="true"
                                    />
                                    {{ t("channels.navigation.add_channels") }}
                                </MenuButton>
                            </div>

                            <transition
                                enter-active-class="transition ease-out duration-100"
                                enter-from-class="transform opacity-0 scale-95"
                                enter-to-class="transform opacity-100 scale-100"
                                leave-active-class="transition ease-in duration-75"
                                leave-from-class="transform opacity-100 scale-100"
                                leave-to-class="transform opacity-0 scale-95"
                            >
                                <MenuItems
                                    ref="floating"
                                    class="absolute z-20 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-hidden"
                                    :style="floatingStyles"
                                >
                                    <div class="py-1">
                                        <MenuItem
                                            v-slot="{ active }"
                                            @click="channelsStore.newChannelDialog = true"
                                        >
                                            <a
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900 outline-hidden'
                                                        : 'text-gray-700',
                                                    'group flex items-center px-4 py-2 text-sm cursor-pointer',
                                                ]"
                                            >
                                                <PlusIcon
                                                    :class="[
                                                        active ? 'text-gray-500' : '',
                                                        'mr-3 size-5 text-gray-400',
                                                    ]"
                                                    aria-hidden="true"
                                                />
                                                {{ t("channels.navigation.menu.create_channel") }}
                                            </a>
                                        </MenuItem>
                                        <MenuItem
                                            v-slot="{ active }"
                                            @click="channelsStore.browseChannelDialog = true"
                                        >
                                            <a
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900 outline-hidden'
                                                        : 'text-gray-700',
                                                    'group flex items-center px-4 py-2 text-sm cursor-pointer',
                                                ]"
                                            >
                                                <GlobeAltIcon
                                                    :class="[
                                                        active ? 'text-gray-500' : '',
                                                        'mr-3 size-5 text-gray-400',
                                                    ]"
                                                    aria-hidden="true"
                                                />
                                                {{ t("channels.navigation.menu.browse_channels") }}
                                            </a>
                                        </MenuItem>
                                    </div>
                                </MenuItems>
                            </transition>
                        </Menu>
                    </li>
                </ul>
            </DisclosurePanel>
        </transition>
    </Disclosure>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref } from "vue";
import {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
} from "@headlessui/vue";
import { LockClosedIcon, GlobeAltIcon } from "@heroicons/vue/24/outline";
import LetterAvatar from "@/components/LetterAvatar.vue";
import { PlusIcon } from "@heroicons/vue/24/solid";
import { EllipsisVerticalIcon, ChevronDownIcon } from "@heroicons/vue/20/solid";
import { useChannelsStore } from "@/store/channels";
import useChatOperations from "@/composables/chat/useChatOperations";
import { useFloating, autoPlacement } from "@floating-ui/vue";
import { usePermissions } from "@/composables/usePermissions";
import { useChatNavigation } from "@/composables/chat/useChatNavigation";

const emit = defineEmits(["rename", "archive", "leave"]);

const { isChannelAdmin } = useChatOperations();
const channelsStore = useChannelsStore();
const { can } = usePermissions();
const {
    route,
    channelsUnreadCount,
    regularChannels,
    unreadByChannel,
    mentionsByChannel,
    openChannel,
} = useChatNavigation();

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
</script>
