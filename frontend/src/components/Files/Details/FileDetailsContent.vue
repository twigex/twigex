<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        v-if="!detailsStore.file"
        class="relative w-full h-full flex grow flex-col items-center justify-center gap-y-5 overflow-y-auto bg-white px-1"
    >
        <button
            type="button"
            class="absolute right-2 top-2 flex items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors cursor-pointer"
            @click="closeDetails"
        >
            <span class="sr-only">Close panel</span>
            <XMarkIcon class="h-5 w-5" aria-hidden="true" />
        </button>
        <div class="text-center">
            <svg
                class="mx-auto h-12 w-12 text-gray-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                aria-hidden="true"
            >
                <path
                    vector-effect="non-scaling-stroke"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z"
                />
            </svg>
            <h3 class="mt-2 text-sm font-semibold text-gray-900">
                {{ t("files.details.title") }}
            </h3>
            <p class="mt-1 text-sm text-gray-500">
                {{ t("files.details.description") }}
            </p>
        </div>
    </div>
    <div v-if="detailsStore.file" class="w-full h-full flex grow flex-col overflow-hidden bg-white">
        <div class="shrink-0 border-b border-gray-200 px-4 py-4 sm:px-6">
            <div class="flex items-center justify-between gap-x-3">
                <div class="flex min-w-0 items-center gap-x-3">
                    <component
                        v-if="fileType !== 'image'"
                        :is="fileIcon"
                        :class="['h-8 w-8 flex-none', fileColor]"
                    />
                    <img
                        v-else
                        class="h-8 w-8 flex-none rounded object-cover"
                        :src="`/api/files/thumbnails/${detailsStore.file.id}`"
                    />
                    <div class="min-w-0">
                        <h1
                            class="text-sm font-semibold text-gray-900 truncate"
                            :title="detailsStore.file.name"
                        >
                            {{ detailsStore.file.name }}
                        </h1>
                        <p class="mt-0.5 text-xs text-gray-500 truncate">
                            {{
                                fileTypes[detailsStore.file.type] === undefined
                                    ? fileTypes["not-found"].fileType
                                    : fileTypes[detailsStore.file.type].fileType
                            }}
                            <template v-if="!isFolder">
                                ·
                                {{ convertSize(detailsStore.file.size) }}
                            </template>
                        </p>
                    </div>
                </div>
                <button
                    type="button"
                    class="shrink-0 flex items-center justify-center rounded-md p-1.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors cursor-pointer"
                    @click="closeDetails"
                >
                    <span class="sr-only">Close panel</span>
                    <XMarkIcon class="h-5 w-5" aria-hidden="true" />
                </button>
            </div>
        </div>

        <div class="shrink-0 px-4 pt-4 sm:px-6">
            <div class="sm:hidden">
                <select
                    id="tabs"
                    name="tabs"
                    :value="activeTabName"
                    class="block w-full rounded-md border-gray-300 focus:border-indigo-500 focus:ring-indigo-500 cursor-pointer"
                    @change="(e) => updateActiveTab(tabs.find((t) => t.name === e.target.value))"
                >
                    <option v-for="tab in tabs" :key="tab.name" :value="tab.name">
                        {{ tab.name }}
                    </option>
                </select>
            </div>
            <div class="hidden sm:block">
                <nav class="-mb-px flex gap-x-6 border-b border-gray-200" aria-label="Tabs">
                    <a
                        v-for="tab in tabs"
                        :key="tab.name"
                        :class="[
                            tab.current
                                ? 'border-indigo-500 text-indigo-600'
                                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300',
                            'whitespace-nowrap border-b-2 py-2 px-1 text-sm font-medium cursor-pointer',
                        ]"
                        :aria-current="tab.current ? 'page' : undefined"
                        @click="updateActiveTab(tab)"
                    >
                        {{ tab.name }}
                    </a>
                </nav>
            </div>
        </div>

        <div class="flex-1 overflow-y-auto">
            <div v-if="tabs[0].current" class="flex flex-col px-4 sm:px-6 pb-5">
                <dl class="mt-4 divide-y divide-gray-100 text-sm">
                    <div class="flex justify-between gap-x-4 py-2.5 leading-6">
                        <dt class="font-medium text-gray-900">
                            {{ t("files.details.type") }}
                        </dt>
                        <dd class="text-gray-500 truncate">
                            {{
                                fileTypes[detailsStore.file.type] === undefined
                                    ? fileTypes["not-found"].fileType
                                    : fileTypes[detailsStore.file.type].fileType
                            }}
                        </dd>
                    </div>
                    <div v-if="!isFolder" class="flex justify-between gap-x-4 py-2.5 leading-6">
                        <dt class="font-medium text-gray-900">
                            {{ t("files.details.size") }}
                        </dt>
                        <dd class="text-gray-500">
                            {{ convertSize(detailsStore.file.size) }}
                        </dd>
                    </div>
                    <div class="flex justify-between gap-x-4 py-2.5 leading-6">
                        <dt class="font-medium text-gray-900">
                            {{ t("files.details.owner") }}
                        </dt>
                        <dd class="text-gray-500 truncate">
                            {{ detailsStore.details?.owner }}
                        </dd>
                    </div>
                    <div class="flex justify-between gap-x-4 py-2.5 leading-6">
                        <dt class="font-medium text-gray-900">
                            {{ t("files.details.modified_at") }}
                        </dt>
                        <dd class="text-gray-500">
                            {{ useDateOperations().getDateAndTime(detailsStore.file.modified) }}
                        </dd>
                    </div>
                    <div class="flex justify-between gap-x-4 py-2.5 leading-6">
                        <dt class="font-medium text-gray-900">
                            {{ t("files.details.created_at") }}
                        </dt>
                        <dd class="text-gray-500">
                            {{ useDateOperations().getDateAndTime(detailsStore.file.created) }}
                        </dd>
                    </div>
                </dl>

                <div v-if="settingsStore.metadata.length > 0" class="mt-6">
                    <div class="flex items-center justify-between mb-1">
                        <h3 class="text-sm font-semibold leading-6 text-gray-900">
                            {{ t("files.details.metadata") }}
                        </h3>
                        <button
                            type="button"
                            @click="open = true"
                            class="rounded-md p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors cursor-pointer"
                            :title="t('files.details.edit_metadata')"
                        >
                            <span class="sr-only">{{ t("files.details.edit_metadata") }}</span>
                            <PencilSquareIcon class="h-4 w-4" aria-hidden="true" />
                        </button>
                    </div>
                    <FileMetadata :file="detailsStore.file" />
                </div>

                <div
                    v-if="
                        detailsStore.details &&
                        detailsStore.details.sharedUsers != null &&
                        detailsStore.details.sharedUsers.length > 0
                    "
                    class="mt-6"
                >
                    <h3 class="text-sm font-semibold leading-6 text-gray-900 mb-2">
                        {{ t("files.details.users_with_access") }}
                    </h3>
                    <ul
                        role="list"
                        class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200 bg-white"
                    >
                        <li
                            v-for="user in detailsStore.details.sharedUsers"
                            :key="user.id"
                            class="relative flex justify-between gap-x-4 px-3 py-2 hover:bg-gray-50"
                        >
                            <div class="flex min-w-0 gap-x-3 items-center">
                                <div class="h-9 w-9 flex-none">
                                    <UserAvatar :user="user" />
                                </div>
                                <div class="min-w-0 flex-auto">
                                    <p
                                        class="text-sm font-semibold leading-5 text-gray-900 truncate"
                                    >
                                        {{ user.name + " " + user.lastname }}
                                    </p>
                                    <p class="mt-0.5 truncate text-xs leading-5 text-gray-500">
                                        {{ user.email }}
                                    </p>
                                </div>
                            </div>
                            <div class="flex shrink-0 items-center">
                                <Menu
                                    v-if="canManage || user.id === userStore.user.id"
                                    as="div"
                                    class="relative inline-block text-left"
                                >
                                    <div>
                                        <MenuButton
                                            class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100 cursor-pointer"
                                        >
                                            <EllipsisHorizontalIcon
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
                                                <MenuItem
                                                    v-if="canManage"
                                                    @click="openShareDialog(user)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t("files.details.menu.edit_permissions")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                                <MenuItem
                                                    @click="unshareUser(user.id)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            canManage
                                                                ? t(
                                                                      "files.details.menu.remove_access",
                                                                  )
                                                                : t("files.details.menu.leave")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                            </div>
                                        </MenuItems>
                                    </transition>
                                </Menu>
                            </div>
                        </li>
                    </ul>
                </div>

                <div
                    v-if="
                        detailsStore.details &&
                        detailsStore.details.sharedGroups != null &&
                        detailsStore.details.sharedGroups.length > 0
                    "
                    class="mt-6"
                >
                    <h3 class="text-sm font-semibold leading-6 text-gray-900 mb-2">
                        {{ t("files.details.groups_with_access") }}
                    </h3>
                    <ul
                        role="list"
                        class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200 bg-white"
                    >
                        <li
                            v-for="group in detailsStore.details.sharedGroups"
                            :key="group.id"
                            class="relative flex justify-between gap-x-4 px-3 py-2 hover:bg-gray-50"
                        >
                            <div
                                class="flex min-w-0 gap-x-3 items-center cursor-pointer"
                                @click="openGroupMembers(group)"
                            >
                                <div
                                    class="h-9 w-9 flex-none rounded-full bg-indigo-100 flex items-center justify-center"
                                >
                                    <UserGroupIcon
                                        class="h-5 w-5 text-indigo-600"
                                        aria-hidden="true"
                                    />
                                </div>
                                <div class="min-w-0 flex-auto">
                                    <p
                                        class="text-sm font-semibold leading-5 text-gray-900 truncate"
                                    >
                                        {{ group.name }}
                                    </p>
                                    <p class="mt-0.5 truncate text-xs leading-5 text-gray-500">
                                        {{
                                            t("files.details.group_meta", {
                                                count: group.member_count ?? 0,
                                            })
                                        }}
                                    </p>
                                </div>
                            </div>
                            <div class="flex shrink-0 items-center">
                                <Menu
                                    v-if="canManage"
                                    as="div"
                                    class="relative inline-block text-left"
                                >
                                    <div>
                                        <MenuButton
                                            class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100 cursor-pointer"
                                        >
                                            <EllipsisHorizontalIcon
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
                                                <MenuItem
                                                    @click="openGroupShareDialog(group)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t("files.details.menu.edit_permissions")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                                <MenuItem
                                                    @click="unshareGroupAccess(group.id)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t(
                                                                "files.details.menu.remove_group_access",
                                                            )
                                                        }}</a
                                                    >
                                                </MenuItem>
                                            </div>
                                        </MenuItems>
                                    </transition>
                                </Menu>
                            </div>
                        </li>
                    </ul>
                </div>

                <div
                    v-if="
                        detailsStore.details &&
                        detailsStore.details.sharedLinks != null &&
                        detailsStore.details.sharedLinks.length > 0
                    "
                    class="mt-6"
                >
                    <h3 class="text-sm font-semibold leading-6 text-gray-900 mb-2">Shared links</h3>
                    <ul
                        role="list"
                        class="divide-y divide-gray-100 rounded-lg ring-1 ring-gray-200 bg-white"
                    >
                        <li
                            v-for="link in detailsStore.details.sharedLinks"
                            :key="link.id"
                            class="relative flex justify-between gap-x-4 px-3 py-2 hover:bg-gray-50"
                        >
                            <div class="flex min-w-0 gap-x-3 items-center">
                                <span
                                    class="inline-flex h-9 w-9 items-center justify-center rounded-full bg-gray-500"
                                >
                                    <LinkIcon class="h-5 w-5 text-white" aria-hidden="true" />
                                </span>
                                <div class="min-w-0 flex-auto">
                                    <p
                                        class="text-sm font-semibold leading-5 text-gray-900 truncate"
                                    >
                                        {{ accessLabel(link) }}
                                    </p>
                                    <div
                                        class="mt-0.5 flex items-center gap-1.5 text-xs leading-5 text-gray-500"
                                    >
                                        <LockClosedIcon
                                            v-if="link.passwordProtected"
                                            class="h-3.5 w-3.5 shrink-0"
                                            :aria-label="
                                                t('files.share_dialog.link.password_protected')
                                            "
                                        />
                                        <span v-if="link.expiration" class="shrink-0">
                                            {{ t("files.share_dialog.expires") }}
                                            {{
                                                useDateOperations().getDateAndTime(link.expiration)
                                            }}
                                        </span>
                                        <span class="truncate">{{ link.name }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="flex shrink-0 items-center">
                                <Menu as="div" class="relative inline-block text-left">
                                    <div>
                                        <MenuButton
                                            class="flex items-center rounded-full bg-gray-100 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-100 cursor-pointer"
                                        >
                                            <span class="sr-only">Open options</span>
                                            <EllipsisHorizontalIcon
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
                                                <MenuItem
                                                    @click="editLink(link)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t("files.share_dialog.link.edit_title")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                                <MenuItem
                                                    @click="copyLink(link.id)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t("files.share_dialog.link.copy_link")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                                <MenuItem
                                                    @click="deleteLink(link.id)"
                                                    v-slot="{ active }"
                                                >
                                                    <a
                                                        :class="[
                                                            active
                                                                ? 'bg-gray-100 text-gray-900'
                                                                : 'text-gray-700',
                                                            'block px-4 py-2 text-sm cursor-pointer',
                                                        ]"
                                                        >{{
                                                            t("files.share_dialog.link.delete")
                                                        }}</a
                                                    >
                                                </MenuItem>
                                            </div>
                                        </MenuItems>
                                    </transition>
                                </Menu>
                            </div>
                        </li>
                    </ul>
                </div>

                <MetadataDialog :open="open" @close="open = false" />
                <GroupMembersDialog
                    v-model="groupMembersOpen"
                    :group-id="groupMembersGroup?.id ?? ''"
                    :group-name="groupMembersGroup?.name ?? ''"
                />
                <ConfirmDialog
                    :open="confirm.open"
                    :title="confirm.title"
                    :message="confirm.message"
                    :confirm-label="t('files.details.menu.remove_access')"
                    @confirm="runConfirm"
                    @close="confirm.open = false"
                />
            </div>

            <FileActivities v-if="tabs[1].current" />
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed } from "vue";
import FileActivities from "@/components/Files/Details/FileActivities.vue";
import FileMetadata from "@/components/Files/Metadata/FileMetadata.vue";
import MetadataDialog from "@/components/Files/Metadata/MetadataDialog.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import {
    EllipsisHorizontalIcon,
    LinkIcon,
    LockClosedIcon,
    PencilSquareIcon,
    UserGroupIcon,
    XMarkIcon,
} from "@heroicons/vue/24/outline";
import { useSettingsStore } from "@/store/settings";
import { useDetailsStore } from "@/store/details";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useDialogStore } from "@/store/dialogs";
import { convertSize } from "@/utils/utils";
import fileTypes from "@/constants/fileTypes";
import useDateOperations from "@/composables/useDateOperations.js";
import shareService from "@/services/shareService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const settingsStore = useSettingsStore();
const detailsStore = useDetailsStore();
const userStore = useUserStore();
const dialogStore = useDialogStore();
const alertStore = useAlertStore();
const open = ref(false);
const groupMembersOpen = ref(false);
const groupMembersGroup = ref(null);
const confirm = ref({ open: false, title: "", message: "", action: null });

