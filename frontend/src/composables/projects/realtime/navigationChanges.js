// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { findTreeNode, moveTreeNode, removeTreeNode } from "@/utils/projects/tree";

// navigationChanges applies changes to the workspace's folders and tables,
// and to the workspace itself.
export function navigationChanges(ctx) {
    const { router, workspaceStore, workspaces, realoadUI, foldersStore, workspaceTables } = ctx;

    const isFolderNode = (node) => node.isFolder || node.is_folder;
    const isTableEntry = (node) => !isFolderNode(node);

    function onDeleteFolder(fromSocket) {
        const delId = fromSocket.folder_id;

        if (!delId) return;

        const workspace = workspaces.value.find((ws) => ws.id === fromSocket.workspace_id);

        if (workspace) {
            removeTreeNode(workspace.folders, delId, isFolderNode);
            workspaces.value = [...workspaces.value];
        }

        if (removeTreeNode(foldersStore.value, delId, isFolderNode)) {
            foldersStore.value = [...foldersStore.value];
            realoadUI.value += 1;
        }
    }

    function onDeleteTable(fromSocket) {
        const delId = fromSocket.delete_id || fromSocket.table_id;

        if (!delId) return;

        const workspace = workspaces.value.find((ws) => ws.id === fromSocket.workspace_id);

        if (workspace) {
            if (Array.isArray(workspace.tables)) {
                workspace.tables = workspace.tables.filter((t) => t.id !== delId);
            }

            removeTreeNode(workspace.folders, delId, isTableEntry);
            workspaces.value = [...workspaces.value];
        }

        if (removeTreeNode(foldersStore.value, delId, isTableEntry)) {
            foldersStore.value = [...foldersStore.value];
            realoadUI.value += 1;
        }

        workspaceTables.value = workspaceTables.value.filter((t) => String(t.id) !== String(delId));
    }

    function onCreateFolder(fromSocket) {
        if (fromSocket.client_id != workspaceStore.getConnectionID) {
            const newFolderData = fromSocket.data;

            if (!newFolderData || !newFolderData.id) return;

            const newFolder = {
                id: newFolderData.id,
                name: newFolderData.name,
                isFolder: true,
                is_folder: true,
                children: [],
                modified: new Date().toISOString(),
            };

            if (foldersStore.value) {
                const parent = findTreeNode(foldersStore.value, newFolderData.parent_folder_id);

                if (parent) {
                    if (!parent.children) parent.children = [];
                    parent.children.unshift(newFolder);
                } else {
                    const workspaceIndex = foldersStore.value.findIndex(
                        (item) => item.id === fromSocket.workspace_id,
                    );

                    if (workspaceIndex !== -1) {
                        const workspace = foldersStore.value[workspaceIndex];

                        if (!workspace.children) workspace.children = [];
                        workspace.children.unshift(newFolder);
                        foldersStore.value[workspaceIndex] = {
                            ...workspace,
                        };
                    } else {
                        foldersStore.value.unshift(newFolder);
                    }
                }

                foldersStore.value = [...foldersStore.value];
            }

            if (workspaces.value) {
                const workspaceIndex = workspaces.value.findIndex(
                    (ws) => ws.id === fromSocket.workspace_id,
                );

                if (workspaceIndex !== -1) {
                    const workspace = workspaces.value[workspaceIndex];

                    if (!workspace.folders) workspace.folders = [];
                    workspace.folders.unshift(newFolder);

                    if (workspace.children) {
                        workspace.children.unshift(newFolder);
                    }

                    workspaces.value = [...workspaces.value];
                }
            }

            if (realoadUI) {
                realoadUI.value += 1;
            }
        }
    }

    function onUpdateTableName(fromSocket) {
        if (fromSocket.client_id != workspaceStore.getConnectionID) {
            const tableData = fromSocket.data;

            if (!tableData || !tableData.id) return;

            const rename = (table) => {
                if (!table) return;
                table.name = tableData.name;
                table.display_name = {
                    String: tableData.name,
                    Valid: true,
                };
            };

            const workspaceFolders = workspaceStore.getWorkspaceFolders;

            if (workspaceFolders && workspaceFolders.length > 0) {
                const newFolders = JSON.parse(JSON.stringify(workspaceFolders));

                rename(findTreeNode(newFolders, tableData.id));
                workspaceStore.setWorkspaceFolders(newFolders);
            }

            const workspaces = workspaceStore.getWorkspaces;

            if (workspaces && workspaces.length > 0) {
                const newWorkspaces = JSON.parse(JSON.stringify(workspaces));

                for (let ws of newWorkspaces) {
                    if (ws.id === fromSocket.workspace_id) {
                        rename(findTreeNode(ws.children, tableData.id));
                        if (Array.isArray(ws.tables)) {
                            rename(ws.tables.find((t) => t.id === tableData.id));
                        }

                        break;
                    }
                }

                workspaceStore.setWorkspaces(newWorkspaces);
            }

            const currentReady = workspaceStore.getWorkspacesReady;

            workspaceStore.setWorkspacesReady(0);
            setTimeout(() => {
                workspaceStore.setWorkspacesReady(currentReady + 1);
            }, 50);
        }
    }

    function onMoveItem(fromSocket) {
        if (fromSocket.client_id != workspaceStore.getConnectionID && foldersStore.value?.length) {
            moveTreeNode(
                foldersStore.value,
                fromSocket.item_id,
                fromSocket.target_folder_id || foldersStore.value[0].id,
            );
        }
    }

    function onUpdateFolderName(fromSocket) {
        if (fromSocket.client_id != workspaceStore.getConnectionID) {
            const rename = (nodes) => {
                const folder = findTreeNode(nodes, fromSocket.folder_id);

                if (folder) folder.name = fromSocket.data.name;
            };

            if (foldersStore.value) {
                rename(foldersStore.value);
                foldersStore.value = [...foldersStore.value];
            }

            const workspace = workspaces.value.find((ws) => ws.id === fromSocket.workspace_id);

            if (workspace) {
                if (workspace.folders) {
                    rename(workspace.folders);
                    workspace.folders = [...workspace.folders];
                }

                workspaces.value = [...workspaces.value];
            }

            if (realoadUI) {
                realoadUI.value += 1;
            }
        }
    }

    function onUpdateTables(fromSocket) {
        if (fromSocket.client_id != workspaceStore.getConnectionID) {
            const newTable = fromSocket.data;

            if (!newTable || !newTable.id) return;

            const mainView = newTable.views?.find((view) => view.main_view === true);
            const viewId = mainView ? mainView.id : newTable.views?.[0]?.id;

            const formattedTable = {
                id: newTable.id,
                name: newTable.name,
                display_name: newTable.display_name,
                path: `${fromSocket.workspace_id}/grid/${newTable.id}/view/${viewId}`,
                modified: new Date().toISOString(),
                visible: true,
                view_id: viewId,
                folder_id: newTable.folder_id || fromSocket.workspace_id,
                type: "project",
                isFolder: false,
                children: [],
            };

            const workspaceIndex = workspaces.value.findIndex(
                (ws) => ws.id === fromSocket.workspace_id,
            );

            if (workspaceIndex !== -1) {
                const workspace = workspaces.value[workspaceIndex];

                if (!workspace.tables) workspace.tables = [];
                workspace.tables.push(formattedTable);
                workspace.tables = [...workspace.tables];

                workspaces.value = [...workspaces.value];
            }

            if (foldersStore.value) {
                const folder = findTreeNode(foldersStore.value, formattedTable.folder_id);

                if (folder) {
                    folder.children = [...(folder.children || []), formattedTable];
                } else {
                    const workspaceRootIndex = foldersStore.value.findIndex(
                        (item) => item.id === fromSocket.workspace_id,
                    );

                    if (workspaceRootIndex !== -1) {
                        const workspaceRoot = foldersStore.value[workspaceRootIndex];

                        if (!workspaceRoot.children) workspaceRoot.children = [];
                        workspaceRoot.children.push(formattedTable);
                        workspaceRoot.children = [...workspaceRoot.children];
                    }
                }

                foldersStore.value = [...foldersStore.value];
            }

            if (workspaceTables.value) {
                workspaceTables.value.push(formattedTable);
                workspaceTables.value = [...workspaceTables.value];
            }

            if (realoadUI) {
                realoadUI.value += 1;
            }
        }
    }

    function onDeleteWorkspace(fromSocket) {
        workspaces.value = workspaces.value.filter((ws) => ws.id !== fromSocket.workspace_id);
        if (fromSocket.client_id !== workspaceStore.getConnectionID) {
            router.push({ name: "projects" });
        }
    }

    return {
        onDeleteFolder,
        onDeleteTable,
        onCreateFolder,
        onUpdateTableName,
        onMoveItem,
        onUpdateFolderName,
        onUpdateTables,
        onDeleteWorkspace,
    };
}
