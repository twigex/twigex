<!--
Copyright (c) 2022 Twigex, SIA.
SPDX-License-Identifier: AGPL-3.0-only
-->

<template>
    <div>
        <div>
            <div v-if="selectedWorkspace" class="space-y-10">
                <MembersTable :members="memberRows" :columns="memberColumns">
                    <template #toolbar>
                        <div class="flex flex-wrap items-center justify-between gap-4">
                            <h2 class="text-sm font-semibold text-gray-900">
                                {{ t("projects.workspace_members.member_of") }}
                            </h2>
                            <BaseButton
                                v-if="rolesStore.hasPermissionToAddMemberToWorkspace"
                                size="small"
                                :prepend-icon="PlusIcon"
                                @click="showAddDialog = true"
                            >
                                {{ t("projects.workspace_members.add_members") }}
                            </BaseButton>
                        </div>
                    </template>

                    <template #role="{ item }">
                        <div class="flex flex-wrap gap-1">
                            <span
                                v-for="role in workspaceRoleArray(item.user_id)"
                                :key="role"
                                class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                            >
                                {{ role }}
                            </span>
                        </div>
                    </template>

                    <template #actions="{ item }">
                        <button
                            v-if="rolesLicensed && canAssignUser(item.user_info)"
                            type="button"
                            class="text-indigo-600 hover:text-indigo-500"
                            @click="openRoleDialog(item.user_info)"
                        >
                            {{ t("common.button.edit") }}
                        </button>
                        <button
                            v-if="rolesStore.hasPermissionToDeleteWorkspaceMember"
                            type="button"
                            class="text-red-600 hover:text-red-500"
                            @click="openDeleteDialog(item.user_info)"
                        >
                            {{ t("common.button.remove") }}
                        </button>
                    </template>
                </MembersTable>

                <DataTable v-if="groupsLicensed" :columns="groupColumns" :items="workspaceGroups">
                    <template #toolbar>
                        <h2 class="text-sm font-semibold text-gray-900">
                            {{ t("projects.workspace_members.groups") }}
                        </h2>
                    </template>

                    <template #empty>
                        {{ t("projects.workspace_members.no_groups") }}
                    </template>

                    <template #name="{ item }">
                        <div class="flex min-w-0 items-center gap-x-3">
                            <div
                                class="flex size-9 shrink-0 items-center justify-center rounded-full bg-gray-100"
                            >
                                <UserGroupIcon class="size-5 text-gray-500" aria-hidden="true" />
                            </div>
                            <span class="truncate font-medium text-gray-900">{{ item.name }}</span>
                        </div>
                    </template>

                    <template #roles="{ item }">
                        <div class="flex flex-wrap gap-1">
                            <span
                                v-for="role in item.roles"
                                :key="role"
                                class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                            >
                                {{ roleDisplayName(role) }}
                            </span>
                        </div>
                    </template>

                    <template #member_count="{ item }">
                        <WorkspaceMemberListPopover :group-id="item.group_id">
                            {{ item.member_count }}
                        </WorkspaceMemberListPopover>
                    </template>

                    <template #actions="{ item }">
                        <div class="flex items-center justify-end gap-x-4 font-medium">
                            <button
                                v-if="canAssignGroup(item)"
                                type="button"
                                class="text-indigo-600 hover:text-indigo-500"
                                @click="openGroupRoleDialog(item)"
                            >
                                {{ t("common.button.edit") }}
                            </button>
                            <button
                                v-if="rolesStore.hasPermissionToDeleteWorkspaceMember"
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click="openDeleteGroupDialog(item)"
                            >
                                {{ t("common.button.remove") }}
                            </button>
                        </div>
                    </template>
                </DataTable>

                <DataTable v-if="rolesLicensed" :columns="roleColumns" :items="roleRows">
                    <template #toolbar>
                        <div class="flex flex-wrap items-center justify-between gap-4">
                            <h2 class="text-sm font-semibold text-gray-900">
                                {{ t("projects.workspace_members.roles_in") }}
                            </h2>
                            <BaseButton
                                v-if="rolesStore.hasPermissionToCreateRoles"
                                size="small"
                                variant="secondary"
                                :prepend-icon="PlusIcon"
                                @click="goToCreateRole"
                            >
                                {{ t("projects.workspace_members.create_new_role") }}
                            </BaseButton>
                        </div>
                    </template>

                    <template #empty>
                        {{ t("projects.workspace_members.no_roles_created_yet") }}
                    </template>

                    <template #display_name="{ item }">
                        <div class="min-w-0">
                            <div class="truncate font-medium text-gray-900">{{ item.label }}</div>
                            <div v-if="item.about" class="mt-1 max-w-md truncate text-gray-500">
                                {{ item.about }}
                            </div>
                        </div>
                    </template>

                    <template #permissions="{ item }">
                        <div class="flex max-w-xl flex-wrap gap-1 whitespace-normal">
                            <span
                                v-for="permission in item.permissions"
                                :key="permission"
                                class="inline-flex items-center rounded-md bg-gray-50 px-1.5 py-0.5 text-xs font-medium text-gray-600 ring-1 ring-inset ring-gray-500/10"
                            >
                                {{ permission }}
                            </span>
                            <span
                                v-if="item.permissions.length === 0"
                                class="text-xs text-gray-400"
                            >
                                {{ t("projects.workspace_members.no_permissions") }}
                            </span>
                        </div>
                    </template>

                    <template #actions="{ item }">
                        <div class="flex items-center justify-end gap-x-4 font-medium">
                            <button
                                v-if="canEditRole(item)"
                                type="button"
                                class="text-indigo-600 hover:text-indigo-500"
                                @click="editRole(item)"
                            >
                                {{ t("common.button.edit") }}
                            </button>
                            <button
                                v-if="
                                    rolesStore.hasPermissionToDeleteRoles &&
                                    item.name !== 'admin' &&
                                    item.name !== 'user'
                                "
                                type="button"
                                class="text-red-600 hover:text-red-500"
                                @click="openDeleteRoleDialog(item)"
                            >
                                {{ t("common.button.delete") }}
                            </button>
                        </div>
                    </template>
                </DataTable>
            </div>

            <ConfirmDialog
                :open="showConfirm"
                :title="deleteTitle"
                :message="deleteMessage"
                :confirm-label="t('common.button.delete')"
                @confirm="confirmDelete"
                @close="showConfirm = false"
            />

            <MemberRolesDialog
                v-model:open="showRoleConfirm"
                v-model:roles="selectedRoles"
                :title="roleTitle"
                :message="roleMessage"
                :options="assignableRoleOptions"
                :role-name="roleDisplayName"
                @confirm="confirmRoleUpdate"
            />

            <AddMemberGroupDialog
                v-model="showAddDialog"
                :workspace="selectedWorkspace"
                :existing-member-ids="existingMemberIds"
                @add="onAddUsers"
                @add-groups="onAddGroups"
            />
        </div>
    </div>
