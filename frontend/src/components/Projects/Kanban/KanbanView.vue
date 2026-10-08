<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="kanban-container">
        <div
            v-if="kanbanLoading"
            class="absolute inset-0 z-50 flex items-center justify-center bg-white"
        >
            <div class="flex flex-col items-center gap-y-3">
                <BaseSpinner size="lg" class="text-indigo-600" :label="t('common.label.loading')" />
                <p class="text-sm text-gray-500">{{ t("common.label.loading") }}</p>
            </div>
        </div>
        <div class="kanban-wrapper" style="display: inline-flex; align-items: flex-start">
            <draggable
                :list="titles"
                item-key="id"
                handle=".column-title"
                group="kanban-columns"
                :disabled="!rolesStore.hasPermissionToUpdateWorkspaceView"
                ghost-class="kanban-column-ghost"
                :force-fallback="true"
                :fallback-on-body="true"
                class="kanban"
                style="margin-top: 10px; margin-bottom: 3px"
                @change="updateOrder(titles)"
            >
                <template #item="{ element: title, index }">
                    <div
                        class="column"
                        :style="{
                            width: parseInt(title.width) + 'px',
                            '--column-color': title.color || '#e5e7eb',
                        }"
                    >
                        <div
                            :style="{
                                height: '4px',
                                borderRadius: '9px 9px 0 0',
                                backgroundColor: title.color || '#e5e7eb',
                                margin: '-4px -1px 0 -1px',
                            }"
                        ></div>
                        <div class="title-container group">
                            <template v-if="editableTitle.index === index">
                                <input
                                    v-model="editableTitle.value"
                                    @blur="saveTitle(index)"
                                    @keyup.enter="saveTitle(index)"
                                    @keyup.esc="resetTitleEdit"
                                    class="block w-full min-w-0 rounded-md border-0 px-2 py-1 text-base font-medium text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                />
                            </template>
                            <template v-else>
                                <div class="title-wrapper-container">
                                    <div class="title-wrapper">
                                        <h2
                                            class="column-title flex items-center text-1xl font-medium tracking-tight text-gray-900 ellipsis-title"
                                            :title="title.title"
                                            @dblclick="
                                                !(
                                                    singleSelectValue === 'status' &&
                                                    ['Completed', 'Cancelled'].includes(title.title)
                                                ) && editTitle(index, title.title)
                                            "
                                        >
                                            <div v-if="title.display_name == ''">
                                                {{
                                                    singleSelectValue === "status"
                                                        ? getStatusDisplayName({
                                                              name: title.title,
                                                          }) || title.title
                                                        : title.title
                                                }}
                                            </div>
                                            <div v-else>
                                                {{
                                                    singleSelectValue === "status"
                                                        ? getStatusDisplayName({
                                                              name: title.display_name,
                                                          }) || title.display_name
                                                        : title.display_name
                                                }}
                                            </div>
                                        </h2>
                                        <span
                                            v-if="
                                                getColumnById(title.id)?.totalCount > KANBAN_PAGE &&
                                                getColumnById(title.id).tasks.length <
                                                    getColumnById(title.id).totalCount
                                            "
                                            class="ml-2 text-xs text-gray-400"
                                            :title="`${getColumnById(title.id).totalCount} tasks total`"
                                        >
                                            {{ getColumnById(title.id).tasks.length }} /
                                            {{ getColumnById(title.id).totalCount }}
                                        </span>
                                    </div>
                                </div>
                            </template>

                            <div class="flex items-center space-x-2" style="margin-top: 8px">
                                <EllipsisHorizontalIcon
                                    v-if="
                                        title.title != 'Unassigned' &&
                                        !(
                                            singleSelectValue === 'status' &&
                                            ['Completed', 'Cancelled'].includes(title.title)
                                        )
                                    "
                                    class="h-4 w-4 text-gray-700 opacity-0 group-hover:opacity-100 transition-opacity duration-150 hover:text-gray-900"
                                    aria-hidden="true"
                                    @click="toggleItemMenuPopover($event, title, 'singleField')"
                                    style="margin-bottom: 6px"
                                />

                                <button
                                    class="resize-button"
                                    @mousedown="startResize(index)"
                                    style="margin-bottom: 6px; margin-right: 2px"
                                >
                                    <svg
                                        xmlns="http://www.w3.org/2000/svg"
                                        fill="none"
                                        viewBox="0 0 25 25"
                                        stroke-width="2"
                                        stroke="currentColor"
                                        class="w-4 h-4"
                                        style="cursor: ew-resize"
                                    >
                                        <path
                                            stroke-linecap="round"
                                            stroke-linejoin="round"
                                            d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"
                                        />
                                    </svg>
                                </button>
                            </div>
                        </div>

                        <div class="scrollable-column" style="margin-bottom: 0">
                            <div
                                v-for="item in taskLists.filter(
                                    (list) => list.title === title.title,
                                )"
                                :key="item.id"
                                class="task-list"
                            >
                                <div class="item-wrapper" style="min-height: 90px">
                                    <draggable
                                        :list="item.tasks"
                                        item-key="id"
                                        group="kanban"
                                        :disabled="!canMoveCards"
                                        ghost-class="kanban-task-ghost"
                                        :force-fallback="true"
                                        :fallback-on-body="true"
                                        style="min-height: 600px"
                                        @start="onDragStart"
                                        @change="(event) => onKanbanItemChange(event, item)"
                                    >
                                        <template #item="{ element }">
                                            <KanbanCard
                                                v-if="
                                                    element.name &&
                                                    !Number(element.deleted_at) &&
                                                    isTaskVisible(element)
                                                "
                                                :task="element"
                                                :fields="cardFields"
                                                :shown="shownFields"
                                                :subtasks="subtaskMap[element.id]?.length || 0"
                                                @open="showFullView"
                                                @assign="openAssignDialog"
                                                @menu="openCardMenu"
                                                @go-to-table="goToTable"
                                            />
                                        </template>
                                    </draggable>
                                </div>
                            </div>
                            <button
                                v-if="
                                    getColumnById(title.id)?.next &&
                                    getColumnById(title.id).tasks.length <
                                        getColumnById(title.id).totalCount
                                "
                                class="w-full mt-2 py-1 text-xs text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded"
                                @click="loadMoreTasks(title.id)"
                            >
                                {{ t("projects.kanban_view.load_more") }}
                                {{
                                    Math.min(
                                        KANBAN_PAGE,
                                        getColumnById(title.id).totalCount -
                                            getColumnById(title.id).tasks.length,
                                    )
                                }}
                                ({{
                                    getColumnById(title.id).totalCount -
                                    getColumnById(title.id).tasks.length
                                }}
                                {{ t("projects.kanban_view.remaining") }})
                            </button>
                        </div>
                        <div style="margin-top: 6px; margin-bottom: 10px" class="mx-3">
                            <button
                                @click="openTaskDialog(title)"
                                type="button"
                                class="inline-flex justify-center items-center gap-x-1.5 rounded-md bg-gray-100 px-2.5 py-1.5 text-sm font-semibold text-gray-600 shadow-sm hover:bg-gray-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gray-500 w-full"
                            >
                                {{ t("projects.kanban_view.add_task") }}
                                <PlusIcon class="h-5 w-5 text-gray-900" aria-hidden="true" />
                            </button>
                        </div>
                    </div>
                </template>
            </draggable>

            <button
                @click="openSectionDialog"
                style="margin: 11px"
                type="button"
                class="inline-flex justify-center items-center gap-x-1.5 rounded-md bg-gray-200 px-2.5 py-1.5 text-sm font-semibold text-gray-700 shadow-sm hover:bg-gray-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-gray-500"
            >
                {{ t("projects.kanban_view.add_section") }}
                <PlusIcon class="h-5 w-5 text-gray-900" aria-hidden="true" />
            </button>
        </div>
        <FormDialog
            :open="sectionDialog"
            :title="t('projects.section_name')"
            :confirm-label="t('common.button.create')"
            :confirm-disabled="!sectionName.trim()"
            @confirm="addSection"
            @close="closeSectionModal"
        >
            <input
                v-model="sectionName"
                type="text"
                name="value"
                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                :placeholder="t('projects.kanban_view.section_name_placeholder')"
            />
        </FormDialog>
        <ShareWorkspaceDialog
            :open="assignDialogOpen"
            :selectedWorkspace="{ id: route.params.id }"
            :selectedTask="assignCard || {}"
            :assignFieldName="assignField"
            :selectedAssigneTable="{ id: route.params.tid }"
            :selectedUserID="assignCard?.[assignField] || null"
            @close="assignDialogOpen = false"
            @update:assignee="onAssigned"
        />
        <RowMenu
            :visible="itemPopoverVisible && parentComp === 'item'"
            :anchor="itemPopoverAnchor"
            :task="activePopoverItem"
            @close="
                itemPopoverVisible = false;
                activePopoverItem = null;
            "
        />
        <FieldMenu
            :visible="itemPopoverVisible && parentComp === 'singleField'"
            :anchor="itemPopoverAnchor"
            :field="activePopoverItem"
            is-option
            @close="
                itemPopoverVisible = false;
                activePopoverItem = null;
            "
        />
    </div>
