<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="bg-white p-1 flex min-w-0 flex-1 items-center justify-between">
        <div class="flex flex-row gap-x-2 items-center">
            <!-- Project Creation Menu -->
            <ProjectCreationMenu />

            <!-- Divider: Full Height -->
            <div v-if="route.name != 'assigned-to-me'" class="w-px bg-gray-300 self-stretch"></div>

            <!-- Breadcrumb Navigation -->
            <nav class="flex pl-2" aria-label="Breadcrumb">
                <ol role="list" class="flex items-center">
                    <!-- Show all breadcrumbs if 3 or fewer -->
                    <template v-if="breadcrumbs.length <= 3">
                        <li
                            v-for="(page, index) in breadcrumbs"
                            :key="page.id"
                            class="flex items-center"
                        >
                            <ChevronRightIcon
                                v-if="index > 0"
                                class="size-5 shrink-0 text-gray-400 mx-1"
                                aria-hidden="true"
                            />
                            <span
                                :class="[crumbClass, 'text-sm font-medium cursor-pointer']"
                                :data-move-target="page.id"
                                @click="navigateTo(page.id, page.name)"
                            >
                                {{ page.name }}
                            </span>
                        </li>
                    </template>

                    <!-- If breadcrumbs are long, collapse middle ones -->
                    <template v-else>
                        <!-- First breadcrumb -->
                        <li class="flex items-center">
                            <span
                                :class="[crumbClass, 'text-sm font-medium cursor-pointer']"
                                @click="navigateTo(breadcrumbs[0].id, breadcrumbs[0].name)"
                                :data-move-target="breadcrumbs[0].id"
                            >
                                {{ breadcrumbs[0].name }}
                            </span>
                            <ChevronRightIcon
                                class="size-5 shrink-0 text-gray-400 mx-1"
                                aria-hidden="true"
                            />
                        </li>

                        <!-- Dropdown for middle folders -->
                        <Menu as="div" class="relative inline-block text-left">
                            <MenuButton
                                class="inline-flex w-full justify-center gap-x-1.5 rounded-md bg-white px-3 py-1 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50"
                            >
                                ...
                            </MenuButton>

                            <transition
                                enter-active-class="transition ease-out duration-100"
                                enter-from-class="transform opacity-0 scale-95"
                                enter-to-class="transform opacity-100 scale-100"
                                leave-active-class="transition ease-in duration-75"
                                leave-from-class="transform opacity-100 scale-100"
                                leave-to-class="transform opacity-0 scale-95"
                            >
                                <MenuItems
                                    class="absolute left-0 z-50 mt-2 w-56 origin-top-left rounded-md bg-white shadow-lg ring-1 ring-black/5 focus:outline-none"
                                >
                                    <div
                                        v-for="page in breadcrumbs.slice(1, breadcrumbs.length - 1)"
                                        :key="page.id"
                                        class="py-1"
                                    >
                                        <MenuItem v-slot="{ active }">
                                            <span
                                                :class="[
                                                    active
                                                        ? 'bg-gray-100 text-gray-900 outline-none'
                                                        : 'text-gray-700',
                                                    'block px-4 py-2 text-sm cursor-pointer',
                                                ]"
                                                @click="navigateTo(page.id, page.name)"
                                            >
                                                {{ page.name }}
                                            </span>
                                        </MenuItem>
                                    </div>
                                </MenuItems>
                            </transition>
                        </Menu>

                        <!-- Last breadcrumb -->
                        <li class="flex items-center">
                            <ChevronRightIcon
                                class="size-5 shrink-0 text-gray-400 mx-1"
                                aria-hidden="true"
                            />
                            <span
                                :class="[crumbClass, 'text-sm font-medium cursor-pointer']"
                                @click="
                                    navigateTo(
                                        breadcrumbs[breadcrumbs.length - 1].id,
                                        breadcrumbs[breadcrumbs.length - 1].name,
                                    )
                                "
                                :data-move-target="breadcrumbs[breadcrumbs.length - 1].id"
                            >
                                {{ breadcrumbs[breadcrumbs.length - 1].name }}
                            </span>
                        </li>
                    </template>
                </ol>
            </nav>
        </div>

        <WorkspaceShareButton v-if="route.name === 'project-folder'" />
    </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import { ChevronRightIcon } from "@heroicons/vue/20/solid";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import ProjectCreationMenu from "../Menu/ProjectCreationMenu.vue";
import WorkspaceShareButton from "../Members/WorkspaceShareButton.vue";

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const crumbClass =
    "text-gray-500 hover:text-gray-700 data-[drop-target]:rounded-sm data-[drop-target]:text-indigo-600 data-[drop-target]:ring-1 data-[drop-target]:ring-indigo-500 data-[drop-target]:ring-offset-2";

const workspaces = computed({
    get() {
        return workspaceStore.getWorkspaces || []; // Ensure it's always an array
    },
    set(value) {
        workspaceStore.setWorkspaces(value);
    },
});

const breadcrumbs = computed(() => {
    const folderId = route.params.fid;
    const workspaceId = route.params.id; // Get the current workspace ID

    if (!workspaces.value || workspaces.value.length === 0) return [];

    const currentWorkspace = workspaces.value.find((ws) => ws.id === workspaceId);

    if (!currentWorkspace) return [];

    // Get the full breadcrumb path
    const path = findBreadcrumbPath(workspaceStore.getWorkspaceFolders, folderId) || [];

    const firstPathItem = path.length > 0 ? path[0].name : null;
    const isWorkspaceAlreadyInPath = firstPathItem === currentWorkspace.title;

    return isWorkspaceAlreadyInPath
        ? path.filter((page) => page && page.name) // If already included, return only path
        : [
              { name: currentWorkspace.title, id: workspaceId },
              ...path.filter((page) => page && page.name),
          ];
});

// Find the breadcrumb path from root to the current folder/project
function findBreadcrumbPath(folders, targetId, path = []) {
    for (const folder of folders) {
        if (folder.id === targetId) return [...path, { name: folder.name, id: folder.id }];

        if (folder.children) {
            const childPath = findBreadcrumbPath(folder.children, targetId, [
                ...path,
                { name: folder.name, id: folder.id },
            ]);

            if (childPath.length) return childPath;
        }
    }

    return [];
}

const selectedWorkspace = computed({
    get() {
        return workspaceStore.getWorkspace;
    },
    set(value) {
        workspaceStore.setWorkspace(value);
    },
});

// Navigate when clicking breadcrumb (folder name only)
function navigateTo(folderId) {
    if (folderId) {
        router.push({
            name: "project-folder",
            params: { id: selectedWorkspace.value.id, fid: folderId },
        });
    } else {
        router.push({
            name: "project-folder",
            params: { id: selectedWorkspace.value.id, fid: selectedWorkspace.value.id },
        });
    }
}
</script>
