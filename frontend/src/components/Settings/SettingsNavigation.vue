<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="lg:inset-y-0 lg:z-50 lg:flex lg:flex-col h-full w-full">
        <div v-if="!mini" class="flex grow flex-col gap-y-5 overflow-y-auto bg-white pb-4 mt-2">
            <nav class="flex flex-1 flex-col" aria-label="Sidebar">
                <ul role="list" class="flex flex-1 flex-col gap-y-7">
                    <li>
                        <div class="text-xs font-semibold leading-6 text-gray-400">
                            {{ t("settings.navigation.user_settings") }}
                        </div>
                        <ul role="list" class="space-y-1">
                            <li
                                v-for="item in userNavigation"
                                :key="item.name"
                                class="cursor-pointer"
                            >
                                <a
                                    @click="open(item)"
                                    :class="[
                                        route.matched[2].name == item.path
                                            ? 'bg-gray-50 text-indigo-600'
                                            : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50',
                                        'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold',
                                    ]"
                                >
                                    <component
                                        :is="item.icon"
                                        :class="[
                                            route.matched[2].name == item.path
                                                ? 'text-indigo-600'
                                                : 'text-gray-400 group-hover:text-indigo-600',
                                            'h-6 w-6 shrink-0',
                                        ]"
                                        aria-hidden="true"
                                    />
                                    {{ item.name }}
                                </a>
                            </li>
                        </ul>
                    </li>
                    <li v-if="adminNavigation.length > 0">
                        <div class="text-xs font-semibold leading-6 text-gray-400">
                            {{ t("settings.navigation.admin_settings") }}
                        </div>
                        <ul role="list" class="mt-2 space-y-1">
                            <li
                                v-for="item in adminNavigation"
                                :key="item.name"
                                class="cursor-pointer"
                                @click="open(item)"
                            >
                                <a
                                    :class="[
                                        route.matched[2].name == item.path
                                            ? 'bg-gray-50 text-indigo-600'
                                            : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50',
                                        'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold cursor-pointer',
                                    ]"
                                >
                                    <component
                                        :is="item.icon"
                                        :class="[
                                            route.matched[2].name == item.path
                                                ? 'text-indigo-600'
                                                : 'text-gray-400 group-hover:text-indigo-600',
                                            'h-6 w-6 shrink-0',
                                        ]"
                                        aria-hidden="true"
                                    />
                                    <span class="truncate">{{ item.name }}</span>
                                </a>
                            </li>
                        </ul>
                    </li>
                </ul>
            </nav>
        </div>

        <nav v-else class="flex flex-col items-center gap-y-1 pt-2" aria-label="Sidebar">
            <template v-for="(group, index) in [userNavigation, adminNavigation]" :key="index">
                <div v-if="index > 0 && group.length > 0" class="my-1 h-px w-6 bg-gray-200" />
                <a
                    v-for="item in group"
                    :key="item.path"
                    :title="item.name"
                    :class="[
                        route.matched[2]?.name == item.path
                            ? 'bg-gray-50 text-indigo-600'
                            : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                        'flex size-10 cursor-pointer items-center justify-center rounded-md',
                    ]"
                    @click="open(item)"
                >
                    <component :is="item.icon" class="size-5" aria-hidden="true" />
                    <span class="sr-only">{{ item.name }}</span>
                </a>
            </template>
        </nav>
    </div>
</template>

<script setup>
import { computed } from "vue";
import { t } from "@/i18n/index";
import { useRouter, useRoute } from "vue-router";
import { useUserStore } from "@/store/user";
import { usePermissions } from "@/composables/usePermissions";
import {
    ShieldCheckIcon,
    UserIcon,
    BellAlertIcon,
    UserGroupIcon,
    TagIcon,
    LockClosedIcon,
    EnvelopeIcon,
    ChatBubbleLeftRightIcon,
    GlobeEuropeAfricaIcon,
    DocumentTextIcon,
    CircleStackIcon,
    IdentificationIcon,
    KeyIcon,
    UsersIcon,
    ShieldExclamationIcon,
    DocumentCheckIcon,
    PresentationChartBarIcon,
    ClipboardDocumentListIcon,
} from "@heroicons/vue/24/outline";

defineProps({
    mini: {
        type: Boolean,
        default: false,
    },
});

const router = useRouter();
const route = useRoute();
const userStore = useUserStore();
const { can } = usePermissions();

const isAdmin = computed(() => userStore.user?.role === "system_admin");

const userNavigation = [
    { name: t.value("settings.navigation.profile_info"), path: "profile", icon: UserIcon },
    { name: t.value("settings.navigation.password"), path: "password", icon: KeyIcon },
    {
        name: t.value("settings.navigation.notifications"),
        path: "notifications",
        icon: BellAlertIcon,
    },
    {
        name: t.value("settings.navigation.user_security"),
        path: "user-security",
        icon: ShieldCheckIcon,
    },
];

const adminNavigation = computed(() => {
    const items = [];

    if (can("create_users") || can("delete_users") || can("edit_users")) {
        items.push({ name: t.value("settings.navigation.users"), path: "users", icon: UsersIcon });
    }

    if (isAdmin.value) {
        items.push({
            name: t.value("settings.navigation.groups"),
            path: "groups",
            icon: UserGroupIcon,
        });
    }

    if (can("manage_channels")) {
        items.push({
            name: t.value("settings.navigation.channels"),
            path: "channels",
            icon: ChatBubbleLeftRightIcon,
        });
    }

    if (can("manage_projects")) {
        items.push({
            name: t.value("settings.navigation.projects"),
            path: "projects-settings",
            icon: ClipboardDocumentListIcon,
        });
    }

    if (can("manage_collimato")) {
        items.push({
            name: t.value("settings.navigation.collimato"),
            path: "collimato-settings",
            icon: PresentationChartBarIcon,
        });
    }

    if (can("manage_roles")) {
        items.push({
            name: t.value("settings.navigation.roles"),
            path: "roles",
            icon: IdentificationIcon,
        });
    }

    if (isAdmin.value) {
        items.push(
            { name: t.value("settings.navigation.smtp"), path: "smtp", icon: EnvelopeIcon },
            {
                name: t.value("settings.navigation.authorization"),
                path: "authorization",
                icon: LockClosedIcon,
            },
            {
                name: t.value("settings.navigation.security"),
                path: "security",
                icon: ShieldExclamationIcon,
            },
            { name: t.value("settings.navigation.metadata"), path: "metadata", icon: TagIcon },
            {
                name: t.value("settings.navigation.license"),
                path: "license",
                icon: DocumentCheckIcon,
            },
            {
                name: t.value("settings.navigation.language"),
                path: "language",
                icon: GlobeEuropeAfricaIcon,
            },
            {
                name: t.value("settings.navigation.office"),
                path: "office-settings",
                icon: DocumentTextIcon,
            },
            {
                name: t.value("settings.navigation.storage"),
                path: "storage-settings",
                icon: CircleStackIcon,
            },
        );
    }

    return items;
});

function open(item) {
    router.push({ name: item.path });
}
</script>

<style scoped>
.item {
    cursor: pointer;
    line-height: 1.5;
}
</style>
