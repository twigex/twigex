<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div v-if="loaded" class="w-full h-[93vh] flex flex-col bg-white">
        <div class="flex-grow overflow-y-auto">
            <div class="flex w-full overflow-y-auto justify-center">
                <div class="w-1/2 py-4">
                    <!-- Name & Description -->
                    <div>
                        <label
                            for="name"
                            class="block text-sm font-medium leading-6 text-gray-900"
                            >{{ t("projects.create_role.title") }}</label
                        >
                        <div class="my-2">
                            <input
                                v-model="name"
                                type="text"
                                name="name"
                                id="name"
                                :disabled="isBuiltIn"
                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                :placeholder="t('projects.create_role.title')"
                                :class="
                                    v$.name.$error
                                        ? 'ring-red-600 focus:ring-red-600'
                                        : 'ring-gray-300 focus:ring-indigo-600'
                                "
                            />
                            <p v-if="v$.name.$error" class="mt-2 text-sm text-red-600">
                                {{ t("common.error.required_field") }}
                            </p>
                        </div>

                        <label
                            for="description"
                            class="block text-sm font-medium leading-6 text-gray-900"
                            >{{ t("projects.create_role.description") }}</label
                        >
                        <div class="mt-2 mb-6">
                            <textarea
                                v-model="description"
                                rows="3"
                                name="description"
                                id="description"
                                :disabled="isBuiltIn"
                                class="block w-full rounded-md border-0 py-1.5 text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6"
                                :placeholder="t('projects.create_role.description')"
                            />
                        </div>
                    </div>

                    <!-- Section A: Capabilities -->
                    <h3 class="text-sm font-semibold text-gray-900 mb-1">
                        {{ t("projects.create_role.section.workspace_actions") }}
                    </h3>
                    <p class="text-xs text-gray-500 mb-4">
                        {{ t("projects.create_role.section.workspace_actions_hint") }}
                    </p>
                    <div class="space-y-4">
                        <div
                            v-for="group in permissionGroups.filter(
                                (g) => g.key !== 'rows' && g.key !== 'other' && g.items.length > 0,
                            )"
                            :key="group.key"
                            class="rounded-lg border border-gray-200 overflow-hidden"
                        >
                            <div
                                class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                            >
                                <span class="text-sm font-semibold text-gray-900">{{
                                    group.label
                                }}</span>
                            </div>
                            <div class="divide-y divide-gray-100">
                                <label
                                    v-for="perm in group.items"
                                    :key="perm.value"
                                    :for="'perm-' + perm.value"
                                    class="flex items-start gap-x-3 px-4 py-3 cursor-pointer"
                                    :class="
                                        permissions.includes(perm.value)
                                            ? 'bg-indigo-50'
                                            : 'hover:bg-gray-50'
                                    "
                                >
                                    <input
                                        :id="'perm-' + perm.value"
                                        v-model="permissions"
                                        :value="perm.value"
                                        type="checkbox"
                                        class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                    <span class="min-w-0">
                                        <span class="block text-sm font-medium text-gray-900">{{
                                            perm.label
                                        }}</span>
                                        <span class="block text-xs text-gray-500 mt-0.5">{{
                                            perm.description
                                        }}</span>
                                    </span>
                                </label>
                            </div>
                        </div>
                    </div>

                    <!-- Section B: View Types -->
                    <div class="mt-8 mb-6">
                        <h3 class="text-sm font-semibold text-gray-900 mb-1">
                            {{ t("projects.create_role.view_types.title") }}
                        </h3>
                        <p class="text-xs text-gray-500 mb-3">
                            {{ t("projects.create_role.view_types.description") }}
                        </p>
                        <div class="flex flex-wrap gap-4">
                            <label
                                v-for="vt in viewTypeOptions"
                                :key="vt.value"
                                class="flex items-center gap-2 cursor-pointer"
                            >
                                <input
                                    type="checkbox"
                                    :checked="hasViewPerm(vt.value)"
                                    @change="toggleViewPerm(vt.value, $event.target.checked)"
                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                />
                                <span class="text-sm font-medium text-gray-900">{{
                                    vt.label
                                }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Section C: Row Permissions -->
                    <div class="mt-8 mb-6">
                        <h3 class="text-sm font-semibold text-gray-900 mb-1">
                            {{ t("projects.create_role.section.rows") }}
                        </h3>
                        <p class="text-xs text-gray-500 mb-4">
                            {{ t("projects.create_role.section.rows_hint") }}
                        </p>

                        <!-- Mode toggle -->
                        <div class="flex gap-3 mb-5">
                            <button
                                type="button"
                                @click="tableRestrictEnabled = false"
                                :class="[
                                    'px-3 py-1.5 rounded-md text-sm font-medium border transition-colors',
                                    !tableRestrictEnabled
                                        ? 'bg-indigo-600 text-white border-indigo-600'
                                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50',
                                ]"
                            >
                                {{ t("projects.create_role.section.rows_workspace_level") }}
                            </button>
                            <button
                                type="button"
                                @click="tableRestrictEnabled = true"
                                :class="[
                                    'px-3 py-1.5 rounded-md text-sm font-medium border transition-colors',
                                    tableRestrictEnabled
                                        ? 'bg-indigo-600 text-white border-indigo-600'
                                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50',
                                ]"
                            >
                                {{ t("projects.create_role.section.rows_per_table") }}
                            </button>
                        </div>

                        <!-- Workspace-level row permissions -->
                        <div
                            v-if="!tableRestrictEnabled"
                            class="rounded-lg border border-gray-200 overflow-hidden"
                        >
                            <div class="grid grid-cols-1 sm:grid-cols-2">
                                <label
                                    v-for="perm in permissionGroups
                                        .find((g) => g.key === 'rows')
                                        .items.concat(
                                            permissionGroups.find((g) => g.key === 'other').items,
                                        )"
                                    :key="perm.value"
                                    :for="'perm-' + perm.value"
                                    class="flex items-start gap-x-3 px-4 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 sm:[&:nth-last-child(-n+2)]:border-b-0 sm:even:border-l sm:even:border-l-gray-100"
                                    :class="
                                        permissions.includes(perm.value)
                                            ? 'bg-indigo-50'
                                            : 'hover:bg-gray-50'
                                    "
                                >
                                    <input
                                        :id="'perm-' + perm.value"
                                        v-model="permissions"
                                        :value="perm.value"
                                        type="checkbox"
                                        class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                    <span class="min-w-0">
                                        <span class="block text-sm font-medium text-gray-900">{{
                                            perm.label
                                        }}</span>
                                        <span class="block text-xs text-gray-500 mt-0.5">{{
                                            perm.description
                                        }}</span>
                                    </span>
                                </label>
                            </div>
                        </div>

                        <!-- Per-table matrix -->
                        <div v-else class="overflow-x-auto rounded-md border border-gray-200">
                            <table class="min-w-full text-sm">
                                <thead class="bg-gray-50 border-b border-gray-200">
                                    <!-- Column headers -->
                                    <tr>
                                        <th
                                            class="py-2 pl-3 pr-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider min-w-[140px]"
                                        >
                                            {{ t("projects.create_role.per_table.table_header") }}
                                        </th>
                                        <th
                                            v-for="action in tableActions"
                                            :key="action.value"
                                            class="px-3 py-2 text-center text-xs font-semibold text-gray-500 uppercase tracking-wider"
                                        >
                                            {{ action.label }}
                                        </th>
                                    </tr>
                                    <!-- Select-all row -->
                                    <tr class="border-t border-gray-200 bg-gray-100">
                                        <td class="py-1.5 pl-3 pr-4 text-xs text-gray-400 italic">
                                            {{ t("projects.create_role.per_table.select_all") }}
                                        </td>
                                        <td
                                            v-for="action in tableActions"
                                            :key="action.value"
                                            class="px-3 py-1.5 text-center"
                                        >
                                            <input
                                                type="checkbox"
                                                :checked="allTablesHavePerm(action.value)"
                                                :indeterminate="
                                                    !allTablesHavePerm(action.value) &&
                                                    someTablesHavePerm(action.value)
                                                "
                                                @change="
                                                    toggleAllTablesPerm(
                                                        action.value,
                                                        $event.target.checked,
                                                    )
                                                "
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer"
                                            />
                                        </td>
                                    </tr>
                                </thead>
                                <tbody
                                    v-for="table in workspaceTables"
                                    :key="table.id"
                                    class="divide-y divide-gray-100 bg-white"
                                >
                                    <tr class="hover:bg-gray-50">
                                        <td class="py-2 pl-3 pr-4 whitespace-nowrap">
                                            <div class="flex items-center gap-2">
                                                <input
                                                    type="checkbox"
                                                    :checked="allActionsForTable(table.id)"
                                                    :indeterminate="
                                                        !allActionsForTable(table.id) &&
                                                        someActionsForTable(table.id)
                                                    "
                                                    @change="
                                                        toggleAllActionsForTable(
                                                            table.id,
                                                            $event.target.checked,
                                                        )
                                                    "
                                                    class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer shrink-0"
                                                />
                                                <span
                                                    class="font-medium text-gray-900 text-sm truncate max-w-[240px]"
                                                    :title="
                                                        table.display_name?.String || table.name
                                                    "
                                                    >{{
                                                        (
                                                            table.display_name?.String || table.name
                                                        ).slice(0, 60)
                                                    }}{{
                                                        (table.display_name?.String || table.name)
                                                            .length > 60
                                                            ? "…"
                                                            : ""
                                                    }}</span
                                                >
                                            </div>
                                        </td>
                                        <td
                                            v-for="action in tableActions"
                                            :key="action.value"
                                            class="px-3 py-2 text-center"
                                        >
                                            <input
                                                type="checkbox"
                                                :checked="hasTablePerm(table.id, action.value)"
                                                :disabled="
                                                    action.value !== 'view' &&
                                                    !hasTablePerm(table.id, 'view')
                                                "
                                                @change="
                                                    toggleTablePerm(
                                                        table.id,
                                                        action.value,
                                                        $event.target.checked,
                                                    )
                                                "
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 disabled:opacity-30"
                                            />
                                        </td>
                                    </tr>
                                </tbody>
                                <tbody v-if="workspaceTables.length === 0">
                                    <tr>
                                        <td
                                            :colspan="tableActions.length + 1"
                                            class="py-4 text-center text-sm text-gray-400"
                                        >
                                            {{ t("projects.create_role.per_table.no_tables") }}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <!-- Section D: Table Visibility -->
                    <div class="mt-8 mb-6">
                        <h3 class="text-sm font-semibold text-gray-900 mb-1">
                            {{ t("projects.create_role.visibility.title") }}
                        </h3>
                        <p class="text-xs text-gray-500 mb-4">
                            {{ t("projects.create_role.visibility.description") }}
                        </p>

                        <div class="flex gap-3 mb-5">
                            <button
                                type="button"
                                @click="setTableVisibility(false)"
                                :class="[
                                    'px-3 py-1.5 rounded-md text-sm font-medium border transition-colors',
                                    !tableVisibilityCustom
                                        ? 'bg-indigo-600 text-white border-indigo-600'
                                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50',
                                ]"
                            >
                                {{ t("projects.create_role.visibility.all_visible") }}
                            </button>
                            <button
                                type="button"
                                @click="setTableVisibility(true)"
                                :class="[
                                    'px-3 py-1.5 rounded-md text-sm font-medium border transition-colors',
                                    tableVisibilityCustom
                                        ? 'bg-indigo-600 text-white border-indigo-600'
                                        : 'bg-white text-gray-700 border-gray-300 hover:bg-gray-50',
                                ]"
                            >
                                {{ t("projects.create_role.visibility.custom") }}
                            </button>
                        </div>

                        <div
                            v-if="tableVisibilityCustom"
                            class="overflow-x-auto rounded-md border border-gray-200"
                        >
                            <table class="min-w-full text-sm">
                                <thead class="bg-gray-50 border-b border-gray-200">
                                    <tr>
                                        <th
                                            class="py-2 pl-3 pr-4 text-left text-xs font-semibold text-gray-500 uppercase tracking-wider"
                                        >
                                            {{ t("projects.create_role.per_table.table_header") }}
                                        </th>
                                        <th
                                            class="px-3 py-2 text-center text-xs font-semibold text-gray-500 uppercase tracking-wider w-24"
                                        >
                                            {{
                                                t("projects.create_role.visibility.visible_header")
                                            }}
                                        </th>
                                    </tr>
                                </thead>
                                <!-- Folders with their tables -->
                                <tbody
                                    v-for="folder in workspaceFolders"
                                    :key="'folder-' + folder.id"
                                    class="divide-y divide-gray-100 bg-white"
                                >
                                    <tr class="bg-gray-50 hover:bg-gray-100">
                                        <td class="py-2 pl-3 pr-4 whitespace-nowrap">
                                            <div class="flex items-center gap-2">
                                                <FolderIcon
                                                    class="h-4 w-4 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                                <span class="font-medium text-gray-700 text-sm">{{
                                                    folder.name
                                                }}</span>
                                                <EyeSlashIcon
                                                    v-if="
                                                        folderVisibilityState(folder) ===
                                                        'unchecked'
                                                    "
                                                    class="h-3.5 w-3.5 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                            </div>
                                        </td>
                                        <td class="px-3 py-2 text-center">
                                            <input
                                                type="checkbox"
                                                :checked="
                                                    folderVisibilityState(folder) === 'checked'
                                                "
                                                :indeterminate="
                                                    folderVisibilityState(folder) ===
                                                    'indeterminate'
                                                "
                                                @change="
                                                    toggleFolderVisibility(
                                                        folder,
                                                        $event.target.checked,
                                                    )
                                                "
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer"
                                            />
                                        </td>
                                    </tr>
                                    <tr
                                        v-for="table in folderTables(folder)"
                                        :key="table.id"
                                        class="hover:bg-gray-50"
                                    >
                                        <td class="py-2 pl-8 pr-4 whitespace-nowrap">
                                            <div class="flex items-center gap-2">
                                                <CircleStackIcon
                                                    class="h-4 w-4 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                                <span
                                                    class="font-medium text-gray-900 text-sm"
                                                    :title="
                                                        table.display_name?.String || table.name
                                                    "
                                                >
                                                    {{
                                                        (
                                                            table.display_name?.String || table.name
                                                        ).slice(0, 60)
                                                    }}{{
                                                        (table.display_name?.String || table.name)
                                                            .length > 60
                                                            ? "…"
                                                            : ""
                                                    }}
                                                </span>
                                                <EyeSlashIcon
                                                    v-if="!hasTableVisibility(table.id)"
                                                    class="h-3.5 w-3.5 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                            </div>
                                        </td>
                                        <td class="px-3 py-2 text-center">
                                            <input
                                                type="checkbox"
                                                :checked="hasTableVisibility(table.id)"
                                                @change="
                                                    toggleTableVisibility(
                                                        table.id,
                                                        $event.target.checked,
                                                    )
                                                "
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer"
                                            />
                                        </td>
                                    </tr>
                                </tbody>
                                <!-- Root tables (not in any folder) -->
                                <tbody
                                    v-for="table in rootTables"
                                    :key="table.id"
                                    class="divide-y divide-gray-100 bg-white"
                                >
                                    <tr class="hover:bg-gray-50">
                                        <td class="py-2 pl-3 pr-4 whitespace-nowrap">
                                            <div class="flex items-center gap-2">
                                                <CircleStackIcon
                                                    class="h-4 w-4 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                                <span
                                                    class="font-medium text-gray-900 text-sm"
                                                    :title="
                                                        table.display_name?.String || table.name
                                                    "
                                                >
                                                    {{
                                                        (
                                                            table.display_name?.String || table.name
                                                        ).slice(0, 60)
                                                    }}{{
                                                        (table.display_name?.String || table.name)
                                                            .length > 60
                                                            ? "…"
                                                            : ""
                                                    }}
                                                </span>
                                                <EyeSlashIcon
                                                    v-if="!hasTableVisibility(table.id)"
                                                    class="h-3.5 w-3.5 text-gray-400 shrink-0"
                                                    aria-hidden="true"
                                                />
                                            </div>
                                        </td>
                                        <td class="px-3 py-2 text-center">
                                            <input
                                                type="checkbox"
                                                :checked="hasTableVisibility(table.id)"
                                                @change="
                                                    toggleTableVisibility(
                                                        table.id,
                                                        $event.target.checked,
                                                    )
                                                "
                                                class="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600 cursor-pointer"
                                            />
                                        </td>
                                    </tr>
                                </tbody>
                                <tbody v-if="workspaceTables.length === 0">
                                    <tr>
                                        <td
                                            colspan="2"
                                            class="py-4 text-center text-sm text-gray-400"
                                        >
                                            {{ t("projects.create_role.per_table.no_tables") }}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="p-2 flex items-center justify-end gap-x-6 bg-gray-100 md:mb-0">
            <button
                @click="
                    router.push({
                        name: 'workspace-members',
                        params: { workspaceId: route.params.id },
                    })
                "
                type="button"
                class="text-sm font-semibold leading-6 text-gray-900"
            >
                {{ t("common.button.cancel") }}
            </button>
            <button
                @click="createWorkspaceRole"
                class="rounded-md px-3 py-2 text-sm font-semibold text-white shadow-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 bg-indigo-600 hover:bg-indigo-500 focus-visible:outline-indigo-600"
            >
                {{
                    route.params.rid == undefined && !createdRoleId
                        ? t("common.button.create")
                        : t("common.button.update")
                }}
            </button>
        </div>
        <ErrorDialog v-model="errorDialog" :message="error" />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, watch, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import ErrorDialog from "@/components/Collimato/Dialogs/ErrorDialog.vue";
