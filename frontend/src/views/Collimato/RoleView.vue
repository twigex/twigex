<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="flex flex-col h-full overflow-hidden">
        <!-- Loading -->
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <template v-else>
            <!-- Header -->
            <div class="shrink-0 flex items-center gap-x-3 border-b px-4 py-3">
                <button
                    type="button"
                    @click="router.push({ name: 'collimato-users' })"
                    class="shrink-0 rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                >
                    <ChevronLeftIcon class="h-5 w-5" aria-hidden="true" />
                </button>
                <div>
                    <h1 class="text-base font-semibold text-gray-900">
                        {{
                            readOnly
                                ? t("collimato.roles.view_title")
                                : route.params.id
                                  ? t("collimato.roles.edit_title")
                                  : t("collimato.roles.create_title")
                        }}
                    </h1>
                    <p class="text-xs text-gray-500">{{ t("collimato.roles.page_description") }}</p>
                </div>
            </div>

            <!-- Scrollable content -->
            <div class="flex-1 overflow-y-auto">
                <fieldset
                    :disabled="readOnly"
                    class="mx-auto min-w-0 max-w-3xl px-6 py-6 space-y-8"
                >
                    <!-- Role name & description -->
                    <div class="space-y-4">
                        <div>
                            <label
                                for="role-name"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.roles.role_name") }}
                            </label>
                            <div class="mt-2">
                                <input
                                    v-model="name"
                                    type="text"
                                    id="role-name"
                                    :disabled="isBuiltIn"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset placeholder:text-gray-400 focus:ring-2 focus:ring-inset sm:leading-6"
                                    :class="
                                        v$.name.$error
                                            ? 'ring-red-600 focus:ring-red-600'
                                            : 'ring-gray-300 focus:ring-indigo-600'
                                    "
                                    placeholder="Role name"
                                />
                                <p v-if="v$.name.$error" class="mt-1.5 text-sm text-red-600">
                                    {{ t("common.error.required_field") }}
                                </p>
                            </div>
                        </div>
                        <div>
                            <label
                                for="role-description"
                                class="block text-sm font-medium leading-6 text-gray-900"
                            >
                                {{ t("collimato.roles.role_description") }}
                            </label>
                            <div class="mt-2">
                                <textarea
                                    v-model="description"
                                    id="role-description"
                                    rows="3"
                                    :disabled="isBuiltIn"
                                    class="block w-full rounded-md border-0 py-1.5 px-3 text-sm text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:leading-6"
                                    placeholder="Role description"
                                />
                            </div>
                        </div>
                    </div>

                    <!-- Auto-update toggle -->
                    <div class="rounded-lg border border-gray-200 overflow-hidden">
                        <SwitchGroup as="div" class="flex items-center justify-between px-4 py-4">
                            <span class="flex flex-col">
                                <SwitchLabel
                                    as="span"
                                    class="text-sm font-medium text-gray-900"
                                    passive
                                >
                                    {{ t("collimato.roles.auto_update") }}
                                </SwitchLabel>
                                <SwitchDescription as="span" class="text-sm text-gray-500 mt-0.5">
                                    {{ t("collimato.roles.auto_update_description") }}
                                </SwitchDescription>
                            </span>
                            <Switch
                                v-model="auto_update"
                                :class="[
                                    auto_update ? 'bg-indigo-600' : 'bg-gray-200',
                                    'relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-indigo-600 focus:ring-offset-2',
                                ]"
                            >
                                <span
                                    aria-hidden="true"
                                    :class="[
                                        auto_update ? 'translate-x-5' : 'translate-x-0',
                                        'pointer-events-none inline-block size-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out',
                                    ]"
                                />
                            </Switch>
                        </SwitchGroup>
                    </div>

                    <!-- Permission groups -->
                    <div class="space-y-6">
                        <div
                            v-for="group in permissionGroups"
                            :key="group.key"
                            class="rounded-lg border border-gray-200 overflow-hidden"
                        >
                            <!-- Group header -->
                            <div
                                class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                            >
                                <span class="text-sm font-semibold text-gray-900">{{
                                    group.label
                                }}</span>
                                <button
                                    v-if="!readOnly"
                                    type="button"
                                    @click="toggleGroup(group)"
                                    class="text-xs text-indigo-600 hover:text-indigo-500"
                                >
                                    {{
                                        isGroupFullySelected(group)
                                            ? t("collimato.roles.deselect_all")
                                            : t("collimato.roles.select_all")
                                    }}
                                </button>
                            </div>

                            <!-- Permission items -->
                            <div
                                class="grid grid-cols-1 sm:grid-cols-2 divide-y sm:divide-y-0 sm:divide-x divide-gray-100"
                            >
                                <label
                                    v-for="item in group.items"
                                    :key="item.value"
                                    class="flex items-start gap-x-3 px-4 py-3 hover:bg-gray-50 cursor-pointer border-b border-gray-100 last:border-b-0 sm:[&:nth-last-child(-n+2)]:border-b-0"
                                >
                                    <input
                                        type="checkbox"
                                        v-model="permissions"
                                        :value="item.value"
                                        class="mt-0.5 h-4 w-4 shrink-0 rounded border-gray-300 text-indigo-600 focus:ring-indigo-600"
                                    />
                                    <span class="min-w-0">
                                        <span class="block text-sm font-medium text-gray-900">{{
                                            item.label
                                        }}</span>
                                        <span class="block text-xs text-gray-500 mt-0.5">{{
                                            item.desc
                                        }}</span>
                                    </span>
                                </label>
                            </div>
                        </div>
                    </div>

                    <!-- Data model permissions -->
                    <div class="rounded-lg border border-gray-200 overflow-hidden">
                        <div
                            class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                        >
                            <div>
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("collimato.roles.table.model.title") }}
                                </h3>
                                <p class="text-xs text-gray-500 mt-0.5">
                                    {{ t("collimato.roles.table.model.description") }}
                                </p>
                            </div>
                            <button
                                v-if="!readOnly"
                                type="button"
                                @click="modelPermissionDialogOpen = true"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-indigo-500"
                            >
                                <PlusIcon class="-ml-0.5 h-3.5 w-3.5" aria-hidden="true" />
                                {{ t("collimato.roles.table.model.button.add_data_permission") }}
                            </button>
                        </div>
                        <ul role="list" class="divide-y divide-gray-100">
                            <li
                                v-if="tablePermissions.length === 0"
                                class="px-4 py-4 text-center text-sm text-gray-400"
                            >
                                {{ t("collimato.roles.table.model.empty") }}
                            </li>
                            <li
                                v-for="item in tablePermissions"
                                :key="item.name"
                                class="flex items-center justify-between px-4 py-2.5"
                                :class="
                                    pendingDelete.table === item ? 'bg-red-50' : 'hover:bg-gray-50'
                                "
                            >
                                <template v-if="pendingDelete.table === item">
                                    <span class="text-sm text-red-700"
                                        >Remove "{{ item.name }}"?</span
                                    >
                                    <div class="flex items-center gap-x-2">
                                        <button
                                            type="button"
                                            @click="pendingDelete.table = null"
                                            class="text-sm text-gray-500 hover:text-gray-700"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="button"
                                            @click="
                                                tablePermissions.splice(
                                                    tablePermissions.indexOf(item),
                                                    1,
                                                );
                                                pendingDelete.table = null;
                                            "
                                            class="rounded px-2.5 py-1 text-xs font-semibold bg-red-600 text-white hover:bg-red-700"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                </template>
                                <template v-else>
                                    <div class="flex items-center gap-x-3">
                                        <CircleStackIcon
                                            class="h-4 w-4 shrink-0 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        <span class="text-sm text-gray-900">{{ item.name }}</span>
                                    </div>
                                    <button
                                        v-if="!readOnly"
                                        type="button"
                                        @click="pendingDelete.table = item"
                                        class="rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    >
                                        <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                    </button>
                                </template>
                            </li>
                        </ul>
                    </div>

                    <!-- Column restrictions -->
                    <div
                        v-if="hasCollimatoRoles"
                        class="rounded-lg border border-gray-200 overflow-hidden"
                    >
                        <div
                            class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                        >
                            <div>
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("collimato.roles.table.columns.title") }}
                                </h3>
                                <p class="text-xs text-gray-500 mt-0.5">
                                    {{ t("collimato.roles.table.columns.description") }}
                                </p>
                            </div>
                            <button
                                v-if="!readOnly"
                                type="button"
                                @click="columnPermissionDialogOpen = true"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-indigo-500"
                            >
                                <PlusIcon class="-ml-0.5 h-3.5 w-3.5" aria-hidden="true" />
                                {{ t("collimato.roles.table.columns.button.add_restriction") }}
                            </button>
                        </div>
                        <ul role="list" class="divide-y divide-gray-100">
                            <li
                                v-if="columnPermissions.length === 0"
                                class="px-4 py-4 text-center text-sm text-gray-400"
                            >
                                {{ t("collimato.roles.table.columns.empty") }}
                            </li>
                            <li
                                v-for="item in columnPermissions"
                                :key="`${item.table}-${item.field}`"
                                class="flex items-center justify-between px-4 py-2.5"
                                :class="
                                    pendingDelete.column === item ? 'bg-red-50' : 'hover:bg-gray-50'
                                "
                            >
                                <template v-if="pendingDelete.column === item">
                                    <span class="text-sm text-red-700 truncate"
                                        >Remove "{{ item.table }} / {{ item.field }}"?</span
                                    >
                                    <div class="flex shrink-0 items-center gap-x-2 ml-3">
                                        <button
                                            type="button"
                                            @click="pendingDelete.column = null"
                                            class="text-sm text-gray-500 hover:text-gray-700"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="button"
                                            @click="
                                                columnPermissions.splice(
                                                    columnPermissions.indexOf(item),
                                                    1,
                                                );
                                                pendingDelete.column = null;
                                            "
                                            class="rounded px-2.5 py-1 text-xs font-semibold bg-red-600 text-white hover:bg-red-700"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                </template>
                                <template v-else>
                                    <div class="flex items-center gap-x-3 min-w-0">
                                        <TableCellsIcon
                                            class="h-4 w-4 shrink-0 text-gray-400"
                                            aria-hidden="true"
                                        />
                                        <span class="text-sm text-gray-900 truncate">{{
                                            item.table
                                        }}</span>
                                        <span class="text-gray-300">/</span>
                                        <span class="text-sm text-gray-500 truncate">{{
                                            item.field
                                        }}</span>
                                    </div>
                                    <button
                                        v-if="!readOnly"
                                        type="button"
                                        @click="pendingDelete.column = item"
                                        class="shrink-0 rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    >
                                        <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                    </button>
                                </template>
                            </li>
                        </ul>
                    </div>

                    <!-- Row restrictions -->
                    <div
                        v-if="hasCollimatoRoles"
                        class="rounded-lg border border-gray-200 overflow-hidden"
                    >
                        <div
                            class="flex items-center justify-between px-4 py-3 bg-gray-50 border-b border-gray-200"
                        >
                            <div>
                                <h3 class="text-sm font-semibold text-gray-900">
                                    {{ t("collimato.roles.table.rows.title") }}
                                </h3>
                                <p class="text-xs text-gray-500 mt-0.5">
                                    {{ t("collimato.roles.table.rows.description") }}
                                </p>
                            </div>
                            <button
                                v-if="!readOnly"
                                type="button"
                                @click="rowPermissionDialogOpen = true"
                                class="inline-flex items-center gap-x-1.5 rounded-md bg-indigo-600 px-3 py-1.5 text-xs font-semibold text-white shadow-sm hover:bg-indigo-500"
                            >
                                <PlusIcon class="-ml-0.5 h-3.5 w-3.5" aria-hidden="true" />
                                {{ t("collimato.roles.table.rows.button.add_row_restriction") }}
                            </button>
                        </div>
                        <ul role="list" class="divide-y divide-gray-100">
                            <li
                                v-if="rowPermissions.length === 0"
                                class="px-4 py-4 text-center text-sm text-gray-400"
                            >
                                {{ t("collimato.roles.table.rows.empty") }}
                            </li>
                            <li
                                v-for="item in rowPermissions"
                                :key="`${item.member}-${item.operator}`"
                                class="flex items-center justify-between px-4 py-2.5"
                                :class="
                                    pendingDelete.row === item ? 'bg-red-50' : 'hover:bg-gray-50'
                                "
                            >
                                <template v-if="pendingDelete.row === item">
                                    <span class="text-sm text-red-700 truncate"
                                        >Remove "{{ item.member }}"?</span
                                    >
                                    <div class="flex shrink-0 items-center gap-x-2 ml-3">
                                        <button
                                            type="button"
                                            @click="pendingDelete.row = null"
                                            class="text-sm text-gray-500 hover:text-gray-700"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            type="button"
                                            @click="
                                                rowPermissions.splice(
                                                    rowPermissions.indexOf(item),
                                                    1,
                                                );
                                                pendingDelete.row = null;
                                            "
                                            class="rounded px-2.5 py-1 text-xs font-semibold bg-red-600 text-white hover:bg-red-700"
                                        >
                                            Remove
                                        </button>
                                    </div>
                                </template>
                                <template v-else>
                                    <div class="flex items-center gap-x-2 min-w-0 flex-1">
                                        <span class="text-sm font-medium text-gray-900 truncate">{{
                                            item.member
                                        }}</span>
                                        <span
                                            class="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600"
                                            >{{ item.operator }}</span
                                        >
                                        <span class="text-sm text-gray-500 truncate">{{
                                            Array.isArray(item.values)
                                                ? item.values.join(", ")
                                                : item.values
                                        }}</span>
                                    </div>
                                    <button
                                        v-if="!readOnly"
                                        type="button"
                                        @click="pendingDelete.row = item"
                                        class="shrink-0 rounded p-1 text-gray-400 hover:bg-red-50 hover:text-red-600"
                                    >
                                        <TrashIcon class="h-4 w-4" aria-hidden="true" />
                                    </button>
                                </template>
                            </li>
                        </ul>
                    </div>
                </fieldset>
            </div>

            <!-- Footer -->
            <div
                class="shrink-0 flex items-center justify-end gap-x-3 border-t bg-gray-50 px-6 py-4"
            >
                <button
                    type="button"
                    @click="router.push({ name: 'collimato-users' })"
                    class="rounded-md px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900"
                >
                    {{ t("common.button.cancel") }}
                </button>
                <button
                    v-if="!readOnly"
                    type="button"
                    @click="createWorkspaceRole"
                    class="rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
                >
                    {{ route.params.id ? t("common.button.update") : t("common.button.create") }}
                </button>
            </div>
        </template>

        <TablePermissionDialog
            v-model="modelPermissionDialogOpen"
            :data-models="dataModels"
            @add="addModelPermissions"
        />
        <ColumnPermissionDialog
            v-model="columnPermissionDialogOpen"
            @add="addColumnPermissions"
            :data-models="selectedDataModels"
        />
        <RowPermissionDialog
            v-model="rowPermissionDialogOpen"
            @add-filter="addRowFilter"
            @close="rowPermissionDialogOpen = false"
            :data-models="selectedDataModels"
        />
        <ErrorDialog v-model="errorDialog" :message="error" />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted, reactive } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useVuelidate } from "@vuelidate/core";
