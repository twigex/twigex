// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// viewChanges applies changes to a table's views and saved filters.
export function viewChanges(ctx) {
    const { route, router, workspaceStore, realoadUI, typesCreated, workspaceTables } = ctx;

    function onCreateGridView(fromSocket) {
        if (
            fromSocket.client_id !== workspaceStore.getConnectionID &&
            route.params.tid === fromSocket.table_id
        ) {
            const newView = fromSocket.data;

            const currentViewTypes = workspaceStore.getViewTypes || [];

            workspaceStore.setViewTypes([
                ...currentViewTypes,
                {
                    name: newView.name,
                    id: newView.id,
                    view_type: "grid",
                    is_public: newView.is_public,
                },
            ]);

            const tableIndex = workspaceTables.value?.findIndex(
                (t) => String(t.id) === String(fromSocket.table_id),
            );

            if (tableIndex !== -1 && workspaceTables.value[tableIndex]) {
                if (!workspaceTables.value[tableIndex].views) {
                    workspaceTables.value[tableIndex].views = [];
                }

                workspaceTables.value[tableIndex].views.push(newView);
                workspaceTables.value = [...workspaceTables.value];
            }

            if (realoadUI) {
                realoadUI.value += 1;
            }
        }
    }

    function onFilterSaved(fromSocket) {
        const tableIdx = workspaceTables.value.findIndex(
            (t) => String(t.id) === String(fromSocket.table_id),
        );

        if (tableIdx !== -1) {
            const table = workspaceTables.value[tableIdx];

            const patchViews = (arr) => {
                if (!Array.isArray(arr)) return arr;

                return arr.map((v) => {
                    if (String(v.id) !== String(fromSocket.view_id)) return v;

                    let currentFilters;

                    try {
                        currentFilters = Array.isArray(v.filters)
                            ? v.filters
                            : JSON.parse(v.filters || "[]");
                    } catch {
                        currentFilters = [];
                    }

                    let updatedFilters;

                    if (fromSocket.type === "FILTER_CREATED") {
                        updatedFilters = [...currentFilters, fromSocket.filter];
                    } else {
                        updatedFilters = currentFilters.map((f) =>
                            f.id === fromSocket.filter.id ? { ...f, ...fromSocket.filter } : f,
                        );
                    }

                    return {
                        ...v,
                        filters: JSON.stringify(updatedFilters),
                    };
                });
            };

            const nextViews = patchViews(table.views);

            workspaceTables.value[tableIdx] = {
                ...table,
                views: nextViews,
            };
            workspaceTables.value = [...workspaceTables.value];
        }

        if (
            route.params.tid === fromSocket.table_id &&
            route.params.fid === fromSocket.view_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID
        ) {
            const currentFilters = workspaceStore.getFilters || [];

            if (fromSocket.type === "FILTER_CREATED") {
                workspaceStore.setFilters([...currentFilters, fromSocket.filter]);
            } else {
                const updatedFilters = currentFilters.map((f) =>
                    f.id === fromSocket.filter.id ? { ...f, ...fromSocket.filter } : f,
                );

                workspaceStore.setFilters(updatedFilters);
            }
        }
    }

    function onDeleteView(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id !== workspaceStore.getConnectionID
        ) {
            typesCreated.value = typesCreated.value.filter(
                (item) => item.id !== fromSocket.delete_id,
            );
            if (route.params.fid == fromSocket.delete_id) {
                router.push({
                    name: "grid-view",
                    params: {
                        id: route.params.id,
                        tid: route.params.tid,
                    },
                });
            }
        }
    }

    function onViewVisibility(fromSocket) {
        if (route.params.tid === fromSocket.table_id) {
            typesCreated.value = typesCreated.value.map((view) =>
                view.id === fromSocket.view_id
                    ? { ...view, is_public: fromSocket.is_public }
                    : view,
            );
        }
    }

    function onViewCreated(fromSocket) {
        if (
            route.params.tid === fromSocket.table_id &&
            fromSocket.client_id != workspaceStore.getConnectionID
        ) {
            typesCreated.value.push(fromSocket.data);
        }
    }

    return { onCreateGridView, onFilterSaved, onDeleteView, onViewVisibility, onViewCreated };
}