</template>

<script setup>
import { t } from "@/i18n/index.js";
import { useRouter } from "vue-router";
import BaseButton from "@/components/BaseButton.vue";
import DataTable from "@/components/DataTable.vue";
import MembersTable from "@/components/MembersTable.vue";
import { useUserStore } from "@/store/user";
import { useWorkspaceStore } from "@/store/workspaces";
import workspaceService from "@/services/workspaceService";
import { computed, ref, watch } from "vue";
import { useAlertStore } from "@/store/alerts";
import { extractErrorMessage } from "@/utils/errors";
import { PlusIcon } from "@heroicons/vue/20/solid";
import { UserGroupIcon } from "@heroicons/vue/24/outline";
import { useWorkspaceRolesStore } from "@/store/workspaceRoles";
import { useSettingsStore } from "@/store/settings";
import AddMemberGroupDialog from "@/components/Projects/Members/AddMemberGroupDialog.vue";
import WorkspaceMemberListPopover from "@/components/Projects/Members/WorkspaceMemberListPopover.vue";
import MemberRolesDialog from "@/components/Projects/Members/MemberRolesDialog.vue";
import ConfirmDialog from "@/components/ConfirmDialog.vue";
import { roleLabel, roleDescription, roleNameLabel } from "@/utils/roleLabels";
const rolesStore = useWorkspaceRolesStore();
const settingsStore = useSettingsStore();

// Membership is not a paid feature, so only the role and group sections are gated.
const rolesLicensed = computed(() => settingsStore.getLicenseFeature("workspace_roles"));
const groupsLicensed = computed(() => settingsStore.getLicenseFeature("groups"));

const memberColumns = computed(() =>
    rolesLicensed.value
        ? [
              {
                  key: "role",
                  label: t.value("members_table.roles"),
                  sortable: true,
                  hiddenBelow: "sm",
              },
          ]
        : [],
);

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
    { key: "display_name", label: t.value("data_table.role"), sortable: true, sortKey: "label" },
    { key: "permissions", label: t.value("data_table.permissions"), hiddenBelow: "md" },
    { key: "actions", label: t.value("data_table.actions"), srOnly: true, align: "right" },
]);

