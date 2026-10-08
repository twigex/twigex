<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div
        :class="
            embedded
                ? 'space-y-5'
                : 'max-h-[80vh] min-w-[700px] space-y-5 overflow-y-auto rounded-md bg-white p-4'
        "
    >
        <div class="flex items-center justify-between gap-x-3">
            <h2 class="text-base font-semibold text-gray-900">
                {{ title || t("projects.filter_builder.title") }}
            </h2>
            <div class="flex items-center gap-x-2">
                <BaseButton
                    v-if="groups.length === 0"
                    type="button"
                    variant="secondary"
                    :prepend-icon="PlusIcon"
                    @click="addFilter"
                >
                    {{ t("projects.filter_builder.add_condition") }}
                </BaseButton>
                <BaseButton
                    type="button"
                    variant="secondary"
                    :prepend-icon="PlusIcon"
                    @click="addGroup"
                >
                    {{ t("projects.filter_builder.add_group") }}
                </BaseButton>
            </div>
        </div>

        <div v-if="groups.length === 0" class="space-y-3">
            <div v-for="(filter, index) in flatFilters" :key="index" class="space-y-1">
                <div v-if="index > 0" class="flex justify-center mb-2">
                    <BaseButton
                        type="button"
                        variant="secondary"
                        class="min-w-[4rem]"
                        @click="
                            filter.operatorBetween = filter.operatorBetween === 'AND' ? 'OR' : 'AND'
                        "
                    >
                        {{
                            filter.operatorBetween === "OR"
                                ? t("projects.filter_builder.or")
                                : t("projects.filter_builder.and")
                        }}
                    </BaseButton>
                </div>
                <FilterRow
                    :filter="filter"
                    @remove="removeFlatFilter(index)"
                    :users="matchedUsers"
                    :search-users="searchUsers"
                />
            </div>
        </div>

        <div v-else class="space-y-6">
            <div v-for="(group, groupIndex) in groups" :key="groupIndex">
                <div class="border border-gray-300 p-4 rounded space-y-3">
                    <div class="flex items-center justify-between">
                        <div class="text-sm font-semibold text-gray-600">
                            {{ t("projects.filter_builder.group") }}
                            {{ groupIndex + 1 }}
                        </div>
                        <BaseButton
                            type="button"
                            variant="secondary"
                            :prepend-icon="TrashIcon"
                            :aria-label="t('projects.filter_builder.remove_group')"
                            :title="t('projects.filter_builder.remove_group')"
                            color="!shadow-none text-gray-400 hover:bg-red-50 hover:text-red-600"
                            @click="removeGroup(groupIndex)"
                        />
                    </div>

                    <div
                        v-for="(filter, filterIndex) in group.filters"
                        :key="filterIndex"
                        class="space-y-1"
                    >
                        <div v-if="filterIndex > 0" class="flex justify-center mb-2">
                            <BaseButton
                                type="button"
                                variant="secondary"
                                class="min-w-[4rem]"
                                @click="
                                    filter.operatorBetween =
                                        filter.operatorBetween === 'AND' ? 'OR' : 'AND'
                                "
                            >
                                {{
                                    filter.operatorBetween === "OR"
                                        ? t("projects.filter_builder.or")
                                        : t("projects.filter_builder.and")
                                }}
                            </BaseButton>
                        </div>
                        <FilterRow
                            :filter="filter"
                            @remove="removeFilterFromGroup(groupIndex, filterIndex)"
                            :users="matchedUsers"
                            :search-users="searchUsers"
                            :statusOptions="allStatusOptions"
                        />
                    </div>

                    <BaseButton
                        type="button"
                        variant="soft"
                        :prepend-icon="PlusIcon"
                        @click="addFilterToGroup(groupIndex)"
                    >
                        {{ t("projects.filter_builder.add_condition") }}
                    </BaseButton>
                </div>

                <div
                    v-if="groupIndex < groups.length - 1"
                    class="flex justify-center items-center my-2 mt-5"
                >
                    <BaseButton
                        type="button"
                        variant="secondary"
                        class="min-w-[4rem]"
                        @click="
                            groups[groupIndex].nextRelation =
                                groups[groupIndex].nextRelation === 'AND' ? 'OR' : 'AND'
                        "
                    >
                        {{
                            groups[groupIndex].nextRelation === "OR"
                                ? t("projects.filter_builder.or")
                                : t("projects.filter_builder.and")
                        }}
                    </BaseButton>
                </div>
            </div>
        </div>

        <div v-if="!embedded" class="flex items-center justify-between gap-2 pt-4 border-t">
            <BaseButton type="button" variant="secondary" @click="clearAppliedFilters">
                {{ t("projects.filter_builder.clear") }}
            </BaseButton>
            <div class="flex items-center gap-x-3">
                <BaseButton type="button" variant="secondary" @click="emit('close')">
                    {{ t("common.button.cancel") }}
                </BaseButton>
                <BaseButton
                    v-if="!loadedFilter?.name"
                    type="button"
                    variant="secondary"
                    @click="openSaveFilterDialog"
                >
                    {{ t("projects.filter_builder.save_as_filter") }}
                </BaseButton>
                <template v-else-if="isEdited">
                    <BaseButton type="button" variant="secondary" @click="saveAsNew">
                        {{ t("projects.filter_builder.save_as_new_filter_ellipsis") }}
                    </BaseButton>
                    <BaseButton type="button" variant="secondary" @click="saveChanges">
                        {{ t("projects.filter_builder.save_changes") }}
                    </BaseButton>
                </template>
                <BaseButton type="button" @click="apply">
                    {{ t("common.button.apply") }}
                </BaseButton>
            </div>
        </div>
        <TransitionRoot appear :show="showFilterDialog" as="template">
            <Dialog as="div" @close="closeFilterDialog" class="relative z-50">
                <TransitionChild
                    as="template"
                    enter="duration-300 ease-out"
                    enter-from="opacity-0"
                    enter-to="opacity-100"
                    leave="duration-200 ease-in"
                    leave-from="opacity-100"
                    leave-to="opacity-0"
                >
                    <div class="fixed inset-0 bg-gray-500 bg-opacity-75 transition-opacity" />
                </TransitionChild>

                <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
                    <div
                        class="flex min-h-full items-end justify-center p-4 text-center sm:items-center sm:p-0"
                    >
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
                                class="relative w-full transform overflow-hidden rounded-lg bg-white px-4 pb-4 pt-5 text-left shadow-xl transition-all sm:my-8 sm:max-w-lg sm:p-6"
                            >
                                <DialogTitle
                                    as="h3"
                                    class="text-base font-semibold leading-6 text-gray-900"
                                >
                                    {{
                                        dialogMode === "save"
                                            ? t("projects.filter_builder.save_filter")
                                            : t("projects.filter_builder.update_filter")
                                    }}
                                </DialogTitle>

                                <div class="mt-4">
                                    <label
                                        class="block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("projects.filter_builder.filter_name") }}
                                    </label>
                                    <input
                                        v-model="dialogFilterName"
                                        type="text"
                                        class="mt-2 block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                        :placeholder="
                                            t('projects.filter_builder.error.enter_filter_name')
                                        "
                                    />
                                </div>

                                <div class="mt-4">
                                    <label
                                        class="mb-2 block text-sm font-medium leading-6 text-gray-900"
                                    >
                                        {{ t("projects.filter_builder.filter_type") }}
                                    </label>
                                    <div class="space-y-2">
                                        <label class="flex items-center">
                                            <input
                                                type="radio"
                                                v-model="dialogFilterType"
                                                value="private"
                                                class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                            <span class="ml-2 text-sm text-gray-700">{{
                                                t("projects.filter_builder.private_only_me")
                                            }}</span>
                                        </label>
                                        <label class="flex items-center">
                                            <input
                                                type="radio"
                                                v-model="dialogFilterType"
                                                value="collaborative"
                                                class="h-4 w-4 border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                            />
                                            <span class="ml-2 text-sm text-gray-700">{{
                                                t(
                                                    "projects.filter_builder.collaborative_all_workspace_members",
                                                )
                                            }}</span>
                                        </label>
                                    </div>
                                </div>

                                <div
                                    v-if="
                                        dialogMode === 'save' &&
                                        route.name !== 'detailed-task-report'
                                    "
                                    class="mt-4 flex items-center gap-x-2"
                                >
                                    <input
                                        id="filterMakeDefault"
                                        v-model="dialogMakeDefault"
                                        type="checkbox"
                                        class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                    <label for="filterMakeDefault" class="text-sm text-gray-700">
                                        {{
                                            t(
                                                "projects.filter_builder.set_as_default_filter_for_this_view",
                                            )
                                        }}
                                    </label>
                                </div>

                                <div class="mt-5 sm:mt-6 sm:flex sm:flex-row-reverse">
                                    <BaseButton
                                        type="button"
                                        class="w-full sm:ml-3 sm:w-auto"
                                        :is-disabled="!dialogFilterName"
                                        :is-loading="isDialogSubmitting"
                                        @click="handleFilterDialogSubmit"
                                    >
                                        {{
                                            isDialogSubmitting
                                                ? dialogMode === "save"
                                                    ? t("projects.filter_builder.saving")
                                                    : t("projects.filter_builder.updating")
                                                : dialogMode === "save"
                                                  ? t("common.button.save")
                                                  : t("common.button.update")
                                        }}
                                    </BaseButton>
                                    <BaseButton
                                        type="button"
                                        variant="secondary"
                                        class="mt-3 w-full sm:mt-0 sm:w-auto"
                                        @click="closeFilterDialog"
                                    >
                                        {{ t("common.button.cancel") }}
                                    </BaseButton>
                                </div>
                            </DialogPanel>
                        </TransitionChild>
                    </div>
                </div>
            </Dialog>
        </TransitionRoot>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted, toRaw, watch } from "vue";
