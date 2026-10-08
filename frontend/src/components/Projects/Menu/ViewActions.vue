<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ConfirmDialog
        :open="confirming"
        :title="t('projects.popover_menu.delete_title_view')"
        :message="t('projects.popover_menu.delete_message_view')"
        :confirm-label="t('common.button.delete')"
        @confirm="deleteView"
        @close="confirming = false"
    />

    <RenameDialog
        v-model:name="name"
        :open="renaming"
        :title="t('projects.popover_menu.edit_view_name')"
        @save="saveName"
        @close="closeRename"
    />
</template>

<script setup>
import { ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { t } from "@/i18n/index.js";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import RenameDialog from "@/components/Projects/Menu/RenameDialog.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

// What a view's own menu in the toolbar does: rename it, delete it, and make
// it public or private.
const route = useRoute();
const router = useRouter();
const workspaceStore = useWorkspaceStore();

const renaming = ref(false);
const confirming = ref(false);
const name = ref("");
const target = ref(null);

function patchView(id, patch) {
    workspaceStore.setViewTypes(
        (workspaceStore.getViewTypes || []).map((v) => (v.id === id ? { ...v, ...patch } : v)),
    );
}

function rename(view) {
    target.value = view;
    name.value = view.name || "";
    renaming.value = true;
}

function closeRename() {
    renaming.value = false;
    target.value = null;
}

async function saveName() {
    const view = target.value;

    if (!view) return;

    try {
        await workspaceService.updateView({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: view.id,
            name: name.value,
        });

        patchView(view.id, { name: name.value });
        const selected = workspaceStore.getSelectedViewData;

        if (selected?.id === view.id) {
            workspaceStore.setSelectedViewData({ ...selected, name: name.value });
            workspaceStore.setSelectedView(name.value);
        }

        closeRename();
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

function remove(view) {
    target.value = view;
    confirming.value = true;
}

async function deleteView() {
    const view = target.value;

    try {
        if (view?.id) {
            await workspaceService.deleteWorkspaceTableView({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                view_id: view.id,
            });
            workspaceStore.setViewTypes(
                (workspaceStore.getViewTypes || []).filter((v) => v.id !== view.id),
            );

            if (route.params.fid == view.id) {
                router.push({
                    name: "grid-view",
                    params: { id: route.params.id, tid: route.params.tid },
                });
            }
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        confirming.value = false;
        target.value = null;
    }
}

async function toggleVisibility(view) {
    if (!view?.id) return;

    const isPublic = view.is_public === false;

    try {
        await workspaceService.setViewVisibility({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: view.id,
            is_public: isPublic,
        });
        patchView(view.id, { is_public: isPublic });
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

defineExpose({ rename, remove, toggleVisibility });
</script>