</template>

<script setup>
import { ref, onBeforeUnmount, watch, computed, reactive, nextTick } from "vue";
import { useItemMenu } from "@/composables/projects/useItemMenu";
import { t } from "@/i18n/index.js";
import BaseSpinner from "@/components/BaseSpinner.vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import draggable from "vuedraggable";
import { useRoute } from "vue-router";
import { useGoToTable } from "@/composables/projects/useGoToTable";
import { useTaskCompletion } from "@/composables/projects/useTaskCompletion";
import { useWorkspaceStore } from "@/store/workspaces";
import { optionLookup, withOptions } from "@/utils/projects/rows";
import { KANBAN_PAGE, useKanbanBoard } from "@/composables/projects/useKanbanBoard";
import workspaceService from "@/services/workspaceService";
import { PlusIcon, EllipsisHorizontalIcon } from "@heroicons/vue/20/solid";
import KanbanCard from "@/components/Projects/Kanban/KanbanCard.vue";
import RowMenu from "@/components/Projects/Menu/RowMenu.vue";
import FieldMenu from "@/components/Projects/Menu/FieldMenu.vue";
import { useUserStore } from "@/store/user";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import FormDialog from "@/components/FormDialog.vue";
import ShareWorkspaceDialog from "@/components/Projects/Members/ShareWorkspaceDialog.vue";
import { getStatusDisplayName } from "@/utils/projects/cells";