import FilterRow from "./FilterRow.vue";
import BaseButton from "@/components/BaseButton.vue";
import { PlusIcon, TrashIcon } from "@heroicons/vue/20/solid";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { useRoute } from "vue-router";
import { useUserStore } from "@/store/user";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import {
    filterConditionsKey,
    filtersForSave,
    withoutFilterOptions,
} from "@/utils/projects/filterCombine";
import { useTableFilters } from "@/composables/projects/useTableFilters";
import { TransitionRoot, TransitionChild, Dialog, DialogPanel, DialogTitle } from "@headlessui/vue";

const showFilterDialog = ref(false);
const dialogMode = ref("save");
const dialogFilterName = ref("");
const dialogFilterType = ref("private");
const isDialogSubmitting = ref(false);
const dialogMakeDefault = ref(false);
const allStatusOptions = ref(null);

const props = defineProps({
    title: { type: String, default: "" },
    // Starts a new filter from no conditions, leaving the loaded one aside.
    creating: { type: Boolean, default: false },
    // Part of a larger form, which applies the conditions itself: no footer,
    // and every edit is passed up as change.
    embedded: { type: Boolean, default: false },
    // The task report's workspaces to look for people in, when a form has
    // chosen them before the report has run.
    workspaceIds: { type: Array, default: null },
});