import { useAlertStore } from "@/store/alerts";
import workspaceService from "@/services/workspaceService";
import { extractErrorMessage } from "@/utils/errors";
import { rolePermissionGroups } from "@/utils/projects/rolePermissions";
import { isBuiltInRole, roleDescription, roleLabel } from "@/utils/roleLabels";
import { EyeSlashIcon, FolderIcon, CircleStackIcon } from "@heroicons/vue/24/outline";

const route = useRoute();
const router = useRouter();

const workspaceTables = ref([]);
const workspaceFolders = ref([]);
const errorDialog = ref(false);
const error = ref("");
const loaded = ref(false);
const permissions = ref([]);
const name = ref("");
const description = ref("");
const isBuiltIn = ref(false);
const tableRestrictEnabled = ref(false);
const createdRoleId = ref(null);

const rules = { name: { required } };
const v$ = useVuelidate(rules, { name });

const permissionGroups = computed(() => rolePermissionGroups(t.value, tableRestrictEnabled.value));

const viewTypeOptions = [
    {
        value: "grid",
        label: t.value("projects.create_role.view_types.grid"),
    },
    {
        value: "kanban",
        label: t.value("projects.create_role.view_types.kanban"),
    },
    {
        value: "calendar",
        label: t.value("projects.create_role.view_types.calendar"),
    },
    {
        value: "gantt",
        label: t.value("projects.create_role.view_types.gantt"),
    },
];