const userStore = useUserStore();
const { goToTable } = useGoToTable();
const taskCompletion = useTaskCompletion();
const route = useRoute();
const workspaceStore = useWorkspaceStore();
const rolesStore = useWorkspaceRolesStore();
const canMoveCards = computed(() => rolesStore.canRowAction(route.params.tid, "update_task"));
const { kanbanLoading, singleSelectValue, subtaskMap, loadMoreTasks } = useKanbanBoard();

const sectionName = ref("");
const sectionDialog = ref(false);

function openSectionDialog() {
    sectionName.value = "";
    sectionDialog.value = true;
}

function closeSectionModal() {
    sectionDialog.value = false;
}

const {
    visible: itemPopoverVisible,
    anchor: itemPopoverAnchor,
    item: activePopoverItem,
    context: parentComp,
    toggle: toggleItemMenuPopover,
} = useItemMenu({
    onOpen: (item) => {
        item.linked_table_id = selectedView.value.parent_table_id;
    },
});

const taskLists = computed({
    get() {
        return workspaceStore.getKanbanTaskList || [];
    },
    set(value) {
        workspaceStore.setKanbanTaskList(value);
    },
});

const assignDialogOpen = ref(false);
const assignCard = ref(null);
const assignField = ref("assignee");

function openAssignDialog(card, fieldName = "assignee") {
    assignCard.value = card;
    assignField.value = fieldName;
    assignDialogOpen.value = true;
}

