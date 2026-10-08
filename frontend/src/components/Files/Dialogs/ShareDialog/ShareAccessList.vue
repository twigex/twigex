<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <h3 class="text-base font-semibold leading-6 text-gray-600 mb-2">
        {{ t("files.share_dialog.users_with_access") }}
    </h3>
    <ul
        role="list"
        class="max-h-[35vh] divide-y divide-gray-100 overflow-y-auto bg-white shadow-sm ring-1 ring-gray-900/5 sm:rounded-xl"
    >
        <li class="relative flex justify-between gap-x-6 px-4 py-2 hover:bg-gray-50 sm:px-6">
            <div class="flex min-w-0 gap-x-4">
                <div class="h-12 w-12 flex-none">
                    <UserAvatar :user-id="ownerId" />
                </div>
                <div class="min-w-0 flex-auto">
                    <p class="text-sm font-semibold leading-6 text-gray-900">
                        <a>
                            <span class="absolute inset-x-0 -top-px bottom-0" />
                            {{ fileDetails?.owner }}
                        </a>
                    </p>
                    <p class="mt-1 flex text-xs leading-5 text-gray-500">
                        <a class="relative truncate hover:underline">{{
                            fileDetails?.ownerEmail
                        }}</a>
                    </p>
                </div>
            </div>
            <div class="flex shrink-0 items-center gap-x-4">
                <div class="hidden sm:flex sm:flex-col sm:items-end">
                    <p class="text-sm leading-6 text-gray-900">
                        {{ t("files.share_dialog.owner") }}
                    </p>
                </div>
            </div>
        </li>

        <template v-if="fileDetails != null">
            <li
                v-for="user in fileDetails.sharedUsers"
                :key="user.id"
                class="relative flex justify-between gap-x-6 px-4 py-2 hover:bg-gray-50 sm:px-6"
            >
                <div class="flex min-w-0 gap-x-4">
                    <div class="h-12 w-12 flex-none">
                        <UserAvatar :user="user" />
                    </div>
                    <div class="min-w-0 flex-auto">
                        <p class="text-sm font-semibold leading-6 text-gray-900">
                            <a>
                                <span class="absolute inset-x-0 -top-px bottom-0" />
                                {{ user.name + " " + user.lastname }}
                            </a>
                        </p>
                        <p class="mt-1 flex text-xs leading-5 text-gray-500">
                            <a class="relative truncate hover:underline">{{ user.email }}</a>
                        </p>
                    </div>
                </div>
                <div class="flex shrink-0 items-center gap-x-4">
                    <div class="hidden sm:flex sm:flex-col sm:items-end">
                        <DropdownMenu
                            v-if="showUserMenu(user)"
                            size="md"
                            :items="userMenuItems(user)"
                            @select="(item) => emit('menuSelect', item)"
                        >
                            <template #button>
                                <span
                                    class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600"
                                >
                                    <EllipsisHorizontalIcon class="h-5 w-5" aria-hidden="true" />
                                </span>
                            </template>
                        </DropdownMenu>
                    </div>
                </div>
            </li>
        </template>
        <template v-if="fileDetails && (fileDetails.sharedGroups ?? []).length > 0">
            <li
                v-for="group in fileDetails.sharedGroups"
                :key="group.id"
                class="relative flex justify-between gap-x-6 px-4 py-2 hover:bg-gray-50 sm:px-6"
            >
                <div class="flex min-w-0 gap-x-4 cursor-pointer" @click="emit('openGroup', group)">
                    <div
                        class="h-12 w-12 flex-none rounded-full bg-indigo-100 flex items-center justify-center"
                    >
                        <UserGroupIcon class="h-6 w-6 text-indigo-600" aria-hidden="true" />
                    </div>
                    <div class="min-w-0 flex-auto">
                        <p class="text-sm font-semibold leading-6 text-gray-900">
                            {{ group.name }}
                        </p>
                        <p class="mt-1 flex text-xs leading-5 text-gray-500">
                            {{
                                t("files.share_dialog.group_meta", {
                                    count: group.member_count ?? 0,
                                })
                            }}
                        </p>
                    </div>
                </div>
                <div class="flex shrink-0 items-center gap-x-4">
                    <DropdownMenu
                        v-if="showGroupMenu"
                        size="md"
                        :items="groupMenuItems(group)"
                        @select="(item) => emit('menuSelect', item)"
                    >
                        <template #button>
                            <span
                                class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600"
                            >
                                <EllipsisHorizontalIcon class="h-5 w-5" aria-hidden="true" />
                            </span>
                        </template>
                    </DropdownMenu>
                </div>
            </li>
        </template>
    </ul>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import DropdownMenu from "@/components/DropdownMenu.vue";
import { EllipsisHorizontalIcon } from "@heroicons/vue/20/solid";
import { UserGroupIcon } from "@heroicons/vue/24/outline";

defineProps({
    ownerId: {
        type: [String, Number],
        default: null,
    },
    fileDetails: {
        type: Object,
        default: null,
    },
    showUserMenu: {
        type: Function,
        required: true,
    },
    showGroupMenu: {
        type: Boolean,
        default: false,
    },
    userMenuItems: {
        type: Function,
        required: true,
    },
    groupMenuItems: {
        type: Function,
        required: true,
    },
});

const emit = defineEmits(["menuSelect", "openGroup"]);
</script>