const isOwner = computed(() => detailsStore.file?.owner === userStore.user.id);
const canManage = computed(() => isOwner.value || detailsStore.details?.accessLevel === "manager");

function openGroupMembers(group) {
    groupMembersGroup.value = group;
    groupMembersOpen.value = true;
}

const tabs = ref([
    { name: t.value("files.details.tabs.details"), current: true },
    { name: t.value("files.details.tabs.activities"), current: false },
]);

const fileIcon = computed(
    () => fileTypes[detailsStore.file?.type]?.icon ?? fileTypes["not-found"].icon,
);
const fileColor = computed(
    () => fileTypes[detailsStore.file?.type]?.color ?? fileTypes["not-found"].color,
);
const fileType = computed(
    () => fileTypes[detailsStore.file?.type]?.fileType ?? fileTypes["not-found"].fileType,
);
const isFolder = computed(() => fileType.value === "folder");

const activeTabName = computed(() => tabs.value.find((t) => t.current)?.name);

function closeDetails() {
    detailsStore.open = false;
    detailsStore.mobileOpen = false;
}

function openShareDialog(user) {
    dialogStore.openEditShareDialog(detailsStore.file, user, 1);
}

function openGroupShareDialog(group) {
    dialogStore.openEditShareDialog(detailsStore.file, group, 2);
}

