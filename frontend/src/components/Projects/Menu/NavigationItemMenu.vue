<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ItemMenu
        :visible="visible && !renaming && !confirming"
        :anchor="anchor"
        :items="items"
        @close="emit('close')"
    />

    <ConfirmDialog
        :open="confirming"
        :title="deleteTitle"
        :message="deleteMessage"
        :confirm-label="t('common.button.delete')"
        @confirm="deleteTarget"
        @close="confirming = false"
    />

    <RenameDialog
        v-model:name="name"
        :open="renaming"
        :title="renameTitle"
        @save="saveName"
        @close="closeRename"
    />
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Cog6ToothIcon, FolderArrowDownIcon, PencilIcon, TrashIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ItemMenu from "@/components/Projects/Menu/ItemMenu.vue";
import RenameDialog from "@/components/Projects/Menu/RenameDialog.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { findTreeNode, isTableNode } from "@/utils/projects/tree";

// The menu of a workspace, folder or table in the project navigation.
const props = defineProps({
    visible: { type: Boolean, default: false },
    anchor: { type: Object, default: null },
    item: { type: Object, default: null },
});

const emit = defineEmits(["close", "move"]);

const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();

const renaming = ref(false);
const confirming = ref(false);
const name = ref("");
const target = ref(null);
const deleteTitle = ref("");
const deleteMessage = ref("");

const kindOf = (item) =>
    item?.folders || item?.pre_fix ? "workspace" : item?.isFolder ? "folder" : "table";

const RENAME_TITLES = {
    workspace: "projects.popover_menu.edit_workspace_name",
    folder: "projects.popover_menu.edit_folder_name",
    table: "projects.popover_menu.edit_table_name",
};
const renameTitle = computed(() => t.value(RENAME_TITLES[kindOf(target.value)]));

const items = computed(() => {
    const item = props.item;
    const kind = kindOf(item);
    const edit = {
        label: t.value("common.button.edit"),
        icon: PencilIcon,
        action: () => rename(item),
    };
    const remove_ = {
        label: t.value("common.button.delete"),
        icon: TrashIcon,
        action: () => remove(item),
    };
    const move = {
        label: t.value("projects.move_item.move_to"),
        icon: FolderArrowDownIcon,
        action: () => {
            emit("move", item);
            emit("close");
        },
    };

    if (kind === "workspace") {
        const userPerms = item?.user_permissions || [];
        const list = [];

        if (userPerms.includes("update_workspace")) {
            list.push({
                label: t.value("projects.workspace_settings.title"),
                icon: Cog6ToothIcon,
                action: () => {
                    router.push({ name: "workspace-settings", params: { id: item.id } });
                    emit("close");
                },
            });
        }

        if (item?.can_delete) list.push(remove_);

        return list;
    }

    const canEdit =
        kind === "folder"
            ? rolesStore.hasPermissionToUpdateWorkspaceFolder
            : rolesStore.hasPermissionToUpdateWorkspaceTable;
    const canDelete =
        kind === "folder"
            ? rolesStore.hasPermissionToDeleteWorkspaceFolder
            : rolesStore.hasPermissionToDeleteWorkspaceTable;

    const list = [];

    if (canEdit) list.push(move, edit);
    if (canDelete) list.push(remove_);

    return list;
});

function rename(item) {
    target.value = item;
    const kind = kindOf(item);

    name.value =
        kind === "workspace"
            ? item.title
            : kind === "table" && item.display_name?.Valid && item.display_name?.String?.trim()
              ? item.display_name.String
              : item.name;
    renaming.value = true;
}

function closeRename() {
    renaming.value = false;
    target.value = null;
    emit("close");
}