function onAssigned({ task_id, field_name, user_id }) {
    for (const list of taskLists.value) {
        const card = list.tasks?.find((task) => task.id === task_id);

        if (card) card[field_name] = user_id;
    }

    patchOpenTask(task_id, { [field_name]: user_id });
}

// The task panel shows a copy of the card taken when it was opened, so a
// change made on the board is also made to that copy.
function patchOpenTask(taskId, patch) {
    const open = workspaceStore.getSelectedItem;

    if (open && String(open.id) === String(taskId)) {
        workspaceStore.setSelectedItem({ ...open, ...patch });
    }
}

const titles = computed({
    get() {
        return workspaceStore.getKanbanTitles || [];
    },
    set(value) {
        workspaceStore.setKanbanTitles(value);
    },
});

const kanbanDisplayNameUpdateKey = computed({
    get() {
        return workspaceStore.getKanbanDisplayNameUpdateKey;
    },
    set(value) {
        workspaceStore.setKanbanDisplayNameUpdateKey(value);
    },
});

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

let resizing = false;
let initialX = 0;
let initialWidth = 0;
let resizeIndex = -1;

function openTaskDialog(column) {
    workspaceStore.openNewTaskDialog({
        section: singleSelectValue.value,
        singleSelect: String(column.id),
        columnName: column.display_name || column.title,
    });
}

// Saves the board's columns: their order, names, colours, widths and
// visibility, and the fields its cards show. Where cards sit is saved card
// by card with moveCard; the server keeps it through this save.
function saveKanbanOrder() {
    const fieldsVisible = tableHeaders.value.map((header) => ({
        name: header.name,
        visible: header.visible,
    }));

    const transformed = {
        fields: taskLists.value.map((list) => ({
            color: list.color,
            id: list.id,
            name: list.title,
            display_name: list.display_name,
            width: String(list.width),
            visible: list.visible,
        })),
        section: singleSelectValue.value,
        fieldsVisible,
    };

    workspaceService
        .updateView({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: selectedView.value.id,
            order: JSON.stringify(transformed),
            name: selectedView.value.name,
        })
        .then(() => {
            selectedView.value.order = transformed;
        })
        .catch(showFailureAndReload);
}

// The board already shows a change it is saving, so one the server refuses
// is undone by reading the board again.
function showFailureAndReload(error) {
    useAlertStore().showError(extractErrorMessage(error));
    workspaceStore.bumpGridReloadToken();
}

// Puts a task made in the New task dialog in its column, below the cards
// the column has loaded, and keeps it there.
function placeCreatedCard(task) {
    const sectionField = singleSelectValue.value;

    if (!task || !sectionField) return;

    const value = task[sectionField];
    const columnId = String((value && typeof value === "object" ? value.id : value) || "0");
    const list = taskLists.value.find((l) => String(l.id) === columnId);

    if (!list || list.tasks.some((t) => String(t.id) === String(task.id))) return;

    const card = { ...task };

    for (const header of tableHeaders.value) {
        if (header.visible && !(header.name in card)) card[header.name] = "";
    }

    // The board's tables, loaded with it, hold the options a card shows.
    const enriched = withOptions(card, optionLookup(tableHeaders.value, workspaceTables.value));

    if (!enriched[sectionField] || typeof enriched[sectionField] !== "object") {
        const column = _titleMap.value.get(list.id);

        enriched[sectionField] = {
            id: list.id,
            name: (column?.display_name || column?.title || "").trim(),
            color: column?.color || null,
        };
    }

    list.tasks.push(enriched);
    list.totalCount = (list.totalCount || 0) + 1;
    moveCard(list, list.tasks.length - 1).catch((error) => {
        useAlertStore().showError(extractErrorMessage(error));
    });
}