const emit = defineEmits(["apply", "close", "change", "loaded"]);
const workspaceStore = useWorkspaceStore();
const flatFilters = ref([{ field: "", operator: "is", value: "", operatorBetween: "AND" }]);
const userStore = useUserStore();
const route = useRoute();
const groups = ref([]);
const matchedUsers = ref([]);

const filterActive = computed({
    get() {
        return workspaceStore.getFilterActive;
    },
    set(value) {
        workspaceStore.setFilterActive(value);
    },
});

const fullFilter = computed({
    get() {
        return workspaceStore.getFilter;
    },
    set(value) {
        workspaceStore.setFilter(value);
    },
});

const { setViewDefaultFilter } = useTableFilters();

const loadedFilter = computed(() => (props.creating ? null : fullFilter.value));

// Whether the rows differ from the loaded saved filter, so there is
// something to save into it.
const isEdited = computed(
    () =>
        !!loadedFilter.value?.name &&
        filterConditionsKey(collectPayload()) !== filterConditionsKey(loadedFilter.value.filters),
);

// People are searched a page at a time rather than all loaded: a table's
// among its workspace's members, the task report's among the members of the
// workspaces it covers.
async function searchUsers(query) {
    const workspaceId = route.params.id;
    const res = workspaceId
        ? await workspaceService.searchWorkspaceMembers(workspaceId, query, 50, 0)
        : await workspaceService.searchTaskReportMembers(
              props.workspaceIds ?? workspaceStore.getSelectedWorkspaces ?? [],
              query,
              50,
          );
    const users = res.data ?? [];

    userStore.addUsers(users);

    return users;
}

