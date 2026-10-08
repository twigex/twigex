<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="workspace" class="flex flex-none items-center gap-x-3 pr-2">
        <button
            type="button"
            class="flex items-center -space-x-1.5 rounded-full focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
            :title="t('projects.workspace_share.members', { count: members.length })"
            @click="openMembers"
        >
            <span
                v-for="member in shown"
                :key="member.user_id"
                class="inline-block h-7 w-7 rounded-full ring-2 ring-white"
            >
                <UserAvatar :user="member.user" />
            </span>
            <span
                v-if="members.length > shown.length"
                class="inline-flex h-7 w-7 items-center justify-center rounded-full bg-gray-100 text-xs font-medium text-gray-600 ring-2 ring-white"
            >
                +{{ members.length - shown.length }}
            </span>
        </button>

        <button
            v-if="canShare"
            type="button"
            class="inline-flex items-center gap-x-1.5 rounded-md bg-white px-2.5 py-1.5 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
            @click="openMembers"
        >
            <UserPlusIcon class="-ml-0.5 h-4 w-4 text-gray-400" aria-hidden="true" />
            {{ t("projects.workspace_share.share") }}
        </button>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { UserPlusIcon } from "@heroicons/vue/20/solid";
import { t } from "@/i18n/index.js";
import UserAvatar from "@/components/UserAvatar.vue";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";

const shownAvatars = 4;

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const userStore = useUserStore();

const workspace = computed(() =>
    (workspaceStore.getWorkspaces || []).find((ws) => ws.id === route.params.id),
);

const members = computed(() =>
    (workspace.value?.members || []).map((member) => ({
        user_id: member.user_id,
        user: {
            ...(member.user_info || {}),
            ...(userStore.usersMap[member.user_id] || {}),
            id: member.user_id,
        },
    })),
);

const shown = computed(() => members.value.slice(0, shownAvatars));

const canShare = computed(() =>
    (workspace.value?.user_permissions || []).includes("add_member_to_workspace"),
);

function openMembers() {
    router.push({
        name: "workspace-settings",
        params: { id: route.params.id, tab: "members" },
    });
}
</script>
