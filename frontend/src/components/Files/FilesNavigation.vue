<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="lg:inset-y-0 lg:z-50 lg:flex lg:flex-col h-full w-full">
        <!-- Sidebar component, swap this element with another sidebar if you like -->
        <div v-if="!mini" class="flex grow flex-col gap-y-5 overflow-y-auto bg-white">
            <nav class="flex flex-1 flex-col mt-3">
                <ul role="list" class="flex flex-1 flex-col gap-y-7">
                    <li>
                        <ul role="list" class="space-y-1">
                            <li
                                v-for="item in navigation"
                                :key="item.name"
                                class="cursor-pointer items-center"
                            >
                                <a
                                    @click="router.push(item.path)"
                                    :class="[
                                        route.name == item.path
                                            ? 'bg-gray-50 text-indigo-600'
                                            : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50',
                                        'group flex gap-x-3 rounded-md p-2 text-sm leading-6 font-semibold items-center',
                                    ]"
                                >
                                    <component
                                        :is="item.icon"
                                        :class="[
                                            route.name == item.path
                                                ? 'text-indigo-600'
                                                : 'text-gray-400 group-hover:text-indigo-600',
                                            'h-5 w-5 shrink-0',
                                        ]"
                                        aria-hidden="true"
                                    />
                                    {{ item.name }}
                                </a>
                            </li>
                        </ul>
                    </li>
                    <li>
                        <div class="text-xs font-semibold leading-6 text-gray-400">
                            {{ t("files.navigation.your_drives") }}
                        </div>
                        <ul role="list" class="mt-2 space-y-1">
                            <TreeItem
                                v-for="(item, index) in filesStore.treeData"
                                :key="index"
                                class="item"
                                :model="item"
                                :load-children="loadChildren"
                                :open="open"
                            ></TreeItem>
                        </ul>
                    </li>
                </ul>
            </nav>
        </div>

        <!-- The folder tree has no compact form: a drive opens its root, and
             the full sidebar is one click away for the tree. -->
        <nav v-else class="flex flex-col items-center gap-y-1 pt-2">
            <a
                v-for="item in navigation"
                :key="item.name"
                :title="item.name"
                :class="[
                    route.name == item.path
                        ? 'bg-gray-50 text-indigo-600'
                        : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                    'flex size-10 cursor-pointer items-center justify-center rounded-md',
                ]"
                @click="router.push(item.path)"
            >
                <component :is="item.icon" class="size-5" aria-hidden="true" />
                <span class="sr-only">{{ item.name }}</span>
            </a>

            <div class="my-1 h-px w-6 bg-gray-200" />

            <a
                v-for="drive in filesStore.treeData"
                :key="drive.id"
                :title="drive.name"
                :class="[
                    route.params.id === drive.id
                        ? 'bg-gray-50 text-indigo-600'
                        : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                    'flex size-10 cursor-pointer items-center justify-center rounded-md',
                ]"
                @click="open(drive)"
            >
                <ServerIcon class="size-5" aria-hidden="true" />
                <span class="sr-only">{{ drive.name }}</span>
            </a>
        </nav>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";

import TreeItem from "@/components/Files/TreeItem.vue";
import { onMounted } from "vue";
import { useRouter } from "vue-router";
import { useRoute } from "vue-router";
import { useFilesStore } from "@/store/files";
import filesService from "@/services/fileService";

import { UsersIcon, StarIcon, ShareIcon, TrashIcon, ServerIcon } from "@heroicons/vue/24/outline";

defineProps({
    mini: {
        type: Boolean,
        default: false,
    },
});

onMounted(() => {
    filesService.getDrive().then((res) => {
        let data = res.data;

        data.forEach((item) => {
            item.children = [];
        });

        filesStore.setTreeData(data);

        //if no parameters are set then redirect to first drive
        if (route.name == "files" && route.params.id == undefined) {
            router.push({ path: `/files/${data[0].id}` });
        }
    });
});

const router = useRouter();
const route = useRoute();
const filesStore = useFilesStore();

const navigation = [
    {
        name: t.value("files.navigation.shared_with_me"),
        path: "shared",
        href: "/files/shared",
        icon: ShareIcon,
        current: true,
    },
    {
        name: t.value("files.navigation.recents"),
        path: "recents",
        href: "/files/recents",
        icon: UsersIcon,
        current: false,
    },
    {
        name: t.value("files.navigation.favorites"),
        path: "favorites",
        href: "/files/favorites",
        icon: StarIcon,
        current: false,
    },
    {
        name: t.value("files.navigation.trash"),
        path: "deleted",
        href: "/files/deleted",
        icon: TrashIcon,
        current: false,
    },
];

function loadChildren(model) {
    filesService.getFilesFromFolder(model.id).then((res) => {
        let data = [];

        res.data.forEach((item) => {
            if (item.isFolder) {
                item.children = [];
                data.push(item);
            }
        });

        model.children = data;
    });
}

function open(file) {
    router.push({ path: `/files/${file.id}` });
}
</script>

<style scoped>
.item {
    cursor: pointer;
    line-height: 1.5;
}
</style>
