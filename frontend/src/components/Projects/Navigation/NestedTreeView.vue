<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ul v-show="isOpen" class="pl-3">
        <li v-for="folder in folders" :key="folder.id" class="list-none truncate">
            <TreeTableItem v-if="isTableNode(folder)" :table="folder" align-with-folders />

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
                @click.stop="toggleFolder(folder.id)"
            >
                <!-- TOGGLE ICON -->
                <span @click.stop="toggleFolder(folder.id)">
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

            <!-- RECURSION -->
            <NestedTreeView
                v-if="!isTableNode(folder) && folder.children?.length"
                :folders="folder.children"
                :parent-id="folder.id"
                :is-open="isFolderOpen(folder.id)"
                :is-folder-open="isFolderOpen"
                @toggle-folder="toggleFolder"
            />
        </li>
    </ul>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useMoveProjectItem } from "@/composables/projects/useMoveProjectItem";
import { isTableNode } from "@/utils/projects/tree";
import TreeTableItem from "./TreeTableItem.vue";

import {
    ChevronDownIcon,
    ChevronRightIcon,
    FolderIcon,
    FolderOpenIcon,
} from "@heroicons/vue/24/outline";
import NestedTreeView from "./NestedTreeView.vue";

const workspaceStore = useWorkspaceStore();
const route = useRoute();
const router = useRouter();
const { pressItem } = useMoveProjectItem();

defineProps({
    folders: {
        type: Array,
        required: true,
    },
    parentId: {
        type: String,
        required: false,
    },
    isOpen: {
        type: Boolean,
        required: true,
    },
    isFolderOpen: {
        type: Function,
        required: true,
    },
});

const emits = defineEmits(["toggle-folder"]);

const selectedWorkspace = computed({
    get() {
        return workspaceStore.getWorkspace;
    },
    set(value) {
        workspaceStore.setWorkspace(value);
    },
});

// Emit toggle event to parent component
const toggleFolder = (folderId) => {
    emits("toggle-folder", folderId);
    router.push({
        name: "project-folder",
        params: { id: selectedWorkspace.value.id, fid: folderId },
    });
};
</script>
