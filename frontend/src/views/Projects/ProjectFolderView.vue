<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex h-full flex-col bg-white">
        <div class="flex min-h-0 flex-1 flex-col overflow-y-auto">
            <div class="flex w-full flex-1 flex-col px-4 py-6 sm:px-6">
                <!-- Columns are at least 15rem; with few items the grid keeps
                     empty tracks rather than stretching the cards. -->
                <ul
                    v-if="paginatedItems.length"
                    role="list"
                    class="grid grid-cols-[repeat(auto-fill,minmax(15rem,1fr))] gap-4"
                >
                    <li
                        v-for="item in paginatedItems"
                        :key="item.id"
                        :class="[
                            'bg-white ring-1 ring-black/5 data-[drop-target]:bg-indigo-50 data-[drop-target]:ring-1 data-[drop-target]:ring-indigo-500',
                            dragged?.id === item.id ? 'opacity-50' : '',
                            'group relative flex cursor-pointer items-center gap-x-3 rounded-lg p-4 shadow-sm transition-shadow hover:shadow-md',
                        ]"
                        :data-move-target="item.id"
                        @click="navigateToFolder(item)"
                        @contextmenu.stop.prevent="togglePopover($event, item)"
                        @pointerdown="pressItem($event, item)"
                    >
                        <div
                            :class="[
                                item.isFolder
                                    ? 'bg-indigo-50 text-indigo-600'
                                    : 'bg-sky-50 text-sky-600',
                                'flex h-10 w-10 flex-none items-center justify-center rounded-lg',
                            ]"
                        >
                            <FolderIcon v-if="item.isFolder" class="h-6 w-6" aria-hidden="true" />
                            <TableCellsIcon v-else class="h-6 w-6" aria-hidden="true" />
                        </div>
                        <div class="min-w-0 flex-auto">
                            <p
                                class="truncate text-sm font-semibold text-gray-900"
                                :title="itemName(item)"
                            >
                                {{ itemName(item) }}
                            </p>
                            <p class="mt-0.5 truncate text-xs text-gray-500">
                                {{ itemDetail(item) }}
                            </p>
                        </div>
                        <ItemActionsMenu
                            v-if="canManage(item)"
                            :item="item"
                            :can-edit="canEditItem(item)"
                            :can-delete="canDeleteItem(item)"
                            @move="movingItem = $event"
                            @edit="itemActions?.rename($event)"
                            @delete="itemActions?.remove($event)"
                        />
                    </li>
                </ul>

                <div v-else-if="folder" class="m-auto text-center">
                    <FolderOpenIcon class="mx-auto h-12 w-12 text-gray-300" aria-hidden="true" />
                    <h3 class="mt-2 text-sm font-semibold text-gray-900">
                        {{ t("projects.folder_view.empty_title") }}
                    </h3>
                    <p class="mt-1 text-sm text-gray-500">
                        {{ t("projects.folder_view.empty_hint") }}
                    </p>
                </div>
            </div>
        </div>

        <div v-if="totalItems > itemsPerPage" class="flex-shrink-0 bg-white">
            <PaginationProjectFolder
                :currentPage="currentPage"
                :totalItems="totalItems"
                :itemsPerPage="itemsPerPage"
                @update:currentPage="updateCurrentPage"
            />
        </div>

        <NavigationItemMenu
            ref="itemActions"
            :visible="popoverVisible"
            :anchor="popoverAnchor"
            :item="activePopoverItem"
            @close="
                popoverVisible = false;
                activePopoverItem = null;
            "
            @move="movingItem = $event"
        />

        <MoveItemDialog :open="!!movingItem" :item="movingItem" @close="movingItem = null" />
    </div>
</template>

<script setup>
import { ref, shallowRef, computed, watch, nextTick } from "vue";
import { anchorFromEvent } from "@/composables/useAnchoredPopup";
import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import NavigationItemMenu from "@/components/Projects/Menu/NavigationItemMenu.vue";
import ItemActionsMenu from "@/components/Projects/Menu/ItemActionsMenu.vue";

import { t } from "@/i18n/index.js";
import { FolderIcon, FolderOpenIcon, TableCellsIcon } from "@heroicons/vue/24/outline";
import PaginationProjectFolder from "@/components/Projects/Navigation/PaginationProjectFolder.vue";
import MoveItemDialog from "@/components/Projects/Dialogs/MoveItemDialog.vue";
import { useMoveProjectItem } from "@/composables/projects/useMoveProjectItem";
import { findTreeNode } from "@/utils/projects/tree";

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();
const folder = ref(null);
const currentPage = ref(1);
const itemsPerPage = ref(20);

const movingItem = ref(null);
const { dragged, pressItem } = useMoveProjectItem();

function canEditItem(item) {
    return item.isFolder
        ? rolesStore.hasPermissionToUpdateWorkspaceFolder
        : rolesStore.hasPermissionToUpdateWorkspaceTable;
}

function canDeleteItem(item) {
    return item.isFolder
        ? rolesStore.hasPermissionToDeleteWorkspaceFolder
        : rolesStore.hasPermissionToDeleteWorkspaceTable;
}

function canManage(item) {
    return canEditItem(item) || canDeleteItem(item);
}

const itemActions = ref(null);

function itemName(item) {
    return item.display_name?.String || item.name;
}

function itemDetail(item) {
    if (!item.isFolder) return t.value("projects.folder_view.table");
    const count = item.children?.length || 0;

    return t.value("projects.folder_view.folder_items", { count });
}

