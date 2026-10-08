<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <nav class="flex flex-1 flex-col px-2">
        <div class="min-w-0 flex flex-row justify-between text-sm leading-6 py-2">
            <ul role="list" class="space-y-6">
                <li
                    v-for="(activityItem, activityItemIdx) in detailsStore.activities"
                    :key="activityItem.id"
                    class="relative flex gap-x-4"
                >
                    <div
                        :class="[
                            activityItemIdx === detailsStore.activities.length - 1
                                ? 'h-6'
                                : '-bottom-6',
                            'absolute left-0 top-0 flex w-6 justify-center',
                        ]"
                    >
                        <div class="w-px bg-gray-200" />
                    </div>
                    <template v-if="activityItem.Type === 'file_download'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.file_download") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>

                    <template v-else-if="activityItem.Type === 'file_share'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <div class="flex flex-col w-full">
                            <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                                <span class="font-medium text-gray-900">{{
                                    activityItem.UserName
                                }}</span>
                                {{ t("files.activities.shared") }}

                                <span class="font-medium text-gray-900">{{
                                    activityItem.Parameters.name
                                }}</span>
                            </p>
                            <ul role="list" class="divide-y divide-gray-100 w-full">
                                <li
                                    v-for="person in getSharedUsers(activityItem.Parameters.users)"
                                    :key="person.email"
                                    class="border rounded-xl hover:bg-gray-100 flex gap-x-4 px-3 py-2 cursor-pointer w-full"
                                >
                                    <div class="h-6 w-6 flex-none">
                                        <UserAvatar :user="person" />
                                    </div>
                                    <div class="min-w-0">
                                        <p class="text-sm font-semibold leading-6 text-gray-900">
                                            {{ person.name + " " + person.lastname }}
                                        </p>
                                        <p class="mt-1 truncate text-xs leading-5 text-gray-500">
                                            {{ person.email }}
                                        </p>
                                    </div>
                                </li>
                            </ul>
                        </div>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_create'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.created") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_upload'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.file_upload") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_public_upload'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            {{ t("files.activities.file_public_upload") }}
                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_public_edit'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            {{ t("files.activities.file_public_edit") }}
                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_delete'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.deleted") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_restore'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.restored") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_rename'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.renamed_file") }}
                            <strong>{{ activityItem.Parameters.from }}</strong>
                            {{ t("files.activities.to") }}
                            <strong>{{ activityItem.Parameters.to }}</strong>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_update'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.updated") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>

                    <template v-else-if="activityItem.Type === 'file_view'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.viewed") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else-if="activityItem.Type === 'file_meta'">
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ t("files.activities.edited_metadata") }}

                            <span class="font-medium text-gray-900">{{
                                activityItem.Parameters.name
                            }}</span>
                        </p>
                    </template>
                    <template v-else>
                        <div
                            class="relative flex h-6 w-6 flex-none items-center justify-center bg-white"
                        >
                            <div
                                class="h-1.5 w-1.5 rounded-full bg-gray-100 ring-1 ring-gray-300"
                            />
                        </div>
                        <p class="flex-auto py-0.5 text-xs leading-5 text-gray-500">
                            <span class="font-medium text-gray-900">{{
                                activityItem.UserName
                            }}</span>
                            {{ activityItem.Type }}

                            <span class="font-medium text-gray-900">{{
                                t("files.activities.file")
                            }}</span>
                        </p>
                    </template>
                </li>
            </ul>
        </div>
    </nav>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { useDetailsStore } from "@/store/details";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useUsers } from "@/composables/useUser";

const detailsStore = useDetailsStore();
const userStore = useUserStore();

// Trigger a lazy load so shared-with user names resolve in the activity list.
useUsers(() => (detailsStore.activities ?? []).flatMap((a) => a.Parameters?.users ?? []));

function getSharedUsers(users) {
    return (users ?? []).map((user) => userStore.getUserById(user)).filter(Boolean);
}
</script>