function hasViewPerm(type) {
    return permissions.value.includes("views_" + type);
}

function toggleViewPerm(type, checked) {
    const perm = "views_" + type;

    if (checked) {
        if (!permissions.value.includes(perm)) {
            permissions.value.push(perm);
        }
    } else {
        permissions.value = permissions.value.filter((p) => p !== perm);
    }
}

const tableActions = [
    {
        value: "view",
        label: t.value("projects.create_role.per_table.action.view"),
    },
    {
        value: "create_task",
        label: t.value("projects.create_role.per_table.action.create_task"),
    },
    {
        value: "update_task",
        label: t.value("projects.create_role.per_table.action.update_task"),
    },
    {
        value: "delete_task",
        label: t.value("projects.create_role.per_table.action.delete_task"),
    },
    {
        value: "create_fields",
        label: t.value("projects.create_role.per_table.action.create_fields"),
    },
    {
        value: "edit_fields",
        label: t.value("projects.create_role.per_table.action.edit_fields"),
    },
    {
        value: "manage_views",
        label: t.value("projects.create_role.per_table.action.manage_views"),
    },
    {
        value: "show_assigned_tasks_only",
        label: t.value("projects.create_role.per_table.action.assigned_only"),
    },
];

// tablePerms holds per-table permissions as {table_id, action} objects.
// These are saved/loaded via the dedicated table-permissions endpoint,
// separate from the workspace-level permissions array.
const tablePerms = ref([]);

