<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        class="md:border-l w-full h-full flex grow flex-col overflow-y-auto border-gray-200 bg-white"
    >
        <form class="flex h-full flex-col divide-y divide-gray-200 bg-white w-full">
            <div class="h-0 flex-1 max-w-96 overflow-y-auto">
                <div class="border-b border-gray-200 px-4 py-4 sm:px-6">
                    <div class="flex items-center justify-between">
                        <div class="min-w-0">
                            <h1 class="text-sm font-semibold text-gray-900 truncate">
                                {{
                                    channelStore.currentChannel.type == channelTypes.Direct
                                        ? getChannelName(channelStore.currentChannel)
                                        : channelStore.currentChannel.displayname
                                }}
                            </h1>
                            <p
                                v-if="channelStore.currentChannel.type != channelTypes.Direct"
                                class="mt-0.5 text-xs text-gray-500 line-clamp-2"
                            >
                                {{
                                    channelStore.currentChannel.description == ""
                                        ? t("channels.details.no_description")
                                        : channelStore.currentChannel.description
                                }}
                            </p>
                            <p v-else class="mt-0.5 text-xs text-gray-500">
                                {{ t("channels.details.direct_message_with") }}
                                {{ getChannelName(channelStore.currentChannel) }}
                            </p>
                        </div>
                        <button
                            type="button"
                            class="hidden md:flex ml-3 shrink-0 items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
                            @click="channelStore.details = false"
                        >
                            <span class="sr-only">Close panel</span>
                            <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                        </button>
                    </div>
                </div>
                <div class="flex flex-1 flex-col justify-between">
                    <div class="divide-y divide-gray-200 px-4 sm:px-6">
                        <div class="space-y-6 pb-5 pt-6">
                            <div>
                                <h3 class="text-sm font-medium leading-6 text-gray-900">
                                    {{ t("channels.details.channel_members") }}
                                </h3>
                                <div class="mt-2">
                                    <div class="flex space-x-2">
                                        <ul role="list" class="divide-y divide-gray-100 w-full">
                                            <li
                                                v-for="user in directMembers"
                                                :key="user.id"
                                                class="flex items-center justify-between gap-x-6 py-2"
                                            >
                                                <div class="flex min-w-0 gap-x-4">
                                                    <div class="min-h-12 min-w-12 h-12 w-12">
                                                        <UserAvatar :user="user" status />
                                                    </div>

                                                    <div class="min-w-0 flex-auto">
                                                        <p
                                                            class="text-sm font-semibold leading-6 text-gray-900 truncate"
                                                        >
                                                            {{ user.name + " " + user.lastname }}
                                                            <span
                                                                v-if="
                                                                    user.channel_role ==
                                                                        channelRoles.Admin &&
                                                                    channelStore.currentChannel
                                                                        .type != channelTypes.Direct
                                                                "
                                                                class="inline-flex items-center rounded-md bg-gray-50 px-2 py-1 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                                                            >
                                                                admin
                                                            </span>
                                                        </p>
                                                        <p
                                                            class="mt-1 truncate text-xs leading-5 text-gray-500"
                                                        >
                                                            {{ user.email }}
                                                        </p>
                                                    </div>
                                                </div>
                                                <Menu
                                                    v-if="
                                                        isAdmin() &&
                                                        channelStore.currentChannel.type !=
                                                            channelTypes.Direct
                                                    "
                                                    as="div"
                                                    class="relative inline-block text-left"
                                                >
                                                    <div>
                                                        <MenuButton
                                                            class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100"
                                                        >
                                                            <EllipsisVerticalIcon
                                                                class="h-5 w-5"
                                                                aria-hidden="true"
                                                            />
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
                                                            class="absolute right-0 z-10 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                                                        >
                                                            <div class="py-1">
                                                                <MenuItem v-slot="{ active }">
                                                                    <a
                                                                        @click="
                                                                            changeUserRole(user)
                                                                        "
                                                                        :class="[
                                                                            active
                                                                                ? 'bg-gray-100 text-gray-900'
                                                                                : 'text-gray-700',
                                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                                        ]"
                                                                    >
                                                                        {{
                                                                            user.channel_role ==
                                                                            channelRoles.Admin
                                                                                ? t(
                                                                                      "channels.details.make_channel_member",
                                                                                  )
                                                                                : t(
                                                                                      "channels.details.make_channel_admin",
                                                                                  )
                                                                        }}</a
                                                                    >
                                                                </MenuItem>
                                                                <MenuItem v-slot="{ active }">
                                                                    <a
                                                                        @click="removeUser(user)"
                                                                        :class="[
                                                                            active
                                                                                ? 'bg-gray-100 text-gray-900'
                                                                                : 'text-gray-700',
                                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                                        ]"
                                                                        >{{
                                                                            t(
                                                                                "channels.details.remove_from_channel",
                                                                            )
                                                                        }}</a
                                                                    >
                                                                </MenuItem>
                                                            </div>
                                                        </MenuItems>
                                                    </transition>
                                                </Menu>
                                            </li>
                                        </ul>
                                    </div>
                                </div>
                            </div>

                            <div v-if="channelStore.currentChannel.type != channelTypes.Direct">
                                <h3 class="text-sm font-medium leading-6 text-gray-900">
                                    {{ t("channels.details.channel_groups") }}
                                </h3>
                                <div class="mt-2">
                                    <p
                                        v-if="channelGroups.length === 0"
                                        class="text-xs text-gray-500"
                                    >
                                        {{ t("channels.details.no_groups") }}
                                    </p>
                                    <ul v-else role="list" class="divide-y divide-gray-100 w-full">
                                        <li
                                            v-for="cg in channelGroups"
                                            :key="cg.group_id"
                                            class="flex items-center justify-between gap-x-6 py-2"
                                        >
                                            <div
                                                class="flex min-w-0 gap-x-4 items-center cursor-pointer"
                                                @click="openGroupMembers(cg)"
                                            >
                                                <div
                                                    class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-50"
                                                >
                                                    <UserGroupIcon
                                                        class="h-5 w-5 text-indigo-600"
                                                        aria-hidden="true"
                                                    />
                                                </div>
                                                <div class="min-w-0 flex-auto">
                                                    <p
                                                        class="text-sm font-semibold leading-6 text-gray-900 truncate"
                                                    >
                                                        {{ cg.name }}
                                                    </p>
                                                    <p
                                                        class="mt-0.5 truncate text-xs leading-5 text-gray-500"
                                                    >
                                                        {{
                                                            t("channels.details.group_meta", {
                                                                count: cg.member_count ?? 0,
                                                            })
                                                        }}
                                                    </p>
                                                </div>
                                            </div>
                                            <Menu
                                                v-if="isAdmin()"
                                                as="div"
                                                class="relative inline-block text-left"
                                            >
                                                <div>
                                                    <MenuButton
                                                        class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100"
                                                    >
                                                        <EllipsisVerticalIcon
                                                            class="h-5 w-5"
                                                            aria-hidden="true"
                                                        />
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
                                                        class="absolute right-0 z-10 mt-2 w-56 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
                                                    >
                                                        <div class="py-1">
                                                            <MenuItem v-slot="{ active }">
                                                                <a
                                                                    @click="removeGroup(cg)"
                                                                    :class="[
                                                                        active
                                                                            ? 'bg-gray-100 text-gray-900'
                                                                            : 'text-gray-700',
                                                                        'block px-4 py-2 text-sm cursor-pointer',
                                                                    ]"
                                                                    >{{
                                                                        t(
                                                                            "channels.details.remove_group",
                                                                        )
                                                                    }}</a
                                                                >
                                                            </MenuItem>
                                                        </div>
                                                    </MenuItems>
                                                </transition>
                                            </Menu>
                                        </li>
                                    </ul>

                                    <div v-if="isAdmin()">
                                        <div class="mt-4 flex border-t border-gray-100 pt-3">
                                            <button
                                                @click="
                                                    channelStore.openAddUserDialog(
                                                        channelStore.currentChannel,
                                                    )
                                                "
                                                type="button"
                                                class="text-sm font-semibold leading-6 text-indigo-600 hover:text-indigo-500"
                                            >
                                                <span aria-hidden="true">+</span>
                                                {{ t("channels.details.add_user") }}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </form>

        <GroupMembersDialog
            v-model="groupMembersOpen"
            :group-id="groupMembersGroup?.group_id ?? ''"
            :group-name="groupMembersGroup?.name ?? ''"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { computed, ref, watch } from "vue";
