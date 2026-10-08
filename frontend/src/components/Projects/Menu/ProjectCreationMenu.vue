<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <!-- Dropdown Menu -->
        <Menu as="div" class="relative inline-block text-left">
            <div v-if="route.name != 'assigned-to-me'">
                <!-- Plus Button -->
                <MenuButton
                    class="rounded-full bg-indigo-600 p-2 text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    <PlusIcon class="h-5 w-5" aria-hidden="true" />
                </MenuButton>
            </div>

            <!-- Menu Dropdown -->
            <transition
                enter-active-class="transition ease-out duration-100"
                enter-from-class="transform opacity-0 scale-95"
                enter-to-class="transform opacity-100 scale-100"
                leave-active-class="transition ease-in duration-75"
                leave-from-class="transform opacity-100 scale-100"
                leave-to-class="transform opacity-0 scale-95"
            >
                <MenuItems
                    class="absolute left-0 z-50 mt-2 w-56 origin-top-left bg-white shadow-lg ring-1 ring-black/5 focus:outline-none rounded-md"
                >
                    <div class="py-1">
                        <!-- Create New Project -->
                        <MenuItem v-slot="{ active }">
                            <div
                                @click="openTableModal"
                                :class="[
                                    active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                    'group flex items-center px-4 py-2 text-sm cursor-pointer',
                                ]"
                            >
                                <TableCellsIcon
                                    class="mr-3 h-5 w-5 text-sky-500 group-hover:text-sky-600"
                                    aria-hidden="true"
                                />

                                {{ t("projects.menu.project_creation_menu.new_table") }}
                            </div>
                        </MenuItem>

                        <!-- Create New Folder -->
                        <MenuItem v-slot="{ active }">
                            <div
                                @click="openFolderModal(route.params.id)"
                                :class="[
                                    active ? 'bg-gray-100 text-gray-900' : 'text-gray-700',
                                    'group flex items-center px-4 py-2 text-sm cursor-pointer',
                                ]"
                            >
                                <FolderIcon
                                    class="mr-3 h-5 w-5 text-gray-500 group-hover:text-gray-600"
                                    aria-hidden="true"
                                />
                                {{ t("projects.menu.project_creation_menu.new_folder") }}
                            </div>
                        </MenuItem>
                    </div>
                </MenuItems>
            </transition>
        </Menu>

        <FormDialog
            :open="tableDialog"
            :title="t('projects.menu.project_creation_menu.create_workspace_table')"
            :confirm-label="t('common.button.create')"
            @confirm="createWorkspaceTable(currentWorkspace)"
            @close="closeTableModal"
        >
            <input
                v-model="workspaceTableName"
                type="text"
                name="value"
                id="value"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.menu.project_creation_menu.placeholder_table_name')"
            />
        </FormDialog>

        <FormDialog
            :open="folderDialog"
            :title="t('projects.menu.project_creation_menu.create_folder')"
            :confirm-label="t('common.button.create')"
            @confirm="createWorkspaceFolder(currentWorkspace)"
            @close="closeFolderModal"
        >
            <input
                v-model="workspaceFolderName"
                type="text"
                name="value"
                id="value"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.menu.project_creation_menu.placeholder_folder_name')"
            />
        </FormDialog>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed, ref } from "vue";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { PlusIcon, FolderIcon, TableCellsIcon } from "@heroicons/vue/20/solid";
import { useRoute } from "vue-router";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { findTreeNode } from "@/utils/projects/tree";
import FormDialog from "@/components/FormDialog.vue";

const workspaceStore = useWorkspaceStore();
const route = useRoute();
const createSubtasks = ref(false);

const currentWorkspace = computed({
    get() {
        return workspaceStore.getCurrentWorkspace;
    },
    set(value) {
        workspaceStore.setCurrentWorkspace(value);
    },
});

const workspaceTableName = ref("");
const workspaceFolderName = ref("");

const workspaces = computed({
    get() {
        return workspaceStore.getWorkspaces || []; // Ensure it's always an array
    },
    set(value) {
        workspaceStore.setWorkspaces(value);
    },
});

