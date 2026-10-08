// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index.js";
import { PRESET_LABELS } from "@/composables/projects/useTaskReportFilters";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { useUserStore } from "@/store/user";
import { extractErrorMessage } from "@/utils/errors";
import { useWorkspaceStore } from "@/store/workspaces";

// A table's filter presets. Working one out and putting it in the store are
// separate, so a caller whose page changed meanwhile can drop the result.
export function useTableFilters() {
    const workspaceStore = useWorkspaceStore();
    const userStore = useUserStore();

    async function loadTablePreset(preset, workspaceId, tableId) {
        if (preset === "all_tasks") return { preset, flatFilters: [], active: false };
        if (preset === "my_tasks") {
            return {
                preset,
                flatFilters: [
                    {
                        field: "assignee",
                        operator: "is",
                        value: userStore.user.id,
                        operatorBetween: "AND",
                    },
                ],
                active: true,
            };
        }

        // Each table has its own Done and Closed options.
        const res = await workspaceService.getTableStatusTypes(workspaceId, tableId);
        const flatFilters = (res.data.done_ids || []).map((id) => ({
            field: "status",
            operator: "is_not",
            value: id,
            operatorBetween: "AND",
        }));

        if (preset === "late_tasks") {
            flatFilters.push({
                field: "due_date",
                operator: "is",
                value: "Overdue",
                operatorBetween: "AND",
            });

            return { preset, flatFilters, active: true };
        }

        return { preset, flatFilters, active: flatFilters.length > 0 };
    }

    function storeTablePreset({ preset, flatFilters, active }) {
        workspaceStore.setFilter(null);
        workspaceStore.setSavedFlatFilters(flatFilters);
        workspaceStore.setSavedGroups([]);
        workspaceStore.setDefaultFlatFilters(flatFilters);
        workspaceStore.setDefaultGroupFilters([]);
        workspaceStore.setFilterActive(active);
        workspaceStore.setFastFilter(
            preset,
            PRESET_LABELS[preset] ? t.value(PRESET_LABELS[preset]) : "",
        );
    }

    // A view has at most one default filter, which loads when it opens.
    async function setViewDefaultFilter(filter, isDefault, { workspaceId, tableId, viewId }) {
        try {
            await workspaceService.updateFilterActiveStatus({
                workspace_id: workspaceId,
                table_id: tableId,
                view_id: viewId,
                filter_id: filter.id,
                is_active: isDefault,
            });
        } catch (error) {
            useAlertStore().showError(
                extractErrorMessage(
                    error,
                    t.value("projects.filter_builder.failed_to_set_default_filter"),
                ),
            );

            return false;
        }

        workspaceStore.setFilters(
            (workspaceStore.getFilters || []).map((f) =>
                f.table_id === tableId && f.view_id === viewId
                    ? { ...f, is_active: isDefault && f.id === filter.id }
                    : f,
            ),
        );
        useAlertStore().showSuccess(
            t.value(
                isDefault
                    ? "projects.filter_builder.default_filter_set_successfully"
                    : "projects.filter_builder.default_filter_removed",
            ),
        );

        return true;
    }

    return { loadTablePreset, storeTablePreset, setViewDefaultFilter };
}
