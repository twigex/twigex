// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import {
    FolderIcon,
    ChartPieIcon,
    ChatBubbleLeftEllipsisIcon,
    SquaresPlusIcon,
} from "@heroicons/vue/24/outline";
import { useUserStore } from "@/store/user";
import { useSettingsStore } from "@/store/settings";
import { useChannelsStore } from "@/store/channels";
import { usePermissionsStore } from "@/store/permissions";

const SECTIONS = [
    {
        name: "files",
        displayname: "Files",
        icon: FolderIcon,
        path: "files",
    },
    {
        name: "chat",
        displayname: "Chat",
        icon: ChatBubbleLeftEllipsisIcon,
        path: "chat",
    },
    {
        name: "collimato",
        displayname: "Collimato",
        icon: ChartPieIcon,
        path: "collimato",
    },
    {
        name: "projects",
        displayname: "Projects",
        icon: SquaresPlusIcon,
        path: "projects-home",
    },
];

export const sectionGuards = {
    files: "view_files",
    chat: "view_channels",
    collimato: "view_collimato",
    projects: "view_projects",
};

export function useSections() {
    const router = useRouter();
    const route = useRoute();
    const userStore = useUserStore();
    const settingsStore = useSettingsStore();
    const channelsStore = useChannelsStore();
    const permissionsStore = usePermissionsStore();

    function show(item) {
        if (item.name === "collimato" && settingsStore.collimatoStatus != true) {
            return false;
        }

        const required = sectionGuards[item.name];

        if (required && !permissionsStore.permissions.includes(required)) {
            return false;
        }

        return true;
    }

    const sections = computed(() => SECTIONS.filter((item) => show(item)));

    const totalUnreadPosts = computed(() => {
        if (!userStore.user) {
            return 0;
        }

        return channelsStore.getTotalUnreadPosts(userStore.user.id);
    });

    function isActive(item) {
        return route.matched[1]?.name == item.name;
    }

    function open(item) {
        if (route.name === item.path) return;

        if (route.matched[1]?.name == item.path) {
            return;
        }

        router.push({ name: item.path });
    }

    return { sections, totalUnreadPosts, isActive, open };
}