const USER_FIELDS = new Set(["assignee", "default_assignee", "created_by"]);

function chosenUserIds() {
    const all = [...flatFilters.value, ...groups.value.flatMap((g) => g.filters || [])];

    return [
        ...new Set(
            all
                .filter((f) => USER_FIELDS.has(f.field) && typeof f.value === "string" && f.value)
                .map((f) => f.value),
        ),
    ];
}

// The first page of people, and any a condition already names, loaded by id
// so it shows their name.
let memberLoad = 0;

async function fetchMembers() {
    const load = ++memberLoad;

    try {
        const users = await searchUsers("");
        const chosen = chosenUserIds();

        await userStore.ensureUsers(chosen);
        if (load !== memberLoad) return;

        const listed = new Set(users.map((u) => u.id));

        matchedUsers.value = [
            ...users,
            ...chosen
                .filter((id) => !listed.has(id))
                .map((id) => userStore.getUserById(id))
                .filter(Boolean),
        ];
    } catch {
        if (load === memberLoad) matchedUsers.value = userStore.users;
    }
}

watch(
    () => props.workspaceIds,
    () => fetchMembers(),
    { deep: true },
);

// Status conditions are stored by id, and each table has its own ids for a
// status, so the report matches them by name. Without the names the
// conditions load as saved.
async function reportStatusNames() {
    try {
        const res = await workspaceService.getStatusList({ workspace_ids: [] });

        return res.data.name_to_ids || {};
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));

        return null;
    }
}

onMounted(async () => {
    // For DTR: resolve status names before rendering filters to avoid UUID flicker.
    // For other routes: set immediately (linkedOptions resolved in FilterRow from store).
    let rawFilters = props.creating
        ? [{ field: "", operator: "is", value: "", operatorBetween: "AND" }]
        : JSON.parse(JSON.stringify(workspaceStore.getSavedFlatFilters || []));

    const savedGroups = props.creating ? [] : workspaceStore.getSavedGroups || [];

    groups.value = savedGroups.map((group) => ({
        filters: group.filters || [],
        nextRelation: group.relationToNext || "AND",
    }));

    const nameToIds = route.name === "detailed-task-report" ? await reportStatusNames() : null;

    if (nameToIds) {
        const statusOptionsArray = Object.entries(nameToIds).map(([name, ids]) => ({
            id: ids,
            name: name,
        }));

        workspaceStore.setStatusOptions(statusOptionsArray);

        allStatusOptions.value = statusOptionsArray;

        let processedFilters = rawFilters.map((f) => {
            if (f.field === "status") {
                const filterValues = f.values || (Array.isArray(f.value) ? f.value : null);

                if (filterValues && filterValues.length > 0) {
                    let statusName = null;

                    for (const [name, ids] of Object.entries(nameToIds)) {
                        if (ids.includes(filterValues[0])) {
                            statusName = name;
                            break;
                        }
                    }

                    if (statusName) {
                        return {
                            ...f,
                            values: nameToIds[statusName],
                            value: nameToIds[statusName],
                            linkedOptions: statusOptionsArray,
                        };
                    }
                }

                return {
                    ...f,
                    linkedOptions: statusOptionsArray,
                };
            }

            return { ...f };
        });

        flatFilters.value = processedFilters;

        if (groups.value.length > 0) {
            groups.value = groups.value.map((group) => ({
                ...group,
                filters: group.filters.map((f) => {
                    if (f.field === "status") {
                        const filterValues = f.values || (Array.isArray(f.value) ? f.value : null);

                        if (filterValues && filterValues.length > 0) {
                            let statusName = null;

                            for (const [name, ids] of Object.entries(nameToIds)) {
                                if (ids.includes(filterValues[0])) {
                                    statusName = name;
                                    break;
                                }
                            }

                            if (statusName) {
                                return {
                                    ...f,
                                    values: nameToIds[statusName],
                                    value: nameToIds[statusName],
                                    linkedOptions: statusOptionsArray,
                                };
                            }
                        }

                        return {
                            ...f,
                            linkedOptions: statusOptionsArray,
                        };
                    }

                    return f;
                }),
            }));
        }
    } else {
        flatFilters.value = rawFilters;
    }

    // The conditions as loaded, once status values have their names, are what
    // an embedding form compares edits against.
    if (props.embedded) emit("loaded", collectPayload());

    fetchMembers();
});