function hasTablePerm(tableId, action) {
    return tablePerms.value.some((p) => p.table_id === tableId && p.action === action);
}

function allTablesHavePerm(action) {
    return (
        workspaceTables.value.length > 0 &&
        workspaceTables.value.every((t) => hasTablePerm(t.id, action))
    );
}

function someTablesHavePerm(action) {
    return workspaceTables.value.some((t) => hasTablePerm(t.id, action));
}

function toggleAllTablesPerm(action, checked) {
    for (const table of workspaceTables.value) {
        if (checked && action !== "view" && !hasTablePerm(table.id, "view")) {
            toggleTablePerm(table.id, "view", true);
        }

        toggleTablePerm(table.id, action, checked);
    }
}

function allActionsForTable(tableId) {
    return tableActions.every((a) => hasTablePerm(tableId, a.value));
}

function someActionsForTable(tableId) {
    return tableActions.some((a) => hasTablePerm(tableId, a.value));
}

function toggleAllActionsForTable(tableId, checked) {
    if (checked) {
        for (const action of tableActions) {
            if (action.value === "show_assigned_tasks_only") continue;
            toggleTablePerm(tableId, action.value, true);
        }
    } else {
        toggleTablePerm(tableId, "view", false);
    }
}

function toggleTablePerm(tableId, action, checked) {
    if (checked) {
        if (!tablePerms.value.some((p) => p.table_id === tableId && p.action === action)) {
            tablePerms.value.push({ table_id: tableId, action });
        }
    } else {
        if (action === "view") {
            // removing view access also removes all other actions for this table
            tablePerms.value = tablePerms.value.filter((p) => p.table_id !== tableId);
        } else {
            tablePerms.value = tablePerms.value.filter(
                (p) => !(p.table_id === tableId && p.action === action),
            );
        }
    }
}

