// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

import { computed } from "vue";
import { countSubtasks } from "@/utils/projects/rows";

// How many subtasks each task in a list has, counted once per change.
export function useSubtaskCounts(rows) {
    const counts = computed(() => countSubtasks(rows.value));
    const hasChildren = (id) => counts.value.has(String(id));
    const getSubtaskCount = (id) => counts.value.get(String(id)) || 0;

    return { hasChildren, getSubtaskCount };
}
