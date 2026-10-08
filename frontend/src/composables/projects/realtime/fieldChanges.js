// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { newFieldColumn } from "@/utils/projects/rows";

// fieldChanges applies changes to a table's fields: added, renamed, moved,
// resized, given a formula or deleted.
export function fieldChanges(ctx) {
    const {
        route,
        workspaceStore,
        setFullViewHeaders,
        tableData,
        realoadUI,
        columnWidths,
        tableHeaders,
        taskLists,
        titles,
        workspaceTables,
        removeTaskRow,
    } = ctx;

    // The creating tab adjusts the new header before showing it; other tabs need
    // the same header.
    function newFieldHeader(change) {
        const header = { ...change.fieldData };

        if (header.header_usage === "single select") header.single_select = true;
        header.parent_table_id = change.linkedTableID || null;
        header.header_type = String(header.header_type || "").toUpperCase();
        if (change.fieldType === "file" && !header.header_usage) header.header_usage = "file";
        if (change.formula) header.formula = change.formula;

        return header;
    }

    function removeField(tableId, fieldId, own) {
        const tIdx = workspaceTables.value.findIndex((t) => String(t.id) === String(tableId));
        const table = tIdx === -1 ? null : workspaceTables.value[tIdx];
        const header = (table?.headers || []).find((h) => String(h.id) === String(fieldId));

        if (header) {
            workspaceTables.value[tIdx] = {
                ...table,
                headers: table.headers.filter((h) => h !== header),
                data_base: (table.data_base || []).map(({ [header.name]: _, ...rest }) => rest),
            };
            workspaceTables.value = [...workspaceTables.value];
        }

        if (own || String(tableId) !== String(route.params.tid)) return;

        const idx = tableHeaders.value.findIndex((h) => String(h.id) === String(fieldId));

        if (idx === -1) return;
        const name = tableHeaders.value[idx].name;

        tableHeaders.value = tableHeaders.value.filter((_, i) => i !== idx);
        columnWidths.value = columnWidths.value.filter((_, i) => i !== idx);
        tableData.value = tableData.value.map(({ [name]: _, ...rest }) => rest);
    }

    // A calculation is worked out in the browser from its field's formula, so a
    // new formula only has to reach the field's header.
    function applyFieldFormula(tableId, fieldName, formula) {
        const patch = (headers) =>
            (headers || []).map((h) => (h.name === fieldName ? { ...h, formula } : h));

        if (String(route.params.tid) === String(tableId))
            tableHeaders.value = patch(tableHeaders.value);
        workspaceTables.value = workspaceTables.value.map((t) =>
            String(t.id) === String(tableId) && Array.isArray(t.headers)
                ? { ...t, headers: patch(t.headers) }
                : t,
        );
    }

    function onNewField(fromSocket) {
        const f = { ...(fromSocket.fieldData || {}) };
        const name = f.name;
        let headerType = String(f.header_type || "").toUpperCase();

        const linkedId = f.linked_id ? String(f.linked_id) : null;
        const linkedTable = f.table_data || f.tableData || null;

        if (linkedId && linkedTable) {
            const hasLinked =
                Array.isArray(workspaceTables.value) &&
                workspaceTables.value.some((t) => String(t.id) === linkedId);

            if (!hasLinked) {
                const safeLinkedTable = {
                    ...linkedTable,
                    data_base: Array.isArray(linkedTable.data_base) ? linkedTable.data_base : [],
                    headers: Array.isArray(linkedTable.headers) ? linkedTable.headers : [],
                    views: Array.isArray(linkedTable.views) ? linkedTable.views : [],
                    options: Array.isArray(linkedTable.options) ? linkedTable.options : [],
                };

                workspaceTables.value = [...(workspaceTables.value ?? []), safeLinkedTable];
            }
        }

        const column = newFieldColumn(headerType);

        headerType = column.headerType;

        const tableIdx = workspaceTables.value.findIndex(
            (t) => String(t.id) === String(fromSocket.table_id),
        );

        if (tableIdx !== -1) {
            const table = workspaceTables.value[tableIdx];
            const rows = Array.isArray(table.data_base) ? table.data_base : [];
            const nextRows = rows.map((row) => ({
                ...row,
                [name]: column.empty(),
            }));

            const hasHeaders = Array.isArray(table.headers);
            const headerToPush = {
                ...(f || {}),
                name,
                header_type: headerType,
                display_name: f.display_name ?? name,
                width: "150",
            };
            const nextHeaders = hasHeaders
                ? table.headers.some((h) => String(h.name) === String(name))
                    ? [...table.headers]
                    : [...table.headers, headerToPush]
                : table.headers;

            const patchViewsOrder = (arr) => {
                if (!Array.isArray(arr)) return arr;

                return arr.map((v) => {
                    let parsed = Array.isArray(v?.order)
                        ? v.order
                        : typeof v?.order === "string" && v.order.trim()
                          ? (() => {
                                try {
                                    return JSON.parse(v.order);
                                } catch {
                                    return null;
                                }
                            })()
                          : null;

                    if (Array.isArray(parsed)) {
                        const exists = parsed.some((o) => String(o?.name) === String(name));

                        if (!exists)
                            parsed.push({
                                name,
                                display_name: f.display_name ?? name,
                                visible: true,
                                width: "150",
                            });
                        else
                            parsed = parsed.map((o) =>
                                String(o?.name) === String(name)
                                    ? {
                                          ...o,
                                          width:
                                              o.width == null || o.width === ""
                                                  ? "150"
                                                  : String(o.width),
                                      }
                                    : o,
                            );

                        return {
                            ...v,
                            order: JSON.stringify(parsed),
                        };
                    }

                    return v;
                });
            };

            const nextOptions = patchViewsOrder(table.options);
            const nextViews = patchViewsOrder(table.views);

            workspaceTables.value[tableIdx] = {
                ...table,
                data_base: nextRows,
                ...(hasHeaders ? { headers: nextHeaders } : {}),
                ...(Array.isArray(nextOptions) ? { options: nextOptions } : {}),
                ...(Array.isArray(nextViews) ? { views: nextViews } : {}),
            };
            workspaceTables.value = [...workspaceTables.value];
        }

        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id != workspaceStore.getConnectionID
        ) {
            const newFieldData = { ...fromSocket.fieldData };
            const column = newFieldColumn(newFieldData.header_type);

            newFieldData.header_type = column.headerType;

            tableHeaders.value = [...tableHeaders.value, newFieldData];

            if (setFullViewHeaders.value?.length > 0) {
                setFullViewHeaders.value = [...setFullViewHeaders.value, newFieldData];
            }

            tableData.value = tableData.value.map((item) => ({
                ...item,
                [newFieldData.name]: column.empty(),
            }));

            columnWidths.value = [...columnWidths.value, 150];
        } else if (
            fromSocket.linkedTableID == route.params.tid &&
            fromSocket.client_id != workspaceStore.getConnectionID
        ) {
            const newFieldData = { ...fromSocket.fieldData };

            newFieldData.id = fromSocket.fieldData.linked_header_id;
            const column = newFieldColumn(newFieldData.header_type);

            newFieldData.header_type = column.headerType;
            newFieldData.name = fromSocket.fieldNameInSecondTable;
            newFieldData.linked_id = newFieldData.parent_link_table_id;

            tableHeaders.value = [...tableHeaders.value, newFieldData];

            if (setFullViewHeaders.value?.length > 0) {
                setFullViewHeaders.value = [...setFullViewHeaders.value, newFieldData];
            }

            tableData.value = tableData.value.map((item) => ({
                ...item,
                [newFieldData.name]: column.empty(),
            }));

            columnWidths.value = [...columnWidths.value, 150];
        }
    }

    function onDisplayNameUpdate(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID &&
            fromSocket.view_id === route.params.fid
        ) {
            const d = fromSocket.data || {};
            const updated_field = d.updated_field;
            const display_name = d.display_name;
            const viewId = d.view_id;
            const currentViewId = route.params.fid;

            if (viewId === currentViewId) {
                const updatedHeaders = tableHeaders.value.map((header) => {
                    if (header.name === updated_field) {
                        return {
                            ...header,
                            display_name: display_name,
                            ...(typeof d.visible === "boolean" ? { visible: d.visible } : {}),
                        };
                    }

                    return header;
                });

                tableHeaders.value = [...updatedHeaders];
            }
        }
    }

    function onHeaderOrderUpdate(fromSocket) {
        const tableIdx = workspaceTables.value.findIndex(
            (t) => String(t.id) === String(fromSocket.table_id),
        );

        if (tableIdx !== -1) {
            const table = workspaceTables.value[tableIdx];

            const newOrderNames = Array.isArray(fromSocket.value)
                ? fromSocket.value.map((h) => h.name)
                : [];

            const nextHeaders = Array.isArray(table.headers)
                ? (() => {
                      const map = Object.fromEntries((table.headers || []).map((h) => [h.name, h]));

                      return newOrderNames.map((n) => map[n]).filter(Boolean);
                  })()
                : table.headers;

            const patchViewOrder = (arr) => {
                if (!Array.isArray(arr)) return arr;

                return arr.map((v) => {
                    if (String(v.id) !== String(fromSocket.view_id)) return v;

                    let parsed;

                    try {
                        parsed = Array.isArray(v.order) ? v.order : JSON.parse(v.order || "[]");
                    } catch {
                        parsed = [];
                    }

                    const prevMap = new Map((parsed || []).map((o) => [String(o.name), o]));
                    const incoming = Array.isArray(fromSocket.value) ? fromSocket.value : [];

                    const nextOrder = incoming.map((item) => {
                        const prev = prevMap.get(String(item.name)) || {};
                        const widthStr =
                            item.width != null && item.width !== ""
                                ? String(item.width)
                                : prev.width != null && prev.width !== ""
                                  ? String(prev.width)
                                  : "150";

                        return {
                            ...prev,
                            name: item.name,
                            width: widthStr,
                            display_name: prev.display_name ?? item.display_name ?? item.name,
                            visible:
                                typeof item.visible === "boolean"
                                    ? item.visible
                                    : typeof prev.visible === "boolean"
                                      ? prev.visible
                                      : true,
                        };
                    });

                    return {
                        ...v,
                        order: JSON.stringify(nextOrder),
                    };
                });
            };

            const nextOptions = patchViewOrder(table.options);
            const nextViews = patchViewOrder(table.views);

            workspaceTables.value[tableIdx] = {
                ...table,
                ...(Array.isArray(nextHeaders) ? { headers: nextHeaders } : {}),
                ...(Array.isArray(nextOptions) ? { options: nextOptions } : {}),
                ...(Array.isArray(nextViews) ? { views: nextViews } : {}),
            };
            workspaceTables.value = [...workspaceTables.value];
        }

        if (
            route.params.tid === fromSocket.table_id &&
            route.params.fid === fromSocket.view_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID
        ) {
            const newOrder = fromSocket.value.map((h) => h.name);

            // Capture original header names and their widths before changing anything
            const originalHeaderNames = tableHeaders.value.map((h) => h.name);
            const widthMap = {};

            originalHeaderNames.forEach((name, index) => {
                widthMap[name] = columnWidths.value[index];
            });

            // Reorder headers
            const headerMap = Object.fromEntries(tableHeaders.value.map((h) => [h.name, h]));

            tableHeaders.value = newOrder.map((name) => headerMap[name]).filter(Boolean);

            // Reorder rows
            tableData.value = tableData.value.map((row) => {
                const reorderedRow = {};

                newOrder.forEach((name) => {
                    reorderedRow[name] = row[name];
                });

                return reorderedRow;
            });

            // Apply columnWidths from saved mapping
            columnWidths.value = newOrder.map((name) => widthMap[name] ?? 100); // Default width fallback
        }
    }

    function onDeleteSingleField(fromSocket) {
        removeTaskRow(
            fromSocket.linked_table_id,
            fromSocket.delete_id,
            true,
            fromSocket.client_id === workspaceStore.getConnectionID,
        );

        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID
        ) {
            const isKanbanView =
                route.params.viewType === "kanban" ||
                route.name === "kanban" ||
                route.path?.includes("kanban");

            if (isKanbanView) {
                const sectionIndex = taskLists.value.findIndex(
                    (section) => String(section.id) === String(fromSocket.delete_id),
                );

                if (sectionIndex !== -1) {
                    const deletedSection = taskLists.value[sectionIndex];

                    let unassignedSection = taskLists.value.find(
                        (section) => String(section.id) === "0",
                    );

                    if (!unassignedSection) {
                        unassignedSection = {
                            id: "0",
                            title: "Unassigned",
                            tasks: [],
                        };
                        taskLists.value.push(unassignedSection);
                    }

                    if (deletedSection.tasks && deletedSection.tasks.length > 0) {
                        unassignedSection.tasks = [
                            ...(unassignedSection.tasks || []),
                            ...deletedSection.tasks,
                        ];
                    }

                    taskLists.value = taskLists.value.filter(
                        (section) => String(section.id) !== String(fromSocket.delete_id),
                    );
                }

                if (titles.value) {
                    titles.value = titles.value.filter(
                        (title) => String(title.id) !== String(fromSocket.delete_id),
                    );
                }

                taskLists.value = [...taskLists.value];
                if (titles.value) titles.value = [...titles.value];
            }

            if (realoadUI) {
                realoadUI.value += 1;
            }
        }
    }

    function onDeleteField(fromSocket) {
        const own = fromSocket.client_id === workspaceStore.getConnectionID;

        removeField(fromSocket.table_id, fromSocket.field_id, own);
        const linked = fromSocket.linked_data;

        if (linked?.table_id && linked?.parent_field_id) {
            removeField(linked.table_id, linked.parent_field_id, own);
        }
    }

    function onGridHeaderUpdate(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            route.params.fid === fromSocket.view_id &&
            fromSocket.client_id != workspaceStore.getConnectionID
        ) {
            let columnName, newWidth;

            if (fromSocket.column_name && fromSocket.width !== undefined) {
                columnName = fromSocket.column_name;
                newWidth = parseInt(fromSocket.width);
            } else if (Array.isArray(fromSocket.headers) && fromSocket.headers.length > 0) {
                columnName = fromSocket.headers[0].name;
                newWidth = parseInt(fromSocket.headers[0].width);
            } else {
                return;
            }

            const columnIndex = tableHeaders.value.findIndex(
                (header) => header.name === columnName,
            );

            if (columnIndex !== -1) {
                tableHeaders.value[columnIndex].width = String(newWidth);
                columnWidths.value[columnIndex] = newWidth;

                const styleWidth = `${newWidth}px`;
                let elements = document.querySelectorAll(`[data-column-name="${columnName}"]`);

                if (elements.length === 0) {
                    elements = document.querySelectorAll(
                        `.grid-column[data-column-name="${columnName}"], .grid-item-text[data-column-name="${columnName}"]`,
                    );
                }

                if (elements.length === 0 && columnIndex !== -1) {
                    elements = document.querySelectorAll(`[data-column-index="${columnIndex}"]`);
                }

                elements.forEach((el) => {
                    el.style.width = styleWidth;
                    el.style.minWidth = styleWidth;
                });
            }
        }
    }

    return {
        newFieldHeader,
        removeField,
        applyFieldFormula,
        onNewField,
        onDisplayNameUpdate,
        onHeaderOrderUpdate,
        onDeleteSingleField,
        onDeleteField,
        onGridHeaderUpdate,
    };
}
