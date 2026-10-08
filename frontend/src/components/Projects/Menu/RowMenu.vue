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
        :title="t('projects.popover_menu.delete_title_item')"
        :message="
            hasSubtasks
                ? t('projects.popover_menu.delete_message_item_with_subtasks')
                : t('projects.popover_menu.delete_message_item')
        "
        :confirm-label="t('common.button.delete')"
        @confirm="deleteTask"
        @close="confirming = false"
    />

    <RenameDialog
        v-model:name="name"
        :open="renaming"
        :title="t('projects.popover_menu.edit_task_name')"
        @save="saveName"
        @close="closeRename"
    />
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute } from "vue-router";
import { PencilIcon, TrashIcon } from "@heroicons/vue/24/solid";
import { t } from "@/i18n/index.js";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import ItemMenu from "@/components/Projects/Menu/ItemMenu.vue";
import RenameDialog from "@/components/Projects/Menu/RenameDialog.vue";
import workspaceService from "@/services/workspaceService";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { withoutTask } from "@/utils/projects/rows";

// The menu of a task, a grid row or a board card: rename it or delete it.
const props = defineProps({
    visible: { type: Boolean, default: false },
    anchor: { type: Object, default: null },
    task: { type: Object, default: null },
});

const emit = defineEmits(["close", "taskDeleted"]);

const route = useRoute();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();

const renaming = ref(false);
const confirming = ref(false);
const hasSubtasks = ref(false);
const name = ref("");
const target = ref(null);

const allowed = (action) => rolesStore.canRowAction(route.params.tid, action);

const items = computed(() => [
    {
        label: t.value("common.button.edit"),
        icon: PencilIcon,
        disabled: !allowed("update_task"),
        action: openRename,
    },
    {
        label: t.value("common.button.delete"),
        icon: TrashIcon,
        disabled: !allowed("delete_task"),
        action: openDelete,
    },
]);

function openRename() {
    target.value = props.task;
    name.value = target.value?.name || "";
    renaming.value = true;
}

function closeRename() {
    renaming.value = false;
    target.value = null;
    emit("close");
}

// A board's card does not hold its subtasks, so the server is asked whether
// the delete takes any with it.
async function openDelete() {
    target.value = props.task;
    emit("close");
    hasSubtasks.value = false;
    try {
        const { data } = await workspaceService.getSubtasks({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            task_id: target.value.id,
        });

        hasSubtasks.value = Array.isArray(data) && data.length > 0;
    } catch {
        hasSubtasks.value = false;
    }

    confirming.value = true;
}

// Every list a task is shown in: the grid's rows, the board's columns and
// the task panel.
function patchTask(id, patch) {
    workspaceStore.setTableData(
        (workspaceStore.getTableData || []).map((row) =>
            row.id === id ? { ...row, ...patch } : row,
        ),
    );
    const columns = workspaceStore.getKanbanTaskList;

    if (columns?.length) {
        workspaceStore.setKanbanTaskList(
            columns.map((list) => ({
                ...list,
                tasks: list.tasks.map((task) => (task.id === id ? { ...task, ...patch } : task)),
            })),
        );
    }

    const selected = workspaceStore.getSelectedItem;

    if (selected?.id === id) workspaceStore.setSelectedItem({ ...selected, ...patch });
}

async function saveName() {
    const task = target.value;

    if (!task?.id) {
        useAlertStore().showError(t.value("projects.popover_menu.task_not_found"));
        closeRename();

        return;
    }

    const newName = name.value.trim();

    if (!newName) {
        useAlertStore().showError(t.value("projects.popover_menu.task_name_cant_be_empty"));

        return;
    }

    try {
        await workspaceService.updateTask({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            task_id: task.id,
            field: "name",
            value: newName,
            field_id:
                (workspaceStore.getTableHeaders || []).find((header) => header.name === "name")
                    ?.id ?? null,
        });
        patchTask(task.id, { name: newName });
        closeRename();
    } catch (error) {
        target.value = null;
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.popover_menu.error_updating_task")),
        );
    }
}

async function deleteTask() {
    const task = target.value;

    try {
        if (task?.id) {
            await workspaceService.deleteWorkspaceItem({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                item_id: task.id,
            });
            workspaceStore.setTableData(withoutTask(workspaceStore.getTableData || [], task.id));
            workspaceStore.setKanbanTaskList(
                (workspaceStore.getKanbanTaskList || []).map((list) => ({
                    ...list,
                    tasks: list.tasks.filter((t) => t.id !== task.id),
                })),
            );
            emit("taskDeleted", task.id);
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