watch(
    () => workspaceStore.createdTask,
    (task) => placeCreatedCard(task),
);

const startResize = (index) => {
    resizing = true;
    initialX = event.clientX;
    initialWidth = parseInt(titles.value[index].width);
    resizeIndex = index;
    window.addEventListener("mousemove", resize);
    window.addEventListener("mouseup", stopResize);
};

const resize = (event) => {
    if (resizing) {
        const deltaX = event.clientX - initialX;
        const newWidth = Math.max(135, initialWidth + deltaX); // Enforce minimum width of 135px

        // Create a new object with updated width to trigger Vue's reactivity
        const updatedTitle = {
            ...titles.value[resizeIndex],
            width: newWidth.toString(),
        };

        // Only the header row is updated while dragging: the board takes its
        // widths from the titles, and rebuilding the columns on every mouse
        // move would render every card again.
        titles.value.splice(resizeIndex, 1, updatedTitle);
    }
};

const stopResize = () => {
    if (resizing) {
        resizing = false;
        window.removeEventListener("mousemove", resize);
        window.removeEventListener("mouseup", stopResize);

        const widths = new Map(titles.value.map((title) => [title.id, title.width]));

        taskLists.value = taskLists.value.map((list) =>
            widths.has(list.id) && widths.get(list.id) !== list.width
                ? { ...list, width: widths.get(list.id) }
                : list,
        );
        saveKanbanOrder();
    }
};

onBeforeUnmount(() => {
    window.removeEventListener("mousemove", resize);
    window.removeEventListener("mouseup", stopResize);
});

// A card draws its name, dates, assignee and status in their own places.
const cardFields = computed(() =>
    tableHeaders.value.filter(
        (header) =>
            header.visible &&
            !["name", "start_date", "due_date", "status", "assignee"].includes(header.name) &&
            header.header_usage !== "default_name",
    ),
);

const shownFields = computed(
    () => new Set(tableHeaders.value.filter((h) => h.visible).map((h) => h.name)),
);

function openCardMenu(event, task) {
    toggleItemMenuPopover(event, task, "item");
}