// sync visibility section with per-table toggle
watch(tableRestrictEnabled, (enabled) => {
    if (enabled) {
        // Switching to per-table mode: auto-switch visibility to Custom so the
        // visibility section accurately reflects per-table access (not "All visible").
        // Pre-fills all tables as visible; admin can uncheck any they want to hide.
        if (!tableVisibilityCustom.value) {
            setTableVisibility(true);
        }
    } else {
        // Remove per-table action entries; preserve visibility (view/hidden) entries
        tablePerms.value = tablePerms.value.filter(
            (p) => p.action === "view" || p.action === "hidden",
        );
    }
});

// Table visibility section (independent of per-table rows toggle)
const tableVisibilityCustom = ref(false);

function setTableVisibility(enabled) {
    tableVisibilityCustom.value = enabled;
    if (!enabled) {
        // Always clear view/hidden visibility entries regardless of per-table mode state
        tablePerms.value = tablePerms.value.filter(
            (p) => p.action !== "view" && p.action !== "hidden",
        );
    } else {
        // switch to custom: default all tables to visible
        for (const table of workspaceTables.value) {
            // remove hidden sentinel if present
            tablePerms.value = tablePerms.value.filter(
                (p) => !(p.table_id === table.id && p.action === "hidden"),
            );
            if (!tablePerms.value.some((p) => p.table_id === table.id && p.action === "view")) {
                tablePerms.value.push({ table_id: table.id, action: "view" });
            }
        }
    }
}

