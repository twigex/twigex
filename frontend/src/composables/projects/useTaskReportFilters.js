// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { t } from "@/i18n/index.js";
import workspaceService from "@/services/workspaceService";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";

export const PRESET_LABELS = {
    open_tasks: "projects.menu.task_filter_menu.filter_options.open_tasks",
    my_tasks: "projects.menu.task_filter_menu.filter_options.my_tasks",
    all_tasks: "projects.menu.task_filter_menu.filter_options.all_tasks",
    late_tasks: "projects.menu.task_filter_menu.filter_options.late_tasks",
};

// Puts the task report's filters in the store, from one of its presets or a
// saved filter. The caller loads the report.
export function useTaskReportFilters() {
    const workspaceStore = useWorkspaceStore();
    const userStore = useUserStore();

    function setFlatFilters(flatFilters, active) {
        workspaceStore.setSavedFlatFilters(flatFilters);
        workspaceStore.setSavedGroups([]);
        workspaceStore.setDefaultFlatFilters(flatFilters);
        workspaceStore.setDefaultGroupFilters([]);
        workspaceStore.setFilterActive(active);
    }

    // Only the latest choice is applied: an answer for one picked before it
    // must not overwrite it.
    let choice = 0;

    // applyPreset reports whether it was applied, which it is not when a
    // later choice came first.
    async function applyPreset(preset) {
        const mine = ++choice;
        const labelKey = PRESET_LABELS[preset];
        const choose = (flatFilters, active) => {
            workspaceStore.setFilter(null);
            workspaceStore.setFastFilter(preset, labelKey ? t.value(labelKey) : "");
            setFlatFilters(flatFilters, active);
        };

        if (preset === "all_tasks") {
            choose([], false);

            return true;
        }

        if (preset === "my_tasks") {
            choose(
                [
                    {
                        field: "assignee",
                        operator: "is",
                        value: userStore.user.id,
                        operatorBetween: "AND",
                    },
                ],
                true,
            );

            return true;
        }

        // Done and Closed are matched by name: each table has its own status
        // options, and the server widens a name to every table's ids for it.
        const res = await workspaceService.getStatusList();

        if (mine !== choice) return false;

        const flatList = res.data.flat_list || [];
        const nameToIds = res.data.name_to_ids || {};
        const doneClosedNames = new Set(
            flatList
                .filter((s) => s.status_type === "Done" || s.status_type === "Closed")
                .map((s) => s.name),
        );
        const statusFilters = [...doneClosedNames].map((name) => {
            const ids = nameToIds[name] || [];

            return {
                field: "status",
                operator: "is_not",
                value: ids[0] || "",
                values: ids,
                operatorBetween: "AND",
            };
        });

        if (preset === "late_tasks") {
            choose(
                [
                    ...statusFilters,
                    { field: "due_date", operator: "is", value: "Overdue", operatorBetween: "AND" },
                ],
                true,
            );
        } else {
            choose(statusFilters, statusFilters.length > 0);
        }

        return true;
    }

    function applySavedFilter(filter) {
        choice++;
        workspaceStore.setFilter(filter);
        workspaceStore.setFastFilter("saved_filter", filter.name);
        workspaceStore.setSavedFlatFilters(filter.filters?.flatFilters || []);
        workspaceStore.setSavedGroups(filter.filters?.groups || []);
        workspaceStore.setFilterActive(true);
    }

    return { applyPreset, applySavedFilter };
}