function removeFlatFilter(index) {
    flatFilters.value.splice(index, 1);
}

function addGroup() {
    if (groups.value.length === 0 && flatFilters.value.length > 0) {
        groups.value.push({
            filters: [...flatFilters.value],
            nextRelation: "AND",
        });
        flatFilters.value = [];
    }

    groups.value.push({
        filters: [
            {
                field: "",
                operator: "is",
                value: "",
                operatorBetween: "AND",
            },
        ],
        nextRelation: "AND",
    });
}

function addFilter() {
    const newFilter = {
        field: "",
        operator: "is",
        value: "",
        operatorBetween: "AND",
    };

    const statusOptions = toRaw(workspaceStore.getStatusOptions);

    if (route.name === "detailed-task-report" && statusOptions && statusOptions.length > 0) {
        newFilter.linkedOptions = statusOptions.map((opt) => ({
            id: Array.isArray(opt.id) ? [...opt.id] : opt.id,
            name: opt.name,
        }));
    }

    if (groups.value.length === 0) {
        const current = toRaw(flatFilters.value);

        flatFilters.value = [...current, newFilter];
    } else {
        const group = toRaw(groups.value[groups.value.length - 1]);

        group.filters = [...(group.filters || []), newFilter];
    }
}

function addFilterToGroup(index) {
    const newFilter = {
        field: "",
        operator: "is",
        value: "",
        operatorBetween: "AND",
    };

    const statusOptions = workspaceStore.getStatusOptions;

    if (route.name === "detailed-task-report" && statusOptions && statusOptions.length > 0) {
        newFilter.linkedOptions = statusOptions;
    }

    groups.value[index].filters.push(newFilter);
}

// With no group left, the conditions go back to a plain list.
function removeGroup(index) {
    groups.value.splice(index, 1);
    if (groups.value.length === 0) {
        flatFilters.value = [{ field: "", operator: "is", value: "", operatorBetween: "AND" }];
    }
}

function removeFilterFromGroup(groupIndex, filterIndex) {
    groups.value[groupIndex].filters.splice(filterIndex, 1);
}

watch(
    [flatFilters, groups],
    () => {
        if (props.embedded) emit("change", collectPayload());
    },
    { deep: true },
);

function collectPayload() {
    return filtersForSave(flatFilters.value, groups.value);
}

function apply() {
    const payload = collectPayload();

    workspaceStore.setSavedFlatFilters(payload.flatFilters);
    workspaceStore.setSavedGroups(payload.groups);

    filterActive.value = payload.flatFilters.length > 0 || payload.groups.length > 0;

    emit("apply", payload);
}