const showConfirm = ref(false);
const deleteUserId = ref(null);
const deleteGroupRef = ref(null);
const deleteTitle = ref("");
const deleteMessage = ref("");

const openDeleteDialog = (user) => {
    deleteUserId.value = user.id;
    deleteGroupRef.value = null;
    deleteTitle.value = `${t.value("common.button.remove")} ${user.name} ${user.lastname}?`;
    deleteMessage.value = `${t.value("projects.workspace_members.are_you_sure_you_want_to_remove")} ${user.name} ${t.value("projects.workspace_members.from_the_workspace")} "${selectedWorkspace.value.title}"? ${t.value("projects.workspace_members.this_action_cannot_be_undone")}`;
    showConfirm.value = true;
};

const openDeleteGroupDialog = (group) => {
    deleteGroupRef.value = group;
    deleteUserId.value = null;
    deleteTitle.value = `${t.value("common.button.remove")} ${group.name}?`;
    deleteMessage.value = `${t.value("projects.workspace_members.are_you_sure_you_want_to_remove")} ${group.name} ${t.value("projects.workspace_members.from_the_workspace")} "${selectedWorkspace.value.title}"? ${t.value("projects.workspace_members.this_action_cannot_be_undone")}`;
    showConfirm.value = true;
};

const goToCreateRole = () => {
    router.push({
        name: "create-role",
        params: { id: selectedWorkspace.value.id },
    });
};

const selectedRoleGroup = ref(null);

const assignableRoleOptions = computed(() => {
    if (rolesStore.isWorkspaceAdmin) return roleOptions.value;

    return roleOptions.value.filter((name) => name !== "admin");
});

function memberRoles(userId) {
    const member = selectedWorkspace.value?.members?.find((m) => m.user_id === userId);

    return (member?.role ?? "").split(" ").filter(Boolean);
}

function canAssignUser(user) {
    if (!rolesStore.hasPermissionToUpdateRoles) return false;
    if (rolesStore.isWorkspaceAdmin) return true;

    return user.id !== userStore.user?.id && !memberRoles(user.id).includes("admin");
}

function canAssignGroup(group) {
    if (!rolesStore.hasPermissionToUpdateRoles) return false;
    if (rolesStore.isWorkspaceAdmin) return true;

    return !(group.roles ?? []).includes("admin");
}

function canEditRole(role) {
    if (!rolesStore.hasPermissionToUpdateRoles) return false;
    if (rolesStore.isWorkspaceAdmin) return true;

    return role.name !== "admin" && !rolesStore.roleNames.includes(role.name);
}

const openGroupRoleDialog = (group) => {
    selectedRoleGroup.value = group;
    selectedRoles.value = group.roles?.length ? [...group.roles] : ["user"];
    roleTitle.value = `${t.value("projects.workspace_members.change_roles_for")} ${group.name}`;
    roleMessage.value = `${t.value("projects.workspace_members.change_roles_for")} ${group.name} ${t.value("projects.workspace_members.in_workspace")} "${selectedWorkspace.value.title}"?`;
    showRoleConfirm.value = true;
};

const showRoleConfirm = ref(false);
const roleTitle = ref("");
const roleMessage = ref("");
const roleOptions = ref(["user", "admin"]);
const selectedRoles = ref([]);
const selectedRoleUser = ref(null);

const openRoleDialog = (user) => {
    selectedRoleUser.value = user;
    selectedRoleGroup.value = null;
    const member = selectedWorkspace.value.members.find((m) => m.user_id === user.id);

    // Parse existing SPACE-separated roles
    const existingRoles = member?.role
        ? member.role
              .split(" ")
              .map((r) => r.trim())
              .filter((r) => r)
        : ["user"];

    selectedRoles.value = existingRoles;

    roleMessage.value = `${t.value("projects.workspace_members.change_roles_for")} ${user.name} ${t.value("projects.workspace_members.in_workspace")} "${selectedWorkspace.value.title}"?`;
    roleTitle.value = `${t.value("projects.workspace_members.change_roles_for")} ${user.name}`;
    showRoleConfirm.value = true;
};

