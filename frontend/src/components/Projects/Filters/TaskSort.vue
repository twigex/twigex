<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        :class="
            embedded
                ? 'space-y-5'
                : 'max-h-[80vh] min-w-[300px] space-y-5 overflow-y-auto rounded-md bg-white p-4'
        "
    >
        <h2 v-if="!embedded" class="text-base font-semibold text-gray-900">
            {{ t("projects.task_sort.sort_options") }}
        </h2>

        <div class="space-y-3">
            <div v-for="(sort, index) in sortOptions" :key="index" class="flex items-center gap-3">
                <BaseSelect
                    v-model="sort.field"
                    class="min-w-0 flex-1"
                    :options="fieldOptions"
                    :placeholder="t('projects.task_sort.select_column')"
                />

                <BaseButton
                    type="button"
                    variant="secondary"
                    class="w-24 flex-none"
                    @click="toggleDirection(index)"
                >
                    {{ sort.direction === "asc" ? "A → Z" : "Z → A" }}
                </BaseButton>
                <BaseButton
                    v-if="sortOptions.length > 1"
                    type="button"
                    variant="secondary"
                    :prepend-icon="TrashIcon"
                    :aria-label="t('common.button.remove')"
                    color="!shadow-none text-gray-400 hover:bg-red-50 hover:text-red-600"
                    @click="removeSort(index)"
                />
            </div>

            <div class="flex items-center gap-x-3">
                <BaseButton type="button" variant="soft" :prepend-icon="PlusIcon" @click="addSort">
                    {{ t("projects.task_sort.add_sort_options") }}
                </BaseButton>
                <BaseButton v-if="embedded" type="button" variant="secondary" @click="resetRows">
                    {{ t("projects.filter_builder.clear") }}
                </BaseButton>
            </div>
        </div>

        <div
            v-if="!embedded"
            class="flex items-center justify-between gap-x-3 border-t border-gray-200 pt-4"
        >
            <BaseButton type="button" variant="secondary" @click="clearSort">
                {{ t("projects.filter_builder.clear") }}
            </BaseButton>
            <div class="flex gap-x-3">
                <BaseButton type="button" variant="secondary" @click="$emit('close')">
                    {{ t("common.button.cancel") }}
                </BaseButton>
                <BaseButton type="button" @click="applySort">
                    {{ t("projects.task_sort.sort") }}
                </BaseButton>
            </div>
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { computed } from "vue";
import { PlusIcon, TrashIcon } from "@heroicons/vue/20/solid";
import BaseButton from "@/components/BaseButton.vue";
import BaseSelect from "@/components/BaseSelect.vue";
import { useWorkspaceStore } from "@/store/workspaces";
import { useRoute } from "vue-router";
import workspaceService from "@/services/workspaceService";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";

const workspaceStore = useWorkspaceStore();
const route = useRoute();

// Part of a larger form, which applies the sort itself.
defineProps({
    embedded: { type: Boolean, default: false },
});

const emit = defineEmits(["close", "sort"]);

// Computed table data binding
// Table headers
const tableHeaders = computed(() => workspaceStore.getTableHeaders || []);

const fieldOptions = computed(() =>
    tableHeaders.value.map((header) => ({
        value: header.name,
        label: header.display_name || header.name,
    })),
);

// Sort options from store
const sortOptions = computed({
    get() {
        return workspaceStore.getSortOptions;
    },
    set(value) {
        workspaceStore.setSortOptions(value);
    },
});

// Add sort option
function addSort() {
    sortOptions.value.push({ field: "", direction: "asc" });
}

// Remove sort option
function removeSort(index) {
    sortOptions.value.splice(index, 1);
}

// Toggle direction
function toggleDirection(index) {
    const option = sortOptions.value[index];

    option.direction = option.direction === "asc" ? "desc" : "asc";
}

const sortActive = computed({
    get() {
        return workspaceStore.getSortActive;
    },
    set(value) {
        workspaceStore.setSortActive(value);
    },
});

// Only a table's view keeps its sort; the task report has none to save to.
function saveViewSort(sorts) {
    if (!route.params.tid) return;
    workspaceService
        .saveGridSort({
            workspace_id: route.params.id,
            table_id: route.params.tid,
            view_id: route.params.fid,
            sort: JSON.stringify(sorts),
        })
        .catch((error) => {
            useAlertStore().showError(extractErrorMessage(error));
        });
}

function resetRows() {
    sortOptions.value = [{ field: "", direction: "asc" }];
}

function clearSort() {
    resetRows();
    saveViewSort([]);
    sortActive.value = false;
    workspaceStore.bumpGridReloadToken();
    emit("sort", []);
    emit("close");
}

function applySort() {
    const validSorts = sortOptions.value.filter((s) => s.field);

    if (!validSorts.length) return;

    saveViewSort(validSorts);

    sortActive.value = true;
    // Signal GridView to reload page 1 from server with the new sort
    workspaceStore.bumpGridReloadToken();
    emit("sort", validSorts);
    emit("close");
}
</script>