function saveAsNew() {
    dialogMode.value = "save";
    dialogFilterName.value = fullFilter.value?.name ? `${fullFilter.value.name} (copy)` : "";
    dialogFilterType.value = fullFilter.value?.is_private ? "private" : "collaborative";
    dialogMakeDefault.value = false;
    showFilterDialog.value = true;
}

function openSaveFilterDialog() {
    dialogMode.value = "save";
    dialogFilterName.value = "";
    dialogFilterType.value = "private";
    dialogMakeDefault.value = false;
    showFilterDialog.value = true;
}

// Saves the conditions into the loaded filter, keeping its name and who sees it.
function saveChanges() {
    if (!fullFilter.value) return;

    dialogMode.value = "update";
    dialogFilterName.value = fullFilter.value.name;
    dialogFilterType.value = fullFilter.value.is_private ? "private" : "collaborative";
    handleFilterDialogSubmit();
}

function closeFilterDialog() {
    showFilterDialog.value = false;
    dialogFilterName.value = "";
    dialogFilterType.value = "private";
    isDialogSubmitting.value = false;
}

const isReport = () => route.name === "detailed-task-report";

// filterTarget is where a filter is saved: the view open, or the report.
function filterTarget() {
    return isReport()
        ? { workspace_id: "all", table_id: "detailed-task-report", view_id: "detailed-task-report" }
        : { workspace_id: route.params.id, table_id: route.params.tid, view_id: route.params.fid };
}

async function saveNewFilter(data, payload) {
    const response = await workspaceService.saveFilter(data);
    const newFilter = {
        id: response.data?.id || `temp-${Date.now()}`,
        filter_id: response.data?.filter_id || `temp-${Date.now()}`,
        name: data.name,
        is_private: data.is_private,
        table_id: data.table_id,
        view_id: data.view_id,
        filters: payload,
        is_active: false,
    };

    workspaceStore.setFilters([...(workspaceStore.getFilters || []), newFilter]);
    workspaceStore.setFilter(newFilter);

    if (dialogMakeDefault.value && !isReport()) {
        await setViewDefaultFilter(newFilter, true, {
            workspaceId: route.params.id,
            tableId: route.params.tid,
            viewId: route.params.fid,
        });
    }
}

async function updateLoadedFilter(data, payload) {
    const loaded = fullFilter.value;

    await workspaceService.updateFilter({
        ...data,
        filter_id: loaded.filter_id,
        saved_filter_id: loaded.id,
        is_active: true,
    });

    const changes = { name: data.name, is_private: data.is_private, filters: payload };

    workspaceStore.setFilter({ ...loaded, ...changes });
    workspaceStore.setFilters(
        (workspaceStore.getFilters || []).map((f) =>
            f.id === loaded.id ? { ...f, ...changes } : f,
        ),
    );
    useAlertStore().showSuccess(
        t.value("projects.filter_builder.default_filter_updated_successfully"),
    );
}

async function handleFilterDialogSubmit() {
    if (!dialogFilterName.value) {
        useAlertStore().showError(t.value("projects.filter_builder.dfilter_name_is_required"));

        return;
    }

    isDialogSubmitting.value = true;
    const saving = dialogMode.value === "save";
    const payload = filtersForSave(flatFilters.value, groups.value, { report: isReport() });
    const data = {
        ...filterTarget(),
        name: dialogFilterName.value,
        type: dialogFilterType.value,
        filters: withoutFilterOptions(payload),
        is_private: dialogFilterType.value === "private",
    };

    try {
        if (saving) await saveNewFilter(data, payload);
        else await updateLoadedFilter(data, payload);
        closeFilterDialog();
        apply();
    } catch (error) {
        const fallback = saving
            ? "projects.filter_builder.failed_to_save_filter"
            : "projects.filter_builder.failed_to_update_filter";

        useAlertStore().showError(extractErrorMessage(error, t.value(fallback)));
    } finally {
        isDialogSubmitting.value = false;
    }
}

function clearAppliedFilters() {
    flatFilters.value = [];
    groups.value = [];

    const payload = {
        groups: [],
        flatFilters: [],
    };

    emit("apply", payload);
}
</script>