import { required } from "@vuelidate/validators";
import { useSettingsStore } from "@/store/settings";
import { useCollimatoStore } from "@/store/collimato";
import { roleDescription } from "@/utils/roleLabels";
import { Switch, SwitchDescription, SwitchGroup, SwitchLabel } from "@headlessui/vue";
import { PlusIcon, TrashIcon } from "@heroicons/vue/20/solid";
import { ChevronLeftIcon, CircleStackIcon, TableCellsIcon } from "@heroicons/vue/24/outline";
import collimatoService from "@/services/collimatoService";
import TablePermissionDialog from "@/components/Collimato/Dialogs/TablePermissionDialog.vue";
import ColumnPermissionDialog from "@/components/Collimato/Dialogs/ColumnPermissionDialog.vue";
import RowPermissionDialog from "@/components/Collimato/Dialogs/RowPermissionDialog.vue";
import ErrorDialog from "@/components/Collimato/Dialogs/ErrorDialog.vue";

const settingsStore = useSettingsStore();
const collimatoStore = useCollimatoStore();
const route = useRoute();
const router = useRouter();

const loaded = ref(false);
const errorDialog = ref(false);
const error = ref("");
const name = ref("");
const description = ref("");
const auto_update = ref(false);
const permissions = ref([]);
const tablePermissions = ref([]);
const columnPermissions = ref([]);
const rowPermissions = ref([]);
const modelPermissionDialogOpen = ref(false);
const columnPermissionDialogOpen = ref(false);
const rowPermissionDialogOpen = ref(false);
const pendingDelete = reactive({ table: null, column: null, row: null });
const dataModels = ref([]);
const hasCollimatoRoles = computed(() => settingsStore.getLicenseFeature("collimato_roles"));
const savedName = ref("");

