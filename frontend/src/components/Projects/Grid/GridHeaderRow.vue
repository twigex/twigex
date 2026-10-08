<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="grid-container sticky top-0 z-[2]">
        <div class="header-row">
            <div class="grid-row" style="display: flex; align-items: center">
                <draggable
                    v-if="isDataLoaded"
                    v-model="headers"
                    group="headers"
                    item-key="name"
                    :animation="150"
                    handle=".header-drag-handle"
                    :filter="'.disable-drag-only'"
                    :prevent-on-filter="false"
                    @end="emit('reorder', $event)"
                    @change="emit('reorder', $event)"
                    :move="onHeaderMove"
                    class="flex"
                    :disabled="isDragDisabled"
                >
                    <template #item="{ element: header, index }">
                        <div
                            :style="getColumnStyle(index)"
                            :class="[
                                (header.visible ?? true) ? 'grid-column' : 'grid-column2',
                                'relative select-none bg-gray-50 hover:bg-gray-100',
                                header.name === 'id' || header.name === 'name'
                                    ? 'disable-drag-only'
                                    : '',
                            ]"
                            :data-column-index="index"
                        >
                            <div
                                v-if="header.visible"
                                class="header-cell group"
                                @dblclick="
                                    !isDefaultField(header) &&
                                    canEditFields &&
                                    handleHeaderDoubleClick(index, header)
                                "
                            >
                                <template v-if="editableHeader.index === index">
                                    <input
                                        v-model="editableHeader.value"
                                        @blur="saveHeaderValue(index)"
                                        @keyup.enter="saveHeaderValue(index)"
                                        @keyup.esc="cancelHeaderEdit"
                                        class="block w-full min-w-0 rounded-md border-0 px-2 py-0.5 text-sm font-medium text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                    />
                                </template>
                                <template v-else>
                                    <label
                                        :class="[
                                            'block text-xs font-medium text-gray-500 header-label',
                                            header.name === 'id' || header.name === 'name'
                                                ? 'non-draggable'
                                                : 'header-drag-handle',
                                        ]"
                                        :title="getHeaderDisplayName(header)"
                                        :style="
                                            header.name === 'id' || header.name === 'name'
                                                ? 'cursor: default;'
                                                : 'cursor: grab;'
                                        "
                                    >
                                        {{ getHeaderDisplayName(header) }}
                                    </label>
                                </template>

                                <div class="flex items-center pr-1.5">
                                    <button
                                        v-if="!isDefaultField(header) && canManageFields"
                                        class="invisible-button"
                                        @click="emit('open-menu', $event, header)"
                                        title="Field options"
                                    >
                                        <EllipsisHorizontalIcon
                                            class="h-4 w-4 text-gray-700 opacity-0 group-hover:opacity-100 transition-opacity duration-150 hover:text-gray-900"
                                            aria-hidden="true"
                                        />
                                    </button>
                                    <button
                                        class="group/handle absolute inset-y-0 right-0 z-10 flex w-1.5 cursor-col-resize items-center justify-center border-0 bg-transparent p-0 outline-none"
                                        @mousedown="
                                            canEditFields && emit('resize-start', index, $event)
                                        "
                                        title="Drag to resize"
                                    >
                                        <span
                                            class="h-[calc(100%-10px)] w-0.5 rounded-full bg-indigo-500 opacity-0 transition duration-150 group-hover/handle:opacity-100 group-active/handle:bg-indigo-600 group-active/handle:opacity-100"
                                        ></span>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </template>
                </draggable>

                <div
                    v-if="!isDialog && canManageFields"
                    class="plus-icon-container"
                    style="margin-left: 6px"
                >
                    <BaseButton
                        type="button"
                        size="small"
                        variant="secondary"
                        :prepend-icon="PlusIcon"
                        :aria-label="t('projects.grid_view.add_field')"
                        @click="emit('add-field', $event)"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { reactive } from "vue";
import { useRoute } from "vue-router";
import draggable from "vuedraggable";
import { t } from "@/i18n/index.js";
import BaseButton from "@/components/BaseButton.vue";
import { PlusIcon } from "@heroicons/vue/24/outline";
import { EllipsisHorizontalIcon } from "@heroicons/vue/20/solid";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const headers = defineModel("headers", { type: Array, default: () => [] });