const activePopoverItem = ref(null); // Tracks which item the popover is active for
const popoverVisible = ref(false); // Tracks popover visibility
const popoverAnchor = shallowRef(null);

function togglePopover(event, item) {
    if (!canManage(item)) {
        return;
    }

    if (popoverVisible.value && activePopoverItem.value === item) {
        popoverVisible.value = false;
        activePopoverItem.value = null;
    } else {
        popoverVisible.value = true;
        activePopoverItem.value = item; // Set workspace or table
        popoverAnchor.value = anchorFromEvent(event);
    }
}

const workspacesReady = computed(() => workspaceStore.getWorkspaceFolders.length > 0);

const currentWorkspace = computed({
    get() {
        return workspaceStore.getCurrentWorkspace;
    },
    set(value) {
        workspaceStore.setCurrentWorkspace(value);
    },
});

watch(
    () => workspaceStore.getWorkspacesReady,
    () => {
        refreshFolderTree({ fid: route.params.fid, id: route.params.id });
    },
);

async function refreshFolderTree({ fid, id }) {
    const isReady = workspaceStore.getWorkspacesReady;

    if (!isReady) {
        const newWorkspace = workspaceStore.getWorkspaces.find((ws) => ws.id === id);

        if (newWorkspace) {
            if (!newWorkspace.children || newWorkspace.children.length === 0) {
                const defaultProject = workspaceStore.getWorkspaceFolders.find(
                    (table) => table.workspace_id === id,
                );

                if (defaultProject) {
                    newWorkspace.children = [defaultProject];
                }
            }

            folder.value = {
                id: "root",
                name: newWorkspace.name || "New Workspace",
                isFolder: true,
                children: newWorkspace.children || [],
            };

            return;
        }

        folder.value = {
            id: "root",
            name: "New Workspace",
            isFolder: true,
            children: [],
        };

        return;
    }

    await nextTick();

    const currentWorkspace = workspaceStore.getWorkspaces.find((ws) => ws.id === id);

    if (!currentWorkspace) {
        folder.value = {
            id: "root",
            name: "New Workspace",
            isFolder: true,
            children: [],
        };

        return;
    }

    if (fid) {
        const foundFolder = findTreeNode(workspaceStore.getWorkspaceFolders, fid);

        if (foundFolder) {
            folder.value = foundFolder;
        } else {
            folder.value = null;
        }
    } else {
        folder.value = {
            id: "root",
            name: currentWorkspace.name || "Workspace Root",
            isFolder: true,
            children: currentWorkspace.children || [],
        };
    }
}

watch(
    () => [
        route.params.fid,
        route.params.id,
        workspacesReady.value,
        currentWorkspace.value,
        workspaceStore.getWorkspacesReady,
    ],
    async ([fid, id, isReady]) => {
        if (!isReady) {
            const newWorkspace = workspaceStore.getWorkspaces.find((ws) => ws.id === id);

            if (newWorkspace) {
                if (!newWorkspace.children || newWorkspace.children.length === 0) {
                    const defaultProject = workspaceStore.getWorkspaceFolders.find(
                        (table) => table.workspace_id === id,
                    );

                    if (defaultProject) {
                        newWorkspace.children = [defaultProject];
                    }
                }

                folder.value = {
                    id: "root",
                    name: newWorkspace.name || "New Workspace",
                    isFolder: true,
                    children: newWorkspace.children || [],
                };

                return;
            }

            folder.value = {
                id: "root",
                name: "New Workspace",
                isFolder: true,
                children: [],
            };

            return;
        }

        await nextTick();

        const currentWorkspace = workspaceStore.getWorkspaces.find((ws) => ws.id === id);

        if (!currentWorkspace) {
            folder.value = {
                id: "root",
                name: "New Workspace",
                isFolder: true,
                children: [],
            };

            return;
        }

        if (fid) {
            const foundFolder = findTreeNode(workspaceStore.getWorkspaceFolders, fid);

            if (foundFolder) {
                folder.value = foundFolder;
            } else {
                folder.value = null;
            }
        } else {
            folder.value = {
                id: "root",
                name: currentWorkspace.name || "Workspace Root",
                isFolder: true,
                children: currentWorkspace.children || [],
            };
        }
    },
    { deep: true, immediate: true },
);

const totalItems = computed(() => {
    return folder.value ? folder.value.children.length : 0;
});

const paginatedItems = computed(() => {
    if (!folder.value) return [];
    const start = (currentPage.value - 1) * itemsPerPage.value;

    return folder.value.children.slice(start, start + itemsPerPage.value);
});

const selectedWorkspace = computed({
    get() {
        return workspaceStore.getWorkspace;
    },
    set(value) {
        workspaceStore.setWorkspace(value);
    },
});

// Navigate to subfolder and update route
function navigateToFolder(item) {
    if (item.isFolder) {
        // Navigate to folder
        router.push({
            name: "project-folder",
            params: { id: selectedWorkspace.value.id, fid: item.id },
        });
    } else {
        workspaceStore.setSortOptions([{ field: "", direction: "asc" }]);
        // Navigate to project using the correct path
        router.push(`/projects/${selectedWorkspace.value.id}/grid/${item.id}/view/${item.view_id}`);
    }
}

// Find the folder by ID in the folder tree

watch(
    () => workspaceStore.getWorkspacesReady,
    () => {
        refreshFolderTree({ fid: route.params.fid, id: route.params.id });
    },
);

// Update the current page
const updateCurrentPage = (page) => {
    currentPage.value = page;
};
</script>
