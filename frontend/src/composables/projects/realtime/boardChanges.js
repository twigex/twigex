// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// boardChanges applies changes to a board's columns and the order of its
// cards.
export function boardChanges(ctx) {
    const {
        route,
        workspaceStore,
        tableHeaders,
        taskLists,
        titles,
        selectedView,
        workspaceTables,
    } = ctx;

    // A board's columns changed: their order, names or widths, or which fields
    // its cards show. Cards stay put.
    function applyKanbanColumns(saved) {
        let order;

        try {
            order = typeof saved === "string" ? JSON.parse(saved) : saved;
        } catch {
            return;
        }

        if (Array.isArray(order?.fieldsVisible)) {
            const shown = new Map(order.fieldsVisible.map((f) => [f.name, f.visible]));

            tableHeaders.value = tableHeaders.value.map((h) =>
                shown.has(h.name) && shown.get(h.name) !== h.visible
                    ? { ...h, visible: shown.get(h.name) }
                    : h,
            );
        }

        const columns = Array.isArray(order?.fields) ? order.fields : [];

        if (columns.length === 0) return;

        const titlesById = new Map(titles.value.map((t) => [String(t.id), t]));
        const reordered = columns
            .map((c) => {
                const title = titlesById.get(String(c.id));

                if (title) {
                    if (c.width !== undefined) title.width = c.width;
                    if (c.display_name !== undefined) title.display_name = c.display_name;
                }

                return title;
            })
            .filter(Boolean);

        if (reordered.length === titles.value.length) titles.value = reordered;
    }

    // Someone moved a card. It is placed after the card named, when this board
    // has that card loaded; otherwise it now sits below what this board has
    // loaded of the column, and is only taken out of where it was.
    function moveKanbanCardLocally(taskId, columnId, afterId) {
        let card = null;

        for (const list of taskLists.value) {
            const i = (list.tasks || []).findIndex((t) => String(t.id) === String(taskId));

            if (i !== -1) {
                card = list.tasks[i];
                list.tasks.splice(i, 1);
                break;
            }
        }

        const target = taskLists.value.find((l) => String(l.id) === String(columnId));

        if (!card || !target) return;
        target.tasks = target.tasks || [];
        if (!afterId) {
            target.tasks.unshift(card);

            return;
        }

        const after = target.tasks.findIndex((t) => String(t.id) === String(afterId));

        if (after !== -1) target.tasks.splice(after + 1, 0, card);
    }

    function onNewKanbanSection(fromSocket) {
        const sd = fromSocket.data || {};
        const normalizedSection = {
            id: sd.id,
            name: sd.name ?? "",
            display_name:
                typeof sd.display_name === "string" && sd.display_name.trim() !== ""
                    ? sd.display_name
                    : (sd.name ?? ""),
            visible: typeof sd.visible === "boolean" ? sd.visible : true,
            width: String(sd.width ?? "300"),
            color: sd.color || "#FFFFFF",
            order: Array.isArray(sd.order) ? sd.order : [],
        };

        const parentIdx = workspaceTables.value.findIndex(
            (t) => String(t.id) === String(fromSocket.parent_table_id),
        );

        if (parentIdx !== -1) {
            const pTable = workspaceTables.value[parentIdx];
            const pRows = Array.isArray(pTable.data_base) ? pTable.data_base : [];
            const existsInParent =
                pRows.some((r) => String(r.id) === String(normalizedSection.id)) ||
                pRows.some((r) => String(r.name) === String(normalizedSection.name));

            if (!existsInParent) {
                const optionRow = {
                    id: normalizedSection.id,
                    name: normalizedSection.name,
                    display_name: normalizedSection.display_name,
                    color: normalizedSection.color,
                    width: normalizedSection.width,
                    visible: normalizedSection.visible,
                };

                workspaceTables.value[parentIdx] = {
                    ...pTable,
                    data_base: pRows.concat(optionRow),
                };
            }
        }

        const tableIdx = workspaceTables.value.findIndex(
            (t) => String(t.id) === String(fromSocket.table_id),
        );
        const safeParseOrder = (ord) => {
            if (typeof ord === "string") {
                try {
                    const parsed = JSON.parse(ord);

                    if (Array.isArray(parsed))
                        return {
                            fields: parsed,
                            section: undefined,
                        };
                    if (parsed && typeof parsed === "object")
                        return {
                            fields: Array.isArray(parsed.fields) ? parsed.fields : [],
                            section: parsed.section,
                        };
                } catch {}

                return { fields: [], section: undefined };
            }

            if (Array.isArray(ord)) return { fields: ord, section: undefined };
            if (ord && typeof ord === "object")
                return {
                    fields: Array.isArray(ord.fields) ? ord.fields : [],
                    section: ord.section,
                };

            return { fields: [], section: undefined };
        };
        const sectionMatches = (sectionFieldName) =>
            sectionFieldName != null && String(sectionFieldName) === String(fromSocket.field);
        const byViewId = (v) =>
            fromSocket.view_id != null && String(v?.id) === String(fromSocket.view_id);
        const byParentId = (v) =>
            v?.parent_table_id != null &&
            String(v.parent_table_id) === String(fromSocket.parent_table_id);
        const collectTargets = (arr) => {
            if (!Array.isArray(arr)) return { targets: [], parsedById: new Map() };
            const parsedById = new Map();
            const explicit = [];
            const parent = [];
            const bySection = [];
            const fallbackKanban = [];

            arr.forEach((v) => {
                // A grid view's order is a list of columns too; given the
                // new option as a column, it would be a board's order.
                if (v?.view_type !== "kanban") return;
                const parsed = safeParseOrder(v.order);

                parsedById.set(String(v?.id ?? Math.random()), parsed);
                const mView = byViewId(v);
                const mParent = byParentId(v);
                const mSection = sectionMatches(parsed.section);
                const looksKanban =
                    Array.isArray(parsed.fields) &&
                    parsed.fields.every(
                        (f) => f && typeof f === "object" && ("id" in f || "name" in f),
                    );

                if (mView) explicit.push(v);
                if (mParent) parent.push(v);
                if (mSection) bySection.push(v);
                if (!parsed.section && looksKanban) fallbackKanban.push(v);
            });
            const targets =
                explicit.length > 0
                    ? explicit
                    : bySection.length > 0
                      ? bySection
                      : parent.length > 0
                        ? parent
                        : fallbackKanban.slice(0, 1);

            return { targets, parsedById };
        };
        const patchList = (arr, targets, parsedMap) => {
            if (!Array.isArray(arr) || targets.length === 0) return { next: arr, changed: false };
            const tset = new Set(targets.map((v) => v));
            let changed = false;
            const next = arr.map((v) => {
                if (!tset.has(v)) return v;
                const parsed = parsedMap.get(String(v?.id)) || safeParseOrder(v.order);
                const existsById = parsed.fields.some(
                    (f) => String(f?.id) === String(normalizedSection.id),
                );
                const existsByName = parsed.fields.some(
                    (f) => String(f?.name) === String(normalizedSection.name),
                );

                if (!existsById && !existsByName) {
                    parsed.fields.push(normalizedSection);
                    changed = true;
                }

                return {
                    ...v,
                    order: JSON.stringify({
                        fields: parsed.fields,
                        section: parsed.section,
                    }),
                };
            });

            return { next, changed };
        };

        if (tableIdx !== -1) {
            const table = workspaceTables.value[tableIdx];
            const { targets: targetViews, parsedById: parsedViews } = collectTargets(
                table.views || [],
            );
            const { targets: targetOptions, parsedById: parsedOptions } = collectTargets(
                table.options || [],
            );
            const { next: nextViews, changed: viewsChanged } = patchList(
                table.views || [],
                targetViews,
                parsedViews,
            );
            const { next: nextOptions, changed: optionsChanged } = patchList(
                table.options || [],
                targetOptions,
                parsedOptions,
            );

            if (viewsChanged || optionsChanged) {
                workspaceTables.value[tableIdx] = {
                    ...table,
                    ...(Array.isArray(nextViews) ? { views: nextViews } : {}),
                    ...(Array.isArray(nextOptions) ? { options: nextOptions } : {}),
                };
            }
        }

        workspaceTables.value = [...workspaceTables.value];

        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID &&
            selectedView.value?.parent_table_id === fromSocket.parent_table_id
        ) {
            const sectionData = fromSocket.data;

            const newSectionId = sectionData.id;
            const newSectionName = sectionData.name;

            let orderObj =
                typeof selectedView.value.order === "string"
                    ? JSON.parse(selectedView.value.order)
                    : selectedView.value.order;

            const newSection = {
                id: newSectionId,
                name: newSectionName,
                display_name: newSectionName,
                visible: true,
                width: String(sectionData.width ?? "300"),
                color: sectionData.color || "#FFFFFF",
                order: [],
            };

            if (Array.isArray(orderObj.fields)) {
                orderObj.fields.push(newSection);
                selectedView.value.order = orderObj;
            }

            if (Array.isArray(titles.value)) {
                titles.value.push({
                    id: newSectionId,
                    title: newSectionName,
                    display_name: newSectionName,
                    color: sectionData.color || "#FFFFFF",
                    width: String(sectionData.width ?? "300"),
                });
            }

            if (Array.isArray(taskLists.value)) {
                taskLists.value.push({
                    id: newSectionId,
                    title: newSectionName,
                    display_name: newSectionName,
                    tasks: [],
                });
            }
        }
    }

    function onKanbanSectionRename(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID &&
            selectedView.value?.parent_table_id === fromSocket.payload?.linked_table_id
        ) {
            const { task_id, value: newName } = fromSocket.payload;
            let orderObj =
                typeof selectedView.value.order === "string"
                    ? JSON.parse(selectedView.value.order)
                    : selectedView.value.order;

            if (Array.isArray(orderObj.fields)) {
                orderObj.fields = orderObj.fields.map((field) => {
                    if (field.id === task_id) {
                        return {
                            ...field,
                            name: newName,
                            display_name: newName,
                        };
                    }

                    return field;
                });
                selectedView.value.order = orderObj;
            }

            if (Array.isArray(titles.value)) {
                titles.value = titles.value.map((title) => {
                    if (title.id === task_id) {
                        return {
                            ...title,
                            title: newName,
                            display_name: newName,
                        };
                    }

                    return title;
                });
            }

            if (Array.isArray(taskLists.value)) {
                taskLists.value.forEach((column) => {
                    column.tasks?.forEach((task) => {
                        if (task.single1?.id === task_id) {
                            task.single1.name = newName;
                        }
                    });
                });

                taskLists.value = taskLists.value.map((col) => {
                    if (col.id === task_id) {
                        return {
                            ...col,
                            title: newName,
                            display_name: newName,
                        };
                    }

                    return col;
                });
            }
        }
    }

    function onKanbanLayout(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            route.params.fid === fromSocket.view_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID
        ) {
            if (fromSocket.type === "KANBAN_COLUMNS") applyKanbanColumns(fromSocket.order);
            else moveKanbanCardLocally(fromSocket.task_id, fromSocket.column, fromSocket.after_id);
        }
    }

    return {
        applyKanbanColumns,
        moveKanbanCardLocally,
        onNewKanbanSection,
        onKanbanSectionRename,
        onKanbanLayout,
    };
}
