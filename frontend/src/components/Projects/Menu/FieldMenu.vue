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
        :title="t('projects.popover_menu.delete_title_field')"
        :message="
            isOption
                ? t('projects.popover_menu.delete_message_single_field')
                : t('projects.popover_menu.delete_message_field')
        "
        :confirm-label="t('common.button.delete')"
        @confirm="deleteTarget"
        @close="confirming = false"
    />

    <RenameDialog
        v-model:name="name"
        :open="renaming"
        :title="
            isOption
                ? t('projects.popover_menu.edit_kanban_section_name')
                : t('projects.popover_menu.edit_field_name')
        "
        @save="saveName"
        @close="closeRename"
    />
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { Cog6ToothIcon, PencilIcon, TrashIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ItemMenu from "@/components/Projects/Menu/ItemMenu.vue";
import RenameDialog from "@/components/Projects/Menu/RenameDialog.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

// The menu of a grid's field, or of a board's column, which is an option of
// the field the board is grouped by.
const props = defineProps({
    visible: { type: Boolean, default: false },
    anchor: { type: Object, default: null },
    field: { type: Object, default: null },
    // An option of a single select field, shown as a board's column.
    isOption: { type: Boolean, default: false },
});

const emit = defineEmits(["close", "editFieldPanel"]);

const DEFAULT_FIELDS = new Set([
    "id",
    "created_at",
    "updated_at",
    "name",
    "status",
    "assignee",
    "description",
    "start_date",
    "due_date",
    "created_by",
]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();

const renaming = ref(false);
const confirming = ref(false);
const name = ref("");
const target = ref(null);

function allowed(action) {
    const permissions = rolesStore.userPermissions;

    if (permissions.length === 0) return true;

    return rolesStore.perTableMode
        ? rolesStore.hasTablePermission(route.params.tid, action)
        : permissions.includes("update_workspace_table") || permissions.includes(action);
}

const items = computed(() => {
    const list = [];

    if (allowed("edit_fields")) {
        list.push({ label: t.value("common.button.edit"), icon: PencilIcon, action: openRename });
    }

    if (allowed("create_fields") && !DEFAULT_FIELDS.has(props.field?.name)) {
        list.push({ label: t.value("common.button.delete"), icon: TrashIcon, action: openDelete });
    }

    if (!props.isOption) {
        list.push({
            label: t.value("projects.grid_view.edit_field"),
            icon: Cog6ToothIcon,
            action: () => {
                emit("editFieldPanel", props.field);
                emit("close");
            },
        });
    }

    return list;
});

function openRename() {
    target.value = props.field;
    name.value = props.isOption
        ? target.value.display_name || target.value.title
        : target.value.display_name || target.value.name;
    renaming.value = true;
}

function closeRename() {
    renaming.value = false;
    target.value = null;
    emit("close");
}

function openDelete() {
    target.value = props.field;
    emit("close");
    confirming.value = true;
}

function saveName() {
    if (props.isOption) renameOption();
    else renameField();
}

async function renameField() {
    const field = target.value;

    try {
        const response = await workspaceService.updateDisplayName({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
            header_name: field.name,
            display_name: name.value,
            visible: field.visible,
        });

        const header = (workspaceStore.getTableHeaders || []).find(
            (h) => h.name === response.data.updated_field,
        );

        if (header) header.display_name = response.data.display_name;
        closeRename();
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

// The board shows the new name at once; one the server refuses comes back
// when the board is read again.
function renameOption() {
    const option = target.value;

    if (!option?.color || !option?.title) {
        closeRename();

        return;
    }

    const title = (workspaceStore.getKanbanTitles || []).find((t) => t.id === option.id);

    if (title) title.display_name = name.value;
    const column = (workspaceStore.getKanbanTaskList || []).find((t) => t.id === option.id);

    if (column) column.display_name = name.value;

    workspaceService
        .updateSingleSelectHeaderName({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            field: "name",
            value: name.value,
            linked_table_id: workspaceStore.getSelectedViewData?.parent_table_id,
            task_id: option.id,
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            workspaceStore.bumpGridReloadToken();
        });

    workspaceStore.setKanbanDisplayNameUpdateKey(1);
    closeRename();
}

async function deleteTarget() {
    const field = target.value;

    try {
        if (field?.id) {
            if (props.isOption) deleteOption(field);
            else await deleteField(field);
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    } finally {
        emit("close");
        confirming.value = false;
        target.value = null;
    }
}

async function deleteField(field) {
    try {
        await workspaceService.deleteWorkspaceTableField({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            field_id: field.id,
            link_both_directions: field.link_both_directions,
            linked_id: field.linked_id,
        });

        const headers = workspaceStore.getTableHeaders || [];
        const index = headers.findIndex((h) => h.id === field.id);

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
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
}

// Its cards move to the column of tasks without an option before the
// option is deleted.
function deleteOption(option) {
    const columns = workspaceStore.getKanbanTaskList || [];
    const column = columns.find((c) => c.id === option.id);

    if (!column) return;

    const unassigned = columns.find((c) => c.id === "0");

    if (unassigned) unassigned.tasks = unassigned.tasks.concat(column.tasks || []);

    workspaceStore.setKanbanTaskList(columns.filter((c) => c.id !== option.id));
    workspaceStore.setKanbanTitles(
        (workspaceStore.getKanbanTitles || []).filter((t) => t.id !== option.id),
    );

    workspaceService
        .deleteWorkspaceTableSingleField({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            field_id: option.id,
            linked_table_id: option.linked_table_id,
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            workspaceStore.bumpGridReloadToken();
        });
}
</script>
