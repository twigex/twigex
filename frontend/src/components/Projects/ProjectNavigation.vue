<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="lg:inset-y-0 lg:z-50 lg:flex lg:flex-col h-full w-full">
        <div v-if="!mini" class="flex grow flex-col gap-y-5 overflow-y-auto bg-white relative">
            <nav class="flex flex-1 flex-col mt-3">
                <ul role="list" class="flex flex-1 flex-col gap-y-2">
                    <li>
                        <div class="text-xs font-semibold leading-6 text-gray-400 mb-2 px-3">
                            {{
                                userStore.user.name +
                                " " +
                                t("projects.project_navigation.workspaces")
                            }}
                        </div>

                        <ul v-if="workspaces.length >= 1" role="list" class="space-y-1">
                            <li v-for="item in workspaceOptions" :key="item.name">
                                <a
                                    @click="goToPath(item.path)"
                                    :class="[
                                        route.name == item.path
                                            ? 'bg-gray-50 text-indigo-600'
                                            : 'text-gray-700 hover:text-indigo-600 hover:bg-gray-50',
                                        'group flex items-center gap-x-3 rounded-md px-3 py-2 text-sm leading-6 font-semibold',
                                    ]"
                                    style="cursor: pointer"
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

                        <ul role="list" class="space-y-1 mt-2">
                            <li v-for="workspace in workspaces" :key="workspace.id">
                                <div
                                    @click="selectWorkspace(workspace)"
                                    :class="[
                                        workspace.open
                                            ? 'bg-gray-50 text-indigo-600'
                                            : 'text-gray-700 hover:bg-gray-50 hover:text-indigo-600',
                                        // keeps justify-between for options button
                                        'group flex items-center justify-between gap-x-3 rounded-md px-3 py-2 text-sm leading-6 font-semibold cursor-pointer',
                                    ]"
                                >
                                    <div
                                        class="flex items-center gap-x-3 overflow-hidden flex-1 min-w-0"
                                    >
                                        <LetterAvatar
                                            :id="workspace.id"
                                            :name="workspace.title"
                                            class="size-6 rounded-md text-[0.625rem]"
                                        />
                                        <span
                                            class="workspace-title truncate text-sm leading-6 font-semibold flex-1 min-w-0"
                                            :class="
                                                workspace.open
                                                    ? 'text-indigo-600'
                                                    : 'text-gray-700 group-hover:text-indigo-600'
                                            "
                                        >
                                            {{ workspace.title }}
                                        </span>
                                    </div>

                                    <button
                                        v-if="
                                            workspace.can_delete ||
                                            (workspace.user_permissions || []).includes(
                                                'update_workspace',
                                            )
                                        "
                                        @click.stop="togglePopover($event, workspace)"
                                        class="opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-gray-600"
                                    >
                                        <EllipsisHorizontalIcon
                                            class="h-4 w-4"
                                            aria-hidden="true"
                                        />
                                    </button>
                                </div>

                                <div
                                    v-if="selectedWorkspace?.id === workspace.id && workspace.open"
                                    class="mt-1 pl-8"
                                >
                                    <TreeView :key="selectedWorkspace.id" />
                                </div>
                            </li>
                        </ul>
                    </li>

                    <li>
                        <button
                            type="button"
                            class="text-sm font-semibold text-gray-700 flex items-center gap-x-2 rounded-md w-full hover:bg-gray-100 px-3 py-2"
                            @click="openModal"
                        >
                            <PlusIcon class="h-5 w-5" aria-hidden="true" />
                            <span>{{ t("projects.projecs_view.button.create_workspace") }}</span>
                        </button>
                    </li>
                </ul>
            </nav>
        </div>

        <nav
            v-else
            class="flex flex-col items-center gap-y-1 pt-2"
            :aria-label="t('projects.project_navigation.workspaces')"
        >
            <a
                v-for="item in workspaceOptions"
                :key="item.path"
                :title="item.name"
                :class="[
                    route.name == item.path
                        ? 'bg-gray-50 text-indigo-600'
                        : 'text-gray-400 hover:bg-gray-50 hover:text-indigo-600',
                    'flex size-10 cursor-pointer items-center justify-center rounded-md',
                ]"
                @click="goToPath(item.path)"
            >
                <component :is="item.icon" class="size-5" aria-hidden="true" />
                <span class="sr-only">{{ item.name }}</span>
            </a>

            <div class="my-1 h-px w-6 bg-gray-200" />

            <button
                v-for="workspace in workspaces"
                :key="workspace.id"
                type="button"
                :title="workspace.title"
                :class="[
                    workspace.open ? 'bg-gray-50' : 'hover:bg-gray-50',
                    'flex size-10 items-center justify-center rounded-md',
                ]"
                @click="selectWorkspace(workspace)"
            >
                <LetterAvatar
                    :id="workspace.id"
                    :name="workspace.title"
                    class="size-7 rounded-md text-xs"
                />
                <span class="sr-only">{{ workspace.title }}</span>
            </button>

            <button
                type="button"
                :title="t('projects.projecs_view.button.create_workspace')"
                class="flex size-10 items-center justify-center rounded-md text-gray-400 hover:bg-gray-50 hover:text-indigo-600"
                @click="openModal"
            >
                <PlusIcon class="size-5" aria-hidden="true" />
                <span class="sr-only">{{
                    t("projects.projecs_view.button.create_workspace")
                }}</span>
            </button>
        </nav>

        <FormDialog
            :open="workspaceCreateDialogOpen"
            :title="t('projects.projecs_view.button.create_workspace')"
            :confirm-label="t('common.button.create')"
            :loading="disableButton"
            @confirm="createWorkspace"
            @close="closeModal"
        >
            <input
                v-model="workspaceName"
                type="text"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.assigned_to_me.workspace_name')"
            />
            <textarea
                v-model="workspaceDescription"
                rows="3"
                class="mt-3 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.project_navigation.workspace_description_placeholder')"
            />
        </FormDialog>

        <teleport to="body">
            <NavigationItemMenu
                :visible="popoverVisible"
                :anchor="popoverAnchor"
                :item="activePopoverItem"
                @close="
                    popoverVisible = false;
                    activePopoverItem = null;
                "
            />
        </teleport>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import NavigationItemMenu from "./Menu/NavigationItemMenu.vue";
