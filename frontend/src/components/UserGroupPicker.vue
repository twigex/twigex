<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div class="rounded-md border border-gray-200 bg-white px-3 py-1">
            <div class="flex flex-wrap items-center gap-2">
                <span
                    v-for="user in selectedUsers"
                    :key="'u-' + user.id"
                    class="inline-flex items-center gap-1 rounded-md bg-gray-100 px-2 py-2 text-xs font-medium text-gray-700"
                >
                    <div class="h-4 w-4 shrink-0">
                        <UserAvatar :user="user" />
                    </div>
                    <span class="truncate max-w-[120px]">
                        {{ user.name }} {{ user.lastname }}
                    </span>
                    <button
                        class="ml-1 text-gray-500 hover:text-gray-700"
                        @click.stop="removeUser(user)"
                    >
                        ×
                    </button>
                </span>

                <span
                    v-for="group in selectedGroups"
                    :key="'g-' + group.id"
                    class="inline-flex items-center gap-1 rounded-md bg-indigo-50 px-2 py-2 text-xs font-medium text-indigo-700"
                >
                    <UserGroupIcon class="h-4 w-4 text-indigo-600" aria-hidden="true" />
                    <span class="truncate max-w-[120px]">
                        {{ group.name }}
                    </span>
                    <button
                        class="ml-1 text-indigo-500 hover:text-indigo-700"
                        @click.stop="removeGroup(group)"
                    >
                        ×
                    </button>
                </span>

                <input
                    class="flex-1 min-w-[18ch] bg-transparent text-sm text-gray-900 placeholder-gray-400 border-0 outline-none focus:outline-none focus:ring-0"
                    spellcheck="false"
                    :placeholder="selectedUsers.length || selectedGroups.length ? '' : placeholder"
                    :value="query"
                    @input="onInput"
                />
            </div>
        </div>

        <div class="mt-4 h-60 overflow-y-auto">
            <div class="text-xs font-semibold uppercase mb-2">
                {{ t("common.label.users") }}
            </div>

            <div v-if="filteredUsers.length === 0" class="py-6 text-center text-sm text-gray-500">
                {{ t("common.label.no_users") }}
            </div>

            <div class="space-y-1">
                <div
                    v-for="user in filteredUsers"
                    :key="user.id"
                    class="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 hover:bg-gray-100"
                    @click="toggleUser(user)"
                >
                    <span
                        class="flex h-[18px] w-[18px] items-center justify-center rounded border"
                        :class="
                            isSelected(user) ? 'bg-indigo-600 border-indigo-600' : 'border-gray-300'
                        "
                    >
                        <svg
                            v-if="isSelected(user)"
                            viewBox="0 0 20 20"
                            class="h-4 w-4 text-white"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.2 7.2a1 1 0 0 1-1.4 0L3.3 9.2a1 1 0 1 1 1.4-1.4l3.1 3.1 6.5-6.5a1 1 0 0 1 1.4 0Z"
                                clip-rule="evenodd"
                            />
                        </svg>
                    </span>

                    <div class="h-8 w-8 shrink-0">
                        <UserAvatar :user="user" />
                    </div>

                    <div class="truncate text-sm">
                        {{ user.name }} {{ user.lastname }}
                        <span v-if="user.email" class="text-gray-500"> ({{ user.email }}) </span>
                    </div>
                </div>
            </div>

            <div class="text-xs font-semibold uppercase mb-2 mt-4">
                {{ t("common.label.groups") }}
            </div>

            <div v-if="filteredGroups.length === 0" class="py-6 text-center text-sm text-gray-500">
                {{ t("common.label.no_groups") }}
            </div>

            <div class="space-y-1">
                <div
                    v-for="group in filteredGroups"
                    :key="group.id"
                    class="flex cursor-pointer items-center gap-3 rounded-md px-3 py-2 hover:bg-gray-100"
                    @click="toggleGroup(group)"
                >
                    <span
                        class="flex h-[18px] w-[18px] items-center justify-center rounded border"
                        :class="
                            isGroupSelected(group)
                                ? 'bg-indigo-600 border-indigo-600'
                                : 'border-gray-300'
                        "
                    >
                        <svg
                            v-if="isGroupSelected(group)"
                            viewBox="0 0 20 20"
                            class="h-4 w-4 text-white"
                            fill="currentColor"
                        >
                            <path
                                fill-rule="evenodd"
                                d="M16.7 5.3a1 1 0 0 1 0 1.4l-7.2 7.2a1 1 0 0 1-1.4 0L3.3 9.2a1 1 0 1 1 1.4-1.4l3.1 3.1 6.5-6.5a1 1 0 0 1 1.4 0Z"
                                clip-rule="evenodd"
                            />
                        </svg>
                    </span>

                    <span
                        class="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-100"
                    >
                        <UserGroupIcon class="h-5 w-5 text-indigo-600" aria-hidden="true" />
                    </span>

                    <div class="min-w-0 flex-1">
                        <div class="truncate text-sm font-medium text-gray-900">
                            {{ group.name }}
                        </div>
                        <div class="truncate text-xs text-gray-500">
                            {{
                                t("common.label.group_meta", {
                                    count: group.member_count ?? 0,
                                })
                            }}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import { ref, computed, onMounted, onBeforeUnmount } from "vue";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import UserAvatar from "@/components/UserAvatar.vue";
import { useGroupSearch } from "@/composables/useGroupSearch";
import { useUserSearch } from "@/composables/useUserSearch";

const props = defineProps({
    selectedUsers: { type: Array, default: () => [] },
    selectedGroups: { type: Array, default: () => [] },
    // User/group ids to hide from results (e.g. already-shared, current user).
    excludeUserIds: { type: Array, default: () => [] },
    excludeGroupIds: { type: Array, default: () => [] },
    placeholder: { type: String, default: "" },
});

const emit = defineEmits(["update:selectedUsers", "update:selectedGroups"]);

const groupSearch = useGroupSearch();
const userSearch = useUserSearch();
const query = ref("");

onMounted(() => {
    userSearch.search("");
    groupSearch.search("");
});

onBeforeUnmount(() => {
    userSearch.reset();
    groupSearch.reset();
});

const filteredUsers = computed(() =>
    userSearch.results.value.filter((u) => !props.excludeUserIds.includes(u.id)),
);

const filteredGroups = computed(() => {
    const q = query.value.toLowerCase().trim();

    // A leading "@" means the user is searching people by handle, no group
    // matches that, so hide groups entirely.
    if (q.startsWith("@")) return [];

    return groupSearch.results.value.filter((g) => !props.excludeGroupIds.includes(g.id));
});

function onInput(e) {
    query.value = e.target.value;
    userSearch.search(e.target.value);

    if (!e.target.value.trim().startsWith("@")) groupSearch.search(e.target.value);
}

function isSelected(user) {
    return props.selectedUsers.some((u) => u.id === user.id);
}

function toggleUser(user) {
    if (isSelected(user)) {
        removeUser(user);
    } else {
        emit("update:selectedUsers", [...props.selectedUsers, user]);
    }
}

function removeUser(user) {
    emit(
        "update:selectedUsers",
        props.selectedUsers.filter((u) => u.id !== user.id),
    );
}

function isGroupSelected(group) {
    return props.selectedGroups.some((g) => g.id === group.id);
}

function toggleGroup(group) {
    if (isGroupSelected(group)) {
        removeGroup(group);
    } else {
        emit("update:selectedGroups", [...props.selectedGroups, group]);
    }
}

function removeGroup(group) {
    emit(
        "update:selectedGroups",
        props.selectedGroups.filter((g) => g.id !== group.id),
    );
}
</script>