import useChatOperations from "@/composables/chat/useChatOperations";
import chatService from "@/services/chatService";
import UserAvatar from "@/components/UserAvatar.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import { channelRoles, channelTypes } from "@/constants/channels";
import { XMarkIcon, EllipsisVerticalIcon } from "@heroicons/vue/20/solid";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";

const { getDirectChannelName } = useChatOperations();

import { useChannelsStore } from "@/store/channels";
import { useUserStore } from "@/store/user";
import { useUsers } from "@/composables/useUser";
import { useAlertStore } from "@/store/alerts";

const channelStore = useChannelsStore();
const userStore = useUserStore();
const alertStore = useAlertStore();

// Trigger a lazy load so the direct-channel counterpart name resolves.
useUsers(() => channelStore.currentChannel?.channel_members?.map((m) => m.user_id) ?? []);

const channelGroups = computed(() => channelStore.currentChannel?.channel_groups ?? []);

const groupMembersOpen = ref(false);
const groupMembersGroup = ref(null);

function openGroupMembers(cg) {
    groupMembersGroup.value = cg;
    groupMembersOpen.value = true;
}

async function loadChannelGroups() {
    const ch = channelStore.currentChannel;

    if (!ch || ch.type === channelTypes.Direct) return;
    if (Array.isArray(ch.channel_groups)) return;
    try {
        const res = await chatService.getChannelGroups(ch.id);

        ch.channel_groups = res.data ?? [];
    } catch {
        ch.channel_groups = [];
    }
}