const props = defineProps({
    columnWidths: { type: Array, default: () => [] },
    tableId: { type: String, default: "" },
    isDialog: { type: Boolean, default: false },
    isDataLoaded: { type: Boolean, default: false },
    isDragDisabled: { type: Boolean, default: true },
    canEditFields: { type: Boolean, default: false },
    canManageFields: { type: Boolean, default: false },
});

const emit = defineEmits(["reorder", "resize-start", "open-menu", "add-field", "update-header"]);

const route = useRoute();

function onHeaderMove(evt) {
    const dragged = evt.dragged?.__vue__?.$props?.element;

    const toIndex = evt.relatedContext?.index;

    if (dragged?.name === "id" || dragged?.name === "name") {
        return false;
    }

    if (toIndex === 0 || toIndex === 1) {
        return false;
    }

    return true;
}

const editableHeader = reactive({
    index: null,
    value: "",
});

const handleHeaderDoubleClick = (index, header) => {
    if (header.display_name == "" || header.display_name == undefined) {
        editableHeader.value = header.name;
    } else {
        editableHeader.value = header.display_name;
    }

    editableHeader.index = index;
    editableHeader.header_name = header.name;
};

const saveHeaderValue = (index) => {
    const valueToSave = editableHeader.value;
    const headerName = headers.value[index].name;
    const previous = headers.value[index].display_name;

    emit("update-header", headerName, { display_name: valueToSave });

    const tableID = props.tableId;

    workspaceService
        .updateDisplayName({
            workspace_id: route.params.id,
            table_id: tableID,
            view_id: route.params.fid,
            header_name: editableHeader.header_name,
            display_name: valueToSave,
            visible: headers.value[index].visible,
        })
        .then(() => {
            editableHeader.index = null;
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
            emit("update-header", headerName, { display_name: previous });
            resetHeaderEditState();
        });
};

const cancelHeaderEdit = () => {
    resetHeaderEditState();
};

const resetHeaderEditState = () => {
    editableHeader.index = null;
    editableHeader.value = "";
};

const defaultFields = new Set([
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

const isDefaultField = (header) => {
    return header && header.name && defaultFields.has(header.name);
};

const DEFAULT_FIELD_TRANSLATION_KEYS = {
    id: "projects.grid_view.id",
    created_at: "projects.grid_view.created_at",
    updated_at: "projects.grid_view.updated_at",
    name: "projects.grid_view.name",
    status: "projects.grid_view.status",
    assignee: "projects.grid_view.assignee",
    description: "projects.grid_view.description",
    start_date: "projects.grid_view.start_date",
    due_date: "projects.grid_view.due_date",
    created_by: "projects.grid_view.created_by",
};

const getHeaderDisplayName = (header) => {
    if (isDefaultField(header)) {
        const translationKey = DEFAULT_FIELD_TRANSLATION_KEYS[header.name];

        if (translationKey) {
            return t.value(translationKey);
        }
    }

    return header.display_name || header.name;
};

const getColumnStyle = (index) => ({
    width: headers.value[index]?.visible ? `${props.columnWidths[index]}px` : "0px",
    overflow: headers.value[index]?.visible ? "visible" : "hidden",
    visibility: headers.value[index]?.visible ? "visible" : "hidden",
});
</script>

<style scoped src="../projectsTable.css"></style>

<style scoped>
.grid-row .plus-icon-container {
    border-bottom: none !important;
}

.grid-row .plus-icon-container {
    margin-left: 1px;
}

.header-row {
    position: sticky;
    top: 0;
    background-color: white;
}

.grid-column {
    will-change: width;
    height: 36px;
    padding: 0 8px;
    display: flex;
    align-items: center;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.grid-column2 {
    box-sizing: border-box;
    padding: 0;
    margin: 0;
    border: none;
    min-width: 0 !important;
    max-width: 0 !important;
    width: 0 !important;
    visibility: hidden;
    overflow: hidden;
    display: none;
}

.grid-column:nth-child(1),
.grid-column:nth-child(2) {
    position: sticky;
    left: 0;
    z-index: 200; /* Higher than data rows */
}

.non-draggable {
    pointer-events: none;
}

.non-draggable .header-drag-handle {
    pointer-events: auto;
}

.disable-drag-only .header-drag-handle {
    pointer-events: none;
    cursor: default;
}
</style>