function hasTableVisibility(tableId) {
    return tablePerms.value.some((p) => p.table_id === tableId && p.action === "view");
}

function toggleTableVisibility(tableId, checked) {
    if (checked) {
        // remove hidden sentinel, then add view
        tablePerms.value = tablePerms.value.filter(
            (p) => !(p.table_id === tableId && p.action === "hidden"),
        );
        if (!tablePerms.value.some((p) => p.table_id === tableId && p.action === "view")) {
            tablePerms.value.push({ table_id: tableId, action: "view" });
        }
    } else {
        // removing view also removes all other per-table actions for this table;
        // save a "hidden" sentinel so backend can distinguish explicitly hidden from unconfigured new tables
        tablePerms.value = tablePerms.value.filter((p) => p.table_id !== tableId);
        tablePerms.value.push({ table_id: tableId, action: "hidden" });
    }
}

const rootTables = computed(() => {
    const folderIds = new Set(
        workspaceFolders.value.flatMap((f) => (f.tables || []).map((t) => t.id)),
    );

    return workspaceTables.value.filter((t) => !folderIds.has(t.id));
});

function folderTables(folder) {
    const ids = new Set((folder.tables || []).map((t) => t.id));

    return workspaceTables.value.filter((t) => ids.has(t.id));
}

function folderVisibilityState(folder) {
    const tables = folderTables(folder);

    if (tables.length === 0) return "checked";
    const visibleCount = tables.filter((t) => hasTableVisibility(t.id)).length;

    if (visibleCount === tables.length) return "checked";
    if (visibleCount === 0) return "unchecked";

    return "indeterminate";
}

