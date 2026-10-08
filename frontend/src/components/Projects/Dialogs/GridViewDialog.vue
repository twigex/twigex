<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <TransitionRoot as="template" :show="showDialog">
        <Dialog as="div" class="relative z-50" @close="$emit('close')">
            <TransitionChild
                as="template"
                enter="ease-out duration-300"
                enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                enter-to="opacity-100 translate-y-0 sm:scale-100"
                leave="ease-in duration-200"
                leave-from="opacity-100 translate-y-0 sm:scale-100"
                leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
            >
                <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
            </TransitionChild>

            <div class="fixed inset-0 z-10 w-screen flex items-center justify-center p-4">
                <TransitionChild
                    as="template"
                    enter="ease-out duration-300"
                    enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                    enter-to="opacity-100 translate-y-0 sm:scale-100"
                    leave="ease-in duration-200"
                    leave-from="opacity-100 translate-y-0 sm:scale-100"
                    leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95"
                >
                    <DialogPanel
                        class="relative flex max-h-[90vh] max-w-[50vw] transform flex-col overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all"
                    >
                        <!-- Header, fixed, never shrinks -->
                        <div class="flex-shrink-0">
                            <DialogTitle
                                as="h3"
                                class="text-base font-semibold leading-6 text-gray-900"
                            >
                                <div class="flex items-center">
                                    {{ props.title || props.field }}
                                    <ArrowTopRightOnSquareIcon
                                        v-if="!singleSelect"
                                        class="ml-2 h-4 w-4 cursor-pointer text-gray-500 hover:text-gray-800"
                                        :title="
                                            t(
                                                'projects.dialogs.grid_view_dialog.go_to_linked_table',
                                            )
                                        "
                                        aria-hidden="true"
                                        @click.stop="redirectToLinkedTable()"
                                    />
                                </div>
                            </DialogTitle>
                        </div>

                        <!-- Grid, no flex-1 so the dialog shrinks to fit content -->
                        <div class="mt-2 overflow-auto">
                            <GridView
                                v-if="isReady"
                                :key="gridKey"
                                :isDialog="true"
                                :tableHeaders="headers"
                                :tableData="data"
                                :columnWidths="columnWidths"
                                @update:tableData="data = $event"
                                @update:tableHeaders="headers = $event"
                                @update:assigneeDialog="onAssigneeChanged($event)"
                                @selected="onItemSelected"
                                :isSingleSelect="props.singleSelect"
                                :newCurrent="newCurrentTableID"
                                :totalCount="dialogTotal"
                                :dialogWorkspaceId="props.workspaceId || route.params.id"
                                :dialogViewId="dialogViewId"
                                :dialogTableId="props.tableId"
                                :dialogSelectedIds="selectedIds"
                            />
                        </div>

                        <!-- Action buttons, pinned at bottom, never scrolls away -->
                        <div
                            class="mt-4 flex justify-end gap-3 flex-shrink-0 border-t border-gray-100 pt-3"
                        >
                            <BaseButton type="button" variant="secondary" @click="$emit('close')">
                                {{ t("common.button.cancel") }}
                            </BaseButton>
                            <BaseButton type="button" @click="linkSelectedItems()">
                                {{ t("common.button.save") }}
                            </BaseButton>
                        </div>
                    </DialogPanel>
                </TransitionChild>
            </div>
        </Dialog>
    </TransitionRoot>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { watch, ref, nextTick, computed } from "vue";
import GridView from "@/components/Projects/Grid/GridView.vue";
import BaseButton from "@/components/BaseButton.vue";
import { useWorkspaceStore } from "@/store/workspaces";
import { fieldLabel } from "@/utils/projects/rows";
import workspaceService from "@/services/workspaceService";
import { useShownFilter } from "@/composables/projects/useShownFilter";
import { useTaskCompletion } from "@/composables/projects/useTaskCompletion";
import { useGoToTable } from "@/composables/projects/useGoToTable";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from "@headlessui/vue";
import { ArrowTopRightOnSquareIcon } from "@heroicons/vue/24/outline";
import { useRoute } from "vue-router";
import { getStatusDisplayName } from "@/utils/projects/cells";
import { moveCardToColumn, setCardField } from "@/utils/projects/board";
const route = useRoute();
const headers = ref([]);
const columnWidths = ref([]);
const data = ref([]);
const dialogTotal = ref(0);
const dialogViewId = ref("");
const selectedItems = ref([]);
const selectedIds = ref([]);
const initialIds = ref([]);
const workspaceStore = useWorkspaceStore();

