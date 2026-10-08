// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { useRoute } from "vue-router";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useWorkspaceStore } from "@/store/workspaces";
import { extractErrorMessage } from "@/utils/errors";
import { usePointerDrag } from "@/composables/usePointerDrag";
import { canMoveInto, findTreeNode, isTableNode, moveTreeNode } from "@/utils/projects/tree";

function itemName(item) {
    return item.display_name?.String || item.name || "";
}

export function useMoveProjectItem() {
    const route = useRoute();
    const workspaceStore = useWorkspaceStore();
    const rolesStore = useWorkspaceRolesStore();
    const { dragged, startPress } = usePointerDrag();

    function roots() {
        return workspaceStore.getWorkspaceFolders;
    }

    function canMove(item) {
        if (!item) return false;

        return isTableNode(item)
            ? rolesStore.hasPermissionToUpdateWorkspaceTable
            : rolesStore.hasPermissionToUpdateWorkspaceFolder;
    }

    function canMoveTo(item, target) {
        return canMove(item) && canMoveInto(roots(), item, target);
    }

    async function moveItem(item, target) {
        if (!canMoveTo(item, target)) return false;

        try {
            await workspaceService.moveWorkspaceItem({
                workspace_id: route.params.id,
                item_id: item.id,
                item_type: isTableNode(item) ? "table" : "folder",
                target_folder_id: target.id === roots()[0]?.id ? "" : target.id,
            });
        } catch (error) {
            useAlertStore().showError(extractErrorMessage(error));

            return false;
        }

        moveTreeNode(roots(), item.id, target.id);

        return true;
    }

    function resolveTarget(id, item) {
        const node = findTreeNode(roots(), id);

        return node && canMoveTo(item, node) ? node : null;
    }

    function pressItem(event, item) {
        if (!canMove(item)) return;

        startPress(event, {
            start: () => ({
                item,
                label: itemName(item),
            }),
            resolveTarget,
            drop: moveItem,
        });
    }

    return {
        dragged,
        canMove,
        moveItem,
        pressItem,
    };
}
