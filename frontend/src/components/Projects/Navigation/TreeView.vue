<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ul v-if="setFolders.length">
        <li v-for="folder in setFolders" :key="folder.id" class="list-none truncate">
            <TreeTableItem
                v-if="isTableNode(folder)"
                :table="folder"
                :align-with-folders="hasFolders"
            />

            <!-- FOLDER ITEM -->
            <div
                v-else
                :data-move-target="folder.id"
                @pointerdown="pressItem($event, folder)"
                :class="[
                    'group flex items-center gap-x-2 rounded-md p-2 pl-3 text-xs leading-2 truncate data-[drop-target]:bg-indigo-50 data-[drop-target]:ring-1 data-[drop-target]:ring-inset data-[drop-target]:ring-indigo-500',
                    {
                        'font-semibold': true,
                        'cursor-pointer': true,
                        'bg-gray-50 text-indigo-600': route.params.fid == folder.id,
                        'text-gray-700 hover:text-indigo-600 hover:bg-gray-50':
                            route.params.fid != folder.id,
                    },
                ]"
                @click.stop="toggleOpen(folder.id)"
            >
                <!-- TOGGLE ICON -->
                <span @click.stop="toggleOpen(folder.id)">
                    <ChevronDownIcon v-if="isFolderOpen(folder.id)" class="h-5 w-5" />
                    <ChevronRightIcon v-else class="h-5 w-5" />
                </span>

                <!-- ICON -->
                <span>
                    <FolderOpenIcon v-if="isFolderOpen(folder.id)" class="h-5 w-5" />
                    <FolderIcon v-else class="h-5 w-5" />
                </span>

                <!-- NAME -->
                {{ folder.name }}
            </div>

            <!-- RECURSIVE CHILD COMPONENT -->
            <NestedTreeView
                v-if="!isTableNode(folder) && folder.children?.length"
                :folders="folder.children"
                :parent-id="folder.id"
                :is-open="isFolderOpen(folder.id)"
                :is-folder-open="isFolderOpen"
                @toggle-folder="toggleOpen"
            />
        </li>
    </ul>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import {
    ChevronDownIcon,
    ChevronRightIcon,
    FolderIcon,
    FolderOpenIcon,
} from "@heroicons/vue/24/outline";
import NestedTreeView from "@/components/Projects/Navigation/NestedTreeView.vue"; // Import the recursive component
import { useMoveProjectItem } from "@/composables/projects/useMoveProjectItem";
import { isTableNode } from "@/utils/projects/tree";
import TreeTableItem from "@/components/Projects/Navigation/TreeTableItem.vue";

const router = useRouter();
const route = useRoute();
const workspaceStore = useWorkspaceStore();
const { pressItem } = useMoveProjectItem();

const setFolders = computed(() => {
    return workspaceStore.getWorkspaceFolders.length
        ? workspaceStore.getWorkspaceFolders[0].children || []
        : [];
});

const hasFolders = computed(() => setFolders.value.some((node) => !isTableNode(node)));

const selectedWorkspace = computed({
    get() {
        return workspaceStore.getWorkspace;
    },
    set(value) {
        workspaceStore.setWorkspace(value);
    },
});

// Check if a folder is open using store
const isFolderOpen = (folderId) => workspaceStore.isWorkspaceOpen(folderId);

// Toggle folder open/close state
const toggleOpen = (folderId) => {
    workspaceStore.toggleWorkspace(folderId);
    router.push({
        name: "project-folder",
        params: { id: selectedWorkspace.value.id, fid: folderId },
    });
};
</script>
