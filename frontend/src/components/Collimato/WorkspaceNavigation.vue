<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="lg:inset-y-0 lg:z-50 lg:flex lg:flex-col h-full w-full">
        <div v-if="!mini" class="flex grow flex-col gap-y-5 overflow-y-auto bg-white pb-4">
            <nav class="flex flex-1 flex-col" aria-label="Sidebar">
                <ul role="list" class="flex flex-1 flex-col gap-y-7 justify-between">
                    <li>
                        <ul role="list" class="space-y-1 mt-3">
                            <div v-for="item in navigation" :key="item.path">
                                <li v-if="showNavigation(item.path)" class="cursor-pointer">
                                    <a
                                        @click="open(item)"
                                        :class="[
                                            route.name == item.path
                                                ? 'bg-gray-50 text-indigo-600'
                                                : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50',
                                            'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold',
                                        ]"
                                    >
                                        <component
                                            :is="item.icon"
                                            class="text-gray-400 group-hover:text-indigo-600 h-6 w-6 shrink-0"
                                        />
                                        {{ t(item.key) }}
                                    </a>
                                </li>
                            </div>
                        </ul>
                    </li>

                    <a
                        @click="router.push({ name: 'workspaces' })"
                        class="text-gray-700 hover:text-indigo-600 hover:bg-indigo-100 bg-indigo-50 flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold cursor-pointer"
                    >
                        <ArrowLeftIcon
                            class="text-gray-400 group-hover:text-indigo-600 h-6 w-6 shrink-0"
                        />
                        {{ t("collimato.workspace_navigation.back_to_workspaces") }}
                    </a>
                </ul>
            </nav>
        </div>

        <nav
            v-else
            class="flex flex-1 flex-col items-center justify-between pt-2"
            aria-label="Sidebar"
        >
            <div class="flex flex-col items-center gap-y-1">
                <template v-for="item in navigation" :key="item.path">
                    <a
                        v-if="showNavigation(item.path)"
                        :title="t(item.key)"
                        :class="[
                            route.name == item.path
                                ? 'bg-gray-50 text-indigo-600'
                                : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                            'flex size-10 cursor-pointer items-center justify-center rounded-md',
                        ]"
                        @click="open(item)"
                    >
                        <component :is="item.icon" class="size-5" aria-hidden="true" />
                        <span class="sr-only">{{ t(item.key) }}</span>
                    </a>
                </template>
            </div>

            <a
                :title="t('collimato.workspace_navigation.back_to_workspaces')"
                class="flex size-10 cursor-pointer items-center justify-center rounded-md bg-indigo-50 text-gray-400 hover:bg-indigo-100 hover:text-indigo-600"
                @click="router.push({ name: 'workspaces' })"
            >
                <ArrowLeftIcon class="size-5" aria-hidden="true" />
                <span class="sr-only">{{
                    t("collimato.workspace_navigation.back_to_workspaces")
                }}</span>
            </a>
        </nav>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { useRouter, useRoute } from "vue-router";
import { useCollimatoStore } from "@/store/collimato";
import { useWorkspaceAccess } from "@/composables/collimato/useWorkspaceAccess";
import {
    PresentationChartLineIcon,
    ChartPieIcon,
    DocumentIcon,
    WifiIcon,
    UserGroupIcon,
    ArrowLeftIcon,
} from "@heroicons/vue/24/outline";

defineProps({
    mini: {
        type: Boolean,
        default: false,
    },
});

const router = useRouter();
const route = useRoute();
const collimatoStore = useCollimatoStore();
const { canOpen } = useWorkspaceAccess();

const navigation = [
    {
        key: "collimato.workspace_navigation.dashboards",
        path: "dashboards",
        icon: PresentationChartLineIcon,
    },
    {
        key: "collimato.workspace_navigation.charts",
        path: "charts",
        icon: ChartPieIcon,
    },
    {
        key: "collimato.workspace_navigation.data",
        path: "data",
        icon: DocumentIcon,
    },
    {
        key: "collimato.workspace_navigation.connections",
        path: "connections",
        icon: WifiIcon,
    },
    {
        key: "collimato.workspace_navigation.collimato-users",
        path: "collimato-users",
        icon: UserGroupIcon,
    },
];

function showNavigation(path) {
    return collimatoStore.user !== null && canOpen(path, route.params.workspaceId);
}

function open(item) {
    router.push({ name: item.path });
}
</script>