const props = defineProps({
    isOpen: Boolean,
    workspaceId: String,
    tableId: String,
    currentLinkedIDs: Array,
    singleSelect: Boolean,
    field: String,
    editItemLink: Object,
    linkedID: String,
    currentLocalTableID: String,
    linkedItemsForCheckbox: Array,
    isAtFirstLinkedLevel: Boolean,
    title: { type: String, default: "" },
});

const getSelectedItem = computed({
    get() {
        return workspaceStore.getSelectedItem;
    },
    set(value) {
        workspaceStore.setSelectedItem(value);
    },
});

const onItemSelected = (value, item) => {
    selectedItems.value = value;
    if (item?.id) {
        const id = String(item.id);

        if (props.singleSelect) {
            selectedIds.value = item.itemSelected ? [id] : [];
        } else if (item.itemSelected) {
            if (!selectedIds.value.includes(id)) selectedIds.value = [...selectedIds.value, id];
        } else {
            selectedIds.value = selectedIds.value.filter((selected) => selected !== id);
        }
    }

    if (Array.isArray(value) && value.length === 1) {
        const selectedId = value[0].id;
        const currentTableId = props.tableId;

        if (
            storedTableDataForNestedDialog.value &&
            storedTableDataForNestedDialog.value[currentTableId]
        ) {
            storedTableDataForNestedDialog.value[currentTableId].editItemId = selectedId;
        }

        const index = tableHistory.value.indexOf(currentTableId);
        const parentId = tableHistory.value[index - 1];

        if (storedTableDataForNestedDialog.value[parentId]) {
            storedTableDataForNestedDialog.value[parentId].editItemId = props.editItemLink?.id;
        }
    }
};

const newCurrentTableID = ref(null);

const tableHistory = computed({
    get() {
        return workspaceStore.getTestStore;
    },
    set(value) {
        workspaceStore.setTestStore(value);
    },
});

const { matchesShownFilter } = useShownFilter();
const taskCompletion = useTaskCompletion();
const checkIfTaskShouldBeVisible = (taskId) =>
    matchesShownFilter(route.params.id, route.params.tid, taskId);

// A task the filter hid is no longer among the rows, so when it matches again
// the page is read anew, which puts it where it sorts.
const applyFilterVisibility = async (taskId) => {
    if (!workspaceStore.filterActive) return;

    const visible = await checkIfTaskShouldBeVisible(taskId);
    const i = tableData.value.findIndex((t) => t.id === taskId);

    if (!visible && i !== -1) {
        tableData.value.splice(i, 1);
    } else if (visible && i === -1) {
        workspaceStore.bumpGridReloadToken();
    }
};

const showSelectedItem = (taskId, field, value, extra = {}) => {
    if (taskId !== getSelectedItem.value?.id) return;
    workspaceStore.setSelectedItem({ ...getSelectedItem.value, [field]: value, ...extra });
};

const showInGrid = ({ taskId, field, value }) => {
    const updated_at = Date.now();

    tableData.value = tableData.value.map((row) =>
        row.id === taskId ? { ...row, [field]: value, updated_at } : row,
    );
    showSelectedItem(taskId, field, value, { updated_at });
    applyFilterVisibility(taskId);
};

// showInNestedDialog updates the row in the dialog this one was opened from,
// whose rows are its own. It reports whether that dialog had the row.
const showInNestedDialog = ({ taskId, field, value }) => {
    const id = String(taskId);
    const parentTable = Object.values(storedTableDataForNestedDialog.value).find((table) =>
        table?.data?.some((item) => String(item.id) === id),
    );

    if (!parentTable) return false;

    parentTable.data.find((item) => String(item.id) === id)[field] = value;
    data.value = [...parentTable.data];
    showSelectedItem(taskId, field, value);

    const i = tableData.value.findIndex((item) => String(item.id) === id);

    if (i !== -1) {
        tableData.value[i] = { ...tableData.value[i], [field]: value };
        tableData.value = [...tableData.value];
    }

    return true;
};

const boardGroupField = () => {
    const order = selectedView.value?.order;

    return order ? JSON.parse(order)?.section || null : null;
};