const confirmRoleUpdate = async () => {
    if (!selectedWorkspace.value) return;

    try {
        if (selectedRoleGroup.value) {
            await workspaceService.updateWorkspaceGroupRoles(
                selectedWorkspace.value.id,
                selectedRoleGroup.value.group_id,
                selectedRoles.value,
            );
            selectedRoleGroup.value.roles = [...selectedRoles.value];
        } else if (selectedRoleUser.value) {
            const rolesString = selectedRoles.value.join(" ");

            await workspaceService.updateMemberRole({
                workspace_id: selectedWorkspace.value.id,
                user_id: selectedRoleUser.value.id,
                role: rolesString,
            });
            const member = selectedWorkspace.value.members.find(
                (m) => m.user_id === selectedRoleUser.value.id,
            );

            if (member) member.role = rolesString;
        }

        showRoleConfirm.value = false;
        selectedRoleUser.value = null;
        selectedRoleGroup.value = null;
        selectedRoles.value = [];
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.workspace_members.failed_to_update")),
        );
        showRoleConfirm.value = false;
    }
};

const userStore = useUserStore();
const workspaceStore = useWorkspaceStore();

const props = defineProps({
    workspaceId: { type: String, required: true },
});

const selectedWorkspace = ref(null);

const workspaces = computed(() => workspaceStore.getWorkspaces || []);

const router = useRouter();