function addSection() {
    workspaceService
        .addFieldValue({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            name: sectionName.value,
            field: singleSelectValue.value,
        })
        .then((response) => {
            const newSection = {
                id: response.data.id,
                title: response.data.name,
                color: response.data.color,
                display_name: response.data.display_name,
                width: response.data.width,
                tasks: [],
            };

            taskLists.value.push(newSection);
            titles.value.push(newSection);
            saveKanbanOrder();
            closeSectionModal();
            sectionName.value = "";
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

const tableHeaders = computed({
    get() {
        return workspaceStore.getTableHeaders;
    },
    set(value) {
        workspaceStore.setTableHeaders(value);
    },
});

const updateOrder = (titles) => {
    const newTaskListsOrder = [];

    titles.forEach((title) => {
        const correspondingTaskList = taskLists.value.find((taskList) => taskList.id === title.id);

        if (correspondingTaskList) {
            newTaskListsOrder.push(correspondingTaskList);
        }
    });
    taskLists.value = newTaskListsOrder;

    saveKanbanOrder();
};

const editableTitle = reactive({
    index: null,
    value: "",
});

function editTitle(index, title) {
    editableTitle.index = index;
    editableTitle.value = title;
}

const showFullTask = computed({
    get() {
        return workspaceStore.getFullTask;
    },
    set(value) {
        workspaceStore.setFullTask(value);
    },
});

const selectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const setFullViewHeaders = computed({
    get() {
        return workspaceStore.getFullViewHeaders;
    },
    set(value) {
        workspaceStore.setFullViewHeaders(value);
    },
});

const setTableID = computed({
    get() {
        return workspaceStore.getTableID;
    },
    set(value) {
        workspaceStore.setTableID(value);
    },
});

function showFullView(item) {
    setFullViewHeaders.value = tableHeaders.value;
    setTableID.value = route.params.tid;

    // Clone item so we don't mutate original
    const normalizedItem = { ...item };

    // Normalize TINYINT bool fields to true/false
    for (const header of tableHeaders.value) {
        if (header.header_type === "TINYINT" && header.header_usage === "bool") {
            const value = normalizedItem[header.name];

            normalizedItem[header.name] = value == 1 || value === "1" ? true : false;
        }
    }

    selectedItem.value = normalizedItem;
    showFullTask.value = true;
}

const _columnMap = computed(() => new Map(taskLists.value.map((l) => [l.id, l])));
const getColumnById = (id) => _columnMap.value.get(id);

const _titleMap = computed(() => new Map(titles.value.map((t) => [t.id, t])));

function saveTitle(index) {
    titles.value[index].display_name = editableTitle.value;
    taskLists.value[index].display_name = editableTitle.value;

    var payload = {
        workspace_id: route.params.id,
        table_id: route.params.tid,
        field: "name",
        value: editableTitle.value,
        linked_table_id: selectedView.value.parent_table_id,
        task_id: titles.value[index].id,
    };

    workspaceService
        .updateSingleSelectHeaderName(payload)
        .then(() => {
            saveKanbanOrder();
        })
        .catch(showFailureAndReload);

    resetTitleEdit();
}

function resetTitleEdit() {
    editableTitle.index = null;
    editableTitle.value = "";
}

// Saves where one card now sits: in the column, after the card above it,
// or first. The column list already holds the card at index.
function moveCard(list, index) {
    const card = list.tasks[index];

    if (!card) return Promise.resolve();

    return workspaceService.moveKanbanCard({
        workspace_id: route.params.id,
        table_id: route.params.tid,
        view_id: selectedView.value.id,
        task_id: card.id,
        column: list.id,
        after_id: index > 0 ? list.tasks[index - 1]?.id || "" : "",
    });
}

async function onKanbanItemChange({ added, moved, removed }, item) {
    if (removed) {
        item.totalCount = Math.max(0, (item.totalCount || 0) - 1);
    }

    if (added) {
        item.totalCount = (item.totalCount || 0) + 1;

        // Cancelling puts the dragged card back by reading the board again.
        const completion =
            singleSelectValue.value === "status"
                ? await taskCompletion.beforeStatusChange({
                      workspaceId: route.params.id,
                      tableId: route.params.tid,
                      taskId: added.element.id,
                      statusId: item.id,
                  })
                : undefined;

        if (completion === null) {
            workspaceStore.bumpGridReloadToken();

            return;
        }

        workspaceService
            .updateTask({
                workspace_id: route.params.id,
                table_id: route.params.tid,
                task_id: added.element.id,
                field: singleSelectValue.value,
                value: item.id,
            })
            .then(() => {
                const task = added.element;
                const sectionField = singleSelectValue.value || "single_select";
                const isKanbanByStatus = sectionField === "status";

                if (isKanbanByStatus) {
                    const statusColumn = titles.value.find((col) => col.id === item.id);

                    task.status = {
                        id: item.id,
                        name: statusColumn?.display_name || statusColumn?.title || "Unknown",
                        color: statusColumn?.color || null,
                    };
                } else {
                    const col = _titleMap.value.get(item.id);

                    task[sectionField] = {
                        id: item.id,
                        name: (col?.display_name || col?.title || "").trim(),
                        color: col?.color || null,
                    };
                }

                patchOpenTask(task.id, { [sectionField]: task[sectionField] });
                completion?.afterSave();

                return moveCard(item, item.tasks.indexOf(task));
            })
            .catch(showFailureAndReload);
    } else if (moved) {
        moveCard(item, moved.newIndex).catch(showFailureAndReload);
    }
}

// Only the card being dragged is measured, when the drag starts.
const onDragStart = (e) => {
    const h = e.item?.offsetHeight || null;

    if (!h) return;
    e.item.style.setProperty("--drag-h", h + "px");
    nextTick(() => {
        // Ghost element is created by SortableJS after this tick, not before.
        document.querySelectorAll(".kanban-task-ghost").forEach((el) => {
            el.style.height = h + "px";
            el.style.minHeight = h + "px";
        });
    });
};

watch(kanbanDisplayNameUpdateKey, (newValue) => {
    if (newValue) {
        saveKanbanOrder();
    }
});

const workspaceTables = computed({
    get() {
        return workspaceStore.getWorkspaceTables;
    },
    set(value) {
        workspaceStore.setWorkspaceTables(value);
    },
});

function isTaskVisible(task) {
    const hasShowAssignedOnlyPermission = rolesStore.hasPermissionToShowAssignedTasksOnly;

    if (!hasShowAssignedOnlyPermission) {
        return true;
    }

    if (!task.name || Number(task.deleted_at)) return false;

    const assignee =
        task.assignee && typeof task.assignee === "object" ? task.assignee.id : task.assignee;

    return !!assignee && assignee === userStore.user?.id;
}
</script>

<style scoped>
.kanban {
    display: inline-block;
}

.kanban-container {
    position: relative;
    white-space: nowrap;
    overflow-x: auto;
    height: 100%;
    background: #fff;
}

.column {
    position: relative;
    border: 1px solid #e5e7eb;
    border-top-left-radius: 0px;
    border-top-right-radius: 0px;
    border-bottom-left-radius: 10px;
    border-bottom-right-radius: 10px;
    margin: 0 0;
    margin-left: 10px;
    transition:
        margin 0.3s ease,
        border-color 0.2s ease;
    background-color: #ffffff;
    display: inline-block;
    vertical-align: top;
    white-space: normal;
}

.column:hover {
    border-top-color: transparent;
    border-right-color: var(--column-color);
    border-bottom-color: var(--column-color);
    border-left-color: var(--column-color);
}

.column::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: -1px;
    right: -1px;
    height: 60%;
    border-left: 1px solid transparent;
    border-right: 1px solid transparent;
    border-bottom: 1px solid transparent;
    border-bottom-left-radius: 10px;
    border-bottom-right-radius: 10px;
    pointer-events: none;
    opacity: 0;
    transition: opacity 0.2s ease;
}

.column:hover::after {
    opacity: 1;
    border-left-color: color-mix(in srgb, var(--column-color) 30%, transparent);
    border-right-color: color-mix(in srgb, var(--column-color) 30%, transparent);
    border-bottom-color: color-mix(in srgb, var(--column-color) 30%, transparent);
}

.title-container {
    padding: 10px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    min-width: 0;
}

.ellipsis-title {
    display: block;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    flex-shrink: 1;
    min-width: 0;
    width: 100%;
    max-width: 100%;
}

.column-title,
.column-title {
    cursor: grab;
    user-select: none;
}

.kanban-column-ghost {
    opacity: 0.35;
}

.task-list {
    margin-bottom: 20px;
}

.task {
    padding: 10px;
    margin-bottom: 10px;
    background-color: #fff;
    border: 1px solid #e5e7eb;
    box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
    border-radius: 6px;
    cursor: grab;
    user-select: none;
}

.kanban-task-ghost {
    opacity: 0.35;
    background-color: #e0e7ff;
    border-radius: 5px;
    box-sizing: border-box;
    height: var(--drag-h, auto);
    min-height: var(--drag-h, auto);
    overflow: hidden;
}

button:hover:not(.task-status button) {
    background-color: rgba(0, 0, 0, 0.05);
}

button:active {
    background-color: rgba(0, 0, 0, 0.1);
}

.title-wrapper {
    flex: 1;
    display: flex;
    overflow: hidden;
}
.title-wrapper-container {
    display: flex;
    align-items: center;
    overflow: hidden;
    width: 100%;
    box-sizing: border-box;
}

.column-title > div {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: inline-block;
    max-width: 100%;
}
</style>

<style>
.scrollable-column {
    height: calc(100vh - 280px);
    overflow-y: auto;
    position: relative;
    padding-right: 10px;
    padding-left: 10px;
}
</style>