watch(
    () => channelStore.currentChannel?.id,
    () => loadChannelGroups(),
    { immediate: true },
);

// The backend returns only directly-added members (is_direct) in
// channel_members, each hydrated with user_info. Members materialized via an
// attached group are represented by the channel's groups instead. We render
// straight from that payload, so this no longer depends on a global user list.
const directMembers = computed(() => {
    const channel = channelStore.currentChannel;

    if (!channel?.channel_members) return [];

    const out = [];

    for (const member of channel.channel_members) {
        if (!member.is_direct) continue;
        if (channel.type === channelTypes.Direct && member.user_id === userStore.user.id) {
            continue;
        }

        const info = member.user_info;

        if (!info) continue;
        out.push({
            ...info,
            id: member.user_id,
            channel_role: member.role,
            is_direct: member.is_direct,
        });
    }

    return out;
});

function getChannelName(channel) {
    return getDirectChannelName(userStore, channel);
}

function isAdmin() {
    for (let i = 0; i < channelStore.currentChannel.channel_members.length; i++) {
        if (
            channelStore.currentChannel.channel_members[i].user_id === userStore.user.id &&
            channelStore.currentChannel.channel_members[i].role === channelRoles.Admin
        ) {
            return true;
        }
    }

    return false;
}

function removeUser(user) {
    chatService
        .removeUserFromChannel({
            id: channelStore.currentChannel.id,
            user: user.id,
        })
        .then(() => {
            let index = channelStore.currentChannel.channel_members.findIndex(
                (member) => member.user_id === user.id,
            );

            if (index !== -1) {
                channelStore.currentChannel.channel_members.splice(index, 1);
            }
        });
}

function changeUserRole(user) {
    chatService
        .changeUserRole({
            id: channelStore.currentChannel.id,
            user: user.id,
            role:
                user.channel_role === channelRoles.Admin ? channelRoles.Member : channelRoles.Admin,
        })
        .then(() => {
            let index = channelStore.currentChannel.channel_members.findIndex(
                (member) => member.user_id === user.id,
            );

            if (index !== -1) {
                channelStore.currentChannel.channel_members[index].role =
                    user.channel_role === channelRoles.Admin
                        ? channelRoles.Member
                        : channelRoles.Admin;
            }
        });
}

async function removeGroup(cg) {
    const channel = channelStore.currentChannel;

    try {
        await chatService.removeGroupFromChannel(channel.id, cg.group_id);

        channel.channel_groups = (channel.channel_groups ?? []).filter(
            (g) => g.group_id !== cg.group_id,
        );

        try {
            const res = await chatService.getChannels();
            const updated = (res.data ?? []).find((c) => c.id === channel.id);

            if (updated) {
                channel.channel_members = updated.channel_members;
            }
        } catch {}
    } catch {
        alertStore.showError(t.value("channels.details.group_remove_failed"));
    }
}
</script>