function runConfirm() {
    const action = confirm.value.action;

    confirm.value.open = false;
    if (action) {
        action();
    }
}

function messageFor(entry, name, inheritedFrom) {
    if (entry?.inherited && inheritedFrom) {
        return t.value("files.share_dialog.inherited_confirm", {
            folder: inheritedFrom,
        });
    }

    return t.value("files.details.remove_access_confirm", { name });
}

function unshareUser(userId) {
    const user = (detailsStore.details.sharedUsers ?? []).find((u) => u.id == userId);

    confirm.value = {
        open: true,
        title: t.value("files.details.menu.remove_access"),
        message: messageFor(user, user?.name ?? user?.email, user?.grantedBy),
        action: () => {
            shareService
                .unshareFile(detailsStore.file.id, userId)
                .then(() => {
                    detailsStore.details.sharedUsers = detailsStore.details.sharedUsers.filter(
                        (u) => u.id != userId,
                    );
                })
                .catch((error) => {
                    alertStore.showError(extractErrorMessage(error));
                });
        },
    };
}

function unshareGroupAccess(groupId) {
    const group = (detailsStore.details.sharedGroups ?? []).find((g) => g.id == groupId);

    confirm.value = {
        open: true,
        title: t.value("files.details.menu.remove_group_access"),
        message: messageFor(group, group?.name, group?.granted_by),
        action: () => {
            shareService
                .unshareGroup(detailsStore.file.id, groupId)
                .then(() => {
                    detailsStore.details.sharedGroups = detailsStore.details.sharedGroups.filter(
                        (g) => g.id != groupId,
                    );
                })
                .catch((error) => {
                    alertStore.showError(extractErrorMessage(error));
                });
        },
    };
}

function copyLink(token) {
    navigator.clipboard.writeText(`${window.location.origin}/share/${token}`);
    alertStore.showSuccess(t.value("files.share_dialog.link.copied"));
}

function accessLabel(link) {
    if (link.allowUpload && !link.allowView && !link.allowDownload) {
        return t.value("files.share_dialog.link.level_drop");
    }

    if (link.allowUpload) {
        return t.value("files.share_dialog.link.level_collaborate");
    }

    if (link.allowDownload) {
        return t.value("files.share_dialog.link.level_download");
    }

    return t.value("files.share_dialog.link.level_view");
}

function editLink(link) {
    dialogStore.openShareDialogEditLink(detailsStore.file, link);
}

function deleteLink(token) {
    shareService
        .deleteLink(token)
        .then(() => {
            detailsStore.details.sharedLinks = detailsStore.details.sharedLinks.filter(
                (link) => link.id != token,
            );
        })
        .catch((error) => {
            console.error("Failed to delete link:", error);
        });
}

function updateActiveTab(tab) {
    tabs.value.forEach((tab) => {
        tab.current = false;
    });
    tab.current = true;
}
</script>