const isBuiltIn = computed(() => ["workspace_admin", "workspace_user"].includes(savedName.value));

const readOnly = computed(() => {
    if (!route.params.id) return false;
    if (!collimatoStore.hasPermissionToEditRoles) return true;
    if (collimatoStore.isWorkspaceAdmin) return false;

    return (
        savedName.value === "workspace_admin" || collimatoStore.roleNames.includes(savedName.value)
    );
});

const rules = { name: { required } };
const v$ = useVuelidate(rules, { name });

const permissionGroups = [
    {
        key: "connections",
        label: t.value("collimato.roles.sections.connections"),
        items: [
            {
                value: "create_connections",
                label: t.value("collimato.roles.create_connections"),
                desc: t.value("collimato.roles.create_connections_description"),
            },
            {
                value: "view_connections",
                label: t.value("collimato.roles.view_connections"),
                desc: t.value("collimato.roles.view_connections_description"),
            },
            {
                value: "edit_connections",
                label: t.value("collimato.roles.edit_connections"),
                desc: t.value("collimato.roles.edit_connections_description"),
            },
            {
                value: "delete_connections",
                label: t.value("collimato.roles.delete_connections"),
                desc: t.value("collimato.roles.delete_connections_description"),
            },
        ],
    },
    {
        key: "charts",
        label: t.value("collimato.roles.sections.charts"),
        items: [
            {
                value: "create_charts",
                label: t.value("collimato.roles.create_charts"),
                desc: t.value("collimato.roles.create_charts_description"),
            },
            {
                value: "view_charts",
                label: t.value("collimato.roles.view_charts"),
                desc: t.value("collimato.roles.view_charts_description"),
            },
            {
                value: "edit_charts",
                label: t.value("collimato.roles.edit_charts"),
                desc: t.value("collimato.roles.edit_charts_description"),
            },
            {
                value: "delete_charts",
                label: t.value("collimato.roles.delete_charts"),
                desc: t.value("collimato.roles.delete_charts_description"),
            },
        ],
    },
    {
        key: "dashboards",
        label: t.value("collimato.roles.sections.dashboards"),
        items: [
            {
                value: "create_dashboards",
                label: t.value("collimato.roles.create_dashboards"),
                desc: t.value("collimato.roles.create_dashboards_description"),
            },
            {
                value: "view_dashboards",
                label: t.value("collimato.roles.view_dashboards"),
                desc: t.value("collimato.roles.view_dashboards_description"),
            },
            {
                value: "edit_dashboards",
                label: t.value("collimato.roles.edit_dashboards"),
                desc: t.value("collimato.roles.edit_dashboards_description"),
            },
            {
                value: "delete_dashboards",
                label: t.value("collimato.roles.delete_dashboards"),
                desc: t.value("collimato.roles.delete_dashboards_description"),
            },
        ],
    },
    {
        key: "dashboard_filters",
        label: t.value("collimato.roles.sections.dashboard_filters"),
        items: [
            {
                value: "create_dashboard_filters",
                label: t.value("collimato.roles.create_dashboard_filters"),
                desc: t.value("collimato.roles.create_dashboard_filters_description"),
            },
            {
                value: "view_dashboard_filters",
                label: t.value("collimato.roles.view_dashboard_filters"),
                desc: t.value("collimato.roles.view_dashboard_filters_description"),
            },
            {
                value: "edit_dashboard_filters",
                label: t.value("collimato.roles.edit_dashboard_filters"),
                desc: t.value("collimato.roles.edit_dashboard_filters_description"),
            },
            {
                value: "delete_dashboard_filters",
                label: t.value("collimato.roles.delete_dashboard_filters"),
                desc: t.value("collimato.roles.delete_dashboard_filters_description"),
            },
        ],
    },
    {
        key: "data_models",
        label: t.value("collimato.roles.sections.data_models"),
        items: [
            {
                value: "create_datamodels",
                label: t.value("collimato.roles.create_data_models"),
                desc: t.value("collimato.roles.create_data_models_description"),
            },
            {
                value: "view_datamodels",
                label: t.value("collimato.roles.view_data_models"),
                desc: t.value("collimato.roles.view_data_models_description"),
            },
            {
                value: "edit_datamodels",
                label: t.value("collimato.roles.edit_data_models"),
                desc: t.value("collimato.roles.edit_data_models_description"),
            },
            {
                value: "delete_datamodels",
                label: t.value("collimato.roles.delete_data_models"),
                desc: t.value("collimato.roles.delete_data_models_description"),
            },
        ],
    },
    {
        key: "users",
        label: t.value("collimato.roles.sections.users"),
        items: [
            {
                value: "add_users",
                label: t.value("collimato.roles.add_users"),
                desc: t.value("collimato.roles.add_users_description"),
            },
            {
                value: "delete_users",
                label: t.value("collimato.roles.delete_users"),
                desc: t.value("collimato.roles.delete_users_description"),
            },
            {
                value: "assign_roles",
                label: t.value("collimato.roles.assign_roles_to_users"),
                desc: t.value("collimato.roles.assign_roles_to_users_description"),
            },
        ],
    },
    {
        key: "roles",
        label: t.value("collimato.roles.sections.roles"),
        items: [
            {
                value: "create_roles",
                label: t.value("collimato.roles.create_roles"),
                desc: t.value("collimato.roles.create_roles_description"),
            },
            {
                value: "view_roles",
                label: t.value("collimato.roles.view_roles"),
                desc: t.value("collimato.roles.view_roles_description"),
            },
            {
                value: "edit_roles",
                label: t.value("collimato.roles.edit_roles"),
                desc: t.value("collimato.roles.edit_roles_description"),
            },
            {
                value: "delete_roles",
                label: t.value("collimato.roles.delete_roles"),
                desc: t.value("collimato.roles.delete_roles_description"),
            },
        ],
    },
];

