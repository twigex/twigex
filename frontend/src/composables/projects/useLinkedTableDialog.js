// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed, nextTick, ref } from "vue";
import { useRoute } from "vue-router";
import { t } from "@/i18n/index.js";
import { useAlertStore } from "@/store/alerts";
import { useWorkspaceStore } from "@/store/workspaces";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { linkedDialogTitle } from "@/utils/projects/rows";

const storeValue = (store, getter, setter) =>
    computed({
        get: () => store[getter],
        set: (value) => store[setter](value),
    });

// useLinkedTableDialog opens the dialog that links a task to rows of another
// table, from a grid or from a grid inside such a dialog. The dialogs opened
// one from another are kept as a stack in the store, so closing one goes back
// to the field and table of the one beneath it.
export function useLinkedTableDialog({ props, tableHeaders, canEditRow, onLinkedIds }) {
    const route = useRoute();
    const workspaceStore = useWorkspaceStore();
    const rolesStore = useWorkspaceRolesStore();

    const open = ref(false);
    const shown = ref(false);
    const field = ref("");
    const parentTableID = ref(null);
    const title = ref("");
    const singleSelect = ref(false);
    const linkedID = ref(null);
    const editItemLink = ref({});
    const linkedItemsForCheckbox = ref([]);
    let opening = false;

    const stack = storeValue(workspaceStore, "getCatchedDataForDialog", "setCatchedDataForDialog");
    const dialogTableID = storeValue(workspaceStore, "getDialogTableID", "setDialogTableID");
    const setTableID = storeValue(workspaceStore, "getTableID", "setTableID");
    const propsLinkedUpdate = storeValue(
        workspaceStore,
        "getPropsLinkedUpdate",
        "setPropsLinkedUpdate",
    );
    const linkedFieldName = storeValue(workspaceStore, "getLinkedFieldName", "setLinkedFieldName");

    const isAtFirstLinkedLevel = computed(() => stack.value.length <= 1);

    const knownTables = () => [
        ...(workspaceStore.getWorkspaceTablesForLinked || []),
        ...(workspaceStore.getWorkspaceTables || []),
    ];
    const titleFor = (header) => linkedDialogTitle(header, knownTables());

    // The table a cell of this grid belongs to, with the headers its column
    // index is read against.
    function cellTable() {
        if (!props.isDialog) return { id: route.params.tid, headers: tableHeaders.value };

        // A grid in a dialog shows the dialog's table, with these headers. The
        // dialog's cached entry names the table that opened it instead, so a
        // cell here would be read as that table's column and saved to it.
        if (props.dialogTableId && tableHeaders.value?.length) {
            return { id: props.dialogTableId, headers: tableHeaders.value };
        }

        const last = [...stack.value]
            .reverse()
            .find((c) => c.dialogTableID === dialogTableID.value && !c.closed);

        if (last) return { id: last.fromTableID, headers: last.headers };

        return (
            (workspaceStore.getWorkspaceTablesForLinked || []).find(
                (t) => t.id === dialogTableID.value,
            ) || null
        );
    }

    function canViewTarget(header) {
        if (header.single_select) return true;

        // parent_table_id is empty for one-directional links; fall back to the junction table's parent
        let target = header.parent_table_id;

        if (!target) {
            const junction = (workspaceStore.getWorkspaceTablesForLinked || []).find(
                (t) => t.id === header.linked_id,
            );

            target = junction?.parent_table_id || "";
        }

        return rolesStore.hasTablePermission(target, "view");
    }

    function linkedIdsOf(item, header) {
        const value = item[header.name];

        if (!value) return [];
        if (header.single_select) return value.id ? [value.id] : [];

        return Array.isArray(value) ? value.filter((v) => v?.id).map((v) => v.id) : [];
    }

    function openFor(item, colIndex) {
        if (!canEditRow.value) {
            useAlertStore().showError(t.value("projects.grid_view.no_permission_edit_fields"));

            return;
        }

        if (opening) return;

        const table = cellTable();
        const header = table?.headers?.[colIndex];

        if (!header) return;
        if (!canViewTarget(header)) {
            useAlertStore().showError(t.value("projects.grid_view.no_permission_view_table"));

            return;
        }

        opening = true;

        const reused = stack.value.find(
            (c) =>
                c.dialogTableID === header.linked_id &&
                c.fromTableID === table.id &&
                c.colIndex === colIndex &&
                c.itemID === item?.id &&
                c.closed === true,
        );

        if (reused) {
            reused.closed = undefined;
        } else {
            stack.value.push({
                dialogTableID: header.linked_id,
                fromTableID: table.id,
                colIndex,
                itemID: item?.id,
                headers: table.headers,
                closed: undefined,
            });
        }

        if (stack.value.length >= 2) propsLinkedUpdate.value = true;

        editItemLink.value = item;
        singleSelect.value = header.single_select;
        field.value = header.name;
        parentTableID.value = header.single_select ? header.linked_id : header.parent_table_id;
        linkedID.value = header.linked_id;
        dialogTableID.value = header.linked_id;
        setTableID.value = table.id;

        linkedItemsForCheckbox.value = linkedIdsOf(item, header);
        linkedFieldName.value = header.linked_name || "";
        onLinkedIds?.(linkedItemsForCheckbox.value);

        title.value = titleFor(header);
        open.value = true;
        shown.value = true;

        nextTick(() => {
            opening = false;
        });
    }

    function close() {
        open.value = false;

        const openDialogs = stack.value.filter((s) => !s.closed);

        if (!openDialogs.length) return;
        openDialogs[openDialogs.length - 1].closed = true;

        const back = openDialogs[openDialogs.length - 2];
        const backHeader = back?.headers[back.colIndex];

        dialogTableID.value = back ? back.dialogTableID : null;
        setTableID.value = back ? back.fromTableID : null;
        field.value = backHeader?.name || "";
        parentTableID.value = backHeader?.parent_table_id || "";
        linkedID.value = back ? back.dialogTableID : "";
        if (back) title.value = titleFor(backHeader);

        stack.value = stack.value.filter((entry) => !entry.closed);
        if (stack.value.length < 1) propsLinkedUpdate.value = false;
    }

    return {
        open,
        shown,
        field,
        parentTableID,
        title,
        singleSelect,
        linkedID,
        editItemLink,
        linkedItemsForCheckbox,
        isAtFirstLinkedLevel,
        openFor,
        close,
    };
}