// showOnBoard sets the field on the task's card, and moves the card to the
// column of its new value when mayMove and the board is grouped by that field.
const showOnBoard = ({ taskId, field, value }, mayMove) => {
    taskLists.value =
        mayMove && field === boardGroupField()
            ? moveCardToColumn(taskLists.value, taskId, field, value)
            : setCardField(taskLists.value, taskId, field, value);
};

const loadData = async () => {
    if (!props.workspaceId || !props.tableId) return;

    if (
        tableHistory.value.length === 0 ||
        tableHistory.value[tableHistory.value.length - 1] !== props.tableId
    ) {
        tableHistory.value = tableHistory.value.filter((id) => id !== props.tableId);
        tableHistory.value.push(props.tableId);
    }

    newCurrentTableID.value = props.tableId;

    const norm = (v) => (v == null ? "" : String(v).trim());
    const isNonEmpty = (v) => v !== null && v !== undefined && v !== "";

    try {
        const tables = Array.isArray(workspaceTables.value) ? workspaceTables.value : [];

        const table = tables.find((t) => t.id === props.tableId);

        if (!table) return;

        if (!Array.isArray(table.headers)) {
            headers.value = [];
        } else {
            // A board's order is an object, not the list of columns a grid
            // view keeps, and gives no column settings here.
            let parsedOrder = [];

            try {
                const order = JSON.parse(table.options?.[0]?.order || "[]");

                if (Array.isArray(order)) parsedOrder = order;
            } catch {
                parsedOrder = [];
            }

            const baseHeaders = table.headers
                .filter((header) => header && header.name)
                .map((header) => {
                    const orderItem = parsedOrder.find((order) => order.name === header.name);

                    return {
                        ...header,
                        display_name:
                            header.name === "name"
                                ? "Name"
                                : fieldLabel(header, orderItem?.display_name),
                        visible:
                            header.name === "name"
                                ? true
                                : props.singleSelect && header.name.toLowerCase() === "id"
                                  ? false
                                  : orderItem && "visible" in orderItem
                                    ? orderItem.visible
                                    : true,
                    };
                });

            let finalHeaders = baseHeaders.slice();

            const hasStatusTypeInData =
                table.data_base &&
                table.data_base.length > 0 &&
                "status_type" in table.data_base[0];

            if (hasStatusTypeInData) {
                const hasStatusTypeHeader = finalHeaders.some((h) => h.name === "status_type");

                if (!hasStatusTypeHeader) {
                    finalHeaders.push({
                        name: "status_type",
                        display_name: "Status Type",
                        visible: true,
                        header_type: "VARCHAR",
                        header_usage: "text",
                        single_select: false,
                        linked_id: "",
                        parent_table_id: null,
                    });
                }
            }

            if (props.singleSelect) {
                const hasColorHeader = baseHeaders.some((h) => h.name === "color");

                if (!hasColorHeader) {
                    const colorOrderItem = parsedOrder.find((o) => o.name === "color");

                    finalHeaders.push({
                        name: "color",
                        display_name: (colorOrderItem && colorOrderItem.display_name) || "Color",
                        visible:
                            colorOrderItem && "visible" in colorOrderItem
                                ? colorOrderItem.visible
                                : true,
                    });
                }
            }

            headers.value = finalHeaders;

            columnWidths.value = headers.value.map((header) => {
                if (header.name === "color" && props.singleSelect) return 150;
                if (header.name === "status_type") return 120;
                const orderItem = parsedOrder.find((order) => order.name === header.name);
                const width = orderItem && orderItem.width ? Number(orderItem.width) : 350;

                // The name is what a record is picked by, and it shares its
                // column with the checkbox, so a width saved for the table's
                // own grid does not narrow it here.
                return header.name === "name" ? Math.max(width, 300) : width;
            });
        }

        const linkedIDsRaw = Array.isArray(props.linkedItemsForCheckbox)
            ? props.linkedItemsForCheckbox
            : (props.linkedItemsForCheckbox && props.linkedItemsForCheckbox.value) || [];
        const linkedIDs = linkedIDsRaw.map(norm).filter(isNonEmpty);

        selectedIds.value = [...new Set(linkedIDs)];
        initialIds.value = [...selectedIds.value];

        const mapItem = (item) => {
            const newItem = {};

            (table.headers || []).forEach((header) => {
                newItem[header.name] =
                    header.header_type === "TINYINT"
                        ? item[header.name] === "1"
                        : (item[header.name] ?? "");
            });
            const id = norm(item.id);

            newItem.id = id;
            const rawParentId = item.parent_task_id;
            const parent = rawParentId == null ? null : norm(rawParentId) || null;

            newItem.parent_task_id = parent;
            newItem._original_parent_task_id = parent;
            newItem._is_subtask = !!(parent && parent !== id);
            if ("status_type" in item) newItem.status_type = item.status_type;
            if (props.field === "status" && newItem.name) {
                newItem._original_name = newItem.name;
                newItem.name = getStatusDisplayName({ name: newItem.name });
            }

            if (props.singleSelect) {
                let rawColor = item.color ?? item.Color ?? "";

                if (typeof rawColor === "string" && rawColor.startsWith("bg-"))
                    rawColor = "#c43131";
                newItem.color = rawColor;
            }

            newItem.itemSelected = isNonEmpty(id) && linkedIDs.includes(id);

            return newItem;
        };

        if (props.singleSelect) {
            // Single-select options come from table.data_base in the workspace response
            const rawBase = Array.isArray(table.data_base) ? table.data_base : [];
            const baseData = rawBase.map(mapItem);

            baseData.sort((a, b) => {
                if (a.itemSelected && !b.itemSelected) return -1;
                if (!a.itemSelected && b.itemSelected) return 1;

                return (a.name || "").localeCompare(b.name || "", undefined, {
                    sensitivity: "base",
                });
            });
            data.value = baseData;
            dialogTotal.value = baseData.length;
        } else {
            // Linked table. Fetch page 1 with no filters so ALL tasks are available.
            // Grab first view id for fetchPage pagination.
            dialogViewId.value =
                Array.isArray(table.options) && table.options.length
                    ? table.options[0]?.id || ""
                    : "";

            try {
                const pageRes = await workspaceService.getFilteredTableData({
                    workspace_id: route.params.id,
                    table_id: props.tableId,
                    view_id: dialogViewId.value,
                    filters: { groups: [], flatFilters: [], page: 1, limit: 100, sort: [] },
                });
                const rawBase = pageRes?.data?.data_base || [];

                dialogTotal.value = pageRes?.data?.total || rawBase.length;
                const baseData = rawBase.map(mapItem);

                // Put already-linked items first
                baseData.sort((a, b) => {
                    if (a.itemSelected && !b.itemSelected) return -1;
                    if (!a.itemSelected && b.itemSelected) return 1;

                    return 0;
                });
                data.value = baseData;
            } catch {
                data.value = [];
                dialogTotal.value = 0;
            }
        }

        storedTableDataForNestedDialog.value = {
            ...storedTableDataForNestedDialog.value,
            [props.tableId]: {
                data: data.value,
                headers: headers.value,
                columnWidths: columnWidths.value,
                editItemId:
                    (storedTableDataForNestedDialog.value[props.tableId] &&
                        storedTableDataForNestedDialog.value[props.tableId].editItemId) ||
                    (props.editItemLink && props.editItemLink.id) ||
                    null,
                field: props.field,
                linkedID: props.linkedID,
                currentLocalTableID: props.currentLocalTableID,
            },
        };

        await nextTick();
        isReady.value = true;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const onAssigneeChanged = (value) => {
    const index = data.value.findIndex((item) => item.id === value.task_id);

    if (index !== -1) {
        if (Object.hasOwn(data.value[index], value.field_name)) {
            data.value[index][value.field_name] = value.user_id;
        }
    }
};

const storedTableDataForNestedDialog = computed({
    get() {
        return workspaceStore.getStoredTableDataForNestedDialog;
    },
    set(value) {
        workspaceStore.setStoredTableDataForNestedDialog(value);
    },
});

const tableData = computed({
    get() {
        return workspaceStore.getTableData;
    },
    set(value) {
        workspaceStore.setTableData(value);
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

const singleField = computed({
    get() {
        return workspaceStore.getFieldName;
    },
    set(value) {
        workspaceStore.setFieldName(value);
    },
});

const propsLinkedUpdate = computed({
    get() {
        return workspaceStore.getPropsLinkedUpdate;
    },
    set(value) {
        workspaceStore.setPropsLinkedUpdate(value);
    },
});

const workspaceTables = computed({
    get() {
        return workspaceStore.getWorkspaceTables;
    },
    set(value) {
        workspaceStore.setWorkspaceTables(value);
    },
});

const { goToTable } = useGoToTable();

const redirectToLinkedTable = () => {
    if (goToTable(props.tableId)) closeDialog();
};

const selectedView = computed({
    get() {
        return workspaceStore.getSelectedViewData;
    },
    set(value) {
        workspaceStore.setSelectedViewData(value);
    },
});

const isTouchedCheckbox = computed({
    get() {
        return workspaceStore.getIsTouchedCheckbox;
    },
    set(value) {
        workspaceStore.setIsTouchedCheckbox(value);
    },
});

const linkSelectedItems = async () => {
    const fieldName = props.field;

    const selected = new Set(selectedIds.value);
    const initial = new Set(initialIds.value);
    const change = props.singleSelect
        ? { linked_ids: [...selected] }
        : {
              add: [...selected].filter((id) => !initial.has(id)),
              remove: [...initial].filter((id) => !selected.has(id)),
          };

    var tableID = setTableID.value;

    if (setTableID.value === null || setTableID.value === "" || setTableID.value === undefined) {
        tableID = props.currentLocalTableID;
    }

    const completion =
        props.singleSelect && fieldName === "status" && props.editItemLink?.id
            ? await taskCompletion.beforeStatusChange({
                  workspaceId: props.workspaceId,
                  tableId: tableID,
                  taskId: props.editItemLink.id,
                  statusId: [...selected][0],
              })
            : undefined;

    if (completion === null) return;

    workspaceService
        .updateTableLink({
            workspace_id: props.workspaceId,
            table_id: tableID,
            task_id: props.editItemLink.id,
            parent_table_id: props.linkedID,
            single_select: props.singleSelect,
            field: fieldName,
            ...change,
        })
        .then((response) => {
            // Sent before either branch below: a dialog opened from a dialog
            // keeps rows of its own, which a task created there is not in.
            const linked = response.data?.newLinkedItems;
            const saved = props.singleSelect
                ? {
                      taskId: props.editItemLink.id,
                      field: fieldName,
                      value: response.data?.cleared ? null : response.data,
                  }
                : linked?.item_field_name
                  ? {
                        taskId: props.editItemLink.id,
                        field: linked.item_field_name,
                        value: Array.isArray(linked.item_names) ? linked.item_names : [],
                    }
                  : null;

            if (saved) emit("saved", saved);
            completion?.afterSave();

            initialIds.value = [...selected];
            if (!saved) return;

            if (propsLinkedUpdate.value) {
                if (!showInNestedDialog(saved)) return;
                showOnBoard(saved, false);
            } else {
                showInGrid(saved);
                showOnBoard(saved, props.singleSelect);
            }
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });

    closeDialog();
    isTouchedCheckbox.value = false;
};

const taskLists = computed({
    get() {
        return workspaceStore.getKanbanTaskList || [];
    },
    set(value) {
        workspaceStore.setKanbanTaskList(value);
    },
});

// saved tells the grid that opened the dialog what changed, as its rows are
// not the store's when that grid is itself in a dialog.
const emit = defineEmits(["close", "saved"]);

const closeDialog = () => {
    emit("close");
};

const isReady = ref(false);
const loading = ref(false);

async function safeLoad() {
    if (loading.value) return;
    loading.value = true;
    try {
        await loadData();
    } finally {
        loading.value = false;
    }
}

const gridKey = computed(() => {
    const linked = Array.isArray(props.currentLinkedIDs) ? props.currentLinkedIDs.join(",") : "";

    return [props.tableId, props.field ?? "", props.linkedID ?? "", linked].join("|");
});

const showDialog = ref(false);

function resetState() {
    headers.value = [];
    columnWidths.value = [];
    data.value = [];
    isReady.value = false;
}

let openToken = 0;

watch(
    () => [
        props.isOpen,
        props.workspaceId,
        props.tableId,
        props.currentLinkedIDs,
        props.field,
        props.linkedID,
    ],
    async ([newIsOpen]) => {
        if (!newIsOpen) {
            showDialog.value = false;
            isReady.value = false;

            return;
        }

        showDialog.value = false;
        resetState();
        singleField.value = props.field;

        const token = ++openToken;

        try {
            await safeLoad();
        } catch {
            return;
        }

        if (token === openToken) {
            isReady.value = true;
            await nextTick();
            showDialog.value = true;
        }
    },
);
</script>