const selectedDataModels = computed(() =>
    dataModels.value.filter((model) => tablePermissions.value.some((p) => p.name === model.name)),
);

function isGroupFullySelected(group) {
    return group.items.every((item) => permissions.value.includes(item.value));
}

function toggleGroup(group) {
    if (isGroupFullySelected(group)) {
        permissions.value = permissions.value.filter(
            (p) => !group.items.some((item) => item.value === p),
        );
    } else {
        group.items.forEach((item) => {
            if (!permissions.value.includes(item.value)) permissions.value.push(item.value);
        });
    }
}

function addModelPermissions(perms) {
    modelPermissionDialogOpen.value = false;
    perms.forEach((p) => tablePermissions.value.push({ name: p.name }));
}

function addColumnPermissions(permission) {
    columnPermissionDialogOpen.value = false;
    columnPermissions.value.push(permission);
}

function addRowFilter(filter) {
    rowPermissionDialogOpen.value = false;
    rowPermissions.value.push(filter);
}

async function createWorkspaceRole() {
    const isValid = await v$.value.$validate();

    if (!isValid) return;

    const payload = {
        name: name.value,
        display_name: name.value,
        description: description.value,
        permissions: permissions.value,
        table_permissions: tablePermissions.value.map((p) => p.name),
        auto_update: auto_update.value,
    };

    if (isBuiltIn.value) {
        delete payload.name;
        delete payload.display_name;
        delete payload.description;
    }

    // Omitted rather than sent empty when unavailable: the update endpoint treats
    // an absent field as unchanged, so any stored rules survive an edit here
    // instead of being silently stripped.
    if (hasCollimatoRoles.value) {
        payload.column_permissions = columnPermissions.value;
        payload.row_permissions = rowPermissions.value;
    }

    if (!route.params.id) {
        collimatoService.createWorkspaceRole(route.params.workspaceId, payload).then(() => {
            router.push({ name: "collimato-users" });
        });

        return;
    }

    collimatoService
        .updateWorkspaceRole(route.params.workspaceId, route.params.id, payload)
        .then(() => {
            router.push({ name: "collimato-users" });
        });
}

onMounted(async () => {
    if (!hasCollimatoRoles.value) {
        router.push({ name: "collimato-users" });

        return;
    }

    await collimatoService.roleMeta(route.params.workspaceId).then((response) => {
        if (response.data.error) {
            error.value = response.data.error;
            errorDialog.value = true;
            loaded.value = true;

            return;
        }

        dataModels.value = response.data.cubes;
    });

    if (route.params.id) {
        await collimatoService
            .workspaceRoleById(route.params.workspaceId, route.params.id)
            .then((response) => {
                name.value = response.data.name;
                savedName.value = response.data.name;
                description.value = roleDescription(response.data, "collimato");
                permissions.value = response.data.permissions;
                response.data.table_permissions.forEach((p) =>
                    tablePermissions.value.push({ name: p }),
                );
                columnPermissions.value = response.data.column_permissions;
                rowPermissions.value = response.data.row_permissions;
                auto_update.value = response.data.auto_update;
            });
    }

    loaded.value = true;
});
</script>