// The navigation shows the new name at once; one the server refuses shows
// an error.
function saveName() {
    const item = target.value;
    const newName = name.value;
    const kind = kindOf(item);

    if (kind === "workspace") {
        workspaceService
            .updateWorkspace({ workspace_id: item.id, name: newName })
            .then(() => {
                const workspace = (workspaceStore.getWorkspaces || []).find(
                    (ws) => ws.id === item.id,
                );

                if (workspace) workspace.title = newName;
            })
            .catch((error) => useAlertStore().showError(extractErrorMessage(error)));
    } else if (kind === "folder") {
        workspaceService
            .updateWorkspaceFolder({
                workspace_id: route.params.id,
                folder_id: item.id,
                name: newName,
            })
            .then(() => {
                const folders = workspaceStore.getWorkspaceFolders;
                const folder = findTreeNode(folders, item.id);

                if (folder) {
                    folder.name = newName;
                    workspaceStore.setWorkspaceFolders([...folders]);
                }
            })
            .catch((error) => useAlertStore().showError(extractErrorMessage(error)));
    } else {
        workspaceService
            .updateWorkspaceTable({
                workspace_id: route.params.id,
                table_id: item.id,
                name: newName,
            })
            .then(() => {
                setItemName(item, newName);
                if (!renameTableEverywhere(item.id, newName)) {
                    useAlertStore().showInfo(
                        t.value("projects.popover_menu.name_updated_after_refresh"),
                    );
                }
            })
            .catch((error) => useAlertStore().showError(extractErrorMessage(error)));
    }

    closeRename();
}

function renameTableEverywhere(tableId, newName) {
    if (findAndRenameTable(workspaceStore.getWorkspaceFolders, tableId, newName)) return true;
    if (findAndRenameTable(workspaceStore.getCurrentWorkspace, tableId, newName)) return true;

    const workspace = (workspaceStore.getWorkspaces || []).find((ws) => ws.id === route.params.id);

    return findAndRenameTable(workspace, tableId, newName);
}

function findAndRenameTable(structure, tableId, newName) {
    if (!structure) return false;

    if (Array.isArray(structure)) {
        return structure.some((item) => findAndRenameTable(item, tableId, newName));
    }

    if (structure.id === tableId) {
        setItemName(structure, newName);

        return true;
    }

    return ["tables", "children", "folders", "projects"].some(
        (prop) =>
            Array.isArray(structure[prop]) && findAndRenameTable(structure[prop], tableId, newName),
    );
}

function setItemName(item, newName) {
    if (!item) return;

    if (item.name !== undefined) item.name = newName;
    if (item.title !== undefined) item.title = newName;

    if (item.display_name && typeof item.display_name === "object") {
        if ("String" in item.display_name) {
            item.display_name.String = newName;
        } else if ("string" in item.display_name) {
            item.display_name.string = newName;
        }
    } else {
        item.display_name = { String: newName, Valid: true };
    }
}

function countTables(node) {
    return (node?.children || []).reduce(
        (sum, child) => sum + (isTableNode(child) ? 1 : countTables(child)),
        0,
    );
}

async function remove(item) {
    target.value = item;
    const kind = kindOf(item);

    if (kind === "workspace") {
        deleteTitle.value = t.value("projects.popover_menu.delete_title_workspace");
        deleteMessage.value = t.value("projects.popover_menu.delete_message_workspace");
    } else if (kind === "folder") {
        const tables = countTables(item);

        deleteTitle.value = t.value("projects.popover_menu.delete_title_folder");
        deleteMessage.value = tables
            ? t.value("projects.popover_menu.delete_message_folder_with_tables", { count: tables })
            : t.value("projects.popover_menu.delete_message_folder");
    } else {
        deleteTitle.value = t.value("projects.popover_menu.delete_title_table");
        deleteMessage.value = t.value("projects.popover_menu.delete_message_table");
    }

    emit("close");

    if (kind !== "workspace") {
        const warning = await linkingTablesWarning(
            kind,
            item.workspace_id || route.params.id,
            item.id,
        );

        if (warning) deleteMessage.value = `${deleteMessage.value} ${warning}`;
    }

    confirming.value = true;
}

// Tables elsewhere whose links into this one go with it.
async function linkingTablesWarning(kind, workspaceId, id) {
    try {
        const { data } = await workspaceService.getLinkingTables({
            workspace_id: workspaceId,
            table_id: kind === "table" ? id : undefined,
            folder_id: kind === "folder" ? id : undefined,
        });

        if (!Array.isArray(data) || !data.length) return "";

        return t.value(`projects.popover_menu.delete_links_warning_${kind}`, {
            count: data.length,
            names: data.map((table) => table.name).join(", "),
        });
    } catch {
        return "";
    }
}