function toggleFolderVisibility(folder, checked) {
    for (const table of folderTables(folder)) {
        toggleTableVisibility(table.id, checked);
    }
}

async function createWorkspaceRole() {
    const isValid = await v$.value.$validate();

    if (!isValid) {
        return;
    }

    const roleId = route.params.rid ?? createdRoleId.value;

    if (roleId) {
        updateRole(roleId);

        return;
    }

    let res;

    try {
        res = await workspaceService.createWorkspaceRole(route.params.id, {
            name: name.value,
            display_name: name.value,
            description: description.value,
            permissions: permissions.value,
            per_table_mode: tableRestrictEnabled.value,
        });
    } catch (err) {
        showError(extractErrorMessage(err, t.value("projects.create_role.create_failed")));

        return;
    }

    createdRoleId.value = res.data?.id ?? null;
    if (createdRoleId.value && tablePerms.value.length > 0) {
        try {
            await workspaceService.updateRoleTablePermissions(
                route.params.id,
                createdRoleId.value,
                tablePerms.value,
            );
        } catch {
            showError(t.value("projects.create_role.table_permissions_not_saved"));

            return;
        }
    }

    router.push({ name: "workspace-members", params: { workspaceId: route.params.id } });
}

async function updateRole(roleId) {
    const patch = {
        permissions: permissions.value,
        per_table_mode: tableRestrictEnabled.value,
    };

    if (!isBuiltIn.value) {
        patch.name = name.value;
        patch.display_name = name.value;
        patch.description = description.value;
    }

    try {
        await workspaceService.updateWorkspaceRole(route.params.id, roleId, patch);
        await workspaceService.updateRoleTablePermissions(
            route.params.id,
            roleId,
            tablePerms.value,
        );
        router.push({ name: "workspace-members", params: { workspaceId: route.params.id } });
    } catch (err) {
        showError(extractErrorMessage(err, t.value("projects.create_role.failed_to_create_role")));
    }
}

function showError(message) {
    error.value = message;
    errorDialog.value = true;
}

onMounted(async () => {
    try {
        const res = await workspaceService.getTablesForRoleEditor(route.params.id);

        workspaceTables.value = res.data?.tables || [];
        workspaceFolders.value = res.data?.folders || [];
    } catch {
        workspaceTables.value = [];
    }

    if (route.params.rid !== undefined) {
        try {
            const [rolesRes, tablePermsRes] = await Promise.all([
                workspaceService.getWorkspaceRoles(route.params.id),
                workspaceService.getRoleTablePermissions(route.params.id, route.params.rid),
            ]);
            const roleData = (rolesRes.data || []).find((r) => r.id === route.params.rid);

            if (roleData) {
                isBuiltIn.value = isBuiltInRole(roleData.name, "projects");
                name.value = isBuiltIn.value
                    ? roleLabel(roleData, "projects")
                    : roleData.name || roleData.display_name || "";
                description.value = roleDescription(roleData, "projects");
                permissions.value = roleData.permissions || [];
                tablePerms.value = tablePermsRes.data || [];
                tableRestrictEnabled.value = !!roleData.per_table_mode;
                tableVisibilityCustom.value = tablePerms.value.some(
                    (p) => p.action === "view" || p.action === "hidden",
                );
            } else {
                useAlertStore().showError(t.value("common.error.something_went_wrong"));
                router.push({
                    name: "workspace-members",
                    params: { workspaceId: route.params.id },
                });

                return;
            }
        } catch (err) {
            useAlertStore().showError(extractErrorMessage(err));
            router.push({ name: "workspace-members", params: { workspaceId: route.params.id } });

            return;
        }
    } else {
        permissions.value = ["views_grid", "views_kanban", "views_calendar", "views_gantt"];
    }

    loaded.value = true;
});
</script>