import { ref, shallowRef, onMounted, onBeforeUnmount, computed, watch } from "vue";
import { onReconnect } from "@/js/websocket";
import { anchorFromEvent } from "@/composables/useAnchoredPopup";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";
import TreeView from "./Navigation/TreeView.vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

import FormDialog from "@/components/FormDialog.vue";

import { useRouter } from "vue-router";
import { useRoute } from "vue-router";
import userService from "@/services/userService";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";

import {
    HomeIcon,
    PlusIcon,
    ClipboardDocumentCheckIcon,
    ChartBarIcon,
    EllipsisHorizontalIcon,
} from "@heroicons/vue/24/outline";
import LetterAvatar from "@/components/LetterAvatar.vue";

defineProps({
    mini: {
        type: Boolean,
        default: false,
    },
});

const activePopoverItem = ref(null);
const popoverVisible = ref(false);
const popoverAnchor = shallowRef(null);

const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();

const selectedWorkspace = computed({
    get() {
        return workspaceStore.getWorkspace;
    },
    set(value) {
        workspaceStore.setWorkspace(value);
    },
});

function togglePopover(event, item) {
    if (popoverVisible.value && activePopoverItem.value === item) {
        popoverVisible.value = false;
        activePopoverItem.value = null;
    } else {
        popoverVisible.value = true;
        activePopoverItem.value = item;
        popoverAnchor.value = anchorFromEvent(event);
    }
}

const realoadUI = computed({
    get() {
        return workspaceStore.getWorkspacesReady;
    },
    set(value) {
        workspaceStore.setWorkspacesReady(value);
    },
});

const route = useRoute();
const router = useRouter();