async function deleteTarget() {
    const item = target.value;

    try {
        if (item?.id) {
            const kind = kindOf(item);

            if (kind === "workspace") await deleteWorkspace(item);
            else if (kind === "folder") await deleteFolder(item);
            else deleteTable(item);
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        emit("close");
        confirming.value = false;
        target.value = null;
    }
}

async function deleteWorkspace(item) {
    await workspaceService.deleteWorkspace(item.id);

    const workspaces = workspaceStore.getWorkspaces || [];
    const index = workspaces.findIndex((ws) => ws.id === item.id);

    if (index !== -1) workspaces.splice(index, 1);
}

async function deleteFolder(item) {
    await workspaceService.deleteWorkspaceFolder({
        folder_id: item.id,
        workspace_id: route.params.id,
    });

    const without = (folders) =>
        folders.filter((folder) => {
            if (folder.id === item.id) return false;
            if (Array.isArray(folder.children)) folder.children = without(folder.children);

            return true;
        });

    workspaceStore.setWorkspaceFolders(without(workspaceStore.getWorkspaceFolders || []));
    workspaceStore.setWorkspacesReady(workspaceStore.getWorkspacesReady + 1);
}

function deleteTable(item) {
    workspaceService
        .deleteWorkspaceTable({
            workspace_id: item.workspace_id || route.params.id,
            table_id: item.id,
        })
        .then((response) => {
            const workspace = (workspaceStore.getWorkspaces || []).find(
                (ws) => ws.id === route.params.id,
            );

            [workspace, workspaceStore.getWorkspace, workspaceStore.getCurrentWorkspace].forEach(
                (target) => {
                    if (target) removeTableFromWorkspace(target, item.id);
                },
            );
            removeTableFromFolders(workspaceStore.getWorkspaceFolders, item.id);
            workspaceStore.setWorkspacesReady(workspaceStore.getWorkspacesReady + 1);

            // The fields other tables had linking into it went with it.
            for (const { table_id, field_ids } of Array.isArray(response.data)
                ? response.data
                : []) {
                if (route.params.tid !== table_id) continue;
                for (const fieldId of field_ids) removeShownField(fieldId);
            }
        })
        .catch((error) => useAlertStore().showError(extractErrorMessage(error)));
}

function removeShownField(fieldId) {
    const headers = workspaceStore.getTableHeaders || [];
    const index = headers.findIndex((header) => header.id === fieldId);

    if (index === -1) return;

    const removed = headers[index].name;

    headers.splice(index, 1);
    workspaceStore.getColumnsWidths?.splice(index, 1);
    workspaceStore.setTableData(
        (workspaceStore.getTableData || []).map((row) => {
            const { [removed]: _, ...rest } = row;

            return rest;
        }),
    );
}

function removeTableFromWorkspace(workspace, tableId) {
    if (Array.isArray(workspace.tables)) {
        const index = workspace.tables.findIndex((t) => t.id === tableId);

        if (index !== -1) {
            workspace.tables.splice(index, 1);
            workspace.tables = [...workspace.tables];
        }
    }

    if (Array.isArray(workspace.folders)) {
        for (const folder of workspace.folders) {
            if (removeTableFromFolder(folder, tableId)) {
                workspace.folders = [...workspace.folders];
                break;
            }
        }
    }
}

function removeTableFromFolder(folder, tableId) {
    if (Array.isArray(folder.tables)) {
        const index = folder.tables.findIndex((t) => t.id === tableId);

        if (index !== -1) {
            folder.tables.splice(index, 1);
            folder.tables = [...folder.tables];

            return true;
        }
    }

    return (
        Array.isArray(folder.children) &&
        folder.children.some((child) => removeTableFromFolder(child, tableId))
    );
}

function removeTableFromFolders(folders, tableId) {
    if (!Array.isArray(folders)) return;

    for (let i = folders.length - 1; i >= 0; i--) {
        const item = folders[i];

        if (item.id === tableId && item.type === "project") {
            folders.splice(i, 1);
            continue;
        }

        removeTableFromFolders(item.children, tableId);
    }
}

defineExpose({ rename, remove });
</script>