// One request gives both the role picker's names and the roles shown.
const fetchWorkspaceRoles = async (workspaceId) => {
    try {
        const response = await workspaceService.getWorkspaceRoles(workspaceId);

        workspaceRoles.value = response.data || [];
        roleOptions.value = workspaceRoles.value.map((role) => role.name || role.display_name);
    } catch (error) {
        workspaceRoles.value = [];
        roleOptions.value = [];
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const fetchWorkspaceDetails = async (workspaceId) => {
    try {
        const [, response] = await Promise.all([
            fetchUserRolesForWorkspace(workspaceId),
            workspaceStore.fetchWorkspaceNav(workspaceId),
            fetchWorkspaceRoles(workspaceId),
        ]);

        selectedWorkspace.value = response.data;

        // The group count is written onto the workspace just loaded.
        await fetchWorkspaceGroups(workspaceId);
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const selectedWorkspaceUsers = computed(() => {
    if (!selectedWorkspace.value || !Array.isArray(selectedWorkspace.value.members)) return [];

    return selectedWorkspace.value.members
        .map((member) => {
            if (!member.user_id) return null;

            // id must come last, user_info.id may be empty string for newly added members
            return {
                ...(member.user_info || {}),
                ...(userStore.usersMap[member.user_id] || {}),
                id: member.user_id,
            };
        })
        .filter(Boolean);
});

const memberRows = computed(() =>
    selectedWorkspaceUsers.value.map((user) => ({
        user_id: user.id,
        user_info: user,
        role: memberRoles(user.id).join(" "),
    })),
);

const fetchUserRolesForWorkspace = async (workspaceId) => {
    try {
        if (workspaceId) {
            const response = await workspaceService.me(workspaceId);

            if (response.data) {
                rolesStore.setUser(response.data);
            }
        }
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const workspaceRoles = ref([]);

const roleRows = computed(() =>
    workspaceRoles.value.map((role) => ({
        ...role,
        label: roleLabel(role, "projects"),
        about: roleDescription(role, "projects"),
    })),
);

const workspaceRoleObject = computed({
    get() {
        return workspaceStore.getWorkspaceRole;
    },
    set(value) {
        workspaceStore.setWorkspaceRole(value);
    },
});

const editRole = (role) => {
    workspaceRoleObject.value = role;
    router.push({
        name: "update-role",
        params: {
            id: selectedWorkspace.value.id,
            rid: role.id,
        },
    });
};

const roleDisplayName = (name) => {
    const found = workspaceRoles.value.find((r) => r.name === name);

    return found ? roleLabel(found, "projects") : roleNameLabel(name, "projects");
};

// Each member's role labels, built once per change rather than by scanning
// every member and role from each row on every render.
const memberRoleLabels = computed(() => {
    const labels = new Map();
    const roleNames = new Map(workspaceRoles.value.map((r) => [r.name, roleLabel(r, "projects")]));

    for (const member of selectedWorkspace.value?.members || []) {
        labels.set(
            member.user_id,
            (member.role || "")
                .split(" ")
                .map((roleName) => roleName.trim())
                .filter(Boolean)
                .map((roleName) => roleNames.get(roleName) || roleNameLabel(roleName, "projects")),
        );
    }

    return labels;
});

const workspaceRoleArray = (userId) => {
    if (!selectedWorkspace.value) return ["user"];

    return memberRoleLabels.value.get(userId) || ["user"];
};

const openDeleteRoleDialog = (role) => {
    deleteRoleId.value = role.id;
    deleteTitle.value = `${t.value("common.button.delete")} ${roleLabel(role, "projects")}?`;
    deleteMessage.value = `${t.value("projects.workspace_members.are_you_sure_you_want_delete_the_role")} "${roleLabel(role, "projects")}"? ${t.value("projects.workspace_members.this_will_remove_the_role_from_all_users")}`;
    showConfirm.value = true;
};

const confirmDeleteRole = async () => {
    if (!selectedWorkspace.value || !deleteRoleId.value) return;

    try {
        await workspaceService.deleteWorkspaceRole({
            workspace_id: selectedWorkspace.value.id,
            role_id: deleteRoleId.value,
        });

        showConfirm.value = false;
        deleteRoleId.value = null;

        await fetchWorkspaceDetails(selectedWorkspace.value.id);
    } catch (error) {
        useAlertStore().showError(
            extractErrorMessage(error, t.value("projects.workspace_members.failed_to_delete_role")),
        );
    }
};

const confirmDelete = async () => {
    if (!selectedWorkspace.value) return;

    if (deleteUserId.value) {
        await confirmDeleteMember();
    } else if (deleteGroupRef.value) {
        await confirmDeleteGroup();
    } else if (deleteRoleId.value) {
        await confirmDeleteRole();
    }
};

const confirmDeleteGroup = async () => {
    if (!deleteGroupRef.value) return;
    try {
        await workspaceService.removeWorkspaceGroup(
            selectedWorkspace.value.id,
            deleteGroupRef.value.group_id,
        );
        workspaceGroups.value = workspaceGroups.value.filter(
            (g) => g.id !== deleteGroupRef.value.id,
        );
        showConfirm.value = false;
        deleteGroupRef.value = null;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const confirmDeleteMember = async () => {
    try {
        await workspaceService.deleteWorkspaceMember({
            workspace_id: selectedWorkspace.value.id,
            member_id: deleteUserId.value,
        });

        selectedWorkspace.value.members = selectedWorkspace.value.members.filter(
            (member) => member.user_id !== deleteUserId.value,
        );

        showConfirm.value = false;
        deleteUserId.value = null;
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const deleteRoleId = ref(null);

const showAddDialog = ref(false);
const workspaceGroups = ref([]);

const existingMemberIds = computed(
    () => selectedWorkspace.value?.members?.map((m) => m.user_id) ?? [],
);

const fetchWorkspaceGroups = async (workspaceId) => {
    try {
        const res = await workspaceService.getWorkspaceGroups(workspaceId);

        workspaceGroups.value = res.data ?? [];
        // Keep the group_count on the card in sync
        const count = workspaceGroups.value.length;

        if (selectedWorkspace.value?.id === workspaceId) {
            selectedWorkspace.value.group_count = count;
        }

        const storeWs = workspaceStore.getWorkspaces?.find((w) => w.id === workspaceId);

        if (storeWs) storeWs.group_count = count;
    } catch {
        workspaceGroups.value = [];
    }
};

const onAddUsers = async (users) => {
    if (!selectedWorkspace.value || !users.length) return;
    try {
        const response = await workspaceService.addMembers({
            workspace_id: selectedWorkspace.value.id,
            user_ids: users.map((u) => u.id),
        });

        if (!selectedWorkspace.value.members) selectedWorkspace.value.members = [];
        selectedWorkspace.value.members.push(...response.data);

        // A member just added comes back without the user's details.
        userStore.ensureUsers(response.data.map((member) => member.user_id));
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

const onAddGroups = async (groups) => {
    if (!selectedWorkspace.value || !groups.length) return;
    try {
        await workspaceService.addWorkspaceGroups(
            selectedWorkspace.value.id,
            groups.map((g) => g.id),
            ["user"],
        );
        await fetchWorkspaceGroups(selectedWorkspace.value.id);
    } catch (error) {
        useAlertStore().showError(extractErrorMessage(error));
    }
};

watch(
    () => props.workspaceId,
    (workspaceId) => {
        selectedWorkspace.value = workspaces.value.find((w) => w.id === workspaceId) || null;
        fetchWorkspaceDetails(workspaceId);
    },
    { immediate: true },
);
</script>
