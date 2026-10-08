<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div class="h-full overflow-y-auto">
        <div v-if="!loaded" class="flex h-full items-center justify-center">
            <div
                class="loader h-8 w-8 rounded-full border-4 border-indigo-500 border-t-transparent animate-spin"
            ></div>
        </div>

        <div v-else class="space-y-10 px-6 py-6">
            <MembersTable :members="filteredUsers" :columns="userColumns">
                <template #toolbar>
                    <div class="flex flex-wrap items-end justify-between gap-4">
                        <div>
                            <h1 class="text-base font-semibold text-gray-900">
                                {{ t("collimato.users.users_table.title") }}
                            </h1>
                            <p class="mt-1 text-sm text-gray-500">
                                {{ t("collimato.users.users_table.description") }}
                            </p>
                        </div>
                        <div class="flex items-center gap-3">
                            <div v-if="directUsers.length > 0" class="relative w-56">
                                <MagnifyingGlassIcon
                                    class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
                                    aria-hidden="true"
                                />
                                <input
                                    v-model="userSearch"
                                    type="search"
                                    :placeholder="
                                        t('collimato.users.users_table.search_placeholder')
                                    "
                                    class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                />
                            </div>
                            <BaseButton
                                v-if="collimatoStore.hasPermissionToAddUsers"
                                size="small"
                                :prepend-icon="PlusIcon"
                                @click="userDialog = true"
                            >
                                {{ t("collimato.users.users_table.button.add_user") }}
                            </BaseButton>
                        </div>
                    </div>
                </template>

                <template #empty>
                    <template v-if="directUsers.length === 0">
                        <p class="font-medium text-gray-900">
                            {{ t("collimato.users.users_table.empty.title") }}
                        </p>
                        <p class="mt-1">
                            {{ t("collimato.users.users_table.empty.description") }}
                        </p>
                    </template>
                    <template v-else>
                        {{ t("collimato.users.users_table.search_no_results") }}
                        <button
                            type="button"
                            class="ml-2 text-indigo-600 hover:text-indigo-500"
                            @click="userSearch = ''"
                        >
                            {{ t("collimato.users.users_table.search_clear") }}
                        </button>
                    </template>
                </template>

                <template #role="{ item }">
                    <div class="flex flex-wrap gap-1">
                        <span
                            v-for="role in splitRoles(item.role)"
                            :key="role"
                            class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                        >
                            {{ roleNameLabel(role, "collimato") }}
                        </span>
                        <span v-if="!item.role" class="text-xs text-gray-400">
                            {{ t("collimato.users.users_table.no_role") }}
                        </span>
                    </div>
                </template>

                <template #actions="{ item }">
                    <button
                        v-if="canAssignUser(item)"
                        type="button"
                        class="text-indigo-600 hover:text-indigo-500"
                        @click="
                            userToEdit = item;
                            roleDialog = true;
                        "
                    >
                        {{ t("common.button.edit") }}
                    </button>
                    <button
                        v-if="collimatoStore.hasPermissionToDeleteUsers"
                        type="button"
                        class="text-red-600 hover:text-red-500"
                        @click="
                            userToEdit = item;
                            removeUserDialog = true;
                        "
                    >
                        {{ t("common.button.remove") }}
                    </button>
                </template>
            </MembersTable>

            <DataTable
                v-if="collimatoStore.hasPermissionToAddUsers"
                :columns="groupColumns"
                :items="groups"
            >
                <template #toolbar>
                    <div class="flex flex-wrap items-end justify-between gap-4">
                        <div>
                            <h2 class="text-sm font-semibold text-gray-900">
                                {{ t("collimato.users.groups_table.title") }}
                            </h2>
                            <p class="mt-1 text-sm text-gray-500">
                                {{ t("collimato.users.groups_table.description") }}
                            </p>
                        </div>
                        <BaseButton
                            size="small"
                            variant="secondary"
                            :prepend-icon="PlusIcon"
                            @click="groupDialog = true"
                        >
                            {{ t("collimato.users.groups_table.button.add_group") }}
                        </BaseButton>
                    </div>
                </template>

                <template #empty>
                    <p class="font-medium text-gray-900">
                        {{ t("collimato.users.groups_table.empty.title") }}
                    </p>
                    <p class="mt-1">
                        {{ t("collimato.users.groups_table.empty.description") }}
                    </p>
                </template>

                <template #name="{ item }">
                    <button
                        type="button"
                        class="flex min-w-0 items-center gap-x-3 text-left"
                        @click="openGroupMembers(item)"
                    >
                        <div
                            class="flex size-9 shrink-0 items-center justify-center rounded-full bg-gray-100"
                        >
                            <UserGroupIcon class="size-5 text-gray-500" aria-hidden="true" />
                        </div>
                        <span class="truncate font-medium text-gray-900">{{ item.name }}</span>
                    </button>
                </template>

                <template #roles="{ item }">
                    <div class="flex flex-wrap gap-1">
                        <span
                            v-for="role in item.roles || []"
                            :key="role"
                            class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                        >
                            {{ roleNameLabel(role, "collimato") }}
                        </span>
                    </div>
                </template>

                <template #member_count="{ item }">
                    {{ item.member_count ?? 0 }}
                </template>

                <template #actions="{ item }">
                    <div class="flex items-center justify-end gap-x-4 font-medium">
                        <button
                            v-if="canAssignGroup(item)"
                            type="button"
                            class="text-indigo-600 hover:text-indigo-500"
                            @click="
                                groupToEdit = item;
                                groupRoleDialog = true;
                            "
                        >
                            {{ t("common.button.edit") }}
                        </button>
                        <button
                            v-if="collimatoStore.hasPermissionToDeleteUsers"
                            type="button"
                            class="text-red-600 hover:text-red-500"
                            @click="
                                groupToEdit = item;
                                removeGroupDialog = true;
                            "
                        >
                            {{ t("common.button.remove") }}
                        </button>
                    </div>
                </template>
            </DataTable>

            <DataTable
                v-if="collimatoStore.hasPermissionToViewRoles"
                :columns="roleColumns"
                :items="filteredRoles"
            >
                <template #toolbar>
                    <div class="flex flex-wrap items-end justify-between gap-4">
                        <div>
                            <h2 class="text-sm font-semibold text-gray-900">
                                {{ t("collimato.users.roles_table.title") }}
                            </h2>
                            <p class="mt-1 text-sm text-gray-500">
                                {{ t("collimato.users.roles_table.description") }}
                            </p>
                        </div>
                        <div class="flex items-center gap-3">
                            <div v-if="roles.length > 0" class="relative w-56">
                                <MagnifyingGlassIcon
                                    class="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
                                    aria-hidden="true"
                                />
                                <input
                                    v-model="roleSearch"
                                    type="search"
                                    :placeholder="
                                        t('collimato.users.roles_table.search_placeholder')
                                    "
                                    class="block w-full rounded-md border-0 py-1.5 pl-8 pr-3 text-sm text-gray-900 ring-1 ring-inset ring-gray-300 placeholder:text-gray-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600"
                                />
                            </div>
                            <BaseButton
                                v-if="
                                    collimatoStore.hasPermissionToCreateRoles && hasCollimatoRoles
                                "
                                size="small"
                                variant="secondary"
                                :prepend-icon="PlusIcon"
                                @click="router.push({ name: 'collimato-role' })"
                            >
                                {{ t("collimato.users.roles_table.button.new_role") }}
                            </BaseButton>
                        </div>
                    </div>
                </template>

                <template #empty>
                    <template v-if="roles.length === 0">
                        <p class="font-medium text-gray-900">
                            {{ t("collimato.users.roles_table.empty.title") }}
                        </p>
                        <p class="mt-1">
                            {{ t("collimato.users.roles_table.empty.description") }}
                        </p>
                    </template>
                    <template v-else>
                        {{ t("collimato.users.roles_table.search_no_results") }}
                        <button
                            type="button"
                            class="ml-2 text-indigo-600 hover:text-indigo-500"
                            @click="roleSearch = ''"
                        >
                            {{ t("collimato.users.roles_table.search_clear") }}
                        </button>
                    </template>
                </template>

                <template #name="{ item }">
                    <div class="min-w-0">
                        <div class="truncate font-medium text-gray-900">{{ item.label }}</div>
                        <div v-if="item.about" class="mt-1 max-w-md truncate text-gray-500">
                            {{ item.about }}
                        </div>
                    </div>
                </template>

                <template #created_at="{ item }">
                    {{ getDateAndTime(item.created_at) }}
                </template>

                <template #actions="{ item }">
                    <div
                        v-if="hasCollimatoRoles"
                        class="flex items-center justify-end gap-x-4 font-medium"
                    >
                        <button
                            type="button"
                            class="text-indigo-600 hover:text-indigo-500"
                            @click="
                                router.push({
                                    name: 'collimato-role-edit',
                                    params: { id: item.id },
                                })
                            "
                        >
                            {{
                                canEditRole(item)
                                    ? t("common.button.edit")
                                    : t("common.button.view")
                            }}
                        </button>
                        <button
                            v-if="
                                collimatoStore.hasPermissionToDeleteRoles &&
                                !['workspace_admin', 'workspace_user'].includes(item.name)
                            "
                            type="button"
                            class="text-red-600 hover:text-red-500"
                            @click="openDeleteDialog(item)"
                        >
                            {{ t("common.button.delete") }}
                        </button>
                    </div>
                </template>
            </DataTable>
        </div>

        <RemoveUserDialog
            v-model="removeUserDialog"
            @close="removeUserDialog = false"
            @remove="removeUser(userToEdit)"
        />
        <RoleDialog
            v-model="roleDialog"
            :user="userToEdit"
            :roles="assignableRoles"
            @add="updateUserRoles"
        />
        <UserPickerDialog
            v-model="userDialog"
            :title="t('collimato.users.add_user_dialog.title')"
            @confirm="addUserToWorkspace"
        />
        <GroupMembersDialog
            v-model="groupMembersOpen"
            :group-id="groupMembersGroup?.group_id ?? ''"
            :group-name="groupMembersGroup?.name ?? ''"
        />
        <DeleteDialog
            v-model="deleteDialog"
            :title="t('collimato.users.delete_role_dialog.title')"
            :message="t('collimato.users.delete_role_dialog.confirm')"
            @delete="deleteRole"
        />
        <AddGroupDialog
            v-model="groupDialog"
            :roles="assignableRoles"
            @add="addGroupsToWorkspace"
        />
        <RoleDialog
            v-model="groupRoleDialog"
            :user="{ role: (groupToEdit?.roles || []).join(',') }"
            :roles="assignableRoles"
            @add="updateGroupRoles"
        />
        <DeleteDialog
            v-model="removeGroupDialog"
            :title="t('collimato.users.remove_group_dialog.title')"
            :message="t('collimato.users.remove_group_dialog.confirm')"
            @delete="removeGroupFromWorkspace(groupToEdit)"
        />
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { ref, computed, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { PlusIcon, MagnifyingGlassIcon } from "@heroicons/vue/20/solid";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import collimatoService from "@/services/collimatoService";
import useDateOperations from "@/composables/useDateOperations.js";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { useCollimatoStore } from "@/store/collimato";
import UserPickerDialog from "@/components/UserPickerDialog.vue";
import GroupMembersDialog from "@/components/GroupMembersDialog.vue";
import AddGroupDialog from "@/components/Collimato/Dialogs/AddGroupDialog.vue";
import RoleDialog from "@/components/Collimato/Dialogs/RoleDialog.vue";
import RemoveUserDialog from "@/components/Collimato/Dialogs/RemoveUserDialog.vue";
import DeleteDialog from "@/components/Collimato/Dialogs/DeleteDialog.vue";
import { useSettingsStore } from "@/store/settings";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import MembersTable from "@/components/MembersTable.vue";
import { roleLabel, roleDescription, roleNameLabel } from "@/utils/roleLabels";

const route = useRoute();
const router = useRouter();
const alertStore = useAlertStore();
const collimatoStore = useCollimatoStore();
const settingsStore = useSettingsStore();
const hasCollimatoRoles = computed(() => settingsStore.getLicenseFeature("collimato_roles"));
const { getDateAndTime } = useDateOperations();

const loaded = ref(false);
const users = ref([]);
const roles = ref([]);
const groups = ref([]);
const userToEdit = ref({});
const groupToEdit = ref(null);
const removeUserDialog = ref(false);
const userDialog = ref(false);
const roleDialog = ref(false);
const groupDialog = ref(false);
const groupRoleDialog = ref(false);
const groupMembersOpen = ref(false);
const groupMembersGroup = ref(null);

function openGroupMembers(group) {
    groupMembersGroup.value = group;
    groupMembersOpen.value = true;
}

const removeGroupDialog = ref(false);
const deleteDialog = ref(false);
const deleteItem = ref(null);
const userSearch = ref("");
const roleSearch = ref("");

const userColumns = computed(() => [
    { key: "role", label: t.value("members_table.roles"), sortable: true, hiddenBelow: "sm" },
]);

const groupColumns = computed(() => [
    { key: "name", label: t.value("data_table.name"), sortable: true },
    { key: "roles", label: t.value("members_table.roles"), hiddenBelow: "sm" },
    {
        key: "member_count",
        label: t.value("data_table.members"),
        sortable: true,
        hiddenBelow: "md",
    },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const roleColumns = computed(() => [
    { key: "name", label: t.value("data_table.role"), sortable: true, sortKey: "label" },
    { key: "created_at", label: t.value("data_table.created"), sortable: true, hiddenBelow: "md" },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

function splitRoles(role) {
    return (role ?? "")
        .split(",")
        .map((r) => r.trim())
        .filter(Boolean);
}

// Only users added directly to the workspace. Members who materialize via a
// group (via_group) are represented by the Groups list, click a group to view
// its members in the group viewer.
const directUsers = computed(() => users.value.filter((u) => !u.via_group));

const filteredUsers = computed(() => {
    const q = userSearch.value.trim().toLowerCase();

    if (!q) return directUsers.value;

    return directUsers.value.filter(
        (u) =>
            u.user_info?.name?.toLowerCase().includes(q) ||
            u.user_info?.email?.toLowerCase().includes(q) ||
            u.role?.toLowerCase().includes(q),
    );
});

const filteredRoles = computed(() => {
    const q = roleSearch.value.trim().toLowerCase();

    const rows = roles.value.map((r) => ({
        ...r,
        label: roleLabel(r, "collimato"),
        about: roleDescription(r, "collimato"),
    }));

    if (!q) return rows;

    return rows.filter(
        (r) =>
            r.label.toLowerCase().includes(q) ||
            r.name?.toLowerCase().includes(q) ||
            r.about.toLowerCase().includes(q),
    );
});

const assignableRoles = computed(() => {
    if (collimatoStore.isWorkspaceAdmin) return roles.value;

    return roles.value.filter((r) => r.name !== "workspace_admin");
});

function canAssignUser(user) {
    if (!collimatoStore.hasPermissionToAssignRoles) return false;
    if (collimatoStore.isWorkspaceAdmin) return true;

    const isMe = user.user_id === collimatoStore.user?.user_id;
    const isAdmin = (user.role ?? "").split(",").includes("workspace_admin");

    return !isMe && !isAdmin;
}

function canAssignGroup(group) {
    if (!collimatoStore.hasPermissionToAssignRoles) return false;
    if (collimatoStore.isWorkspaceAdmin) return true;

    return !(group.roles ?? []).includes("workspace_admin");
}

function canEditRole(role) {
    if (!collimatoStore.hasPermissionToEditRoles) return false;
    if (collimatoStore.isWorkspaceAdmin) return true;

    return role.name !== "workspace_admin" && !collimatoStore.roleNames.includes(role.name);
}

function openDeleteDialog(item) {
    deleteItem.value = item;
    deleteDialog.value = true;
}

function getWorkspaceUsers() {
    collimatoService.getWorkspaceUsers(route.params.workspaceId).then((response) => {
        users.value = response.data;
    });
}

function getWorkspaceGroups() {
    collimatoService
        .getWorkspaceGroups(route.params.workspaceId)
        .then((response) => {
            groups.value = response.data ?? [];
        })
        .catch(() => {});
}

function addGroupsToWorkspace({ groupIds, roles: groupRoles }) {
    collimatoService
        .addGroupsToWorkspace(route.params.workspaceId, {
            group_ids: groupIds,
            roles: groupRoles,
        })
        .then(() => {
            getWorkspaceGroups();
            // Members materialize live, so refresh the roster too.
            getWorkspaceUsers();
            alertStore.showSuccess(t.value("collimato.users.success.group_added"));
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });

    groupDialog.value = false;
}

function removeGroupFromWorkspace(group) {
    collimatoService
        .removeGroupFromWorkspace(route.params.workspaceId, group.group_id)
        .then(() => {
            getWorkspaceGroups();
            getWorkspaceUsers();
            alertStore.showSuccess(t.value("collimato.users.success.group_removed"));
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });

    removeGroupDialog.value = false;
}

function updateGroupRoles(newRoles) {
    collimatoService
        .updateWorkspaceGroupRoles(route.params.workspaceId, groupToEdit.value.group_id, {
            roles: newRoles,
        })
        .then(() => {
            getWorkspaceGroups();
            getWorkspaceUsers();
            alertStore.showSuccess(t.value("collimato.users.success.group_roles_updated"));
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });

    groupRoleDialog.value = false;
}

function addUserToWorkspace(usersList) {
    if (!collimatoStore.hasPermissionToAddUsers) {
        alertStore.showError(t.value("collimato.users.error.no_permission_add_users"));
        userDialog.value = false;

        return;
    }

    for (const user of usersList) {
        if (users.value.find((u) => u.user_info.email === user.email)) {
            alertStore.showError(
                `${user.email} ${t.value("collimato.users.error.already_in_workspace")}`,
            );

            return;
        }
    }

    collimatoService
        .addUsersToWorkspace(route.params.workspaceId, {
            users: usersList.map((u) => u.id),
        })
        .then(() => {
            getWorkspaceUsers();
            alertStore.showSuccess(t.value("collimato.users.success.user_added"));
        })
        .catch(() => {
            alertStore.showError(t.value("collimato.users.error.failed_to_add_user"));
        });

    userDialog.value = false;
}

function removeUser(user) {
    collimatoService
        .removeUserFromWorkspace(route.params.workspaceId, user.user_info.id)
        .then(() => {
            users.value = users.value.filter((u) => u.user_info.id !== user.user_info.id);
            alertStore.showSuccess(t.value("collimato.users.success.user_removed"));
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });

    removeUserDialog.value = false;
}

function updateUserRoles(newRoles) {
    collimatoService
        .updateUserRoles(route.params.workspaceId, userToEdit.value.id, {
            roles: newRoles,
        })
        .then(() => {
            getWorkspaceUsers();
            alertStore.showSuccess(t.value("collimato.users.success.user_roles_updated"));
        })
        .catch(() => {
            alertStore.showError(t.value("collimato.users.error.user_roles_update_failed"));
        });

    roleDialog.value = false;
}

function deleteRole() {
    collimatoService
        .deleteWorkspaceRole(route.params.workspaceId, deleteItem.value.id)
        .then(() => {
            roles.value = roles.value.filter((r) => r.id !== deleteItem.value.id);
        })
        .catch((error) => {
            alertStore.showError(extractErrorMessage(error));
        });

    deleteDialog.value = false;
}

onMounted(async () => {
    await collimatoService.getWorkspaceUsers(route.params.workspaceId).then((response) => {
        users.value = response.data;
    });

    await collimatoService
        .workspaceRoles(route.params.workspaceId)
        .then((response) => {
            roles.value = response.data;
        })
        .catch(() => {});

    if (collimatoStore.hasPermissionToAddUsers) {
        await collimatoService
            .getWorkspaceGroups(route.params.workspaceId)
            .then((response) => {
                groups.value = response.data ?? [];
            })
            .catch(() => {});
    }

    loaded.value = true;
});
</script>