function createWorkspaceTable(workspace) {
    let folder_id = route.params.fid || workspace.id;

    if (workspace.id === route.params.fid) {
        folder_id = "";
    }

    closeTableModal();

    workspaceService
        .createWorkspaceTable({
            name: workspaceTableName.value,
            id: workspace.id,
            folder_id: folder_id,
            create_subtasks: createSubtasks.value,
        })
        .then((response) => {
            const mainView = response.data.views.find((view) => view.main_view === true);
            const viewId = mainView ? mainView.id : response.data.views[0].id;

            const newTable = {
                id: response.data.id,
                name: response.data.name,
                display_name: response.data.display_name,
                type: "project",
                isFolder: false,
                path: `${workspace.id}/grid/${response.data.id}/view/${viewId}`,
                modified: new Date().toISOString(),
                visible: true,
                view_id: viewId,
            };

            let parent;
            let newFolderID = response.data.folder_id;

            if (newFolderID === undefined || newFolderID === "") {
                newFolderID = workspace.id;
            }

            parent = findTreeNode(workspaceStore.getWorkspaceFolders, newFolderID);

            if (parent) {
                if (!parent.children) parent.children = [];
                parent.children = [...parent.children, newTable];

                workspaceStore.setWorkspaceFolders([...workspaceStore.getWorkspaceFolders]);
            }

            workspace.tables = [...workspace.tables, response.data];

            if (workspace.folders && workspace.folders.length > 0) {
                const mainFolder = workspace.folders.find((folder) => folder.id === workspace.id);

                if (mainFolder) {
                    if (!mainFolder.tables) mainFolder.tables = [];
                    mainFolder.tables = [...mainFolder.tables, response.data];
                }
            }

            workspace.tables = [...workspace.tables];
            workspace.folders = workspace.folders.map((folder) =>
                folder.id === newFolderID
                    ? { ...folder, tables: [...(folder.tables || [])] }
                    : folder,
            );

            workspaces.value = [...workspaces.value];

            if (createSubtasks.value && response.data.subtask_project) {
                const subtask = response.data.subtask_project;

                const subtaskMainView = subtask.views.find((view) => view.main_view === true);
                const subtaskViewId = subtaskMainView ? subtaskMainView.id : subtask.views[0].id;

                const subtaskTable = {
                    id: subtask.id,
                    name: subtask.name,
                    display_name: subtask.display_name,
                    type: "project",
                    isFolder: false,
                    path: `${workspace.id}/grid/${subtask.id}/view/${subtaskViewId}`,
                    modified: new Date().toISOString(),
                    visible: true,
                    view_id: subtaskViewId,
                };

                const parentFolder = findTreeNode(
                    workspaceStore.getWorkspaceFolders,
                    subtask.folder_id || workspace.id,
                );

                if (parentFolder) {
                    if (!parentFolder.children) parentFolder.children = [];
                    parentFolder.children = [...parentFolder.children, subtaskTable];

                    workspaceStore.setWorkspaceFolders([...workspaceStore.getWorkspaceFolders]);
                }

                workspace.tables = [...workspace.tables, subtask];

                const mainFolder = workspace.folders.find((f) => f.id === workspace.id);

                if (mainFolder) {
                    if (!mainFolder.tables) mainFolder.tables = [];
                    mainFolder.tables = [...mainFolder.tables, subtask];
                }

                workspace.tables = [...workspace.tables];
                workspace.folders = workspace.folders.map((folder) =>
                    folder.id === (subtask.folder_id || workspace.id)
                        ? { ...folder, tables: [...(folder.tables || [])] }
                        : folder,
                );
            }
        })
        .catch((error) => {
            useAlertStore().showError(
                extractErrorMessage(
                    error,
                    t.value("projects.menu.project_creation_menu.folder_create_error_msg"),
                ),
            );
            closeFolderModal();
        });
}

function openFolderModal(workspace) {
    currentWorkspace.value = workspace;
    workspaceFolderName.value = "";
    if (workspace.id != undefined) {
        folderDialog.value = true;
    } else {
        folderDialog.value = true;
    }
}

function closeFolderModal() {
    folderDialog.value = false;
}

const folderDialog = ref(false);

const createWorkspaceFolder = () => {
    const parentId = route.params.id === route.params.fid ? "" : route.params.fid;

    workspaceService
        .createWorkspaceFolder({
            name: workspaceFolderName.value,
            id: route.params.id,
            parent_folder_id: parentId,
        })
        .then((response) => {
            closeFolderModal();

            if (!response || !response.data) {
                return;
            }

            let newFolderID = response.data.parent_folder_id || route.params.id;

            let parent = findTreeNode(workspaceStore.getWorkspaceFolders, newFolderID);

            if (!parent) {
                parent = workspaces.value.find((ws) => ws.id === route.params.id);
            }

            if (parent) {
                if (!parent.children) {
                    parent.children = [];
                }

                const newFolder = {
                    id: response.data.id,
                    name: response.data.name,
                    isFolder: true,
                    is_folder: true,
                    children: [],
                    modified: new Date().toISOString(),
                };

                parent.children.unshift(newFolder);

                let updatedWorkspace = workspaceStore.getWorkspaces.find(
                    (ws) => ws.id === route.params.id,
                );

                if (updatedWorkspace) {
                    const isParentWorkspace = parent.id === updatedWorkspace.id;

                    if (isParentWorkspace) {
                        if (!updatedWorkspace.children) {
                            updatedWorkspace.children = [];
                        }

                        const folderExists = updatedWorkspace.children.some(
                            (item) => item.id === newFolder.id,
                        );

                        if (!folderExists) {
                            updatedWorkspace.children.unshift(newFolder);
                        }
                    } else {
                        if (!updatedWorkspace.folders) {
                            updatedWorkspace.folders = [];
                        }

                        updatedWorkspace.folders.forEach((folder) => {
                            if (folder.id === parent.id) {
                                if (!folder.children) {
                                    folder.children = [];
                                } else if (!Array.isArray(folder.children)) {
                                    folder.children = [folder.children].filter(Boolean);
                                }

                                const folderExists = folder.children.some(
                                    (item) => item.id === newFolder.id,
                                );

                                if (!folderExists) {
                                    folder.children = [newFolder, ...folder.children];
                                }
                            }
                        });
                    }

                    workspaceStore.setWorkspaces([
                        ...workspaceStore.getWorkspaces.map((ws) =>
                            ws.id === updatedWorkspace.id ? { ...updatedWorkspace } : ws,
                        ),
                    ]);
                    workspaceStore.setCurrentWorkspace(updatedWorkspace);
                }
            }

            workspaceStore.setWorkspaceFolders([...workspaceStore.getWorkspaceFolders]);
        })
        .catch((error) => {
            useAlertStore().showError(
                extractErrorMessage(
                    error,
                    t.value("projects.menu.project_creation_menu.folder_create_error_msg"),
                ),
            );
            closeFolderModal();
        });
};

function closeTableModal() {
    tableDialog.value = false;
}

const tableDialog = ref(false);

function openTableModal() {
    var workspace = workspaces.value.find((ws) => ws.id === route.params.id);

    currentWorkspace.value = workspace;
    workspaceTableName.value = "";
    if (workspace.id != undefined) {
        tableDialog.value = true;
    } else {
        workspace.id = workspace.path;
        tableDialog.value = true;
    }
}
</script>
