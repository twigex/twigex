<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <ItemMenu
        :visible="visible && !renaming && !confirming"
        :anchor="anchor"
        :items="items"
        :close-parent-menu="closeParentMenu"
        @close="emit('close')"
    />

    <ConfirmDialog
        :open="confirming"
        :title="t('projects.popover_menu.delete_title_saved_filter')"
        :message="t('projects.popover_menu.delete_message_saved_filter')"
        :confirm-label="t('common.button.delete')"
        @confirm="deleteFilter"
        @close="confirming = false"
    />

    <RenameDialog
        v-model:name="name"
        :open="renaming"
        :title="t('projects.popover_menu.edit_filter_name')"
        @save="saveFilter"
        @close="closeRename"
    >
        <div class="mt-4">
            <label class="mb-2 block text-sm font-medium leading-6 text-gray-900">
                {{ t("projects.filter_builder.filter_type") }}
            </label>
            <div class="space-y-2">
                <label class="flex items-center">
                    <input
                        type="radio"
                        v-model="isPrivate"
                        :value="true"
                        class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-600"
                    />
                    <span class="ml-2 text-sm text-gray-700">{{
                        t("projects.filter_builder.private_only_me")
                    }}</span>
                </label>
                <label class="flex items-center">
                    <input
                        type="radio"
                        v-model="isPrivate"
                        :value="false"
                        class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-600"
                    />
                    <span class="ml-2 text-sm text-gray-700">{{
                        t("projects.filter_builder.collaborative_all_workspace_members")
                    }}</span>
                </label>
            </div>
        </div>
    </RenameDialog>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { PencilIcon, TrashIcon, XMarkIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ItemMenu from "@/components/Projects/Menu/ItemMenu.vue";
import RenameDialog from "@/components/Projects/Menu/RenameDialog.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const props = defineProps({
    visible: { type: Boolean, default: false },
    anchor: { type: Object, default: null },
    filter: { type: Object, default: null },
    closeParentMenu: { type: Function, default: null },
});

const emit = defineEmits(["close", "removeDefault"]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();

const renaming = ref(false);
const confirming = ref(false);
const name = ref("");
const isPrivate = ref(true);
// The filter acted on; the menu's own goes when it closes, before a dialog
// is done with it.
const target = ref(null);

const items = computed(() => {
    const filter = props.filter;
    const canManage = filter?.is_private || rolesStore.hasPermissionToUpdateWorkspaceView;
    const list = [
        {
            label: t.value("common.button.edit"),
            icon: PencilIcon,
            disabled: !canManage,
            action: openRename,
        },
        {
            label: t.value("common.button.delete"),
            icon: TrashIcon,
            disabled: !canManage,
            action: openDelete,
        },
    ];

    const canSetDefault = rolesStore.hasPermissionToUpdateWorkspaceView;

    if (filter?.is_active && canSetDefault && route.name !== "detailed-task-report") {
        list.unshift({
            label: t.value("projects.filter_menu.remove_default"),
            icon: XMarkIcon,
            action: () => {
                emit("removeDefault", filter);
                emit("close");
            },
        });
    }

    return list;
});

function openRename() {
    target.value = props.filter;
    name.value = target.value.name || "";
    isPrivate.value = target.value.is_private ?? true;
    renaming.value = true;
}

function closeRename() {
    renaming.value = false;
    target.value = null;
    emit("close");
}

function openDelete() {
    target.value = props.filter;
    emit("close");
    confirming.value = true;
}

async function saveFilter() {
    const filter = target.value;

    if (!filter) return;

    try {
        await workspaceService.updateFilter({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
            filter_id: filter.filter_id,
            saved_filter_id: filter.id,
            name: name.value,
            filters: filter.filters,
            is_private: isPrivate.value,
            is_active: filter.is_active,
        });

        const filters = workspaceStore.getFilters || [];

        workspaceStore.setFilters(
            filters.map((f) =>
                f.id === filter.id ? { ...f, name: name.value, is_private: isPrivate.value } : f,
            ),
        );
        closeRename();
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

async function deleteFilter() {
    const filter = target.value;

    try {
        if (filter?.id) {
            await workspaceService.deleteSavedFilter({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                view_id: route.params.fid,
                filter_id: filter.id,
            });
            workspaceStore.setFilters(
                (workspaceStore.getFilters || []).filter((f) => f.id !== filter.id),
            );
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        emit("close");
        confirming.value = false;
        target.value = null;
    }
}
</script>