const selectWorkspace = async (workspace) => {
    workspaceStore.setSavedFlatFilters([
        { field: "", operator: "is", value: "", operatorBetween: "AND" },
    ]);

    workspaceStore.setSortOptions([{ field: "", direction: "asc" }]);

    workspaceStore.setSavedGroups([]);
    workspaceStore.setFilterActive(false);

    try {
        const [res, meRes] = await Promise.all([
            workspaceStore.fetchWorkspaceNav(workspace.id),
            workspaceService.me(workspace.id).catch(() => null),
        ]);
        const fullWorkspace = res.data;

        if (meRes?.data) {
            rolesStore.setUser(meRes.data);
        }

        selectedWorkspace.value = { ...fullWorkspace };

        workspaceStore.workspaces = workspaceStore.workspaces.map((ws) =>
            ws.id === workspace.id
                ? {
                      ...ws,
                      ...fullWorkspace,
                      open: true,
                      user_permissions: workspace.user_permissions || [],
                  }
                : ws,
        );

        workspaceStore.workspaces.forEach((ws) => {
            if (ws.id !== workspace.id) ws.open = false;
        });

        workspaceStore.setWorkspaces([...workspaceStore.workspaces]);

        loadDummyFolders(selectedWorkspace.value);

        router.push({
            name: "project-folder",
            params: { id: workspace.id, fid: workspace.id }, // Root folder
        });

        realoadUI.value += 1;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const transformFolder = (folder) => {
    return {
        id: folder.id,
        name: folder.name,
        isFolder: folder.is_folder,
        children: [
            ...(folder.children ? folder.children.map(transformFolder) : []),
            ...(folder.tables
                ? folder.tables
                      .filter((table) => !table.linked && !table.single_select)
                      .map(transformTable)
                : []),
        ],
    };
};

const transformTable = (table) => {
    let viewId = null;

    if (Array.isArray(table.options) && table.options.length > 0) {
        viewId = table.options[0]?.id || null;
    } else if (table.options && typeof table.options === "object") {
        viewId = table.options.id || null;
    }

    return {
        id: table.id,
        name: table.name,
        display_name: table.display_name,
        type: "project",
        view_id: viewId,
    };
};

const transformWorkspace = (workspace) => {
    return {
        id: workspace.id,
        name: workspace.title,
        isFolder: workspace.is_folder,
        children: [
            ...(workspace.folders ? workspace.folders.map(transformFolder) : []),
            ...(workspace.tables ? workspace.tables.map(transformTable) : []),
        ],
    };
};

const loadDummyFolders = (workspace) => {
    selectedWorkspaceFolders.value = [];

    const transformedWorkspace = transformWorkspace(workspace);

    const transformedFolders = transformedWorkspace.children;

    const storedWorkspace = workspaceStore.getWorkspaceFolders.find((w) => w.id === workspace.id);

    if (storedWorkspace) {
        selectedWorkspaceFolders.value = [...storedWorkspace.children]; // Force reactivity
    } else {
        workspaceStore.setWorkspaceFolders([...transformedFolders]);
        selectedWorkspaceFolders.value = [...transformedFolders]; // Ensure Vue updates
    }
};

const selectedWorkspaceFolders = computed({
    get() {
        return workspaceStore.getWorkspaceFolders || [];
    },
    set(value) {
        workspaceStore.setWorkspaceFolders(value);
    },
});

// Folders and tables made, moved or deleted while the socket was down are
// not replayed, so the open workspace's tree is read again.
const stopReconnect = onReconnect(async () => {
    const id = selectedWorkspace.value?.id;

    if (!id) return;

    try {
        const res = await workspaceStore.fetchWorkspaceNav(id);

        if (selectedWorkspace.value?.id !== id) return;

        selectedWorkspace.value = { ...res.data };
        workspaceStore.setWorkspaces(
            workspaceStore.workspaces.map((ws) => (ws.id === id ? { ...ws, ...res.data } : ws)),
        );
        workspaceStore.setWorkspaceFolders(transformWorkspace(res.data).children);
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
});

onBeforeUnmount(stopReconnect);

onMounted(async () => {
    try {
        const meResponse = await userService.me();

        userStore.setUser(meResponse.data);

        const res = await workspaceService.getWorkspaces();

        workspaceStore.setWorkspaces(res.data);

        workspaces.value = res.data.map((workspace) => ({
            ...workspace,
            open: false,
        }));

        if (route.name === "project-folder" && route.params.id) {
            const workspace = workspaces.value.find((ws) => ws.id === route.params.id);

            if (workspace) {
                await selectWorkspace(workspace);
            }
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
});

const workspaces = computed({
    get() {
        return workspaceStore.getWorkspaces || [];
    },
    set(value) {
        workspaceStore.setWorkspaces(value);
    },
});

const workspaceCreateDialogOpen = computed({
    get() {
        return workspaceStore.getWorkspaceCreateDialogOpen;
    },
    set(value) {
        workspaceStore.setWorkspaceCreateDialogOpen(value);
    },
});

function closeModal() {
    workspaceCreateDialogOpen.value = false;
}

function openModal() {
    workspaceCreateDialogOpen.value = true;
}

const workspaceName = ref("");
const workspaceDescription = ref("");

watch(workspaceCreateDialogOpen, (open) => {
    if (open) {
        workspaceName.value = "";
        workspaceDescription.value = "";
    }
});

const workspaceOptions = computed(() => {
    const options = [
        {
            name: t.value("projects.project_navigation.home"),
            path: "projects-home",
            icon: HomeIcon,
            current: false,
        },
        {
            name: t.value("projects.project_navigation.assigned_to_me"),
            path: "assigned-to-me",
            icon: ClipboardDocumentCheckIcon,
            current: false,
        },
    ];

    if (userStore.user.role === "system_admin") {
        options.splice(1, 0, {
            name: t.value("projects.project_navigation.detailed_task_report"),
            path: "detailed-task-report",
            icon: ChartBarIcon,
            current: false,
        });
    }

    return options;
});

const goToPath = (path) => {
    if (path === "projects-home" || path === "assigned-to-me" || path === "detailed-task-report") {
        workspaceStore.workspaces.forEach((ws) => {
            ws.open = false;
        });

        // Force Vue to update the store
        workspaceStore.setWorkspaces([...workspaceStore.workspaces]);
    }

    router.push({ name: path });
};

watch(
    () => route.params.id,
    (newWorkspaceId) => {
        if (route.name === "project-folder" && newWorkspaceId) {
            // Skip if this workspace is already selected, prevents double fetch
            // when selectWorkspace() itself pushes to project-folder
            if (selectedWorkspace.value?.id === newWorkspaceId) return;
            if (workspaces.value.length > 0) {
                const workspace = workspaces.value.find((ws) => ws.id === newWorkspaceId);

                if (workspace) {
                    selectWorkspace(workspace);
                    loadDummyFolders(workspace);
                }
            }
        }
    },
    { immediate: true },
);

const disableButton = ref(false);

const foldersStore = computed({
    get() {
        return workspaceStore.getWorkspaceFolders;
    },
    set(value) {
        workspaceStore.setWorkspaceFolders(value);
    },
});

function createWorkspace() {
    disableButton.value = true;

    workspaceService
        .createWorkspace({
            name: workspaceName.value,
            description: workspaceDescription.value.trim(),
        })
        .then((response) => {
            closeModal();
            const newWorkspace = response.data;

            let mappedProjects = newWorkspace.tables
                ? newWorkspace.tables
                      .filter((table) => !(table.single_select || table.linked))
                      .map((table) => ({
                          id: table.id,
                          name: table.name,
                          display_name: table.display_name,
                          path: `projects/${newWorkspace.id}/grid/${table.id}`,
                          type: "project",
                      }))
                : [];

            const filterFolder = (folder) => ({
                id: folder.id,
                name: folder.name,
                isFolder: folder.is_folder,
                children: [
                    ...(folder.children ? folder.children.map(filterFolder) : []),
                    ...(folder.tables
                        ? folder.tables
                              .filter((table) => !(table.single_select || table.linked))
                              .map((table) => ({
                                  id: table.id,
                                  name: table.name,
                                  display_name: table.display_name,
                                  type: "project",
                              }))
                        : []),
                ],
            });

            let filteredFolders = newWorkspace.folders
                ? newWorkspace.folders.map(filterFolder)
                : [];

            const workspaceObject = {
                id: newWorkspace.id,
                title: newWorkspace.title,
                description: newWorkspace.description,
                start_date: newWorkspace.start_date,
                end_date: newWorkspace.end_date,
                pre_fix: newWorkspace.pre_fix,
                open: false,
                children: mappedProjects,
                created_at: newWorkspace.created_at,
                updated_at: newWorkspace.updated_at,
                user_permissions: newWorkspace.user_permissions || [],
                user_roles: newWorkspace.user_roles || [],
                can_delete: newWorkspace.can_delete,
                members: newWorkspace.members || [],
                folders: filteredFolders,
            };

            workspaces.value = [...workspaces.value, workspaceObject];
            workspaceStore.setWorkspaces(workspaces.value);

            const foldersStoreItem = {
                id: newWorkspace.id,
                name: newWorkspace.title || newWorkspace.name,
                isFolder: false,
                children: [...mappedProjects, ...filteredFolders],
            };

            if (!foldersStore.value) {
                foldersStore.value = [];
            }

            foldersStore.value = [foldersStoreItem, ...foldersStore.value];

            hasWorkspaces.value = true;
            disableButton.value = false;
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            disableButton.value = false;
            closeModal();
        });
}

const hasWorkspaces = computed({
    get() {
        return workspaceStore.getHasWorkspaces;
    },
    set(value) {
        workspaceStore.setHasWorkspaces(value);
    },
});
</script>

<style scoped>
.popover {
    position: absolute;
    z-index: 10;
}

.workspace-title {
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
}
</style>
